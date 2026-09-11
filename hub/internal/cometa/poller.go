package cometa

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"ai-hub/hub/internal/atlas"
	"ai-hub/hub/internal/atlasclient"
	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/platform/idgen"
	"ai-hub/hub/internal/providerauth"
	"ai-hub/hub/internal/providersim"
)

// Each replica has bounded workers; PostgreSQL coordinates ownership across replicas.
func RunPoller(ctx context.Context, store *Store, exec *Executor, _ *atlasclient.Client, interval time.Duration, log *slog.Logger) {
	if interval <= 0 {
		interval = time.Second
	}
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			owner := idgen.New()
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					c, err := store.ClaimPoll(ctx, os.Getenv("CELL_ID"), owner)
					if errors.Is(err, sql.ErrNoRows) {
						continue
					}
					if err != nil {
						log.Error("polling claim failed")
						continue
					}
					exec.observePoll(ctx, c)
				}
			}
		}()
	}
	wg.Wait()
}

func (e *Executor) observePoll(ctx context.Context, c PollClaim) {
	requestCtx, cancel := context.WithTimeout(ctx, time.Duration(c.TimeoutSeconds)*time.Second)
	defer cancel()
	r, retryAfter := e.requestPoll(requestCtx, c)
	saveCtx, done := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
	defer done()
	if err := e.store.CompletePoll(saveCtx, c, r, retryAfter); err != nil && !errors.Is(err, ErrPollFence) {
		e.log.Error("polling receipt custody failed", "operation_id", c.Command.CommandID)
	} else if err == nil && (r.Kind == dispatch.FactSucceeded || r.Kind == dispatch.FactFailed) {
		var snapshot atlas.OfferSnapshot
		if json.Unmarshal(c.Command.ConfigSnapshot, &snapshot) == nil {
			e.resolveCapacityPending(ctx, c.Command, snapshot, "poll-attempt:"+c.AttemptID)
		}
	}
}

func (e *Executor) requestPoll(ctx context.Context, c PollClaim) (dispatch.Result, time.Duration) {
	unknown := func(code string) (dispatch.Result, time.Duration) {
		return dispatch.Result{Kind: dispatch.FactUnknown, ProviderRequestID: c.ProviderRequestID, ErrorCode: code}, 0
	}
	var snap atlas.OfferSnapshot
	if json.Unmarshal(c.Command.ConfigSnapshot, &snap) != nil {
		return unknown("poll_snapshot_invalid")
	}
	target, err := atlas.DecodeCatalogData(snap.Target)
	if err != nil || target.AdapterID == "" {
		return unknown("poll_adapter_unavailable")
	}
	if !e.adapters.Supports(target.AdapterID) {
		return unknown("poll_adapter_unavailable")
	}
	var pa atlasclient.ProviderAccount
	if json.Unmarshal(snap.Account.Data, &pa) != nil || snap.Account.ID != c.Command.ProviderAccountID {
		return unknown("poll_account_invalid")
	}
	cred, err := e.atlas.BoundCredential(ctx, c.Command.TenantID, snap.Binding.ID, snap.Binding.Version)
	if err != nil || cred.BindingID != snap.Binding.ID {
		return unknown("poll_binding_unavailable")
	}
	binding, err := atlas.DecodeCatalogData(snap.Binding)
	if err != nil {
		return unknown("poll_binding_invalid")
	}
	endpoint := strings.TrimRight(pa.BaseURL, "/") + "/v1/operations/" + url.PathEscape(c.ProviderRequestID)
	client, err := e.clients.Client(endpoint, time.Duration(c.TimeoutSeconds)*time.Second)
	if err != nil {
		return unknown("poll_egress_refused")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return unknown("poll_request_invalid")
	}
	// Recheck the committed polling fence before obtaining capacity and credentials.
	var owned bool
	err = e.store.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM polling_schedule WHERE operation_id=$1 AND lease_owner=$2 AND epoch=$3 AND lease_expires_at>clock_timestamp()+make_interval(secs=>timeout_seconds) AND deadline_at>clock_timestamp()+make_interval(secs=>timeout_seconds))`, c.Command.CommandID, c.Owner, c.Epoch).Scan(&owned)
	if err != nil || !owned {
		return unknown("poll_fence_expired")
	}
	capacityPermit, capacityEnabled, err := e.acquireCapacity(ctx, c.Command, snap, "STATUS", c.AttemptID, c.Owner)
	if err != nil {
		return unknown("poll_capacity_unavailable")
	}
	pollBudget := time.Duration(c.TimeoutSeconds) * time.Second
	effectiveBudget := effectiveHTTPBudget(ctx, pollBudget, capacityPermit, capacityEnabled)
	if effectiveBudget <= 0 {
		if capacityEnabled {
			e.releaseCapacity(ctx, capacityPermit, "poll-capacity-lease-budget-exhausted")
		}
		return unknown("poll_capacity_fence")
	}
	if effectiveBudget != pollBudget {
		client, err = e.clients.Client(endpoint, effectiveBudget)
		if err != nil {
			if capacityEnabled {
				e.releaseCapacity(ctx, capacityPermit, "poll-egress-refused")
			}
			return unknown("poll_egress_refused")
		}
	}
	capacitySettled := false
	releaseCapacity := func(evidence string) {
		if capacityEnabled && !capacitySettled {
			e.releaseCapacity(ctx, capacityPermit, evidence)
			capacitySettled = true
		}
	}
	var transportStarted time.Time
	settleCapacity := func(result dispatch.Result) dispatch.Result {
		if capacityEnabled && !capacitySettled {
			e.settleCapacity(ctx, capacityPermit, transportStarted, result, false)
			capacitySettled = true
		}
		return result
	}
	if err := e.validateCapacity(ctx, capacityPermit, capacityEnabled); err != nil {
		releaseCapacity("poll-capacity-fence-before-provider-auth")
		return unknown("poll_capacity_fence")
	}
	authBudget := effectiveHTTPBudget(ctx, minDuration(pollBudget, 3*time.Second), capacityPermit, capacityEnabled)
	if authBudget <= 0 {
		releaseCapacity("poll-capacity-lease-auth-budget-exhausted")
		return unknown("poll_capacity_fence")
	}
	authCtx, authCancel := context.WithTimeout(ctx, authBudget)
	authErr := e.tokenCache.Apply(authCtx, client, snap.Account.ID, providerauth.Config{BindingID: cred.BindingID, TenantID: c.Command.TenantID, Environment: os.Getenv("ENVIRONMENT"), SecretVersion: binding.SecretVersion, AuthType: pa.AuthType, Username: pa.AuthUsername, SecretRef: cred.SecretRef, APIKeyHeader: pa.APIKeyHeader, TokenURL: pa.OAuthTokenURL, ClientID: pa.OAuthClientID, ClientSecretRef: cred.SecretRef, MTLSCertificateRef: pa.MTLSCertificateRef, TokenTTLSeconds: pa.TokenTTLSeconds}, req)
	authCancel()
	if authErr != nil {
		releaseCapacity("poll-provider-authentication-failed")
		return unknown("poll_authentication_failed")
	}
	if err := e.validateCapacity(ctx, capacityPermit, capacityEnabled); err != nil {
		releaseCapacity("poll-capacity-fence-before-status")
		return unknown("poll_capacity_fence")
	}
	if _, err = e.store.db.ExecContext(ctx, `UPDATE attempts SET sent_at=clock_timestamp() WHERE attempt_id=$1 AND operation_id=$2`, c.AttemptID, c.Command.CommandID); err != nil {
		releaseCapacity("poll-attempt-unavailable")
		return unknown("poll_attempt_unavailable")
	}
	transportStarted = time.Now()
	if err := e.validateCapacity(ctx, capacityPermit, capacityEnabled); err != nil {
		releaseCapacity("poll-capacity-fence-immediately-before-status")
		return unknown("poll_capacity_fence")
	}
	resp, err := client.Do(req)
	if err != nil {
		r, retryAfter := unknown("poll_transport_failed")
		return settleCapacity(r), retryAfter
	}
	defer resp.Body.Close()
	retryAfter := pollRetryAfter(resp.Header.Get("Retry-After"), time.Now())
	body, err := io.ReadAll(io.LimitReader(resp.Body, 256*1024+1))
	if err != nil || len(body) > 256*1024 {
		r, retryAfter := unknown("poll_body_invalid")
		return settleCapacity(r), retryAfter
	}
	if resp.StatusCode != http.StatusOK {
		r, _ := unknown("poll_http_" + strconv.Itoa(resp.StatusCode))
		return settleCapacity(r), retryAfter
	}
	var result providersim.OperationResult
	if json.Unmarshal(body, &result) != nil || result.ProviderRequestID != c.ProviderRequestID {
		r, retryAfter := unknown("poll_response_invalid")
		return settleCapacity(r), retryAfter
	}
	if result.Status == "PENDING" {
		r, _ := unknown("poll_pending")
		r.ResponseBody = result
		r.RawResponse = append([]byte(nil), body...)
		return settleCapacity(r), retryAfter
	}
	if result.Status != "SUCCEEDED" && result.Status != "FAILED" {
		r, retryAfter := unknown("poll_status_invalid")
		return settleCapacity(r), retryAfter
	}
	normalized, err := normalizeProviderResult(snap, result)
	if err != nil {
		r, retryAfter := unknown("poll_output_contract_failed")
		return settleCapacity(r), retryAfter
	}
	kind := dispatch.FactSucceeded
	if result.Status == "FAILED" {
		kind = dispatch.FactFailed
	}
	return settleCapacity(dispatch.Result{Kind: kind, ProviderRequestID: c.ProviderRequestID, ResponseBody: normalized, RawResponse: append([]byte(nil), body...)}), retryAfter
}

func pollRetryAfter(value string, now time.Time) time.Duration {
	if seconds, err := strconv.ParseInt(value, 10, 32); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	if at, err := http.ParseTime(value); err == nil && at.After(now) {
		return at.Sub(now)
	}
	return 0
}
