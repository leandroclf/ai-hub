package atlas

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"reflect"
	"sort"
	"strconv"
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

func (s *Store) ResolveOffer(ctx context.Context, tenant, application, service string, serviceVersion int, account string) (OfferSnapshot, error) {
	if serviceVersion < 1 {
		return OfferSnapshot{}, ErrNotFound
	}
	// Resolve no banco pelo conjunto elegível. O LIMIT 2 é intencional:
	// basta distinguir zero, uma oferta ou ambiguidade, sem materializar o
	// portfólio inteiro no caminho crítico.
	resources, err := s.listEligibleOffers(ctx, tenant, application, service, serviceVersion, account)
	if err != nil {
		return OfferSnapshot{}, err
	}
	var selected *Resource
	var data CatalogData
	for _, r := range resources {
		d, e := DecodeCatalogData(r)
		if e != nil {
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

func (s *Store) listEligibleOffers(ctx context.Context, tenant, application, service string, serviceVersion int, account string) ([]Resource, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+resourceColumns+` FROM catalog_resources
		WHERE kind='offers' AND state='PUBLISHED' AND (tenant_id=$1 OR tenant_id='')
		  AND data->>'application_id'=$2 AND data->>'target_id'=$3 AND data->>'target_version'=$4
		  AND (NULLIF(data->>'valid_from','') IS NULL OR (data->>'valid_from')::timestamptz <= clock_timestamp())
		  AND (NULLIF(data->>'valid_until','') IS NULL OR (data->>'valid_until')::timestamptz > clock_timestamp())
		  AND ($5='' OR EXISTS (SELECT 1 FROM jsonb_array_elements(COALESCE(data->'routes','[]'::jsonb)) route WHERE route->>'provider_account_id'=$5))
		ORDER BY id,version LIMIT 2`, tenant, application, service, strconv.Itoa(serviceVersion), account)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	resources := make([]Resource, 0, 2)
	for rows.Next() {
		r, err := scanResource(rows)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}
	return resources, rows.Err()
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
	serviceVersion, err := strconv.Atoi(r.URL.Query().Get("service_version"))
	if err != nil || serviceVersion < 1 {
		writeErr(w, http.StatusBadRequest, "invalid_service_version", "service_version deve ser inteiro positivo")
		return
	}
	snapshot, err := h.store.ResolveOffer(r.Context(), tenant, application, r.URL.Query().Get("service_code"), serviceVersion, r.URL.Query().Get("provider_account_id"))
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
	decoded, err := decodeSingleJSON(input)
	if err != nil {
		return nil, errors.New("entrada deve ser objeto JSON")
	}
	if !validJSONType(decoded, "object") {
		return nil, errors.New("entrada deve ser objeto JSON")
	}
	var source map[string]json.RawMessage
	if err := json.Unmarshal(decoded, &source); err != nil {
		return nil, errors.New("entrada deve ser objeto JSON")
	}
	out := source
	if len(mapping) > 0 {
		out = map[string]json.RawMessage{}
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
	if err := validateSchemaDefinition(schema, "$", map[string]bool{}); err != nil {
		return nil, errors.New("schema inválido")
	}
	var contract schemaRule
	if json.Unmarshal(schema, &contract) != nil || contract.Type != "object" {
		return nil, errors.New("schema inválido")
	}
	// Um objeto sem `properties` é um schema válido para a fronteira de
	// execução quando não declara `additionalProperties:false`: ele
	// representa um objeto aberto e conserva os campos devolvidos pelo
	// adapter. A publicação de perfis continua exigindo propriedades
	// tipadas em validSchema; aqui validamos o snapshot já publicado sem
	// inventar uma restrição diferente no caminho quente.
	if contract.Properties == nil {
		contract.Properties = map[string]json.RawMessage{}
	}
	for _, field := range contract.Required {
		if _, ok := out[field]; !ok {
			return nil, errors.New("campo obrigatório ausente: " + field)
		}
	}
	for field, ruleRaw := range contract.Properties {
		value, ok := out[field]
		if !ok {
			continue
		}
		var rule schemaRule
		if json.Unmarshal(ruleRaw, &rule) != nil || rule.Type == "" {
			return nil, errors.New("regra de schema inválida: " + field)
		}
		valid := validJSONType(value, rule.Type)
		if !valid {
			return nil, errors.New("tipo inválido: " + field)
		}
		if len(rule.Enum) > 0 {
			matched := false
			for _, candidate := range rule.Enum {
				if jsonEqual(value, candidate) {
					matched = true
					break
				}
			}
			if !matched {
				return nil, errors.New("valor fora do enum: " + field)
			}
		}
	}
	if encoded, err := json.Marshal(out); err != nil {
		return nil, errors.New("resultado não serializável")
	} else if err := validateJSONSchema(encoded, schema, "$"); err != nil {
		return nil, err
	}
	if len(mapping) == 0 {
		return append(json.RawMessage(nil), input...), nil
	}
	b, err := json.Marshal(out)
	if len(b) > 256*1024 {
		return nil, errors.New("resultado excede limite")
	}
	return b, err
}

// validateJSONSchema cobre o subconjunto declarativo publicado pelo Hub e
// recusa construções não qualificadas. A validação conserva RawMessage até o
// fim; nenhum número passa por float64 e nenhum campo extra é aceito quando o
// contrato o proíbe.
func validateJSONSchema(value, schema json.RawMessage, path string) error {
	var rule schemaRule
	if err := validateSchemaDefinition(schema, path, map[string]bool{}); err != nil || json.Unmarshal(schema, &rule) != nil || rule.Type == "" {
		return errors.New("schema inválido em " + path)
	}
	if len(rule.Enum) > 0 {
		matched := false
		for _, candidate := range rule.Enum {
			if jsonEqual(value, candidate) {
				matched = true
				break
			}
		}
		if !matched {
			return errors.New("valor fora do enum: " + path)
		}
	}
	if len(rule.Minimum) > 0 && compareJSONNumber(value, rule.Minimum) < 0 {
		return errors.New("valor menor que minimum: " + path)
	}
	if len(rule.Maximum) > 0 && compareJSONNumber(value, rule.Maximum) > 0 {
		return errors.New("valor maior que maximum: " + path)
	}
	if !validJSONType(value, rule.Type) {
		return errors.New("tipo inválido: " + path)
	}
	if rule.Type == "object" {
		var fields map[string]json.RawMessage
		if json.Unmarshal(value, &fields) != nil {
			return errors.New("objeto inválido: " + path)
		}
		for _, required := range rule.Required {
			if _, ok := fields[required]; !ok {
				return errors.New("campo obrigatório ausente: " + path + "." + required)
			}
		}
		for name, field := range fields {
			child, ok := rule.Properties[name]
			if !ok {
				if rule.AdditionalProperties != nil && !*rule.AdditionalProperties {
					return errors.New("campo adicional não permitido: " + path + "." + name)
				}
				continue
			}
			if err := validateJSONSchema(field, child, path+"."+name); err != nil {
				return err
			}
		}
	}
	if rule.Type == "array" && len(rule.Items) > 0 {
		var items []json.RawMessage
		if json.Unmarshal(value, &items) != nil {
			return errors.New("array inválido: " + path)
		}
		for i, item := range items {
			if err := validateJSONSchema(item, rule.Items, fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
	}
	return nil
}

func validJSONType(raw json.RawMessage, want string) bool {
	trimmed := bytes.TrimSpace(raw)
	switch want {
	case "string":
		var v string
		return !bytes.Equal(trimmed, []byte("null")) && json.Unmarshal(trimmed, &v) == nil
	case "number", "integer":
		var n json.Number
		if bytes.Equal(trimmed, []byte("null")) || len(trimmed) == 0 || (trimmed[0] != '-' && (trimmed[0] < '0' || trimmed[0] > '9')) || json.Unmarshal(trimmed, &n) != nil || n.String() == "" {
			return false
		}
		return want != "integer" || isExactInteger(n.String())
	case "boolean":
		return bytes.Equal(trimmed, []byte("true")) || bytes.Equal(trimmed, []byte("false"))
	case "object":
		return len(trimmed) > 1 && trimmed[0] == '{'
	case "array":
		return len(trimmed) > 1 && trimmed[0] == '['
	case "null":
		return bytes.Equal(trimmed, []byte("null"))
	default:
		return false
	}
}

// schemaRule is the deliberately small, published JSON Schema dialect used by
// profiles. Unknown keywords are rejected at publication and runtime instead
// of being silently ignored.
type schemaRule struct {
	Type                 string                     `json:"type"`
	Required             []string                   `json:"required"`
	Properties           map[string]json.RawMessage `json:"properties"`
	AdditionalProperties *bool                      `json:"additionalProperties"`
	Items                json.RawMessage            `json:"items"`
	Enum                 []json.RawMessage          `json:"enum"`
	Minimum              json.RawMessage            `json:"minimum"`
	Maximum              json.RawMessage            `json:"maximum"`
}

var supportedSchemaKeywords = map[string]bool{
	"type": true, "required": true, "properties": true,
	"additionalProperties": true, "items": true, "enum": true,
	"minimum": true, "maximum": true,
}

func validateSchemaDefinition(raw json.RawMessage, path string, stack map[string]bool) error {
	var obj map[string]json.RawMessage
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&obj); err != nil || obj == nil {
		return errors.New("schema deve ser objeto")
	}
	for key := range obj {
		if !supportedSchemaKeywords[key] {
			return fmt.Errorf("keyword não suportada em %s: %s", path, key)
		}
	}
	var rule schemaRule
	if err := json.Unmarshal(raw, &rule); err != nil || rule.Type == "" {
		return errors.New("tipo de schema obrigatório")
	}
	validTypes := map[string]bool{"object": true, "array": true, "string": true, "number": true, "integer": true, "boolean": true, "null": true}
	if !validTypes[rule.Type] {
		return errors.New("tipo de schema não suportado")
	}
	for _, field := range rule.Required {
		if !safeField(field) {
			return errors.New("required inválido")
		}
	}
	for key, child := range rule.Properties {
		if !safeField(key) {
			return errors.New("propriedade inválida")
		}
		if err := validateSchemaDefinition(child, path+"."+key, stack); err != nil {
			return err
		}
	}
	if len(rule.Items) > 0 {
		if err := validateSchemaDefinition(rule.Items, path+"[]", stack); err != nil {
			return err
		}
	}
	for name, bound := range map[string]json.RawMessage{"minimum": rule.Minimum, "maximum": rule.Maximum} {
		if len(bound) > 0 && !validJSONType(bound, "number") {
			return fmt.Errorf("%s deve ser número", name)
		}
	}
	return nil
}

func decodeSingleJSON(raw []byte) (json.RawMessage, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var value json.RawMessage
	if err := dec.Decode(&value); err != nil {
		return nil, err
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return nil, errors.New("mais de um documento JSON")
	}
	return value, nil
}

func isExactInteger(raw string) bool {
	r, ok := new(big.Rat).SetString(raw)
	return ok && r.IsInt()
}

func compareJSONNumber(a, b json.RawMessage) int {
	ra, oka := new(big.Rat).SetString(string(bytes.TrimSpace(a)))
	rb, okb := new(big.Rat).SetString(string(bytes.TrimSpace(b)))
	if !oka || !okb {
		return 0
	}
	return ra.Cmp(rb)
}

func jsonEqual(a, b json.RawMessage) bool {
	var left, right any
	la, lb := json.NewDecoder(bytes.NewReader(a)), json.NewDecoder(bytes.NewReader(b))
	la.UseNumber()
	lb.UseNumber()
	return la.Decode(&left) == nil && lb.Decode(&right) == nil && reflect.DeepEqual(left, right)
}
