// Package atlas implementa o plano de controle minimo do hub: catalogo
// de servicos, contas de provedor, vinculos de credencial e contratos
// de cliente (CFG-01, CFG-05, FIN-02), publicando projecoes que
// Orbita/Cometa consultam no bootstrap e periodicamente — nao a cada
// pedido (ARQ-02).
package atlas

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
)

// ErrNotFound e retornado quando um recurso do catalogo nao existe.
var ErrNotFound = errors.New("atlas: recurso nao encontrado")

// ErrCredentialUnavailable e retornado quando a resolucao de
// credencial (SEG-05) nao encontra um vinculo elegivel e ativo — nunca
// cai para outro modo/tenant como fallback implicito.
var ErrCredentialUnavailable = errors.New("atlas: credencial indisponivel para o modo exigido, sem fallback")

// Store encapsula o acesso a hub_control.
type Store struct {
	db *sql.DB
}

// NewStore cria um Store a partir de uma conexao ja aberta.
func NewStore(db *sql.DB) *Store { return &Store{db: db} }

// Service e a versao publicada de um servico do catalogo (CAT-01).
type Service struct {
	Code             string   `json:"code"`
	Version          int      `json:"version"`
	Description      string   `json:"description"`
	Modes            []string `json:"modes"`
	ClientSLASeconds int      `json:"client_sla_seconds"`
	RetryTTLSeconds  int      `json:"retry_ttl_seconds"`
	Published        bool     `json:"published"`
}

// UpsertService publica (ou atualiza um rascunho de) uma versao de
// servico. Versoes publicadas sao imutaveis (CAT-01) — esta referencia
// nao reforca essa imutabilidade fisica no schema, apenas na regra de
// negocio do handler HTTP (nao permite POST sobre published=true).
func (s *Store) UpsertService(ctx context.Context, svc Service) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO services (code, version, description, schema_input, schema_output, modes, client_sla_seconds, retry_ttl_seconds, published)
		VALUES ($1, $2, $3, '{}'::jsonb, '{}'::jsonb, $4, $5, $6, $7)
		ON CONFLICT (code, version) DO UPDATE SET
			description = EXCLUDED.description,
			modes = EXCLUDED.modes,
			client_sla_seconds = EXCLUDED.client_sla_seconds,
			retry_ttl_seconds = EXCLUDED.retry_ttl_seconds,
			published = EXCLUDED.published
	`, svc.Code, svc.Version, svc.Description, pq.Array(svc.Modes), svc.ClientSLASeconds, svc.RetryTTLSeconds, svc.Published)
	if err != nil {
		return fmt.Errorf("atlas: publicar servico: %w", err)
	}
	return nil
}

// GetService le uma versao de servico pelo codigo/versao.
func (s *Store) GetService(ctx context.Context, code string, version int) (Service, error) {
	var svc Service
	row := s.db.QueryRowContext(ctx, `
		SELECT code, version, description, modes, client_sla_seconds, retry_ttl_seconds, published
		FROM services WHERE code = $1 AND version = $2
	`, code, version)
	err := row.Scan(&svc.Code, &svc.Version, &svc.Description, pq.Array(&svc.Modes), &svc.ClientSLASeconds, &svc.RetryTTLSeconds, &svc.Published)
	if errors.Is(err, sql.ErrNoRows) {
		return Service{}, ErrNotFound
	}
	if err != nil {
		return Service{}, fmt.Errorf("atlas: ler servico: %w", err)
	}
	return svc, nil
}

// ProviderAccount e uma conta de provedor homologada (CAT-06, DAD-02).
type ProviderAccount struct {
	ProviderAccountID string `json:"provider_account_id"`
	ProviderID        string `json:"provider_id"`
	Environment       string `json:"environment"`
	BaseURL           string `json:"base_url"`
	ProviderMode      string `json:"provider_mode"`
}

// UpsertProviderAccount cadastra/atualiza uma conta de provedor.
func (s *Store) UpsertProviderAccount(ctx context.Context, pa ProviderAccount) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO provider_accounts (provider_account_id, provider_id, environment, base_url, provider_mode)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (provider_account_id) DO UPDATE SET
			provider_id = EXCLUDED.provider_id,
			environment = EXCLUDED.environment,
			base_url = EXCLUDED.base_url,
			provider_mode = EXCLUDED.provider_mode
	`, pa.ProviderAccountID, pa.ProviderID, pa.Environment, pa.BaseURL, pa.ProviderMode)
	if err != nil {
		return fmt.Errorf("atlas: publicar conta de provedor: %w", err)
	}
	return nil
}

// GetProviderAccount le uma conta de provedor.
func (s *Store) GetProviderAccount(ctx context.Context, id string) (ProviderAccount, error) {
	var pa ProviderAccount
	row := s.db.QueryRowContext(ctx, `
		SELECT provider_account_id, provider_id, environment, base_url, provider_mode
		FROM provider_accounts WHERE provider_account_id = $1
	`, id)
	err := row.Scan(&pa.ProviderAccountID, &pa.ProviderID, &pa.Environment, &pa.BaseURL, &pa.ProviderMode)
	if errors.Is(err, sql.ErrNoRows) {
		return ProviderAccount{}, ErrNotFound
	}
	if err != nil {
		return ProviderAccount{}, fmt.Errorf("atlas: ler conta de provedor: %w", err)
	}
	return pa, nil
}

// CredentialBinding e o vinculo de credencial (CFG-05, DAD-02) — nunca
// contem o segredo em si, apenas a referencia ao cofre.
type CredentialBinding struct {
	BindingID         string `json:"binding_id"`
	CredentialMode    string `json:"credential_mode"` // SHARED_HUB | TENANT_DEDICATED
	TenantID          string `json:"tenant_id"`        // vazio quando SHARED_HUB
	ProviderAccountID string `json:"provider_account_id"`
	SecretRef         string `json:"secret_ref"`
	SettlementParty   string `json:"settlement_party"`
	State             string `json:"state"`
}

// UpsertCredentialBinding cadastra/atualiza um vinculo de credencial.
func (s *Store) UpsertCredentialBinding(ctx context.Context, cb CredentialBinding) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO credential_bindings (binding_id, credential_mode, tenant_id, provider_account_id, secret_ref, settlement_party, state)
		VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6, $7)
		ON CONFLICT (binding_id) DO UPDATE SET
			credential_mode = EXCLUDED.credential_mode,
			tenant_id = EXCLUDED.tenant_id,
			provider_account_id = EXCLUDED.provider_account_id,
			secret_ref = EXCLUDED.secret_ref,
			settlement_party = EXCLUDED.settlement_party,
			state = EXCLUDED.state
	`, cb.BindingID, cb.CredentialMode, cb.TenantID, cb.ProviderAccountID, cb.SecretRef, cb.SettlementParty, cb.State)
	if err != nil {
		return fmt.Errorf("atlas: publicar vinculo de credencial: %w", err)
	}
	return nil
}

// ResolveCredential implementa a resolucao deterministica de SEG-05:
// le a modalidade exigida pelo contrato do tenant e busca exatamente
// o vinculo correspondente, sem fallback implicito entre modos.
func (s *Store) ResolveCredential(ctx context.Context, tenantID, providerAccountID string) (CredentialBinding, error) {
	contract, err := s.GetContract(ctx, tenantID)
	if err != nil {
		return CredentialBinding{}, err
	}

	var query string
	var args []any
	switch contract.CredentialModeRequired {
	case "TENANT_DEDICATED":
		query = `
			SELECT binding_id, credential_mode, COALESCE(tenant_id, ''), provider_account_id, secret_ref, settlement_party, state
			FROM credential_bindings
			WHERE credential_mode = 'TENANT_DEDICATED' AND tenant_id = $1 AND provider_account_id = $2 AND state = 'ATIVO'
		`
		args = []any{tenantID, providerAccountID}
	default: // SHARED_HUB
		query = `
			SELECT binding_id, credential_mode, COALESCE(tenant_id, ''), provider_account_id, secret_ref, settlement_party, state
			FROM credential_bindings
			WHERE credential_mode = 'SHARED_HUB' AND provider_account_id = $1 AND state = 'ATIVO'
			LIMIT 1
		`
		args = []any{providerAccountID}
	}

	var cb CredentialBinding
	row := s.db.QueryRowContext(ctx, query, args...)
	err = row.Scan(&cb.BindingID, &cb.CredentialMode, &cb.TenantID, &cb.ProviderAccountID, &cb.SecretRef, &cb.SettlementParty, &cb.State)
	if errors.Is(err, sql.ErrNoRows) {
		// SEG-05: credencial dedicada ausente/expirada/revogada NAO cai
		// para SHARED_HUB nem para credencial de outro tenant.
		return CredentialBinding{}, ErrCredentialUnavailable
	}
	if err != nil {
		return CredentialBinding{}, fmt.Errorf("atlas: resolver credencial: %w", err)
	}
	return cb, nil
}

// Contract e o contrato comercial/tecnico do tenant (FIN-02, CFG-04).
type Contract struct {
	TenantID               string  `json:"tenant_id"`
	Plan                   string  `json:"plan"`
	UnitPrice              float64 `json:"unit_price"`
	StrictBalance          bool    `json:"strict_balance"`
	ClientSLASeconds       int     `json:"client_sla_seconds"`
	CredentialModeRequired string  `json:"credential_mode_required"`
}

// UpsertContract cadastra/atualiza o contrato de um tenant.
func (s *Store) UpsertContract(ctx context.Context, c Contract) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO contracts (tenant_id, plan, unit_price, strict_balance, client_sla_seconds, credential_mode_required)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (tenant_id) DO UPDATE SET
			plan = EXCLUDED.plan,
			unit_price = EXCLUDED.unit_price,
			strict_balance = EXCLUDED.strict_balance,
			client_sla_seconds = EXCLUDED.client_sla_seconds,
			credential_mode_required = EXCLUDED.credential_mode_required
	`, c.TenantID, c.Plan, c.UnitPrice, c.StrictBalance, c.ClientSLASeconds, c.CredentialModeRequired)
	if err != nil {
		return fmt.Errorf("atlas: publicar contrato: %w", err)
	}
	return nil
}

// GetContract le o contrato de um tenant.
func (s *Store) GetContract(ctx context.Context, tenantID string) (Contract, error) {
	var c Contract
	row := s.db.QueryRowContext(ctx, `
		SELECT tenant_id, plan, unit_price, strict_balance, client_sla_seconds, credential_mode_required
		FROM contracts WHERE tenant_id = $1
	`, tenantID)
	err := row.Scan(&c.TenantID, &c.Plan, &c.UnitPrice, &c.StrictBalance, &c.ClientSLASeconds, &c.CredentialModeRequired)
	if errors.Is(err, sql.ErrNoRows) {
		return Contract{}, ErrNotFound
	}
	if err != nil {
		return Contract{}, fmt.Errorf("atlas: ler contrato: %w", err)
	}
	return c, nil
}

// Ping verifica a conectividade com hub_control (usado em readiness).
func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}
