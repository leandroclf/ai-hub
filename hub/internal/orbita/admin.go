package orbita

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ai-hub/hub/internal/platform/auth"
)

func (h *Handlers) RegisterAdmin(mux *http.ServeMux) {
	mux.HandleFunc("/admin/v1/protocols", h.handleAdminProtocols)
	mux.HandleFunc("/admin/v1/protocols/", h.handleAdminProtocols)
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
	if p.Workload || !p.HasScope("protocols:read") || !p.MFA || !p.HasRole("hub_protocol_reader") {
		auth.Error(w, 403, "forbidden")
		return
	}
	if r.Method != http.MethodGet {
		auth.Error(w, 405, "method_not_allowed")
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
	path := strings.TrimPrefix(r.URL.Path, "/admin/v1/protocols")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) > 2 {
		auth.Error(w, 404, "not_found")
		return
	}
	id := parts[0]
	// Audit is a prerequisite of diagnostic access, including empty/global reads.
	if _, err := h.store.db.ExecContext(r.Context(), "INSERT INTO protocol_access_audit(subject,requested_tenant,resource,action,mfa) VALUES($1,$2,$3,'READ',$4)", p.Subject, tenant, id, p.MFA); err != nil {
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
