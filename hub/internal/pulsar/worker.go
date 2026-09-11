package pulsar

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"ai-hub/hub/internal/cometa"
	"ai-hub/hub/internal/platform/egress"
	"ai-hub/hub/internal/platform/idgen"
	"ai-hub/hub/internal/providerauth"
	"github.com/lib/pq"
)

type DeliveryWorker struct {
	store          *Store
	log            *slog.Logger
	resolver       providerauth.SecretResolver
	clients        *egress.Pool
	capacity       *cometa.CapacityController
	capacityDomain string
}

func NewDeliveryWorker(store *Store, orbitaURL string, log *slog.Logger) *DeliveryWorker {
	domain := strings.TrimSpace(os.Getenv("CAPACITY_WEBHOOK_DOMAIN"))
	var capacity *cometa.CapacityController
	if domain != "" {
		capacity = cometa.NewCapacityController(store.db)
	}
	return &DeliveryWorker{store: store, log: log, resolver: providerauth.AWSVault{}, clients: egress.NewPool(egress.FromEnv()), capacity: capacity, capacityDomain: domain}
}

func (w *DeliveryWorker) Run(ctx context.Context, interval time.Duration) {
	var workers sync.WaitGroup
	for n := 0; n < 4; n++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			owner := idgen.New()
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					d, err := w.store.ClaimDelivery(ctx, os.Getenv("CELL_ID"), owner)
					if errors.Is(err, sql.ErrNoRows) {
						continue
					}
					if err != nil {
						w.log.Error("delivery claim unavailable")
						continue
					}
					w.attempt(ctx, d)
				}
			}
		}()
	}
	workers.Wait()
}

func (w *DeliveryWorker) attempt(ctx context.Context, d ClaimedDelivery) {
	finish := func(status int, code string) {
		save, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
		defer cancel()
		if err := w.store.CompleteDelivery(save, d, status, code); err != nil {
			var dbError *pq.Error
			dbCode := "non_sql"
			if errors.As(err, &dbError) {
				dbCode = string(dbError.Code)
			}
			w.log.Error("delivery receipt custody unavailable", "delivery_id", d.ID, "database_code", dbCode)
		}
	}
	sum := sha256.Sum256(d.Body)
	if hex.EncodeToString(sum[:]) != d.Hash {
		finish(0, "representation_integrity_failed")
		return
	}
	call, cancel := context.WithTimeout(ctx, time.Duration(d.TimeoutSeconds)*time.Second)
	defer cancel()
	started := time.Now()

	var permit cometa.CapacityPermit
	capacityActive := w.capacity != nil && w.capacityDomain != ""
	if capacityActive {
		var err error
		permit, err = w.capacity.Acquire(call, w.capacityDomain, d.AttemptID, d.TenantID, d.CellID, d.Owner, "FETCH")
		if err != nil {
			finish(0, "capacity_unavailable")
			return
		}
		deadline, ok := call.Deadline()
		if !ok || !capacityLeaseCovers(permit.LeaseUntil, deadline) {
			releaseCtx, releaseCancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
			releaseErr := w.capacity.Release(releaseCtx, permit, "webhook-timeout-exceeds-capacity-lease")
			releaseCancel()
			if releaseErr != nil {
				w.log.Error("webhook capacity release unavailable", "delivery_id", d.ID, "error", releaseErr)
			}
			finish(0, "capacity_lease_too_short")
			return
		}
	}
	releaseCapacity := func(evidence string) {
		if !capacityActive {
			return
		}
		releaseCtx, releaseCancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
		defer releaseCancel()
		if err := w.capacity.Release(releaseCtx, permit, evidence); err != nil {
			w.log.Error("webhook capacity release unavailable", "delivery_id", d.ID, "error", err)
		}
	}
	completeCapacity := func(signal string, evidence string, latency time.Duration) {
		if !capacityActive {
			return
		}
		completeCtx, completeCancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
		defer completeCancel()
		if err := w.capacity.CompleteTransport(completeCtx, permit, signal, latency, false, evidence); err != nil {
			w.log.Error("webhook capacity receipt unavailable", "delivery_id", d.ID, "error", err)
		}
	}
	secret, err := w.resolver.Resolve(call, d.SecretRef, d.SecretVersion)
	if err != nil {
		releaseCapacity("webhook-secret-unavailable")
		finish(0, "signing_key_unavailable")
		return
	}
	client, err := w.clients.Client(d.URL, time.Duration(d.TimeoutSeconds)*time.Second)
	if err != nil {
		releaseCapacity("webhook-egress-refused")
		finish(0, "egress_refused")
		return
	}
	request, err := http.NewRequestWithContext(call, http.MethodPost, d.URL, bytes.NewReader(d.Body))
	if err != nil {
		releaseCapacity("webhook-destination-invalid")
		finish(0, "invalid_destination")
		return
	}
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signed := append([]byte(timestamp+"."+d.EventID+"."+d.ID+"."), d.Body...)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Hub-Signature-256", sign(secret.Value, signed))
	request.Header.Set("X-Hub-Timestamp", timestamp)
	request.Header.Set("X-Hub-Event-Id", d.EventID)
	request.Header.Set("X-Hub-Delivery-Id", d.ID)
	response, err := client.Do(request)
	if err != nil {
		completeCapacity("TIMEOUT", "webhook-transport-unconfirmed", time.Since(started))
		finish(0, "transport_unconfirmed")
		return
	}
	defer response.Body.Close()
	completeCapacity(webhookCapacitySignal(response.StatusCode), "webhook-http-receipt", time.Since(started))
	finish(response.StatusCode, "")
}

const capacityLeaseSafetyMargin = time.Second

func capacityLeaseCovers(leaseUntil, deadline time.Time) bool {
	return leaseUntil.After(deadline.Add(capacityLeaseSafetyMargin))
}

func webhookCapacitySignal(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "SUCCESS"
	case status == http.StatusTooManyRequests:
		return "THROTTLED"
	default:
		return "UNAVAILABLE"
	}
}

func sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
