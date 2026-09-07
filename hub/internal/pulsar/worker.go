package pulsar

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

const maxAttempts = 5

// DeliveryWorker envia tentativas de webhook (EXE-08): busca a mesma
// representacao final materializada usada pelo GET (COM-05), assina
// com HMAC-SHA256 e tenta a entrega com retry/backoff.
type DeliveryWorker struct {
	store      *Store
	orbitaURL  string
	httpClient *http.Client
	log        *slog.Logger
}

// NewDeliveryWorker cria um DeliveryWorker.
func NewDeliveryWorker(store *Store, orbitaURL string, log *slog.Logger) *DeliveryWorker {
	return &DeliveryWorker{
		store:     store,
		orbitaURL: orbitaURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		log:        log,
	}
}

// Run consome entregas devidas em loop ate o contexto ser cancelado.
func (w *DeliveryWorker) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			due, err := w.store.DueDeliveries(ctx, time.Now(), 50)
			if err != nil {
				w.log.Error("pulsar: falha ao buscar entregas devidas", "error", err)
				continue
			}
			for _, d := range due {
				w.attempt(ctx, d)
			}
		}
	}
}

func (w *DeliveryWorker) attempt(ctx context.Context, d Delivery) {
	resp, err := w.httpClient.Get(w.orbitaURL + "/internal/protocols/" + d.ProtocolID)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		w.reschedule(ctx, d, "falha ao buscar representacao final na Orbita")
		return
	}
	var body bytes.Buffer
	_, _ = body.ReadFrom(resp.Body)
	resp.Body.Close()

	// O segredo HMAC e resolvido pela URL de destino, armazenado junto
	// do cadastro em webhook_destinations.
	dest, err := w.store.getDestinationByURL(ctx, d.DestinationURL)
	if err != nil {
		w.reschedule(ctx, d, "destino nao encontrado para assinatura")
		return
	}

	sig := sign(dest.HMACSecret, body.Bytes())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.DestinationURL, bytes.NewReader(body.Bytes()))
	if err != nil {
		w.reschedule(ctx, d, "falha ao construir requisicao de entrega")
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature-256", sig)
	req.Header.Set("X-Hub-Event-Id", d.EventID)
	req.Header.Set("X-Hub-Delivery-Id", d.DeliveryID)

	dResp, err := w.httpClient.Do(req)
	if err != nil {
		w.reschedule(ctx, d, "erro de transporte na entrega")
		return
	}
	defer dResp.Body.Close()

	if dResp.StatusCode >= 200 && dResp.StatusCode < 300 {
		// 2xx confirma recebimento HTTP, nao processamento interno do
		// cliente (EXE-08).
		if err := w.store.MarkDelivered(ctx, d.DeliveryID); err != nil {
			w.log.Error("pulsar: falha ao marcar entrega concluida", "error", err)
		}
		return
	}
	w.reschedule(ctx, d, fmt.Sprintf("destino respondeu status %d", dResp.StatusCode))
}

func (w *DeliveryWorker) reschedule(ctx context.Context, d Delivery, reason string) {
	attempts := d.AttemptsCount + 1
	backoff := time.Duration(attempts*attempts) * time.Second // backoff simples e crescente
	next := time.Now().Add(backoff)
	if err := w.store.RetryOrExhaust(ctx, d.DeliveryID, attempts, maxAttempts, next); err != nil {
		w.log.Error("pulsar: falha ao reagendar entrega", "error", err)
	}
	w.log.Warn("pulsar: tentativa de entrega falhou", "delivery_id", d.DeliveryID, "reason", reason, "attempts", attempts)
}

func sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
