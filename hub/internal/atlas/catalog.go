package atlas

import (
	"ai-hub/hub/internal/contracts/economics"
	"ai-hub/hub/internal/platform/egress"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

var ErrConflict = errors.New("versão publicada ou identidade já existente; crie nova versão")
var ErrRevision = errors.New("revisão alterada por outro operador; recarregue e compare")

type Resource struct {
	Kind      string          `json:"kind"`
	ID        string          `json:"id"`
	Version   int             `json:"version"`
	TenantID  string          `json:"tenant_id"`
	Name      string          `json:"name"`
	State     string          `json:"state"`
	Revision  int64           `json:"revision"`
	Data      json.RawMessage `json:"data"`
	Hash      string          `json:"content_hash"`
	Author    string          `json:"author"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type Step struct {
	ID                         string            `json:"id"`
	ServiceID                  string            `json:"service_id"`
	ServiceVersion             int               `json:"service_version"`
	DependsOn                  []string          `json:"depends_on"`
	Required                   bool              `json:"required"`
	InputMapping               map[string]string `json:"input_mapping"`
	CompensationServiceID      string            `json:"compensation_service_id,omitempty"`
	CompensationServiceVersion int               `json:"compensation_service_version,omitempty"`
}

type CatalogData struct {
	AutoWaitSeconds            int               `json:"auto_wait_seconds"`
	Modes                      []string          `json:"modes"`
	InputSchema                json.RawMessage   `json:"input_schema"`
	OutputSchema               json.RawMessage   `json:"output_schema"`
	DataClass                  string            `json:"data_class"`
	AdapterID                  string            `json:"adapter_id"`
	QualificationID            string            `json:"qualification_id"`
	ClientSLASeconds           int               `json:"client_sla_seconds"`
	ProviderSLASeconds         int               `json:"provider_sla_seconds"`
	RetryTTLSeconds            int               `json:"retry_ttl_seconds"`
	FinalizationReserveSeconds int               `json:"finalization_reserve_seconds"`
	SyncHTTPBudgetSeconds      int               `json:"sync_http_budget_seconds"`
	ProviderMode               string            `json:"provider_mode"`
	ApplicationID              string            `json:"application_id"`
	TargetKind                 string            `json:"target_kind"`
	TargetID                   string            `json:"target_id"`
	TargetVersion              int               `json:"target_version"`
	TechnicalProfileID         string            `json:"technical_profile_id"`
	TechnicalProfileVersion    int               `json:"technical_profile_version"`
	PurchaseContractID         string            `json:"purchase_contract_id"`
	PurchaseContractVersion    int               `json:"purchase_contract_version"`
	SaleContractID             string            `json:"sale_contract_id"`
	SaleContractVersion        int               `json:"sale_contract_version"`
	Routes                     []Route           `json:"routes"`
	Steps                      []Step            `json:"steps"`
	MaxParallel                int               `json:"max_parallel"`
	AllowPartial               bool              `json:"allow_partial"`
	Consolidation              string            `json:"consolidation"`
	FailurePolicy              string            `json:"failure_policy"`
	InputMapping               map[string]string `json:"input_mapping"`
	OutputMapping              map[string]string `json:"output_mapping"`
	MediaType                  string            `json:"media_type"`
	Currency                   string            `json:"currency"`
	UnitPrice                  string            `json:"unit_price"`
	SettlementParty            string            `json:"settlement_party"`
	ValidFrom                  *time.Time        `json:"valid_from"`
	ValidUntil                 *time.Time        `json:"valid_until"`
	CellID                     string            `json:"cell_id"`
	CapacityUnits              int               `json:"capacity_units"`
	CredentialMode             string            `json:"credential_mode"`
	ProviderAccountID          string            `json:"provider_account_id"`
	SecretRef                  string            `json:"secret_ref"`
	SecretVersion              string            `json:"secret_version"`
	Environment                string            `json:"environment"`
}

type Route struct {
	ProviderAccountVersion int    `json:"provider_account_version"`
	ProviderAccountID      string `json:"provider_account_id"`
	BindingID              string `json:"binding_id"`
	BindingVersion         int    `json:"binding_version"`
	Priority               int    `json:"priority"`
	EquivalenceID          string `json:"equivalence_id"`
	CapacityDomain         string `json:"capacity_domain"`
}

type Validation struct {
	Valid                 bool              `json:"valid"`
	FieldErrors           map[string]string `json:"field_errors"`
	Layers                [][]string        `json:"layers"`
	EffectiveRetrySeconds int               `json:"effective_retry_seconds"`
	Hash                  string            `json:"content_hash"`
	Fixture               string            `json:"fixture"`
}

func contentHash(v any) string {
	b, _ := json.Marshal(v)
	var canonical any
	if json.Unmarshal(b, &canonical) == nil {
		b, _ = json.Marshal(canonical)
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
func resourceHash(r Resource) string {
	return contentHash([]any{r.Kind, r.ID, r.Version, r.TenantID, r.Name, r.Data})
}
func validKind(k string) bool {
	for _, v := range []string{"clients", "applications", "services", "products", "offers", "technical-profiles", "providers", "provider-accounts", "credential-bindings", "contracts", "policies"} {
		if k == v {
			return true
		}
	}
	return false
}
func DecodeCatalogData(r Resource) (CatalogData, error) {
	var d CatalogData
	err := json.Unmarshal(r.Data, &d)
	return d, err
}

// ValidateResource performs bounded structural validation without external effects.
func ValidateResource(r Resource) Validation {
	v := Validation{Valid: true, FieldErrors: map[string]string{}, Layers: [][]string{}, Hash: resourceHash(r), Fixture: "synthetic-validation-no-provider-calls"}
	bad := func(k, msg string) { v.Valid = false; v.FieldErrors[k] = msg }
	if !validKind(r.Kind) {
		bad("kind", "recurso não suportado")
	}
	if strings.TrimSpace(r.ID) == "" || len(r.ID) > 128 || strings.ContainsAny(r.ID, "/\\") {
		bad("id", "identidade obrigatória com até 128 caracteres e sem barras")
	}
	if r.Version < 1 {
		bad("version", "versão deve ser positiva")
	}
	if strings.TrimSpace(r.Name) == "" {
		bad("name", "nome obrigatório")
	}
	if len(r.Data) > 256*1024 {
		bad("data", "configuração excede 256 KiB")
		return v
	}
	d, err := DecodeCatalogData(r)
	if err != nil {
		bad("data", "objeto JSON inválido")
		return v
	}
	if r.Kind == "applications" || r.Kind == "offers" || r.Kind == "technical-profiles" || r.Kind == "clients" {
		if r.TenantID == "" {
			bad("tenant_id", "cliente obrigatório")
		}
	}
	if d.ValidFrom != nil && d.ValidUntil != nil && !d.ValidFrom.Before(*d.ValidUntil) {
		bad("valid_until", "fim deve ser posterior ao início")
	}
	if d.RetryTTLSeconds < 0 {
		bad("retry_ttl_seconds", "TTL deve ser inteiro não negativo")
	}
	v.EffectiveRetrySeconds = d.RetryTTLSeconds
	if d.ClientSLASeconds > 0 && v.EffectiveRetrySeconds > d.ClientSLASeconds-d.FinalizationReserveSeconds {
		v.EffectiveRetrySeconds = d.ClientSLASeconds - d.FinalizationReserveSeconds
	}
	if r.Kind == "services" || r.Kind == "technical-profiles" || r.Kind == "offers" || r.Kind == "policies" {
		if d.ClientSLASeconds <= 0 {
			bad("client_sla_seconds", "SLA deve ser positivo")
		}
		if d.FinalizationReserveSeconds < 0 || d.FinalizationReserveSeconds >= d.ClientSLASeconds {
			bad("finalization_reserve_seconds", "reserva deve ser menor que SLA")
		}
		if len(d.Modes) == 0 {
			bad("modes", "selecione modalidades")
		}
		for _, m := range d.Modes {
			if m != "SYNC" && m != "ASYNC" && m != "AUTO" {
				bad("modes", "modalidade inválida")
			}
			if m == "AUTO" && (d.AutoWaitSeconds <= 0 || d.AutoWaitSeconds > d.ClientSLASeconds) {
				bad("auto_wait_seconds", "AUTO exige espera positiva limitada pelo SLA")
			}
			if m == "SYNC" && (d.ProviderMode != "sync" || d.SyncHTTPBudgetSeconds <= 0 || d.ProviderSLASeconds+d.FinalizationReserveSeconds > d.SyncHTTPBudgetSeconds) {
				bad("modes", "SYNC exige provedor síncrono e orçamento qualificado compatível")
			}
		}
	}
	if r.Kind == "services" || r.Kind == "technical-profiles" {
		if !validSchema(d.InputSchema) {
			bad("input_schema", "schema de objeto com propriedades tipadas obrigatório")
		}
		if !validSchema(d.OutputSchema) {
			bad("output_schema", "schema de objeto com propriedades tipadas obrigatório")
		}
		if r.Kind == "services" && (d.AdapterID == "" || d.QualificationID == "" || d.DataClass == "") {
			bad("qualification_id", "adapter, qualificação e classe de dados obrigatórios para publicação")
		}
		if r.Kind == "technical-profiles" {
			if d.MediaType != "application/json" {
				bad("media_type", "perfil implementado: application/json")
			}
			for _, m := range []map[string]string{d.InputMapping, d.OutputMapping} {
				if len(m) > 128 {
					bad("mapping", "máximo de 128 campos")
				}
				for dst, src := range m {
					if !safeField(dst) || !safeField(src) {
						bad("mapping", "mapeamento aceita apenas campos simples, sem código ou rede")
					}
				}
			}
		}
	}
	if r.Kind == "clients" {
		if d.CellID == "" || d.CapacityUnits < 1 {
			bad("cell_id", "célula e capacidade positiva obrigatórias")
		}
	}
	if r.Kind == "contracts" {
		if _, _, err := economics.DecodePublished(r.ID, r.Version, r.Data); err != nil {
			bad("economic_policy", err.Error())
		}
		if !regexp.MustCompile(`^[0-9]+(\.[0-9]{1,4})?$`).MatchString(d.UnitPrice) || len(d.UnitPrice) > 20 {
			bad("unit_price", "preço decimal exato não negativo com até quatro casas")
		}
		if !contains([]string{"BRL", "USD", "EUR"}, d.Currency) {
			bad("currency", "moeda não suportada")
		}
		if !contains([]string{"HUB", "CLIENT_DIRECT"}, d.SettlementParty) {
			bad("settlement_party", "pagador explícito obrigatório")
		}
	}
	if r.Kind == "provider-accounts" {
		var account ProviderAccount
		_ = json.Unmarshal(r.Data, &account)
		account.ProviderAccountID = r.ID
		if err := validateProviderAccount(&account); err != nil {
			bad("provider_account", err.Error())
		}
		if err := egress.ValidateURL(account.BaseURL); err != nil {
			bad("base_url", "destino externo não autorizado")
		}
		if account.OAuthTokenURL != "" {
			if err := egress.ValidateURL(account.OAuthTokenURL); err != nil {
				bad("oauth_token_url", "destino de autenticação não autorizado")
			}
		}
	}
	if r.Kind == "products" {
		if len(d.Steps) < 1 || len(d.Steps) > 20 {
			bad("steps", "produto exige entre 1 e 20 passos")
		}
		if d.MaxParallel < 1 || d.MaxParallel > 5 {
			bad("max_parallel", "paralelismo entre 1 e 5")
		}
		if d.Consolidation != "ALL_REQUIRED" {
			bad("consolidation", "política implementada: ALL_REQUIRED")
		}
		if d.FailurePolicy != "STOP" && d.FailurePolicy != "COMPENSATE" {
			bad("failure_policy", "selecione STOP ou COMPENSATE")
		}
		layers, err := PlanDAG(d.Steps)
		if err != nil {
			bad("steps", err.Error())
		} else {
			v.Layers = layers
		}
	}
	if r.Kind == "offers" {
		if d.ApplicationID == "" || d.TargetID == "" || (d.TargetKind != "services" && d.TargetKind != "products") || d.TargetVersion < 1 {
			bad("target_id", "aplicação e serviço/produto versionado obrigatórios")
		}
		if d.TechnicalProfileID == "" || d.TechnicalProfileVersion < 1 || d.PurchaseContractID == "" || d.SaleContractID == "" || d.PurchaseContractVersion < 1 || d.SaleContractVersion < 1 {
			bad("contracts", "perfis e contratos de compra/venda versionados obrigatórios")
		}
		if len(d.Routes) < 1 || len(d.Routes) > 5 {
			bad("routes", "entre 1 e 5 rotas elegíveis")
		}
		for _, rt := range d.Routes {
			if rt.ProviderAccountID == "" || rt.BindingID == "" || rt.BindingVersion < 1 || rt.CapacityDomain == "" {
				bad("routes", "conta, vínculo, versão e capacidade obrigatórios")
			}
			if len(d.Routes) > 1 && rt.EquivalenceID == "" {
				bad("routes", "rotas alternativas exigem equivalência homologada")
			}
		}
	}
	if r.Kind == "credential-bindings" {
		if d.SecretRef == "" || d.SecretVersion == "" || d.ProviderAccountID == "" {
			bad("secret_ref", "referências de conta e versão do segredo obrigatórias")
		}
		if d.CredentialMode != "SHARED_HUB" && d.CredentialMode != "TENANT_DEDICATED" {
			bad("credential_mode", "modo explícito obrigatório")
		}
		if d.CredentialMode == "TENANT_DEDICATED" && r.TenantID == "" {
			bad("tenant_id", "vínculo dedicado exige cliente")
		}
		if d.SettlementParty != "HUB" && d.SettlementParty != "CLIENT_DIRECT" {
			bad("settlement_party", "pagador explícito obrigatório")
		}
	}
	return v
}
func safeField(s string) bool {
	if s == "" || len(s) > 128 {
		return false
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_') {
			return false
		}
	}
	return true
}
func validSchema(raw json.RawMessage) bool {
	var s schemaRule
	if json.Unmarshal(raw, &s) != nil || s.Type != "object" {
		return false
	}
	if s.Properties == nil {
		return false
	}
	return validateSchemaDefinition(raw, "$", map[string]bool{}) == nil
}
func PlanDAG(steps []Step) ([][]string, error) {
	if len(steps) > 20 {
		return nil, errors.New("máximo de 20 passos")
	}
	byID := map[string]Step{}
	for _, s := range steps {
		if !safeField(s.ID) || s.ServiceID == "" || s.ServiceVersion < 1 {
			return nil, fmt.Errorf("passo %s sem identidade/serviço versionado", s.ID)
		}
		if s.CompensationServiceID != "" && s.CompensationServiceVersion < 1 {
			return nil, fmt.Errorf("passo %s sem serviço de compensação versionado", s.ID)
		}
		if _, ok := byID[s.ID]; ok {
			return nil, fmt.Errorf("passo duplicado: %s", s.ID)
		}
		byID[s.ID] = s
	}
	for _, s := range steps {
		for _, dep := range s.DependsOn {
			if _, ok := byID[dep]; !ok {
				return nil, fmt.Errorf("passo %s depende de referência inexistente %s", s.ID, dep)
			}
		}
		for _, source := range s.InputMapping {
			parts := strings.Split(source, ".")
			if len(parts) != 2 || !safeField(parts[1]) {
				return nil, fmt.Errorf("mapeamento inválido no passo %s", s.ID)
			}
			found := false
			for _, dep := range s.DependsOn {
				if dep == parts[0] {
					found = true
				}
			}
			if !found {
				return nil, fmt.Errorf("passo %s usa saída sem dependência %s", s.ID, source)
			}
		}
	}
	done := map[string]bool{}
	layers := [][]string{}
	for len(done) < len(steps) {
		layer := []string{}
		for _, s := range steps {
			if done[s.ID] {
				continue
			}
			ready := true
			for _, dep := range s.DependsOn {
				if !done[dep] {
					ready = false
				}
			}
			if ready {
				layer = append(layer, s.ID)
			}
		}
		if len(layer) == 0 {
			ids := []string{}
			for _, s := range steps {
				if !done[s.ID] {
					ids = append(ids, s.ID)
				}
			}
			return nil, fmt.Errorf("ciclo entre passos: %s", strings.Join(ids, ", "))
		}
		sort.Strings(layer)
		layers = append(layers, layer)
		for _, id := range layer {
			done[id] = true
		}
	}
	return layers, nil
}

const resourceColumns = `kind,id,version,tenant_id,name,state,revision,data,content_hash,author,updated_at`

func scanResource(row interface{ Scan(...any) error }) (Resource, error) {
	var r Resource
	err := row.Scan(&r.Kind, &r.ID, &r.Version, &r.TenantID, &r.Name, &r.State, &r.Revision, &r.Data, &r.Hash, &r.Author, &r.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		err = ErrNotFound
	}
	return r, err
}
func (s *Store) GetResource(ctx context.Context, kind, id string, version int) (Resource, error) {
	return scanResource(s.db.QueryRowContext(ctx, `SELECT `+resourceColumns+` FROM catalog_resources WHERE kind=$1 AND id=$2 AND version=$3`, kind, id, version))
}
func (s *Store) ListResources(ctx context.Context, kind, tenant, query, state, afterID string, afterVersion, limit int) ([]Resource, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+resourceColumns+` FROM catalog_resources WHERE kind=$1 AND ($2='*' OR tenant_id=$2 OR tenant_id='') AND ($3='' OR name ILIKE '%'||$3||'%' OR id ILIKE '%'||$3||'%') AND ($4='' OR state=$4) AND (id,version)>($5,$6) ORDER BY id,version LIMIT $7`, kind, tenant, query, state, afterID, afterVersion, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Resource{}
	for rows.Next() {
		r, e := scanResource(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
func (s *Store) SaveResource(ctx context.Context, r Resource, expected int64, actor string) (Resource, error) {
	r.Hash = resourceHash(r)
	if expected == 0 {
		row := s.db.QueryRowContext(ctx, `INSERT INTO catalog_resources(kind,id,version,tenant_id,name,data,author,content_hash) VALUES($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT DO NOTHING RETURNING `+resourceColumns, r.Kind, r.ID, r.Version, r.TenantID, r.Name, []byte(r.Data), actor, r.Hash)
		out, err := scanResource(row)
		if errors.Is(err, ErrNotFound) {
			return Resource{}, ErrConflict
		}
		return out, err
	}
	out, err := scanResource(s.db.QueryRowContext(ctx, `UPDATE catalog_resources SET name=$4,data=$5,revision=revision+1,author=$6,updated_at=clock_timestamp(),content_hash=$8 WHERE kind=$1 AND id=$2 AND version=$3 AND revision=$7 AND state='DRAFT' RETURNING `+resourceColumns, r.Kind, r.ID, r.Version, r.Name, []byte(r.Data), actor, expected, r.Hash))
	if errors.Is(err, ErrNotFound) {
		cur, e := s.GetResource(ctx, r.Kind, r.ID, r.Version)
		if e != nil {
			return Resource{}, e
		}
		if cur.State != "DRAFT" {
			return Resource{}, ErrConflict
		}
		return Resource{}, ErrRevision
	}
	return out, err
}

// ValidatePublication checks all same-authority references against immutable versions.
func (s *Store) ValidatePublication(ctx context.Context, r Resource) Validation {
	v := ValidateResource(r)
	d, _ := DecodeCatalogData(r)
	if r.Kind == "services" && d.QualificationID != "" {
		var qualified bool
		err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM catalog_qualifications WHERE id=$1 AND adapter_id=$2 AND state='QUALIFIED' AND valid_until>clock_timestamp())`, d.QualificationID, d.AdapterID).Scan(&qualified)
		if err != nil || !qualified {
			v.Valid = false
			v.FieldErrors["qualification_id"] = "evidência de homologação vigente não encontrada"
		}
	}
	if r.Kind == "clients" {
		var ready bool
		err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM catalog_capacity_cells WHERE cell_id=$1 AND state='READY' AND total_units-reserved_units >= $2)`, d.CellID, d.CapacityUnits).Scan(&ready)
		if err != nil || !ready {
			v.Valid = false
			v.FieldErrors["capacity_units"] = "capacidade qualificada insuficiente; solicite provisionamento"
		}
	}

	check := func(kind, id string, version int) {
		ref, err := s.GetResource(ctx, kind, id, version)
		if err != nil || ref.State != "PUBLISHED" || (ref.TenantID != "" && ref.TenantID != r.TenantID) {
			v.Valid = false
			v.FieldErrors[kind+"/"+id] = "referência publicada e autorizada não encontrada"
		}
	}
	if r.Kind == "products" {
		for _, step := range d.Steps {
			check("services", step.ServiceID, step.ServiceVersion)
		}
	}
	if r.Kind == "offers" {
		check(d.TargetKind, d.TargetID, d.TargetVersion)
		check("technical-profiles", d.TechnicalProfileID, d.TechnicalProfileVersion)
		check("contracts", d.PurchaseContractID, d.PurchaseContractVersion)
		check("contracts", d.SaleContractID, d.SaleContractVersion)
		for _, rt := range d.Routes {
			check("credential-bindings", rt.BindingID, rt.BindingVersion)
			check("provider-accounts", rt.ProviderAccountID, rt.ProviderAccountVersion)
			account, err := s.GetResource(ctx, "provider-accounts", rt.ProviderAccountID, rt.ProviderAccountVersion)
			ad, _ := DecodeCatalogData(account)
			if err != nil || (contains(d.Modes, "SYNC") && ad.ProviderMode != "sync") {
				v.Valid = false
				v.FieldErrors["routes/"+rt.ProviderAccountID] = "conta indisponível ou modalidade incompatível"
			}
		}
	}
	return v
}
func contains(values []string, s string) bool {
	for _, v := range values {
		if v == s {
			return true
		}
	}
	return false
}
func (s *Store) PublishResource(ctx context.Context, r Resource, expected int64, actor, reason string, v Validation) (Resource, error) {
	if !v.Valid || v.Hash != resourceHash(r) {
		return Resource{}, errors.New("validação inválida")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Resource{}, err
	}
	defer tx.Rollback()
	out, err := scanResource(tx.QueryRowContext(ctx, `UPDATE catalog_resources SET state='PUBLISHED',revision=revision+1,author=$5,updated_at=clock_timestamp() WHERE kind=$1 AND id=$2 AND version=$3 AND revision=$4 AND state='DRAFT' AND content_hash=$6 RETURNING `+resourceColumns, r.Kind, r.ID, r.Version, expected, actor, v.Hash))
	if errors.Is(err, ErrNotFound) {
		return Resource{}, ErrRevision
	}
	if err != nil {
		return Resource{}, err
	}
	if err = projectPublication(ctx, tx, out, actor); err != nil {
		return Resource{}, err
	}
	validation, _ := json.Marshal(v)
	_, err = tx.ExecContext(ctx, `INSERT INTO catalog_publications(kind,resource_id,resource_version,revision,content_hash,actor,reason,validation) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, r.Kind, r.ID, r.Version, out.Revision, out.Hash, actor, reason, validation)
	if err != nil {
		return Resource{}, err
	}
	if err = tx.Commit(); err != nil {
		return Resource{}, err
	}
	return out, nil
}
func (s *Store) SuspendResource(ctx context.Context, r Resource, expected int64, actor, reason string) (Resource, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Resource{}, err
	}
	defer tx.Rollback()
	out, err := scanResource(tx.QueryRowContext(ctx, `UPDATE catalog_resources SET state='SUSPENDED',revision=revision+1,updated_at=clock_timestamp() WHERE kind=$1 AND id=$2 AND version=$3 AND revision=$4 AND state='PUBLISHED' RETURNING `+resourceColumns, r.Kind, r.ID, r.Version, expected))
	if errors.Is(err, ErrNotFound) {
		return Resource{}, ErrRevision
	}
	if err != nil {
		return Resource{}, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO catalog_publications(kind,resource_id,resource_version,revision,content_hash,actor,reason,validation) VALUES($1,$2,$3,$4,$5,$6,$7,'{"action":"suspend"}')`, r.Kind, r.ID, r.Version, out.Revision, out.Hash, actor, reason)
	if err != nil {
		return Resource{}, err
	}
	return out, tx.Commit()
}
