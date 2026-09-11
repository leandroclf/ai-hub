package cometa

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"ai-hub/hub/internal/atlas"
	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/platform/idgen"
	"ai-hub/hub/internal/providerauth"
	"ai-hub/hub/internal/providersim"
)

// ReconciliationClaim representa uma solicitação administrativa já
// reivindicada. A reivindicação só permite consultar o status externo; não
// existe caminho neste tipo para reenviar a submissão original.
type ReconciliationClaim struct {
	RequestID, ProtocolID, OperationID, ProviderRequestID, Owner string
	Epoch                                                        int64
	Command                                                      dispatch.Command
}

func (s *Store) ClaimReconciliation(ctx context.Context, cell, owner string) (ReconciliationClaim, error) {
	if cell == "" || owner == "" {
		return ReconciliationClaim{}, errors.New("missing reconciliation identity")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ReconciliationClaim{}, err
	}
	defer tx.Rollback()
	var claim ReconciliationClaim
	var raw []byte
	err = tx.QueryRowContext(ctx, `SELECT r.request_id,r.protocol_id,r.claim_epoch+1,o.operation_id,o.provider_request_id,o.command
		FROM protocol_reconciliation_requests r
		JOIN operations o ON o.protocol_id=r.protocol_id AND o.tenant_id=r.tenant_id
		WHERE r.state='OPEN' AND o.cell_id=$1
		  AND o.state IN ('UNKNOWN','ACCEPTED_EXTERNAL','WAITING_FINAL')
		  AND o.provider_request_id IS NOT NULL
		  AND (r.lease_until IS NULL OR r.lease_until<=clock_timestamp())
		ORDER BY r.created_at,r.request_id
		FOR UPDATE OF r SKIP LOCKED LIMIT 1`, cell).Scan(&claim.RequestID, &claim.ProtocolID, &claim.Epoch, &claim.OperationID, &claim.ProviderRequestID, &raw)
	if err != nil {
		return ReconciliationClaim{}, err
	}
	if json.Unmarshal(raw, &claim.Command) != nil || claim.Command.CommandID != claim.OperationID || claim.Command.ProtocolID != claim.ProtocolID || claim.Command.CellID != cell {
		return ReconciliationClaim{}, errors.New("invalid reconciliation command")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE protocol_reconciliation_requests
		SET claim_owner=$2,claim_epoch=$3,lease_until=clock_timestamp()+interval '30 seconds',attempts=attempts+1,last_error=NULL,updated_at=clock_timestamp()
		WHERE request_id=$1 AND state='OPEN'`, claim.RequestID, owner, claim.Epoch); err != nil {
		return ReconciliationClaim{}, err
	}
	claim.Owner = owner
	return claim, tx.Commit()
}

func (s *Store) ReleaseReconciliation(ctx context.Context, claim ReconciliationClaim, reason string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE protocol_reconciliation_requests
		SET claim_owner=NULL,lease_until=NULL,last_error=$4,updated_at=clock_timestamp()
		WHERE request_id=$1 AND state='OPEN' AND claim_owner=$2 AND claim_epoch=$3`, claim.RequestID, claim.Owner, claim.Epoch, reason)
	return err
}

func (s *Store) ResolveReconciliation(ctx context.Context, claim ReconciliationClaim, evidenceID string) error {
	if evidenceID == "" {
		return errors.New("missing reconciliation evidence")
	}
	result, err := s.db.ExecContext(ctx, `UPDATE protocol_reconciliation_requests
		SET state='RESOLVED',evidence_id=$4,claim_owner=NULL,lease_until=NULL,updated_at=clock_timestamp()
		WHERE request_id=$1 AND state='OPEN' AND claim_owner=$2 AND claim_epoch=$3`, claim.RequestID, claim.Owner, claim.Epoch, evidenceID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err == nil && n != 1 {
		return ErrPollFence
	}
	return err
}

// ReconcileExternal consulta o endpoint de status do provedor usando a
// identidade e a credencial congeladas da operação. Não chama o endpoint de
// submissão e só conserva um resultado terminal após validar o contrato.
func (e *Executor) ReconcileExternal(ctx context.Context, claim ReconciliationClaim) (dispatch.Result, error) {
	var snap atlas.OfferSnapshot
	if json.Unmarshal(claim.Command.ConfigSnapshot, &snap) != nil {
		return dispatch.Result{}, errors.New("reconciliation snapshot unavailable")
	}
	target, err := atlas.DecodeCatalogData(snap.Target)
	if err != nil || target.AdapterID == "" {
		return dispatch.Result{}, errors.New("reconciliation adapter unavailable")
	}
	if !e.adapters.Supports(target.AdapterID) {
		return dispatch.Result{}, ErrAdapterUnavailable
	}
	var pa struct {
		BaseURL            string `json:"base_url"`
		ProviderMode       string `json:"provider_mode"`
		AuthType           string `json:"auth_type"`
		APIKeyHeader       string `json:"api_key_header"`
		AuthUsername       string `json:"auth_username"`
		OAuthTokenURL      string `json:"oauth_token_url"`
		OAuthClientID      string `json:"oauth_client_id"`
		MTLSCertificateRef string `json:"mtls_certificate_ref"`
		TokenTTLSeconds    int    `json:"token_ttl_seconds"`
	}
	if json.Unmarshal(snap.Account.Data, &pa) != nil || snap.Account.ID != claim.Command.ProviderAccountID {
		return dispatch.Result{}, errors.New("reconciliation account unavailable")
	}
	cred, err := e.atlas.BoundCredential(ctx, claim.Command.TenantID, snap.Binding.ID, snap.Binding.Version)
	if err != nil {
		return dispatch.Result{}, fmt.Errorf("reconciliation binding unavailable: %w", err)
	}
	if cred.BindingID != snap.Binding.ID {
		return dispatch.Result{}, errors.New("reconciliation binding unavailable: binding changed")
	}
	binding, err := atlas.DecodeCatalogData(snap.Binding)
	if err != nil {
		return dispatch.Result{}, errors.New("reconciliation binding invalid")
	}
	endpoint := strings.TrimRight(pa.BaseURL, "/") + "/v1/operations/" + url.PathEscape(claim.ProviderRequestID)
	providerBudget := providerHTTPBudget(snap, 15*time.Second)
	clientBudget := effectiveHTTPBudget(ctx, providerBudget, CapacityPermit{}, false)
	if clientBudget <= 0 {
		return dispatch.Result{}, fmt.Errorf("reconciliation request budget exhausted")
	}
	client, err := e.clients.Client(endpoint, clientBudget)
	if err != nil {
		return dispatch.Result{}, fmt.Errorf("reconciliation egress unavailable for %q: %w", endpoint, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return dispatch.Result{}, fmt.Errorf("reconciliation request invalid: %w", err)
	}
	capacityPermit, capacityEnabled, err := e.acquireCapacity(ctx, claim.Command, snap, "STATUS", "reconcile-"+claim.RequestID+"-"+idgen.New(), claim.Owner)
	if err != nil {
		return dispatch.Result{}, fmt.Errorf("reconciliation capacity unavailable: %w", err)
	}
	capacitySettled := false
	releaseCapacity := func(evidence string) {
		if capacityEnabled && !capacitySettled {
			e.releaseCapacity(ctx, capacityPermit, evidence)
			capacitySettled = true
		}
	}
	clientBudget = effectiveHTTPBudget(ctx, providerBudget, capacityPermit, capacityEnabled)
	if clientBudget <= 0 {
		releaseCapacity("reconciliation-capacity-lease-budget-exhausted")
		return dispatch.Result{}, fmt.Errorf("reconciliation capacity fence: %w", ErrCapacityFence)
	}
	if clientBudget != providerBudget {
		client, err = e.clients.Client(endpoint, clientBudget)
		if err != nil {
			releaseCapacity("reconciliation-egress-refused")
			return dispatch.Result{}, fmt.Errorf("reconciliation egress unavailable for %q: %w", endpoint, err)
		}
	}
	transportStarted := time.Now()
	settleCapacity := func(result dispatch.Result) dispatch.Result {
		if capacityEnabled && !capacitySettled {
			e.settleCapacity(ctx, capacityPermit, transportStarted, result, false)
			capacitySettled = true
		}
		return result
	}
	unknown := func(code string) dispatch.Result {
		return dispatch.Result{CommandID: claim.OperationID, OperationID: claim.OperationID, ProviderRequestID: claim.ProviderRequestID, Kind: dispatch.FactUnknown, ErrorCode: code}
	}
	if err := e.validateCapacity(ctx, capacityPermit, capacityEnabled); err != nil {
		releaseCapacity("reconciliation-capacity-fence-before-provider-auth")
		return dispatch.Result{}, fmt.Errorf("reconciliation capacity fence: %w", err)
	}
	authBudget := effectiveHTTPBudget(ctx, minDuration(providerBudget, 3*time.Second), capacityPermit, capacityEnabled)
	if authBudget <= 0 {
		releaseCapacity("reconciliation-capacity-lease-auth-budget-exhausted")
		return dispatch.Result{}, fmt.Errorf("reconciliation capacity fence: %w", ErrCapacityFence)
	}
	authCtx, authCancel := context.WithTimeout(ctx, authBudget)
	authErr := e.tokenCache.Apply(authCtx, client, snap.Account.ID, providerauth.Config{BindingID: cred.BindingID, TenantID: claim.Command.TenantID, Environment: os.Getenv("ENVIRONMENT"), SecretVersion: binding.SecretVersion, AuthType: pa.AuthType, Username: pa.AuthUsername, SecretRef: cred.SecretRef, APIKeyHeader: pa.APIKeyHeader, TokenURL: pa.OAuthTokenURL, ClientID: pa.OAuthClientID, ClientSecretRef: cred.SecretRef, MTLSCertificateRef: pa.MTLSCertificateRef, TokenTTLSeconds: pa.TokenTTLSeconds}, req)
	authCancel()
	if authErr != nil {
		releaseCapacity("reconciliation-provider-authentication-failed")
		return dispatch.Result{}, fmt.Errorf("reconciliation authentication unavailable: %w", authErr)
	}
	if err := e.validateCapacity(ctx, capacityPermit, capacityEnabled); err != nil {
		releaseCapacity("reconciliation-capacity-fence-before-status")
		return dispatch.Result{}, fmt.Errorf("reconciliation capacity fence: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return settleCapacity(unknown("reconciliation_transport_failed")), fmt.Errorf("reconciliation status request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return settleCapacity(unknown("reconciliation_status_unavailable")), errors.New("reconciliation status unavailable")
	}
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 256*1024+1))
	if readErr != nil || len(body) > 256*1024 {
		return settleCapacity(unknown("reconciliation_response_invalid")), errors.New("reconciliation response invalid")
	}
	var result providersim.OperationResult
	if json.Unmarshal(body, &result) != nil || result.ProviderRequestID != claim.ProviderRequestID {
		return settleCapacity(unknown("reconciliation_response_invalid")), errors.New("reconciliation response invalid")
	}
	if result.Status == "PENDING" {
		pending := unknown("reconciliation_provider_pending")
		pending.ResponseBody = result
		pending.RawResponse = append([]byte(nil), body...)
		return settleCapacity(pending), nil
	}
	if result.Status != "SUCCEEDED" && result.Status != "FAILED" {
		return settleCapacity(unknown("reconciliation_status_invalid")), errors.New("reconciliation status invalid")
	}
	normalized, err := normalizeProviderResult(snap, result)
	if err != nil {
		return settleCapacity(unknown("reconciliation_output_contract_failed")), errors.New("reconciliation output contract failed")
	}
	kind := dispatch.FactSucceeded
	if result.Status == "FAILED" {
		kind = dispatch.FactFailed
	}
	return settleCapacity(dispatch.Result{CommandID: claim.OperationID, OperationID: claim.OperationID, ProviderRequestID: result.ProviderRequestID, Kind: kind, ResponseBody: normalized, RawResponse: append([]byte(nil), body...)}), nil
}

// RunReconciliationWorker executa somente consultas de status para solicitações
// abertas. A lease permite retomada por outra réplica e a evidência terminal
// é conservada antes de a solicitação ser marcada como resolvida.
func RunReconciliationWorker(ctx context.Context, store *Store, exec *Executor, cell string, interval time.Duration, log *slog.Logger) {
	if interval <= 0 {
		interval = time.Second
	}
	owner := "admin-reconciler-" + idgen.New()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			claim, err := store.ClaimReconciliation(ctx, cell, owner)
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			if err != nil {
				log.Error("reconciliation claim unavailable", "error", err)
				continue
			}
			call, cancel := context.WithTimeout(ctx, 20*time.Second)
			result, reconcileErr := exec.ReconcileExternal(call, claim)
			cancel()
			if reconcileErr != nil {
				if err = store.ReleaseReconciliation(ctx, claim, reconcileErr.Error()); err != nil {
					log.Error("reconciliation release unavailable", "error", err)
				}
				continue
			}
			if result.Kind == dispatch.FactUnknown {
				if err = store.ReleaseReconciliation(ctx, claim, "provider_pending"); err != nil {
					log.Error("reconciliation pending release unavailable", "error", err)
				}
				continue
			}
			durable, conserveErr := store.ConserveObservation(ctx, claim.Command, result, "ADMIN_RECONCILIATION", claim.RequestID)
			if conserveErr != nil {
				_ = store.ReleaseReconciliation(ctx, claim, conserveErr.Error())
				continue
			}
			var snapshot atlas.OfferSnapshot
			if json.Unmarshal(claim.Command.ConfigSnapshot, &snapshot) == nil {
				exec.resolveCapacityPending(ctx, claim.Command, snapshot, durable.EvidenceID)
			}
			if err = store.ResolveReconciliation(ctx, claim, durable.EvidenceID); err != nil {
				log.Error("reconciliation resolution unavailable", "error", err)
			}
		}
	}
}
