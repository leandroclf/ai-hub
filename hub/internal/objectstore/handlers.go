package objectstore

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"ai-hub/hub/internal/platform/auth"
)

type Handlers struct{ catalog *Catalog }

func NewHandlers(c *Catalog) *Handlers { return &Handlers{catalog: c} }
func (h *Handlers) Register(mux *http.ServeMux) {
	mux.HandleFunc("/v1/files", h.handle)
	mux.HandleFunc("/v1/files/", h.handle)
}
func (h *Handlers) handle(w http.ResponseWriter, r *http.Request) {
	scope := "files:read"
	if r.Method != "GET" {
		scope = "files:write"
	}
	tenant, ok := auth.PublicTenant(r.Context(), scope)
	if !ok {
		if _, present := auth.FromContext(r.Context()); !present {
			auth.Error(w, 401, "unauthenticated")
		} else {
			auth.Error(w, 403, "forbidden")
		}
		return
	}
	write := func(status int, v any) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(v)
	}
	fail := func(e error) {
		switch {
		case errors.Is(e, ErrNotFound):
			auth.Error(w, 404, "file_not_found")
		case errors.Is(e, ErrIneligible):
			auth.Error(w, 409, "file_not_ready")
		case errors.Is(e, ErrIntegrity), errors.Is(e, ErrTooLarge), errors.Is(e, ErrPolicy):
			auth.Error(w, 422, "file_validation_failed")
		default:
			auth.Error(w, 503, "object_custody_unavailable")
		}
	}
	decode := func(v any) bool {
		r.Body = http.MaxBytesReader(w, r.Body, 128*1024)
		d := json.NewDecoder(r.Body)
		d.DisallowUnknownFields()
		if d.Decode(v) != nil || d.Decode(&struct{}{}) != io.EOF {
			auth.Error(w, 400, "invalid_body")
			return false
		}
		return true
	}
	path := strings.TrimPrefix(r.URL.Path, "/v1/files")
	switch {
	case path == "" && r.Method == "POST":
		var q UploadRequest
		if !decode(&q) {
			return
		}
		v, e := h.catalog.CreateSession(r.Context(), tenant, q)
		if e != nil {
			fail(e)
			return
		}
		write(201, v)
	case strings.HasSuffix(path, "/complete") && r.Method == "POST":
		id := strings.TrimSuffix(strings.TrimPrefix(path, "/"), "/complete")
		var q struct {
			Parts []CompletedPart `json:"parts"`
		}
		if !decode(&q) {
			return
		}
		v, e := h.catalog.Complete(r.Context(), tenant, id, q.Parts)
		if e != nil {
			fail(e)
			return
		}
		write(200, v)
	case strings.HasSuffix(path, "/download") && r.Method == "GET":
		id := strings.TrimSuffix(strings.TrimPrefix(path, "/"), "/download")
		url, e := h.catalog.Download(r.Context(), tenant, id)
		if e != nil {
			fail(e)
			return
		}
		write(200, map[string]any{"url": url, "expires_in_seconds": 120})
	case strings.HasPrefix(path, "/") && r.Method == "GET":
		v, e := h.catalog.Resolve(r.Context(), tenant, strings.TrimPrefix(path, "/"))
		if e != nil {
			fail(e)
			return
		}
		write(200, v)
	default:
		auth.Error(w, 405, "method_not_allowed")
	}
}
