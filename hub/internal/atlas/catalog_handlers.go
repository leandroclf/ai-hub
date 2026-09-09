package atlas

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"ai-hub/hub/internal/platform/auth"
)

type catalogCursor struct {
	Scope, Kind, Query, State, ID string
	Version                       int
}

func adminPermission(kind, action string) string {
	if kind == "provider-accounts" || kind == "credential-bindings" || kind == "providers" {
		if action == "read" {
			return "integrations:read"
		}
		return "integrations:write"
	}
	if kind == "contracts" {
		if action == "read" {
			return "finance:read"
		}
		return "finance:write"
	}
	return "catalog:" + action
}
func adminScope(r *http.Request) (string, bool) {
	p, ok := auth.FromContext(r.Context())
	if !ok {
		return "", false
	}
	tenant := r.URL.Query().Get("tenant_id")
	if tenant == "" {
		tenant = p.TenantID
	}
	if tenant == "" {
		tenant = "*"
	}
	if tenant != p.TenantID && !auth.Authorize(r.Context(), "admin:cross_tenant", tenant) {
		return "", false
	}
	return tenant, true
}
func (h *Handlers) registerCatalog(mux *http.ServeMux) {
	for _, kind := range []string{"clients", "applications", "services", "products", "offers", "technical-profiles", "providers", "provider-accounts", "credential-bindings", "contracts", "policies"} {
		mux.HandleFunc("/admin/v1/"+kind, h.handleCatalog)
		mux.HandleFunc("/admin/v1/"+kind+"/", h.handleCatalog)
	}
	mux.HandleFunc("/admin/v1/imports", h.handleImports)
	mux.HandleFunc("/v1/offers/resolve", h.handleOfferResolve)
}
func decodeBody(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("JSON inválido ou corpo excede 1 MiB: %w", err)
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("apenas um objeto JSON permitido: %w", err)
	}
	return nil
}
func catalogError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeErr(w, 404, "not_found", "recurso não encontrado no escopo")
	case errors.Is(err, ErrConflict):
		writeErr(w, 409, "immutable_version", err.Error())
	case errors.Is(err, ErrCapacityUnavailable):
		writeErr(w, 409, "capacity_unavailable", err.Error())
	case errors.Is(err, ErrRevision):
		writeErr(w, 412, "revision_conflict", err.Error())
	default:
		writeErr(w, 503, "catalog_unavailable", "catálogo indisponível; preserve a edição e tente novamente")
	}
}
func resourceResponse(w http.ResponseWriter, status int, r Resource) {
	w.Header().Set("ETag", fmt.Sprintf(`"%d"`, r.Revision))
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, status, r)
}
func (h *Handlers) handleCatalog(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.FromContext(r.Context())
	if !ok {
		writeErr(w, 401, "unauthenticated", "sessão necessária")
		return
	}
	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/admin/v1/"), "/"), "/")
	kind := parts[0]
	if !validKind(kind) {
		writeErr(w, 404, "not_found", "recurso desconhecido")
		return
	}
	tenant, ok := adminScope(r)
	if !ok {
		writeErr(w, 403, "forbidden", "escopo não autorizado")
		return
	}
	action := "read"
	if r.Method != "GET" {
		action = "write"
	}
	if len(parts) == 4 && parts[3] == "publish" {
		action = "publish"
	}
	if !auth.Authorize(r.Context(), adminPermission(kind, action), tenant) {
		writeErr(w, 403, "forbidden", "ação não autorizada")
		return
	}
	if len(parts) == 1 && r.Method == "GET" {
		limit := 25
		if raw := r.URL.Query().Get("limit"); raw != "" {
			v, err := strconv.Atoi(raw)
			if err != nil || v < 1 || v > 100 {
				writeErr(w, 400, "invalid_limit", "limit entre 1 e 100")
				return
			}
			limit = v
		}
		q, state := r.URL.Query().Get("q"), r.URL.Query().Get("state")
		if len(q) > 128 {
			writeErr(w, 400, "invalid_filter", "busca limitada a 128 caracteres")
			return
		}
		cursor := catalogCursor{Scope: p.Subject + ":" + tenant, Kind: kind, Query: q, State: state}
		if raw := r.URL.Query().Get("cursor"); raw != "" {
			b, e := base64.RawURLEncoding.DecodeString(raw)
			var prior catalogCursor
			if e != nil || json.Unmarshal(b, &prior) != nil || prior.Scope != cursor.Scope || prior.Kind != kind || prior.Query != q || prior.State != state {
				writeErr(w, 400, "invalid_cursor", "cursor incompatível com identidade, filtro ou escopo")
				return
			}
			cursor = prior
		}
		items, err := h.store.ListResources(r.Context(), kind, tenant, q, state, cursor.ID, cursor.Version, limit+1)
		if err != nil {
			catalogError(w, err)
			return
		}
		next := ""
		if len(items) > limit {
			items = items[:limit]
			last := items[len(items)-1]
			cursor.ID = last.ID
			cursor.Version = last.Version
			b, _ := json.Marshal(cursor)
			next = base64.RawURLEncoding.EncodeToString(b)
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, 200, map[string]any{"items": items, "next_cursor": next})
		return
	}
	if len(parts) == 1 && r.Method == "POST" {
		var resource Resource
		if err := decodeBody(w, r, &resource); err != nil {
			writeErr(w, 400, "invalid_body", err.Error())
			return
		}
		resource.Kind = kind
		if resource.TenantID == "" && tenant != "*" {
			resource.TenantID = tenant
		}
		if resource.TenantID != "" && !auth.Authorize(r.Context(), adminPermission(kind, "write"), resource.TenantID) {
			writeErr(w, 403, "forbidden", "cliente não autorizado")
			return
		}
		if resource.ID == "" || resource.Name == "" || resource.Version < 1 || !json.Valid(resource.Data) || strings.ContainsAny(resource.ID, "/\\") {
			writeErr(w, 422, "invalid_resource", "nome, identidade, versão positiva e dados JSON obrigatórios")
			return
		}
		out, err := h.store.SaveResource(r.Context(), resource, 0, p.Subject)
		if err != nil {
			catalogError(w, err)
			return
		}
		resourceResponse(w, 201, out)
		return
	}
	if len(parts) < 3 || len(parts) > 4 {
		writeErr(w, 404, "not_found", "use recurso/identidade/versão")
		return
	}
	version, err := strconv.Atoi(parts[2])
	if err != nil {
		writeErr(w, 400, "invalid_version", "versão inteira obrigatória")
		return
	}
	resource, err := h.store.GetResource(r.Context(), kind, parts[1], version)
	if err != nil {
		catalogError(w, err)
		return
	}
	if resource.TenantID != "" && (tenant != "*" && tenant != resource.TenantID || !auth.Authorize(r.Context(), adminPermission(kind, action), resource.TenantID)) {
		writeErr(w, 404, "not_found", "recurso não encontrado no escopo")
		return
	}
	if resource.TenantID == "" && r.Method != "GET" && !auth.Authorize(r.Context(), "admin:cross_tenant", "*") {
		writeErr(w, 403, "forbidden", "alteração global exige permissão explícita")
		return
	}
	if len(parts) == 3 && r.Method == "GET" {
		resourceResponse(w, 200, resource)
		return
	}
	if len(parts) == 4 && (parts[3] == "validate" || parts[3] == "simulate") && r.Method == "POST" {
		writeJSON(w, 200, h.store.ValidatePublication(r.Context(), resource))
		return
	}
	expected, err := strconv.ParseInt(strings.Trim(r.Header.Get("If-Match"), `"`), 10, 64)
	if err != nil || expected <= 0 {
		writeErr(w, 428, "revision_required", "If-Match da revisão atual obrigatório")
		return
	}
	if resource.Revision != expected {
		w.Header().Set("ETag", fmt.Sprintf(`"%d"`, resource.Revision))
		writeJSON(w, 412, map[string]any{"error": "revision_conflict", "message": ErrRevision.Error(), "current": resource})
		return
	}
	if len(parts) == 3 && r.Method == "PATCH" {
		if resource.State != "DRAFT" {
			catalogError(w, ErrConflict)
			return
		}
		var patch struct {
			Name string          `json:"name"`
			Data json.RawMessage `json:"data"`
		}
		if err := decodeBody(w, r, &patch); err != nil {
			writeErr(w, 400, "invalid_body", err.Error())
			return
		}
		if patch.Name != "" {
			resource.Name = patch.Name
		}
		if len(patch.Data) > 0 {
			resource.Data = patch.Data
		}
		out, err := h.store.SaveResource(r.Context(), resource, expected, p.Subject)
		if err != nil {
			catalogError(w, err)
			return
		}
		resourceResponse(w, 200, out)
		return
	}
	if len(parts) == 4 && r.Method == "POST" {
		var command struct {
			Reason      string `json:"reason"`
			ContentHash string `json:"content_hash"`
		}
		if err := decodeBody(w, r, &command); err != nil {
			writeErr(w, 400, "invalid_body", err.Error())
			return
		}
		if len(strings.TrimSpace(command.Reason)) < 8 {
			writeErr(w, 422, "reason_required", "explique o motivo com pelo menos 8 caracteres")
			return
		}
		switch parts[3] {
		case "publish":
			if resource.State != "DRAFT" {
				catalogError(w, ErrConflict)
				return
			}
			v := h.store.ValidatePublication(r.Context(), resource)
			if command.ContentHash != v.Hash {
				writeErr(w, 409, "validation_stale", "execute validação e revise o diff desta revisão")
				return
			}
			if !v.Valid {
				if resource.Kind == "clients" && v.FieldErrors["capacity_units"] != "" {
					if err := h.store.requestCapacity(r.Context(), resource, p.Subject); err != nil {
						catalogError(w, err)
						return
					}
				}
				writeJSON(w, 422, v)
				return
			}
			out, err := h.store.PublishResource(r.Context(), resource, expected, p.Subject, command.Reason, v)
			if err != nil {
				catalogError(w, err)
				return
			}
			resourceResponse(w, 200, out)
			return
		case "suspend":
			out, err := h.store.SuspendResource(r.Context(), resource, expected, p.Subject, command.Reason)
			if err != nil {
				catalogError(w, err)
				return
			}
			resourceResponse(w, 200, out)
			return
		}
	}
	writeErr(w, 405, "method_not_allowed", "método ou ação não permitido")
}
