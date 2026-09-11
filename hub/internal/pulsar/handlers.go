package pulsar

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"ai-hub/hub/internal/platform/auth"
	"ai-hub/hub/internal/platform/egress"
	"ai-hub/hub/internal/platform/idgen"
)

type Handlers struct{ store *Store }

func NewHandlers(s *Store) *Handlers { return &Handlers{store: s} }
func (h *Handlers) Register(mux *http.ServeMux) {
	mux.HandleFunc("/internal/destinations", func(w http.ResponseWriter, r *http.Request) { auth.Error(w, 410, "use_versioned_destinations") })
	mux.HandleFunc("/admin/v1/destinations", h.destinations)
	mux.HandleFunc("/admin/v1/deliveries", h.deliveries)
	mux.HandleFunc("/admin/v1/deliveries/", h.deliveries)
}
func respond(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64*1024))
	d.DisallowUnknownFields()
	if d.Decode(v) != nil || d.Decode(&struct{}{}) != io.EOF {
		auth.Error(w, 400, "invalid_body")
		return false
	}
	return true
}
func adminScope(w http.ResponseWriter, r *http.Request) (string, auth.Principal, bool) {
	p, ok := auth.FromContext(r.Context())
	if !ok {
		auth.Error(w, 401, "unauthenticated")
		return "", p, false
	}
	readRole := p.HasRole("hub_admin") || p.HasRole("tenant_reader") || p.HasRole("tenant_operator") || p.HasRole("hub_protocol_reader")
	writeRole := p.HasRole("hub_admin") || p.HasRole("tenant_operator")
	tenant := r.URL.Query().Get("tenant_id")
	if tenant == "" {
		tenant = p.TenantID
	}
	scope := "deliveries:read"
	if r.Method != "GET" {
		scope = "deliveries:write"
	}
	if p.Workload || !p.MFA || tenant == "" || !auth.Authorize(r.Context(), scope, tenant) || (r.Method == "GET" && !readRole) || (r.Method != "GET" && !writeRole) {
		auth.Error(w, 403, "forbidden")
		return "", p, false
	}
	return tenant, p, true
}
func (h *Handlers) destinations(w http.ResponseWriter, r *http.Request) {
	tenant, p, ok := adminScope(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case "GET":
		rows, err := h.store.db.QueryContext(r.Context(), "SELECT id,version,application_id,url,state,max_attempts,timeout_seconds FROM webhook_destination_versions WHERE tenant_id=$1 ORDER BY created_at DESC LIMIT 100", tenant)
		if err != nil {
			auth.Error(w, 503, "destinations_unavailable")
			return
		}
		defer rows.Close()
		items := []map[string]any{}
		for rows.Next() {
			var id, application, url, state string
			var version, max, timeout int
			if rows.Scan(&id, &version, &application, &url, &state, &max, &timeout) != nil {
				auth.Error(w, 503, "destinations_unavailable")
				return
			}
			items = append(items, map[string]any{"id": id, "version": version, "application_id": application, "url": url, "state": state, "max_attempts": max, "timeout_seconds": timeout})
		}
		if rows.Err() != nil {
			auth.Error(w, 503, "destinations_unavailable")
			return
		}
		respond(w, 200, map[string]any{"items": items})
	case "POST":
		var q struct {
			ID            string `json:"id"`
			Version       int    `json:"version"`
			ApplicationID string `json:"application_id"`
			URL           string `json:"url"`
			SecretRef     string `json:"secret_ref"`
			SecretVersion string `json:"secret_version"`
			State         string `json:"state"`
			MaxAttempts   int    `json:"max_attempts"`
			Timeout       int    `json:"timeout_seconds"`
			Reason        string `json:"reason"`
		}
		if !decode(w, r, &q) {
			return
		}
		if q.ID == "" {
			q.ID = idgen.New()
		}
		if q.Version < 1 || q.SecretRef == "" || q.SecretVersion == "" || (q.State != "ACTIVE" && q.State != "SUSPENDED") || q.MaxAttempts < 1 || q.MaxAttempts > 20 || q.Timeout < 1 || q.Timeout > 15 || len(strings.TrimSpace(q.Reason)) < 8 || egress.ValidateURL(q.URL) != nil {
			auth.Error(w, 422, "invalid_destination_policy")
			return
		}
		tx, err := h.store.db.BeginTx(r.Context(), nil)
		if err != nil {
			auth.Error(w, 503, "destination_custody_unavailable")
			return
		}
		defer tx.Rollback()
		if _, err = tx.ExecContext(r.Context(), "SELECT pg_advisory_xact_lock(hashtextextended($1,0))", q.ID); err != nil {
			auth.Error(w, 503, "destination_custody_unavailable")
			return
		}
		var latest int
		var owner, ownerApplication string
		if err = tx.QueryRowContext(r.Context(), "SELECT COALESCE(max(version),0),COALESCE(max(tenant_id),''),COALESCE(max(application_id),'') FROM webhook_destination_versions WHERE id=$1", q.ID).Scan(&latest, &owner, &ownerApplication); err != nil {
			auth.Error(w, 422, "invalid_destination_id")
			return
		}
		if (owner != "" && owner != tenant) || (owner != "" && ownerApplication != q.ApplicationID) || q.Version != latest+1 {
			auth.Error(w, 409, "destination_version_conflict")
			return
		}
		_, err = tx.ExecContext(r.Context(), `INSERT INTO webhook_destination_versions(id,version,application_id,tenant_id,cell_id,url,secret_ref,secret_version,max_attempts,timeout_seconds,state,created_by) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`, q.ID, q.Version, q.ApplicationID, tenant, os.Getenv("CELL_ID"), q.URL, q.SecretRef, q.SecretVersion, q.MaxAttempts, q.Timeout, q.State, p.Subject)
		if err != nil {
			auth.Error(w, 503, "destination_custody_unavailable")
			return
		}
		_, err = tx.ExecContext(r.Context(), "INSERT INTO webhook_audit(tenant_id,subject,action,resource,reason) VALUES($1,$2,'DESTINATION_VERSION',$3,$4)", tenant, p.Subject, q.ID, q.Reason)
		if err != nil || tx.Commit() != nil {
			auth.Error(w, 503, "destination_custody_unavailable")
			return
		}
		respond(w, 201, map[string]any{"id": q.ID, "version": q.Version, "state": q.State})
	default:
		auth.Error(w, 405, "method_not_allowed")
	}
}
func (h *Handlers) deliveries(w http.ResponseWriter, r *http.Request) {
	tenant, p, ok := adminScope(w, r)
	if !ok {
		return
	}
	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/admin/v1/deliveries"), "/"), "/")
	id := parts[0]
	if len(parts) == 2 && parts[1] == "redeliver" && r.Method == "POST" {
		var q struct {
			Reason string `json:"reason"`
		}
		if !decode(w, r, &q) {
			return
		}
		if len(strings.TrimSpace(q.Reason)) < 8 {
			auth.Error(w, 422, "reason_required")
			return
		}
		tx, err := h.store.db.BeginTx(r.Context(), nil)
		if err != nil {
			auth.Error(w, 503, "delivery_unavailable")
			return
		}
		defer tx.Rollback()
		result, err := tx.ExecContext(r.Context(), "UPDATE deliveries SET state='RETRY_SCHEDULED',next_attempt_at=clock_timestamp() WHERE delivery_id=$1 AND tenant_id=$2 AND state='EXHAUSTED' AND representation IS NOT NULL", id, tenant)
		if err != nil {
			auth.Error(w, 503, "delivery_unavailable")
			return
		}
		n, _ := result.RowsAffected()
		if n != 1 {
			auth.Error(w, 409, "delivery_not_exhausted")
			return
		}
		if _, err = tx.ExecContext(r.Context(), "INSERT INTO webhook_audit(tenant_id,subject,action,resource,reason) VALUES($1,$2,'REDELIVER_SAME_BYTES',$3,$4)", tenant, p.Subject, id, q.Reason); err != nil || tx.Commit() != nil {
			auth.Error(w, 503, "delivery_unavailable")
			return
		}
		respond(w, 200, map[string]any{"delivery_id": id, "state": "RETRY_SCHEDULED"})
		return
	}
	if r.Method != "GET" {
		auth.Error(w, 405, "method_not_allowed")
		return
	}
	limit := 25
	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 100 {
			auth.Error(w, 400, "invalid_limit")
			return
		}
		limit = n
	}
	after := r.URL.Query().Get("cursor")
	status := r.URL.Query().Get("status")
	rows, err := h.store.db.QueryContext(r.Context(), `SELECT delivery_id,protocol_id,state,attempts_count,next_attempt_at,destination_id,destination_version,body_sha256 FROM deliveries WHERE tenant_id=$1 AND ($2='' OR delivery_id::text=$2) AND delivery_id::text>$3 AND ($4='' OR state=$4) ORDER BY delivery_id::text LIMIT $5`, tenant, id, after, status, limit+1)
	if err != nil {
		auth.Error(w, 503, "deliveries_unavailable")
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	last := ""
	for rows.Next() {
		var id, protocol, state, hash string
		var dest sql.NullString
		var version sql.NullInt64
		var count int
		var next time.Time
		if rows.Scan(&id, &protocol, &state, &count, &next, &dest, &version, &hash) != nil {
			auth.Error(w, 503, "deliveries_unavailable")
			return
		}
		var destinationID any
		if dest.Valid {
			destinationID = dest.String
		}
		var destinationVersion any
		if version.Valid {
			destinationVersion = version.Int64
		}
		items = append(items, map[string]any{"delivery_id": id, "protocol_id": protocol, "tenant_id": tenant, "state": state, "attempts_count": count, "next_attempt_at": next, "destination_id": destinationID, "destination_version": destinationVersion, "body_sha256": hash})
		if len(items) == limit {
			last = id
		}
	}
	if rows.Err() != nil {
		auth.Error(w, 503, "deliveries_unavailable")
		return
	}
	if id != "" {
		if len(items) != 1 {
			auth.Error(w, 404, "not_found")
			return
		}
		respond(w, 200, items[0])
		return
	}
	next := ""
	if len(items) > limit {
		items = items[:limit]
		next = last
	}
	respond(w, 200, map[string]any{"items": items, "next_cursor": next})
}
