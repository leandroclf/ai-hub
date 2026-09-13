package libra

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ai-hub/hub/internal/platform/auth"
	"ai-hub/hub/internal/queue"
)

type ReserveRequest struct {
	TenantID   string  `json:"tenant_id"`
	ProtocolID string  `json:"protocol_id"`
	Amount     Decimal `json:"amount"`
	Currency   string  `json:"currency"`
}
type Handlers struct {
	store *Store
	log   *slog.Logger
}

func NewHandlers(store *Store, log *slog.Logger) *Handlers { return &Handlers{store: store, log: log} }
func (h *Handlers) Register(mux *http.ServeMux) {
	mux.HandleFunc("/internal/reservations", h.handleReserve)
	mux.HandleFunc("/admin/v1/finance/", h.handleAdmin)
}
func decodeBody(w http.ResponseWriter, r *http.Request, out any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 128*1024)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if d.Decode(out) != nil || d.Decode(&struct{}{}) != io.EOF {
		auth.Error(w, 400, "invalid_body")
		return false
	}
	return true
}

func exactInt64(raw json.RawMessage) (int64, error) {
	value := strings.TrimSpace(string(raw))
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		var text string
		if err := json.Unmarshal(raw, &text); err != nil {
			return 0, err
		}
		value = text
	}
	if value == "" || strings.ContainsAny(value, ".eE") {
		return 0, errors.New("integer required")
	}
	return strconv.ParseInt(value, 10, 64)
}
func respond(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func (h *Handlers) fail(w http.ResponseWriter, e error) {
	switch {
	case errors.Is(e, ErrNotFound):
		auth.Error(w, 404, "not_found")
	case errors.Is(e, ErrLimitExceeded):
		auth.Error(w, 409, "limit_exceeded")
	case errors.Is(e, ErrLimitMissing):
		auth.Error(w, 409, "limit_not_configured")
	case errors.Is(e, ErrConflict):
		auth.Error(w, 409, "immutable_conflict")
	case errors.Is(e, ErrIncomplete):
		auth.Error(w, 409, "financial_obligations_open")
	default:
		h.log.Error("finance request failed", "error", e)
		auth.Error(w, 503, "finance_unavailable")
	}
}
func (h *Handlers) handleReserve(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		auth.Error(w, 405, "method_not_allowed")
		return
	}
	p, ok := auth.FromContext(r.Context())
	if !ok {
		auth.Error(w, 401, "unauthenticated")
		return
	}
	var q ReserveRequest
	if !decodeBody(w, r, &q) {
		return
	}
	if !p.Workload || q.TenantID == "" || !auth.Authorize(r.Context(), "finance:reserve", q.TenantID) {
		auth.Error(w, 403, "forbidden")
		return
	}
	if e := h.store.ReserveExact(r.Context(), q.TenantID, q.ProtocolID, string(q.Amount), q.Currency); e != nil {
		h.fail(w, e)
		return
	}
	respond(w, 201, map[string]string{"status": "reserved"})
}
func (h *Handlers) handleAdmin(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.FromContext(r.Context())
	if !ok {
		auth.Error(w, 401, "unauthenticated")
		return
	}
	// Finance admin is a human, MFA-authenticated surface. Workload tokens
	// keep their separate /internal/reservations authority and must never be
	// able to turn a finance scope into an administrative read or mutation.
	if p.Workload || !p.MFA || (!p.HasRole("hub_admin") && !p.HasRole("hub_protocol_reader") && !p.HasRole("tenant_reader") && !p.HasRole("tenant_operator")) {
		auth.Error(w, 403, "forbidden")
		return
	}
	tenant := r.URL.Query().Get("tenant_id")
	if tenant == "" {
		tenant = p.TenantID
	}
	if tenant == "" {
		auth.Error(w, 422, "tenant_required")
		return
	}
	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/admin/v1/finance/"), "/")
	scope := "finance:read"
	if r.Method != "GET" {
		scope = "finance:write"
		if !p.HasRole("hub_admin") {
			auth.Error(w, 403, "forbidden")
			return
		}
	}
	if strings.HasSuffix(path, "/approve") {
		scope = "finance:approve"
	}
	if !auth.Authorize(r.Context(), scope, tenant) {
		auth.Error(w, 403, "forbidden")
		return
	}
	if _, e := h.store.db.ExecContext(r.Context(), `INSERT INTO finance_action_audit(tenant_id,actor,action,resource) VALUES($1,$2,$3,$4)`, tenant, p.Subject, r.Method, path); e != nil {
		h.fail(w, e)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	kind := r.URL.Query().Get("kind")
	if kind != "" && kind != "REVENUE" && kind != "COST" {
		auth.Error(w, 422, "invalid_kind")
		return
	}
	after := int64(0)
	if cursor := r.URL.Query().Get("cursor"); cursor != "" {
		b, e := base64.RawURLEncoding.DecodeString(cursor)
		var c struct {
			Tenant string `json:"tenant"`
			Filter string `json:"filter"`
			After  int64  `json:"after"`
		}
		if e != nil || json.Unmarshal(b, &c) != nil || c.Tenant != tenant || c.Filter != path+":"+kind || c.After < 0 {
			auth.Error(w, 400, "invalid_cursor")
			return
		}
		after = c.After
	}
	page := func(items any, last int64, count int) {
		next := ""
		if count == limit {
			b, _ := json.Marshal(map[string]any{"tenant": tenant, "filter": path + ":" + kind, "after": last})
			next = base64.RawURLEncoding.EncodeToString(b)
		}
		respond(w, 200, map[string]any{"items": items, "next_cursor": next})
	}
	switch {
	case path == "accounts" && r.Method == "GET":
		v, e := h.store.Accounts(r.Context(), tenant)
		if e != nil {
			h.fail(w, e)
			return
		}
		respond(w, 200, map[string]any{"items": v})
	case path == "facts" && r.Method == "GET":
		v, e := h.store.Facts(r.Context(), tenant, kind, after, limit)
		if e != nil {
			h.fail(w, e)
			return
		}
		last := int64(0)
		if len(v) > 0 {
			last = v[len(v)-1].ID
		}
		page(v, last, len(v))
	case path == "journal" && r.Method == "GET":
		v, e := h.store.Journal(r.Context(), tenant, after, limit)
		if e != nil {
			h.fail(w, e)
			return
		}
		last := int64(0)
		if len(v) > 0 {
			last = v[len(v)-1].ID
		}
		page(v, last, len(v))
	case path == "adjustments" && r.Method == "GET":
		rows, e := h.store.db.QueryContext(r.Context(), `SELECT id,tenant_id,origin_batch_id,reason,prepared_by,COALESCE(approved_by,''),state FROM finance_adjustments WHERE tenant_id=$1 ORDER BY id LIMIT $2`, tenant, limit)
		if e != nil {
			h.fail(w, e)
			return
		}
		defer rows.Close()
		items := []map[string]any{}
		for rows.Next() {
			var id, owner, origin, reason, prepared, approved, state string
			if e = rows.Scan(&id, &owner, &origin, &reason, &prepared, &approved, &state); e != nil {
				h.fail(w, e)
				return
			}
			items = append(items, map[string]any{"id": id, "tenant_id": owner, "origin_batch_id": origin, "reason": reason, "prepared_by": prepared, "approved_by": approved, "state": state})
		}
		if e = rows.Err(); e != nil {
			h.fail(w, e)
			return
		}
		respond(w, 200, map[string]any{"items": items})
	case path == "adjustments" && r.Method == "POST":
		var q struct {
			Origin string `json:"origin_batch_id"`
			Reason string `json:"reason"`
		}
		if !decodeBody(w, r, &q) {
			return
		}
		id := r.Header.Get("Idempotency-Key")
		if id == "" || q.Reason == "" {
			auth.Error(w, 422, "identity_and_reason_required")
			return
		}
		v, e := h.store.PrepareAdjustment(r.Context(), id, tenant, q.Origin, q.Reason, p.Subject)
		if e != nil {
			h.fail(w, e)
			return
		}
		respond(w, 201, v)
	case strings.HasPrefix(path, "adjustments/") && strings.HasSuffix(path, "/approve") && r.Method == "POST":
		id := strings.TrimSuffix(strings.TrimPrefix(path, "adjustments/"), "/approve")
		if e := h.store.ApproveAdjustment(r.Context(), id, tenant, p.Subject); e != nil {
			h.fail(w, e)
			return
		}
		respond(w, 200, map[string]string{"status": "approved"})
	case path == "periods" && r.Method == "POST":
		var q struct {
			Start string `json:"period_start"`
			End   string `json:"period_end"`
		}
		if !decodeBody(w, r, &q) {
			return
		}
		start, e := parsePeriodTime(q.Start)
		if e != nil {
			auth.Error(w, 422, "invalid_period_start")
			return
		}
		end, e := parsePeriodTime(q.End)
		if e != nil {
			auth.Error(w, 422, "invalid_period_end")
			return
		}
		id := r.Header.Get("Idempotency-Key")
		if id == "" {
			auth.Error(w, 422, "idempotency_key_required")
			return
		}
		v, e := h.store.ClosePeriod(r.Context(), id, tenant, p.Subject, start, end)
		if e != nil {
			h.fail(w, e)
			return
		}
		respond(w, 201, v)
	case path == "periods" && r.Method == "GET":
		rows, e := h.store.db.QueryContext(r.Context(), `SELECT id,period_start,period_end,checksum FROM settlement_periods WHERE tenant_id=$1 ORDER BY period_start DESC LIMIT $2`, tenant, limit)
		if e != nil {
			h.fail(w, e)
			return
		}
		defer rows.Close()
		items := []map[string]any{}
		for rows.Next() {
			var id, hash string
			var start, end time.Time
			if e = rows.Scan(&id, &start, &end, &hash); e != nil {
				h.fail(w, e)
				return
			}
			items = append(items, map[string]any{"id": id, "period_start": start, "period_end": end, "checksum": hash})
		}
		if e = rows.Err(); e != nil {
			h.fail(w, e)
			return
		}
		respond(w, 200, map[string]any{"items": items})
	case strings.HasPrefix(path, "exports/") && r.Method == "GET":
		v, e := h.store.Export(r.Context(), tenant, strings.TrimPrefix(path, "exports/"))
		if e != nil {
			h.fail(w, e)
			return
		}
		respond(w, 200, v)
	case strings.HasPrefix(path, "exports/") && strings.HasSuffix(path, "/receipts") && r.Method == "POST":
		var q struct {
			ReceiptID string `json:"receipt_id"`
			Checksum  string `json:"checksum"`
		}
		if !decodeBody(w, r, &q) {
			return
		}
		id := strings.TrimSuffix(strings.TrimPrefix(path, "exports/"), "/receipts")
		if e := h.store.Receipt(r.Context(), tenant, id, q.ReceiptID, q.Checksum, p.Subject); e != nil {
			h.fail(w, e)
			return
		}
		respond(w, 201, map[string]string{"status": "received"})
	case path == "disputes" && r.Method == "POST":
		var q struct {
			FactID     json.RawMessage `json:"fact_id"`
			Amount     Decimal         `json:"amount"`
			Reason     string          `json:"reason"`
			EvidenceID string          `json:"evidence_id"`
		}
		if !decodeBody(w, r, &q) {
			return
		}
		factID, e := exactInt64(q.FactID)
		if e != nil {
			auth.Error(w, 400, "invalid_fact_id")
			return
		}
		id, e := h.store.Dispute(r.Context(), tenant, factID, q.Amount, q.Reason, q.EvidenceID, p.Subject)
		if e != nil {
			h.fail(w, e)
			return
		}
		respond(w, 201, map[string]string{"id": id, "status": "OPEN"})
	case path == "disputes" && r.Method == "GET":
		rows, e := h.store.db.QueryContext(r.Context(), `SELECT id,fact_id,amount::text,reason,evidence_id,state FROM finance_disputes WHERE tenant_id=$1 ORDER BY created_at DESC LIMIT $2`, tenant, limit)
		if e != nil {
			h.fail(w, e)
			return
		}
		defer rows.Close()
		items := []map[string]any{}
		for rows.Next() {
			var id, amount, reason, evidence, state string
			var fact int64
			if e = rows.Scan(&id, &fact, &amount, &reason, &evidence, &state); e != nil {
				h.fail(w, e)
				return
			}
			items = append(items, map[string]any{"id": id, "fact_id": strconv.FormatInt(fact, 10), "amount": amount, "reason": reason, "evidence_id": evidence, "state": state})
		}
		if e = rows.Err(); e != nil {
			h.fail(w, e)
			return
		}
		respond(w, 200, map[string]any{"items": items})
	case path == "quarantine" && r.Method == "GET":
		// finance_quarantine has no tenant_id (some reasons, e.g. INVALID_JSON,
		// never parse an identity) — listing is a nominal global-scope surface,
		// audited above, not a per-tenant read.
		if !p.HasRole("hub_admin") {
			auth.Error(w, 403, "forbidden")
			return
		}
		rows, e := h.store.db.QueryContext(r.Context(), `SELECT id,consumer,event_id,reason,created_at FROM finance_quarantine WHERE disposition='QUARANTINED' ORDER BY created_at LIMIT $1`, limit)
		if e != nil {
			h.fail(w, e)
			return
		}
		defer rows.Close()
		items := []map[string]any{}
		for rows.Next() {
			var id, consumer, eventID, reason string
			var createdAt time.Time
			if e = rows.Scan(&id, &consumer, &eventID, &reason, &createdAt); e != nil {
				h.fail(w, e)
				return
			}
			items = append(items, map[string]any{"id": id, "consumer": consumer, "event_id": eventID, "reason": reason, "created_at": createdAt})
		}
		if e = rows.Err(); e != nil {
			h.fail(w, e)
			return
		}
		respond(w, 200, map[string]any{"items": items})
	case strings.HasPrefix(path, "quarantine/") && strings.HasSuffix(path, "/replay") && r.Method == "POST":
		// R5-DAD-04: replay reprocesses the original economic fact through the
		// same custody path (ApplyEvent/Quarantine); it never re-issues SUBMIT
		// to the provider. Marking REPLAYED only happens after ProcessEnvelope
		// durably accepts or re-quarantines the fact.
		id := strings.TrimSuffix(strings.TrimPrefix(path, "quarantine/"), "/replay")
		consumer, payload, e := h.store.ReplayQuarantine(r.Context(), id, p.Subject)
		if e != nil {
			h.fail(w, e)
			return
		}
		var env queue.Envelope
		if json.Unmarshal(payload, &env) != nil || env.TenantID != tenant {
			auth.Error(w, 409, "quarantine_payload_not_replayable")
			return
		}
		if e = h.store.ProcessEnvelope(r.Context(), consumer, env); e != nil {
			h.fail(w, e)
			return
		}
		if e = h.store.MarkQuarantineReplayed(r.Context(), id, p.Subject); e != nil {
			h.fail(w, e)
			return
		}
		respond(w, 200, map[string]string{"status": "replayed"})
	default:
		auth.Error(w, 404, "not_found")
	}
}

func parsePeriodTime(value string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.Parse("2006-01-02", value); err == nil {
		return t.UTC(), nil
	}
	return time.Time{}, errors.New("invalid period time")
}
