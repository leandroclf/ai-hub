package atlas

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"ai-hub/hub/internal/platform/auth"
)

type ImportItem struct {
	ID         string `json:"id"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	AuthType   string `json:"auth_type"`
	State      string `json:"state"`
	Difference string `json:"difference"`
}
type ImportBatch struct {
	ID              string       `json:"id"`
	SourceHash      string       `json:"source_hash"`
	State           string       `json:"state"`
	Items           []ImportItem `json:"items"`
	ExecutableCount int          `json:"executable_count"`
}

func sameImportItem(a, b ImportItem) bool {
	return a.ID == b.ID && a.Method == b.Method && a.Path == b.Path && a.AuthType == b.AuthType && a.State == b.State
}

// Sanitization uses an allowlist: credentials, headers, examples, scripts,
// environments, URL authority and query values never enter the retained model.
func SanitizeImport(raw json.RawMessage) ([]ImportItem, error) {
	var root map[string]any
	if len(raw) > 1024*1024 || json.Unmarshal(raw, &root) != nil {
		return nil, fmt.Errorf("arquivo JSON inválido ou maior que 1 MiB")
	}
	items := map[string]ImportItem{}
	count := 0
	add := func(method, path, authType string) error {
		method = strings.ToUpper(method)
		if !contains([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}, method) {
			return fmt.Errorf("método HTTP inválido")
		}
		if strings.HasPrefix(path, "{{") {
			if end := strings.Index(path, "}}"); end >= 0 {
				path = path[end+2:]
			}
		}
		if u, err := url.Parse(path); err == nil {
			path = u.EscapedPath()
		} else {
			return fmt.Errorf("path inválido")
		}
		if path == "" {
			path = "/"
		}
		if len(path) > 1024 {
			return fmt.Errorf("path excessivo")
		}
		// Literal high-entropy segments are omitted from inventory; no secret-bearing URL is retained.
		segments := strings.Split(path, "/")
		for i, segment := range segments {
			if len(segment) > 48 {
				segments[i] = "{redacted}"
			}
		}
		path = strings.Join(segments, "/")
		authType = strings.ToUpper(authType)
		if !contains([]string{"NONE", "BASIC", "BEARER", "OAUTH2", "APIKEY", "MTLS"}, authType) {
			authType = "UNSPECIFIED"
		}
		id := contentHash([]string{method, path})[:32]
		items[id] = ImportItem{ID: id, Method: method, Path: path, AuthType: authType, State: "IMPORTED_NOT_EXECUTABLE", Difference: "NEW"}
		count++
		if count > 5000 {
			return fmt.Errorf("lote excede 5000 operações")
		}
		return nil
	}
	var walk func([]any, int) error
	walk = func(list []any, depth int) error {
		if depth > 16 {
			return fmt.Errorf("collection excede profundidade 16")
		}
		for _, entry := range list {
			m, ok := entry.(map[string]any)
			if !ok {
				continue
			}
			if sub, ok := m["item"].([]any); ok {
				if err := walk(sub, depth+1); err != nil {
					return err
				}
			}
			if req, ok := m["request"].(map[string]any); ok {
				method, _ := req["method"].(string)
				path, _ := req["url"].(string)
				if u, ok := req["url"].(map[string]any); ok {
					path, _ = u["raw"].(string)
				}
				authType := "UNSPECIFIED"
				if a, ok := req["auth"].(map[string]any); ok {
					authType, _ = a["type"].(string)
				}
				if err := add(method, path, authType); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if list, ok := root["item"].([]any); ok {
		if err := walk(list, 0); err != nil {
			return nil, err
		}
	} else if paths, ok := root["paths"].(map[string]any); ok {
		for path, entry := range paths {
			if operations, ok := entry.(map[string]any); ok {
				for method := range operations {
					if contains([]string{"get", "post", "put", "patch", "delete", "head", "options"}, method) {
						if err := add(method, path, "UNSPECIFIED"); err != nil {
							return nil, err
						}
					}
				}
			}
		}
	} else {
		return nil, fmt.Errorf("esperado Postman item ou OpenAPI paths")
	}
	out := []ImportItem{}
	for _, item := range items {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
func (h *Handlers) handleImports(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.FromContext(r.Context())
	if !ok {
		writeErr(w, 401, "unauthenticated", "sessão necessária")
		return
	}
	tenant, ok := adminScope(r)
	if !ok {
		writeErr(w, 403, "forbidden", "escopo não autorizado")
		return
	}
	scope := "catalog:read"
	if r.Method == "POST" {
		scope = "catalog:write"
	}
	if !roleAllows(p, scope) || !auth.Authorize(r.Context(), scope, tenant) {
		writeErr(w, 403, "forbidden", "ação não autorizada")
		return
	}
	if r.Method == "GET" {
		rows, err := h.store.db.QueryContext(r.Context(), `SELECT id,source_hash,state,items FROM catalog_imports WHERE tenant_id=$1 ORDER BY created_at DESC LIMIT 100`, tenant)
		if err != nil {
			catalogError(w, err)
			return
		}
		defer rows.Close()
		batches := []ImportBatch{}
		for rows.Next() {
			var b ImportBatch
			var items []byte
			if err = rows.Scan(&b.ID, &b.SourceHash, &b.State, &items); err != nil {
				catalogError(w, err)
				return
			}
			_ = json.Unmarshal(items, &b.Items)
			batches = append(batches, b)
		}
		writeJSON(w, 200, map[string]any{"items": batches, "next_cursor": ""})
		return
	}
	if r.Method != "POST" {
		writeErr(w, 405, "method_not_allowed", "use GET ou POST")
		return
	}
	var request struct {
		Collection json.RawMessage `json:"collection"`
	}
	if err := decodeBody(w, r, &request); err != nil {
		writeErr(w, 400, "invalid_body", err.Error())
		return
	}
	items, err := SanitizeImport(request.Collection)
	if err != nil {
		writeErr(w, 422, "invalid_collection", err.Error())
		return
	}
	hash := contentHash(items)
	id := hash[:32]
	var previous []byte
	err = h.store.db.QueryRowContext(r.Context(), `SELECT items FROM catalog_imports WHERE tenant_id=$1 ORDER BY created_at DESC LIMIT 1`, tenant).Scan(&previous)
	if errors.Is(err, sql.ErrNoRows) {
		previous = nil
	} else if err != nil {
		catalogError(w, err)
		return
	}
	// Compare only the sanitized inventory against the latest prior batch;
	// executable catalog rows are never created or modified by an import.
	var prior []ImportItem
	if len(previous) > 0 && json.Unmarshal(previous, &prior) != nil {
		writeErr(w, 503, "catalog_unavailable", "inventário anterior inválido; preserve o lote e tente novamente")
		return
	}
	priorByID := make(map[string]ImportItem, len(prior))
	for _, item := range prior {
		priorByID[item.ID] = item
	}
	for i := range items {
		old, ok := priorByID[items[i].ID]
		switch {
		case !ok:
			items[i].Difference = "NEW"
		case sameImportItem(old, items[i]):
			items[i].Difference = "UNCHANGED"
		default:
			items[i].Difference = "CHANGED"
		}
	}
	body, _ := json.Marshal(items)
	var storedID string
	err = h.store.db.QueryRowContext(r.Context(), `INSERT INTO catalog_imports(id,source_hash,source_name,actor,tenant_id,items) VALUES($1,$2,'sanitized-inventory',$3,$4,$5) ON CONFLICT(tenant_id,source_hash) DO UPDATE SET source_hash=EXCLUDED.source_hash RETURNING id`, contentHash([]string{tenant, id})[:32], hash, p.Subject, tenant, body).Scan(&storedID)
	if err != nil {
		catalogError(w, err)
		return
	}
	_ = storedID
	writeJSON(w, 201, ImportBatch{ID: contentHash([]string{tenant, id})[:32], SourceHash: hash, State: "STAGED", Items: items})
}
