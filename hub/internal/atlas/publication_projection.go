package atlas

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/lib/pq"
)

// projectPublication updates the compatibility projection in the same Atlas
// transaction as publication; execution snapshots still pin immutable versions.
func projectPublication(ctx context.Context, tx *sql.Tx, r Resource, actor string) error {
	d, _ := DecodeCatalogData(r)
	switch r.Kind {
	case "services":
		_, err := tx.ExecContext(ctx, `INSERT INTO services(code,version,description,schema_input,schema_output,modes,client_sla_seconds,retry_ttl_seconds,published) VALUES($1,$2,$3,$4,$5,$6,$7,$8,true) ON CONFLICT(code,version) DO NOTHING`, r.ID, r.Version, r.Name, []byte(d.InputSchema), []byte(d.OutputSchema), pq.Array(d.Modes), d.ClientSLASeconds, d.RetryTTLSeconds)
		return err
	case "provider-accounts":
		var pa ProviderAccount
		if err := json.Unmarshal(r.Data, &pa); err != nil {
			return err
		}
		pa.ProviderAccountID = r.ID
		if pa.AuthType == "" {
			pa.AuthType = "NONE"
		}
		if pa.TokenTTLSeconds == 0 {
			pa.TokenTTLSeconds = 300
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO provider_accounts(provider_account_id,provider_id,environment,base_url,provider_mode,auth_type,auth_username,auth_secret_ref,oauth_token_url,oauth_client_id,oauth_client_secret_ref,mtls_certificate_ref,token_ttl_seconds) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) ON CONFLICT(provider_account_id) DO UPDATE SET provider_id=EXCLUDED.provider_id,environment=EXCLUDED.environment,base_url=EXCLUDED.base_url,provider_mode=EXCLUDED.provider_mode,auth_type=EXCLUDED.auth_type,auth_username=EXCLUDED.auth_username,auth_secret_ref=EXCLUDED.auth_secret_ref,oauth_token_url=EXCLUDED.oauth_token_url,oauth_client_id=EXCLUDED.oauth_client_id,oauth_client_secret_ref=EXCLUDED.oauth_client_secret_ref,mtls_certificate_ref=EXCLUDED.mtls_certificate_ref,token_ttl_seconds=EXCLUDED.token_ttl_seconds`, pa.ProviderAccountID, pa.ProviderID, pa.Environment, pa.BaseURL, pa.ProviderMode, pa.AuthType, pa.AuthUsername, pa.AuthSecretRef, pa.OAuthTokenURL, pa.OAuthClientID, pa.OAuthClientSecretRef, pa.MTLSCertificateRef, pa.TokenTTLSeconds)
		return err
	case "credential-bindings":
		_, err := tx.ExecContext(ctx, `INSERT INTO credential_bindings(binding_id,credential_mode,tenant_id,provider_account_id,secret_ref,settlement_party,state) VALUES($1,$2,NULLIF($3,''),$4,$5,$6,'ATIVO') ON CONFLICT(binding_id) DO UPDATE SET secret_ref=EXCLUDED.secret_ref,state='ATIVO'`, r.ID, d.CredentialMode, r.TenantID, d.ProviderAccountID, d.SecretRef, d.SettlementParty)
		return err
	case "clients":
		var existingCell, state string
		var existingUnits int
		err := tx.QueryRowContext(ctx, `SELECT cell_id,capacity_units,state FROM catalog_onboardings WHERE tenant_id=$1 FOR UPDATE`, r.TenantID).Scan(&existingCell, &existingUnits, &state)
		if err == nil && state == "ACTIVE" {
			if existingCell != d.CellID || existingUnits != d.CapacityUnits {
				return errors.New("migração de placement exige fluxo de drenagem")
			}
			return nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		result, err := tx.ExecContext(ctx, `UPDATE catalog_capacity_cells SET reserved_units=reserved_units+$2 WHERE cell_id=$1 AND state='READY' AND total_units-reserved_units >= $2`, d.CellID, d.CapacityUnits)
		if err != nil {
			return err
		}
		n, _ := result.RowsAffected()
		if n != 1 {
			return ErrCapacityUnavailable
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO catalog_onboardings(tenant_id,cell_id,capacity_units,state,reason,actor) VALUES($1,$2,$3,'ACTIVE','qualified-capacity-reserved',$4) ON CONFLICT(tenant_id) DO UPDATE SET cell_id=EXCLUDED.cell_id,capacity_units=EXCLUDED.capacity_units,state='ACTIVE',reason=EXCLUDED.reason,actor=EXCLUDED.actor,updated_at=clock_timestamp()`, r.TenantID, d.CellID, d.CapacityUnits, actor)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO placements(tenant_id,cell_id) VALUES($1,$2) ON CONFLICT(tenant_id) DO UPDATE SET cell_id=EXCLUDED.cell_id,updated_at=clock_timestamp()`, r.TenantID, d.CellID)
		return err
	}
	return nil
}

var ErrCapacityUnavailable = errors.New("capacidade qualificada indisponível; provisionamento pendente")

func (s *Store) requestCapacity(ctx context.Context, r Resource, actor string) error {
	d, _ := DecodeCatalogData(r)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `INSERT INTO catalog_onboardings(tenant_id,cell_id,capacity_units,state,reason,actor) VALUES($1,$2,$3,'PROVISIONING','insufficient-qualified-headroom',$4) ON CONFLICT(tenant_id) DO UPDATE SET reason=EXCLUDED.reason,updated_at=clock_timestamp() WHERE catalog_onboardings.state<>'ACTIVE'`, r.TenantID, d.CellID, d.CapacityUnits, actor)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO catalog_provisioning_requests(request_id,tenant_id,cell_id,required_units) VALUES($1,$2,$3,$4) ON CONFLICT DO NOTHING`, fmt.Sprintf("%s:%s:%d", r.TenantID, d.CellID, d.CapacityUnits), r.TenantID, d.CellID, d.CapacityUnits)
	if err != nil {
		return err
	}
	return tx.Commit()
}
