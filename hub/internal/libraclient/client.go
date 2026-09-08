// Package libraclient e o cliente HTTP que a Orbita usa para obter uma
// reserva financeira estrita antes de liberar qualquer passo externo
// (FIN-06: "nenhum passo externo e liberado antes da reserva confirmada").
package libraclient

import (
	"ai-hub/hub/internal/platform/auth"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// ErrLimitExceeded espelha libra.ErrLimitExceeded pelo canal HTTP
// (status 409).
var ErrLimitExceeded = errors.New("libraclient: reserva excederia o limite estrito do tenant")

// Client e o cliente HTTP do Libra.
type Client struct {
	baseURL string
	http    *http.Client
}

// New cria um cliente do Libra.
func New(baseURL string) *Client {
	return &Client{baseURL: baseURL, http: auth.WorkloadClient(baseURL, 3*time.Second)}
}

// Reserve solicita a reserva atomica do valor do contrato para o
// protocolo (FIN-06). Idempotente por protocol_id no lado do Libra.
func (c *Client) Reserve(ctx context.Context, tenantID, protocolID string, amount float64, currency string) error {
	return c.ReserveExact(ctx, tenantID, protocolID, strconv.FormatFloat(amount, 'f', -1, 64), currency)
}

// ReserveExact keeps decimal values textual end to end.
func (c *Client) ReserveExact(ctx context.Context, tenantID, protocolID, amount, currency string) error {
	body, _ := json.Marshal(map[string]any{
		"tenant_id": tenantID, "protocol_id": protocolID, "amount": amount, "currency": currency,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/reservations", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("libraclient: falha de comunicacao com Libra: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusConflict {
		return ErrLimitExceeded
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("libraclient: status inesperado %d", resp.StatusCode)
	}
	return nil
}
