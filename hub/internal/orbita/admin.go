package orbita

import (
	"database/sql"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ai-hub/hub/internal/platform/auth"
	"ai-hub/hub/internal/platform/idgen"
)

func (h *Handlers) RegisterAdmin(mux *http.ServeMux) {
	mux.HandleFunc("/admin/v1/protocols", h.handleAdminProtocols)
	mux.HandleFunc("/admin/v1/protocols/", h.handleAdminProtocols)
	mux.HandleFunc("/admin/v1/sla-reports", h.handleAdminSLAReports)
	mux.HandleFunc("/admin/v1/sla-reports/", h.handleAdminSLAReports)
}

type protocolCursor struct{ Tenant, Subject, ID, Status, From, To string }

func (h *Handlers) handleAdminProtocols(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.FromContext(r.Context())
	if !ok {
		auth.Error(w, 401, "unauthenticated")
		return
	}
	// A consumer token with protocols:read is not an administrative identity.
	// Administrative diagnostics require a nominal, MFA-authenticated role;
	// cross-tenant access is checked separately below.
	if p.Workload || !p.HasScope("protocols:read") || !p.MFA || (!p.HasRole("hub_protocol_reader") && !p.HasRole("hub_admin")) {
		auth.Error(w, 403, "forbidden")
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		auth.Error(w, 405, "method_not_allowed")
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/admin/v1/protocols")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) > 2 {
		auth.Error(w, 404, "not_found")
		return
	}
	tenant := r.URL.Query().Get("tenant_id")
	if tenant == "" {
		tenant = p.TenantID
	}
	cross := tenant != p.TenantID || tenant == "*"
	if tenant == "" || cross && !p.HasScope("admin:cross_tenant") {
		auth.Error(w, 403, "global_reader_mfa_required")
		return
	}
	reason := strings.TrimSpace(r.URL.Query().Get("reason"))
	if r.Method == http.MethodGet && cross && (len(reason) < 8 || len(reason) > 512) {
		auth.Error(w, 422, "reason_required")
		return
	}
	id := parts[0]
	if r.Method == http.MethodPost {
		if id == "" || len(parts) != 2 || parts[1] != "reconcile" || !p.HasScope("protocols:reconcile") || !p.MFA || (!p.HasRole("hub_protocol_reader") && !p.HasRole("hub_admin")) {
			auth.Error(w, 403, "forbidden")
			return
		}
		protocol, err := h.store.GetByID(r.Context(), id)
		if err != nil || (tenant != "*" && protocol.TenantID != tenant) {
			auth.Error(w, 404, "not_found")
			return
		}
		var input struct {
			Reason string `json:"reason"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64*1024))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&input) != nil || decoder.Decode(&struct{}{}) != io.EOF {
			auth.Error(w, 422, "reason_required")
			return
		}
		input.Reason = strings.TrimSpace(input.Reason)
		if len(input.Reason) < 8 || len(input.Reason) > 512 {
			auth.Error(w, 422, "reason_required")
			return
		}
		tx, err := h.store.db.BeginTx(r.Context(), nil)
		if err != nil {
			auth.Error(w, 503, "reconciliation_unavailable")
			return
		}
		defer tx.Rollback()
		requestID := idgen.New()
		var storedRequestID string
		var missingCorrelation bool
		if err = tx.QueryRowContext(r.Context(), `SELECT EXISTS(SELECT 1 FROM operations WHERE protocol_id=$1 AND tenant_id=$2 AND state IN ('SUBMITTING','UNKNOWN','ACCEPTED_EXTERNAL','WAITING_FINAL') AND provider_request_id IS NULL)`, id, protocol.TenantID).Scan(&missingCorrelation); err != nil {
			auth.Error(w, 503, "reconciliation_unavailable")
			return
		}
		requestState := "OPEN"
		lastError := ""
		if missingCorrelation {
			requestState = "REJECTED"
			lastError = "provider_request_id ausente; obter evidência externa antes de reconciliar"
			err = tx.QueryRowContext(r.Context(), `UPDATE protocol_reconciliation_requests SET state='REJECTED',last_error=$3,claim_owner=NULL,lease_until=NULL,updated_at=clock_timestamp() WHERE tenant_id=$1 AND protocol_id=$2 AND state='OPEN' RETURNING request_id`, protocol.TenantID, id, lastError).Scan(&storedRequestID)
			if errors.Is(err, sql.ErrNoRows) {
				err = tx.QueryRowContext(r.Context(), `INSERT INTO protocol_reconciliation_requests(request_id,protocol_id,tenant_id,requested_by,reason,state,last_error) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING request_id`, requestID, id, protocol.TenantID, p.Subject, strings.TrimSpace(input.Reason), requestState, lastError).Scan(&storedRequestID)
			}
		} else {
			err = tx.QueryRowContext(r.Context(), `INSERT INTO protocol_reconciliation_requests(request_id,protocol_id,tenant_id,requested_by,reason,state,last_error) VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT (tenant_id,protocol_id) WHERE state='OPEN' DO UPDATE SET updated_at=clock_timestamp(),reason=EXCLUDED.reason,requested_by=EXCLUDED.requested_by RETURNING request_id`, requestID, id, protocol.TenantID, p.Subject, strings.TrimSpace(input.Reason), requestState, lastError).Scan(&storedRequestID)
		}
		if err != nil {
			auth.Error(w, 503, "reconciliation_unavailable")
			return
		}
		auditAction := "RECONCILIATION_REQUEST"
		if missingCorrelation {
			auditAction = "RECONCILIATION_REJECTED"
		}
		if _, err = tx.ExecContext(r.Context(), "INSERT INTO protocol_access_audit(subject,requested_tenant,resource,action,mfa,reason) VALUES($1,$2,$3,$4,$5,$6)", p.Subject, tenant, id, auditAction, p.MFA, input.Reason); err != nil {
			auth.Error(w, 503, "audit_unavailable")
			return
		}
		if err = tx.Commit(); err != nil {
			auth.Error(w, 503, "audit_unavailable")
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		if missingCorrelation {
			writeJSON(w, http.StatusConflict, map[string]any{"request_id": storedRequestID, "protocol_id": id, "state": requestState, "effect": "no_provider_correlation", "error": lastError})
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"request_id": storedRequestID, "protocol_id": id, "state": requestState, "effect": "no_provider_replay"})
		return
	}
	// Audit is a prerequisite of diagnostic access, including empty/global reads.
	if _, err := h.store.db.ExecContext(r.Context(), "INSERT INTO protocol_access_audit(subject,requested_tenant,resource,action,mfa,reason) VALUES($1,$2,$3,'READ',$4,$5)", p.Subject, tenant, id, p.MFA, reason); err != nil {
		auth.Error(w, 503, "audit_unavailable")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	if id != "" {
		protocol, err := h.store.GetByID(r.Context(), id)
		if err != nil || (tenant != "*" && protocol.TenantID != tenant) {
			auth.Error(w, 404, "not_found")
			return
		}
		if len(parts) == 2 {
			if parts[1] != "timeline" {
				auth.Error(w, 404, "not_found")
				return
			}
			timeline := []map[string]any{{"event": "ACCEPTED", "recorded_at": protocol.AcceptedAt, "protocol_id": id}}
			rows, err := h.store.db.QueryContext(r.Context(), "SELECT action,recorded_at,details FROM protocol_audit WHERE protocol_id=$1 AND tenant_id=$2 ORDER BY id LIMIT 200", id, protocol.TenantID)
			if err != nil {
				auth.Error(w, 503, "timeline_unavailable")
				return
			}
			defer rows.Close()
			for rows.Next() {
				var action string
				var recorded time.Time
				var details json.RawMessage
				if rows.Scan(&action, &recorded, &details) != nil {
					auth.Error(w, 503, "timeline_unavailable")
					return
				}
				timeline = append(timeline, map[string]any{"event": action, "recorded_at": recorded, "details": details})
			}
			if rows.Err() != nil {
				auth.Error(w, 503, "timeline_unavailable")
				return
			}
			if protocol.FinalizedAt.Valid {
				timeline = append(timeline, map[string]any{"event": protocol.Status, "recorded_at": protocol.FinalizedAt.Time, "reason": protocol.TerminalReason.String})
			}
			writeJSON(w, 200, map[string]any{"items": timeline})
			return
		}
		writeJSON(w, 200, map[string]any{"protocol_id": id, "tenant_id": protocol.TenantID, "application_id": protocol.ApplicationID, "cell_id": protocol.CellID, "status": protocol.Status, "mode": protocol.Mode, "accepted_at": protocol.AcceptedAt, "client_deadline_at": protocol.ClientDeadlineAt, "result_version": protocol.ResultVersion, "final_representation": json.RawMessage(protocol.FinalRepresentation)})
		return
	}
	limit := 25
	if value := r.URL.Query().Get("limit"); value != "" {
		n, err := strconv.Atoi(value)
		if err != nil || n < 1 || n > 100 {
			auth.Error(w, 400, "invalid_limit")
			return
		}
		limit = n
	}
	status, from, to := r.URL.Query().Get("status"), r.URL.Query().Get("from"), r.URL.Query().Get("to")
	for _, date := range []string{from, to} {
		if date != "" {
			if _, err := time.Parse("2006-01-02", date); err != nil {
				auth.Error(w, 400, "invalid_date")
				return
			}
		}
	}
	after := ""
	if cursor := r.URL.Query().Get("cursor"); cursor != "" {
		if len(cursor) > 2048 {
			auth.Error(w, 400, "invalid_cursor")
			return
		}
		raw, err := base64.RawURLEncoding.DecodeString(cursor)
		var c protocolCursor
		if err != nil || json.Unmarshal(raw, &c) != nil || c.Tenant != tenant || c.Subject != p.Subject || c.Status != status || c.From != from || c.To != to {
			auth.Error(w, 400, "cursor_outside_scope")
			return
		}
		after = c.ID
	}
	rows, err := h.store.db.QueryContext(r.Context(), `SELECT protocol_id,tenant_id,application_id,status,mode,accepted_at,client_deadline_at FROM protocols WHERE ($1='*' OR tenant_id=$1) AND ($2='' OR status=$2) AND protocol_id::text>$3 AND ($4='' OR accepted_at>=NULLIF($4,'')::date) AND ($5='' OR accepted_at<NULLIF($5,'')::date+interval '1 day') ORDER BY protocol_id::text LIMIT $6`, tenant, status, after, from, to, limit+1)
	if err != nil {
		auth.Error(w, 503, "protocols_unavailable")
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	last := ""
	for rows.Next() {
		var id, t, app, state, mode string
		var accepted, deadline time.Time
		if rows.Scan(&id, &t, &app, &state, &mode, &accepted, &deadline) != nil {
			auth.Error(w, 503, "protocols_unavailable")
			return
		}
		items = append(items, map[string]any{"protocol_id": id, "tenant_id": t, "application_id": app, "status": state, "mode": mode, "accepted_at": accepted, "client_deadline_at": deadline})
		if len(items) == limit {
			last = id
		}
	}
	if rows.Err() != nil {
		auth.Error(w, 503, "protocols_unavailable")
		return
	}
	next := ""
	if len(items) > limit {
		items = items[:limit]
		raw, _ := json.Marshal(protocolCursor{tenant, p.Subject, last, status, from, to})
		next = base64.RawURLEncoding.EncodeToString(raw)
	}
	writeJSON(w, 200, map[string]any{"items": items, "next_cursor": next})
}

type slaReportCursor struct {
	Tenant, Subject, Status, From, To, ID string
}

// handleAdminSLAReports expõe a mesma autoridade persistida usada pelo
// protocolo, sem consultar o provedor. O prazo do provedor é extraído do
// snapshot aceito; o prazo do cliente vem da coluna imutável da admissão.
// Assim, editar o catálogo depois não reescreve uma apuração histórica.
func (h *Handlers) handleAdminSLAReports(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.FromContext(r.Context())
	if !ok {
		auth.Error(w, 401, "unauthenticated")
		return
	}
	if r.Method != http.MethodGet || p.Workload || !p.HasScope("protocols:read") || !p.MFA || (!p.HasRole("hub_protocol_reader") && !p.HasRole("hub_admin")) {
		auth.Error(w, 403, "forbidden")
		return
	}
	tenant := r.URL.Query().Get("tenant_id")
	if tenant == "" {
		tenant = p.TenantID
	}
	if tenant == "" || (tenant != p.TenantID && !p.HasScope("admin:cross_tenant")) {
		auth.Error(w, 403, "global_reader_mfa_required")
		return
	}
	cross := tenant != p.TenantID || tenant == "*"
	reason := strings.TrimSpace(r.URL.Query().Get("reason"))
	if cross && (len(reason) < 8 || len(reason) > 512) {
		auth.Error(w, 422, "reason_required")
		return
	}
	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/admin/v1/sla-reports"), "/")
	queryID := id
	auditAction := "READ"
	if id == "export" {
		auditAction = "EXPORT"
		queryID = ""
	}
	if _, err := h.store.db.ExecContext(r.Context(), "INSERT INTO protocol_access_audit(subject,requested_tenant,resource,action,mfa,reason) VALUES($1,$2,'sla-reports',$3,$4,$5)", p.Subject, tenant, auditAction, p.MFA, reason); err != nil {
		auth.Error(w, 503, "audit_unavailable")
		return
	}
	status, from, to := r.URL.Query().Get("status"), r.URL.Query().Get("from"), r.URL.Query().Get("to")
	for _, date := range []string{from, to} {
		if date != "" {
			if _, err := time.Parse("2006-01-02", date); err != nil {
				auth.Error(w, 400, "invalid_date")
				return
			}
		}
	}
	limit := 25
	if value := r.URL.Query().Get("limit"); value != "" {
		n, err := strconv.Atoi(value)
		if err != nil || n < 1 || n > 100 {
			auth.Error(w, 400, "invalid_limit")
			return
		}
		limit = n
	}
	after := ""
	if cursor := r.URL.Query().Get("cursor"); cursor != "" {
		if len(cursor) > 2048 {
			auth.Error(w, 400, "invalid_cursor")
			return
		}
		raw, err := base64.RawURLEncoding.DecodeString(cursor)
		var c slaReportCursor
		if err != nil || json.Unmarshal(raw, &c) != nil || c.Tenant != tenant || c.Subject != p.Subject || c.Status != status || c.From != from || c.To != to {
			auth.Error(w, 400, "cursor_outside_scope")
			return
		}
		after = c.ID
	}
	if id != "" {
		after = ""
		limit = 1
		if id == "export" {
			limit = 100
		}
	}
	rows, err := h.store.db.QueryContext(r.Context(), `SELECT protocol_id,tenant_id,application_id,status,accepted_at,client_deadline_at,finalized_at,final_event_id,config_snapshot
		FROM protocols
		WHERE ($1='*' OR tenant_id=$1) AND ($2='' OR status=$2) AND ($3='' OR protocol_id::text=$3) AND protocol_id::text>$4
		  AND ($5='' OR accepted_at>=NULLIF($5,'')::date) AND ($6='' OR accepted_at<NULLIF($6,'')::date+interval '1 day')
		ORDER BY protocol_id::text LIMIT $7`, tenant, status, queryID, after, from, to, limit+1)
	if err != nil {
		auth.Error(w, 503, "sla_reports_unavailable")
		return
	}
	var eligible, open, fulfilled, expired, excluded int
	if err := h.store.db.QueryRowContext(r.Context(), `SELECT count(*),
		count(*) FILTER (WHERE status NOT IN ('SUCCEEDED','PARTIALLY_SUCCEEDED','FAILED','EXPIRED','CANCELLED')),
		count(*) FILTER (WHERE status IN ('SUCCEEDED','PARTIALLY_SUCCEEDED')),
		count(*) FILTER (WHERE status='EXPIRED'),
		count(*) FILTER (WHERE status IN ('FAILED','CANCELLED'))
		FROM protocols
		WHERE ($1='*' OR tenant_id=$1)
		  AND ($2='' OR accepted_at>=NULLIF($2,'')::date)
		  AND ($3='' OR accepted_at<NULLIF($3,'')::date+interval '1 day')`, tenant, from, to).
		Scan(&eligible, &open, &fulfilled, &expired, &excluded); err != nil {
		auth.Error(w, 503, "sla_reports_unavailable")
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	last := ""
	generatedAt := time.Now().UTC()
	watermarkAt := time.Time{}
	for rows.Next() {
		var protocolID, protocolTenant, applicationID, state string
		var accepted, clientDeadline time.Time
		var finalized sql.NullTime
		var finalEvent sql.NullString
		var snapshotRaw []byte
		if err := rows.Scan(&protocolID, &protocolTenant, &applicationID, &state, &accepted, &clientDeadline, &finalized, &finalEvent, &snapshotRaw); err != nil {
			auth.Error(w, 503, "sla_reports_unavailable")
			return
		}
		var snapshot struct {
			Target struct {
				Data json.RawMessage `json:"data"`
			} `json:"target"`
		}
		providerSLA := 0
		var targetData struct {
			ProviderSLASeconds int    `json:"provider_sla_seconds"`
			ProviderSLAPolicy  string `json:"provider_sla_policy"`
		}
		providerPolicy := "MONITOR_ONLY"
		if json.Unmarshal(snapshotRaw, &snapshot) == nil {
			_ = json.Unmarshal(snapshot.Target.Data, &targetData)
			providerSLA = targetData.ProviderSLASeconds
			if value := strings.ToUpper(strings.TrimSpace(targetData.ProviderSLAPolicy)); value != "" {
				providerPolicy = value
			}
		}
		providerDeadline := accepted
		if providerSLA > 0 {
			providerDeadline = accepted.Add(time.Duration(providerSLA) * time.Second)
		}
		observedAt := time.Now().UTC()
		if finalized.Valid {
			observedAt = finalized.Time
		}
		if observedAt.After(watermarkAt) {
			watermarkAt = observedAt
		}
		providerBreached := providerSLA > 0 && !observedAt.Before(providerDeadline)
		providerOutcome := "NOT_CONFIGURED"
		if providerSLA > 0 {
			providerOutcome = "ON_TIME"
			if providerBreached {
				providerOutcome = providerPolicy + "_BREACH"
			}
		}
		items = append(items, map[string]any{
			"protocol_id": protocolID, "tenant_id": protocolTenant, "application_id": applicationID, "status": state,
			"accepted_at": accepted, "client_deadline_at": clientDeadline, "provider_sla_seconds": providerSLA,
			"provider_sla_policy": providerPolicy, "provider_sla_outcome": providerOutcome,
			"provider_deadline_at": providerDeadline, "observed_at": observedAt,
			"client_sla_breached":   !observedAt.Before(clientDeadline),
			"provider_sla_breached": providerBreached,
			"finalized_at":          finalized, "final_event_id": finalEvent,
		})
		if len(items) == limit {
			last = protocolID
		}
	}
	if err := rows.Err(); err != nil {
		auth.Error(w, 503, "sla_reports_unavailable")
		return
	}
	if id == "export" {
		if len(items) > limit {
			items = items[:limit]
		}
		var output strings.Builder
		writer := csv.NewWriter(&output)
		if err := writer.Write([]string{"protocol_id", "tenant_id", "application_id", "status", "accepted_at", "client_deadline_at", "provider_sla_seconds", "provider_sla_policy", "provider_sla_outcome", "provider_deadline_at", "observed_at", "client_sla_breached", "provider_sla_breached", "finalized_at", "final_event_id"}); err != nil {
			auth.Error(w, 503, "sla_export_unavailable")
			return
		}
		value := func(raw any) string {
			switch typed := raw.(type) {
			case sql.NullTime:
				if !typed.Valid {
					return ""
				}
				return typed.Time.UTC().Format(time.RFC3339Nano)
			case sql.NullString:
				if !typed.Valid {
					return ""
				}
				return typed.String
			default:
				return fmt.Sprint(raw)
			}
		}
		for _, item := range items {
			row := make([]string, 0, 15)
			for _, field := range []string{"protocol_id", "tenant_id", "application_id", "status", "accepted_at", "client_deadline_at", "provider_sla_seconds", "provider_sla_policy", "provider_sla_outcome", "provider_deadline_at", "observed_at", "client_sla_breached", "provider_sla_breached", "finalized_at", "final_event_id"} {
				row = append(row, value(item[field]))
			}
			if err := writer.Write(row); err != nil {
				auth.Error(w, 503, "sla_export_unavailable")
				return
			}
		}
		writer.Flush()
		if err := writer.Error(); err != nil {
			auth.Error(w, 503, "sla_export_unavailable")
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, 200, map[string]any{"filename": "sla-report.csv", "content_type": "text/csv", "csv": output.String(), "rows": len(items), "limited": true, "max_rows": 100, "tenant_id": tenant})
		return
	}
	if id != "" {
		if len(items) != 1 {
			auth.Error(w, 404, "not_found")
			return
		}
		writeJSON(w, 200, items[0])
		return
	}
	next := ""
	if len(items) > limit {
		items = items[:limit]
		raw, _ := json.Marshal(slaReportCursor{Tenant: tenant, Subject: p.Subject, Status: status, From: from, To: to, ID: last})
		next = base64.RawURLEncoding.EncodeToString(raw)
	}
	w.Header().Set("Cache-Control", "no-store")
	if watermarkAt.IsZero() {
		watermarkAt = generatedAt
	}
	writeJSON(w, 200, map[string]any{"items": items, "next_cursor": next, "generated_at": generatedAt, "watermark_at": watermarkAt, "watermark_lag_seconds": generatedAt.Sub(watermarkAt).Seconds(), "cohort": map[string]int{"eligible": eligible, "open": open, "fulfilled": fulfilled, "expired": expired, "excluded": excluded}})
}
