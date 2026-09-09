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
	"ai-hub/hub/internal/platform/egress"
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
	if err != nil || target.AdapterID != "synthetic-provider" {
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
	client, err := egress.NewClient(endpoint, time.Duration(c.TimeoutSeconds)*time.Second)
	if err != nil {
		return unknown("poll_egress_refused")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return unknown("poll_request_invalid")
	}
	if err = e.tokenCache.Apply(ctx, client, snap.Account.ID, providerauth.Config{BindingID: cred.BindingID, TenantID: c.Command.TenantID, Environment: os.Getenv("ENVIRONMENT"), SecretVersion: binding.SecretVersion, AuthType: pa.AuthType, Username: pa.AuthUsername, SecretRef: cred.SecretRef, APIKeyHeader: pa.APIKeyHeader, TokenURL: pa.OAuthTokenURL, ClientID: pa.OAuthClientID, ClientSecretRef: cred.SecretRef, MTLSCertificateRef: pa.MTLSCertificateRef, TokenTTLSeconds: pa.TokenTTLSeconds}, req); err != nil {
		return unknown("poll_authentication_failed")
	}
	// Recheck the committed fence after credential resolution, before HTTP I/O.
	var owned bool
	err = e.store.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM polling_schedule WHERE operation_id=$1 AND lease_owner=$2 AND epoch=$3 AND lease_expires_at>clock_timestamp()+make_interval(secs=>timeout_seconds) AND deadline_at>clock_timestamp()+make_interval(secs=>timeout_seconds))`, c.Command.CommandID, c.Owner, c.Epoch).Scan(&owned)
	if err != nil || !owned {
		return unknown("poll_fence_expired")
	}
	if _, err = e.store.db.ExecContext(ctx, `UPDATE attempts SET sent_at=clock_timestamp() WHERE attempt_id=$1 AND operation_id=$2`, c.AttemptID, c.Command.CommandID); err != nil {
		return unknown("poll_attempt_unavailable")
	}
	resp, err := client.Do(req)
	if err != nil {
		return unknown("poll_transport_failed")
	}
	defer resp.Body.Close()
	retryAfter := pollRetryAfter(resp.Header.Get("Retry-After"), time.Now())
	body, err := io.ReadAll(io.LimitReader(resp.Body, 256*1024+1))
	if err != nil || len(body) > 256*1024 {
		return unknown("poll_body_invalid")
	}
	if resp.StatusCode != http.StatusOK {
		r, _ := unknown("poll_http_" + strconv.Itoa(resp.StatusCode))
		return r, retryAfter
	}
	var result providersim.OperationResult
	if json.Unmarshal(body, &result) != nil || result.ProviderRequestID != c.ProviderRequestID {
		return unknown("poll_response_invalid")
	}
	if result.Status == "PENDING" {
		r, _ := unknown("poll_pending")
		r.ResponseBody = result
		return r, retryAfter
	}
	if result.Status != "SUCCEEDED" && result.Status != "FAILED" {
		return unknown("poll_status_invalid")
	}
	if _, err = atlas.TransformJSON(body, nil, target.OutputSchema); err != nil {
		return unknown("poll_output_contract_failed")
	}
	kind := dispatch.FactSucceeded
	if result.Status == "FAILED" {
		kind = dispatch.FactFailed
	}
	return dispatch.Result{Kind: kind, ProviderRequestID: c.ProviderRequestID, ResponseBody: result}, retryAfter
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
