package atlas

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"time"

	"ai-hub/hub/internal/platform/auth"
)

type OfferSnapshot struct {
	Account          Resource   `json:"provider_account"`
	Offer            Resource   `json:"offer"`
	Target           Resource   `json:"target"`
	TechnicalProfile Resource   `json:"technical_profile"`
	PurchaseContract Resource   `json:"purchase_contract"`
	SaleContract     Resource   `json:"sale_contract"`
	Binding          Resource   `json:"binding"`
	Services         []Resource `json:"services"`
	SelectedRoute    Route      `json:"selected_route"`
	SelectionReason  string     `json:"selection_reason"`
	Hash             string     `json:"content_hash"`
	ValidUntil       time.Time  `json:"valid_until"`
}

func (s *Store) ResolveOffer(ctx context.Context, tenant, application, service, account string) (OfferSnapshot, error) {
	resources, err := s.ListResources(ctx, "offers", tenant, "", "PUBLISHED", "", 0, 101)
	if err != nil {
		return OfferSnapshot{}, err
	}
	if len(resources) > 100 {
		return OfferSnapshot{}, errors.New("ofertas ambíguas ou excesso no escopo")
	}
	var selected *Resource
	var data CatalogData
	for _, r := range resources {
		if r.TenantID != tenant {
			continue
		}
		d, e := DecodeCatalogData(r)
		if e != nil || d.ApplicationID != application || d.TargetID != service {
			continue
		}
		now := time.Now()
		if d.ValidFrom != nil && now.Before(*d.ValidFrom) || d.ValidUntil != nil && !now.Before(*d.ValidUntil) {
			continue
		}
		if selected != nil {
			return OfferSnapshot{}, errors.New("mais de uma oferta vigente; resolução recusada")
		}
		copy := r
		selected = &copy
		data = d
	}
	if selected == nil {
		return OfferSnapshot{}, ErrNotFound
	}
	validation := s.ValidatePublication(ctx, *selected)
	if !validation.Valid {
		return OfferSnapshot{}, ErrCredentialUnavailable
	}
	routes := append([]Route(nil), data.Routes...)
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Priority == routes[j].Priority {
			return routes[i].ProviderAccountID < routes[j].ProviderAccountID
		}
		return routes[i].Priority < routes[j].Priority
	})
	var route *Route
	for i := range routes {
		if account == "" || routes[i].ProviderAccountID == account {
			route = &routes[i]
			break
		}
	}
	if route == nil {
		return OfferSnapshot{}, ErrNotFound
	}
	snapshot := OfferSnapshot{Offer: *selected, SelectedRoute: *route, SelectionReason: "eligible-published-offer-then-priority", ValidUntil: time.Now().UTC().Add(30 * time.Second), Services: []Resource{}}
	refs := []struct {
		Kind, ID string
		Version  int
		Target   *Resource
	}{{data.TargetKind, data.TargetID, data.TargetVersion, &snapshot.Target}, {"technical-profiles", data.TechnicalProfileID, data.TechnicalProfileVersion, &snapshot.TechnicalProfile}, {"contracts", data.PurchaseContractID, data.PurchaseContractVersion, &snapshot.PurchaseContract}, {"contracts", data.SaleContractID, data.SaleContractVersion, &snapshot.SaleContract}, {"credential-bindings", route.BindingID, route.BindingVersion, &snapshot.Binding}}
	for _, ref := range refs {
		r, e := s.GetResource(ctx, ref.Kind, ref.ID, ref.Version)
		if e != nil || r.State != "PUBLISHED" || (r.TenantID != "" && r.TenantID != tenant) {
			return OfferSnapshot{}, ErrNotFound
		}
		*ref.Target = r
	}
	snapshot.Account, err = s.GetResource(ctx, "provider-accounts", route.ProviderAccountID, route.ProviderAccountVersion)
	if err != nil || snapshot.Account.State != "PUBLISHED" || (snapshot.Account.TenantID != "" && snapshot.Account.TenantID != tenant) {
		return OfferSnapshot{}, ErrNotFound
	}
	binding, _ := DecodeCatalogData(snapshot.Binding)
	if binding.ProviderAccountID != route.ProviderAccountID || binding.CredentialMode == "TENANT_DEDICATED" && snapshot.Binding.TenantID != tenant {
		return OfferSnapshot{}, ErrCredentialUnavailable
	}
	target, _ := DecodeCatalogData(snapshot.Target)
	for _, step := range target.Steps {
		r, e := s.GetResource(ctx, "services", step.ServiceID, step.ServiceVersion)
		if e != nil || r.State != "PUBLISHED" {
			return OfferSnapshot{}, ErrNotFound
		}
		snapshot.Services = append(snapshot.Services, r)
	}
	if data.ValidUntil != nil && data.ValidUntil.Before(snapshot.ValidUntil) {
		snapshot.ValidUntil = *data.ValidUntil
	}
	snapshot.Hash = contentHash(snapshot)
	return snapshot, nil
}
func (h *Handlers) handleOfferResolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeErr(w, 405, "method_not_allowed", "use GET")
		return
	}
	p, ok := auth.FromContext(r.Context())
	if !ok {
		writeErr(w, 401, "unauthenticated", "identidade necessária")
		return
	}
	tenant, application := r.URL.Query().Get("tenant_id"), r.URL.Query().Get("application_id")
	if tenant == "" {
		tenant = p.TenantID
	}
	if application == "" {
		application = p.ApplicationID
	}
	if !auth.Authorize(r.Context(), "catalog:read", tenant) {
		writeErr(w, 403, "forbidden", "oferta fora do escopo")
		return
	}
	if p.Workload {
		var cell string
		if err := h.store.db.QueryRowContext(r.Context(), "SELECT cell_id FROM placements WHERE tenant_id=$1", tenant).Scan(&cell); err != nil || cell != p.CellID {
			auth.Error(w, 403, "tenant_outside_cell")
			return
		}
	} else if application != p.ApplicationID {
		auth.Error(w, 403, "application_outside_scope")
		return
	}
	snapshot, err := h.store.ResolveOffer(r.Context(), tenant, application, r.URL.Query().Get("service_code"), r.URL.Query().Get("provider_account_id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeErr(w, 403, "offer_not_eligible", "nenhuma oferta elegível para aplicação, serviço e conta")
		} else if errors.Is(err, ErrCredentialUnavailable) {
			writeErr(w, 409, "offer_unavailable", "oferta sem vínculo ou contrato elegível")
		} else {
			catalogError(w, err)
		}
		return
	}
	writeJSON(w, 200, snapshot)
}

// TransformJSON implements a bounded declarative projection. No code evaluation,
// filesystem, network, defaults or secret lookups are available to a profile.
func TransformJSON(input json.RawMessage, mapping map[string]string, schema json.RawMessage) (json.RawMessage, error) {
	if len(input) > 256*1024 || len(mapping) > 128 {
		return nil, errors.New("transformação excede limite")
	}
	var source map[string]any
	if json.Unmarshal(input, &source) != nil {
		return nil, errors.New("entrada deve ser objeto JSON")
	}
	out := source
	if len(mapping) > 0 {
		out = map[string]any{}
		for dst, src := range mapping {
			if !safeField(dst) || !safeField(src) {
				return nil, errors.New("campo de mapeamento inválido")
			}
			value, ok := source[src]
			if !ok {
				return nil, errors.New("campo de origem ausente: " + src)
			}
			out[dst] = value
		}
	}
	var contract struct {
		Type       string   `json:"type"`
		Required   []string `json:"required"`
		Properties map[string]struct {
			Type string `json:"type"`
		} `json:"properties"`
	}
	if json.Unmarshal(schema, &contract) != nil || contract.Type != "object" {
		return nil, errors.New("schema inválido")
	}
	for _, field := range contract.Required {
		if _, ok := out[field]; !ok {
			return nil, errors.New("campo obrigatório ausente: " + field)
		}
	}
	for field, rule := range contract.Properties {
		value, ok := out[field]
		if !ok {
			continue
		}
		valid := false
		switch rule.Type {
		case "string":
			_, valid = value.(string)
		case "number":
			_, valid = value.(float64)
		case "integer":
			f, ok := value.(float64)
			valid = ok && f == float64(int64(f))
		case "boolean":
			_, valid = value.(bool)
		case "object":
			_, valid = value.(map[string]any)
		case "array":
			_, valid = value.([]any)
		case "null":
			valid = value == nil
		}
		if !valid {
			return nil, errors.New("tipo inválido: " + field)
		}
	}
	b, err := json.Marshal(out)
	if len(b) > 256*1024 {
		return nil, errors.New("resultado excede limite")
	}
	return b, err
}
