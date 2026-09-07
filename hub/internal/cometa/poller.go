package cometa

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"ai-hub/hub/internal/atlasclient"
	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/providersim"
)

// RunPoller consulta periodicamente o provedor para operacoes em modo
// polling (EXE-05): "polling significa consulta periodica ao
// provedor". Cada operacao tem lease unica por si so nesta referencia
// (uma unica instancia de Cometa local/dev); multiplas replicas
// exigiriam lease/token de posse (nao implementado aqui — placeholder,
// ver auditoria).
func RunPoller(ctx context.Context, store *Store, exec *Executor, atlas *atlasclient.Client, interval time.Duration, log *slog.Logger) {
	client := &http.Client{Timeout: 5 * time.Second}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			due, err := store.DuePolling(ctx, time.Now(), 50)
			if err != nil {
				log.Error("poller: falha ao buscar agenda devida", "error", err)
				continue
			}
			for _, d := range due {
				pollOne(ctx, client, store, exec, atlas, d, log)
			}
		}
	}
}

func pollOne(ctx context.Context, client *http.Client, store *Store, exec *Executor, atlas *atlasclient.Client, d DuePoll, log *slog.Logger) {
	if d.ProviderRequestID == "" {
		return
	}
	pa, err := atlas.ProviderAccount(ctx, d.ProviderAccountID)
	if err != nil {
		log.Warn("poller: conta de provedor nao resolvida", "provider_account_id", d.ProviderAccountID)
		return
	}
	url := fmt.Sprintf("%s/v1/operations/%s", pa.BaseURL, d.ProviderRequestID)
	resp, err := client.Get(url)
	if err != nil {
		// Falha de rede na consulta: nao reinicia a operacao, apenas
		// tenta de novo no proximo ciclo (EXE-05/EXE-06).
		advance(ctx, store, d, log)
		return
	}
	defer resp.Body.Close()
	var result providersim.OperationResult
	_ = json.NewDecoder(resp.Body).Decode(&result)

	if result.Status == "PENDING" {
		advance(ctx, store, d, log)
		return
	}
	exec.ApplyExternalObservation(ctx, dispatch.Command{ProtocolID: d.ProtocolID, ProviderAccountID: d.ProviderAccountID}, d.OperationID, result)
}

func advance(ctx context.Context, store *Store, d DuePoll, log *slog.Logger) {
	next := time.Duration(d.IntervalSeconds) * time.Second
	max := time.Duration(d.MaxIntervalSeconds) * time.Second
	backoff := next * 2
	if backoff > max {
		backoff = max
	}
	if err := store.AdvancePolling(ctx, d.OperationID, time.Now().Add(next)); err != nil {
		log.Error("poller: falha ao reagendar", "error", err, "operation_id", d.OperationID)
	}
	_ = backoff // backoff calculado para evolucao futura do intervalo persistido
}
