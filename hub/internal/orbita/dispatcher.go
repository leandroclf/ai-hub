package orbita

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/queue"
)

// Dispatcher envia o comando de despacho a Cometa: diretamente por
// HTTPS em SYNC (COM-01/COM-06), ou pela fila dedicada em ASYNC/AUTO
// (EXE-15: dispatch_mode QUEUED, nunca DIRECT).
type Dispatcher struct {
	cometaURL string
	client    *http.Client
	queue     *queue.Client
	queueURL  string
}

// NewDispatcher cria um Dispatcher.
func NewDispatcher(cometaURL string, q *queue.Client, queueURL string) *Dispatcher {
	return &Dispatcher{
		cometaURL: cometaURL,
		client:    &http.Client{Timeout: 30 * time.Second},
		queue:     q,
		queueURL:  queueURL,
	}
}

// DispatchDirect envia o comando sincronamente e aguarda a resposta na
// mesma chamada (EXE-14): "Orbita capture/consolida o que faltar,
// valida a projecao, decide o prazo e confirma seu final; responde
// pela conexao original".
func (d *Dispatcher) DispatchDirect(ctx context.Context, cmd dispatch.Command) (dispatch.Result, error) {
	body, err := json.Marshal(cmd)
	if err != nil {
		return dispatch.Result{}, fmt.Errorf("orbita: serializar comando direto: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.cometaURL+"/internal/commands/direct", bytes.NewReader(body))
	if err != nil {
		return dispatch.Result{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		// Timeout de rede entre Orbita e Cometa (COM-06): nao e prova
		// de que o comando deixou de executar. O chamador trata isso
		// como incerteza (UNKNOWN), nunca como novo aceite seguro.
		return dispatch.Result{CommandID: cmd.CommandID, Kind: dispatch.FactUnknown}, err
	}
	defer resp.Body.Close()

	var result dispatch.Result
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return dispatch.Result{CommandID: cmd.CommandID, Kind: dispatch.FactUnknown}, err
	}
	return result, nil
}

// DispatchQueued publica o mesmo conteudo logico do comando na fila
// dedicada da celula (COM-06: "para ASYNC, o mesmo conteudo logico vai
// a fila apos commit de sua intencao").
func (d *Dispatcher) DispatchQueued(ctx context.Context, cmd dispatch.Command) error {
	body, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("orbita: serializar comando em fila: %w", err)
	}
	return d.queue.SendCommand(ctx, d.queueURL, queue.Envelope{
		EventID:          cmd.CommandID,
		Type:             "command.dispatch",
		SchemaVersion:    1,
		Producer:         "orbita",
		TenantID:         cmd.TenantID,
		ProtocolID:       cmd.ProtocolID,
		OccurredAt:       time.Now().UTC(),
		RecordedAt:       time.Now().UTC(),
		AggregateVersion: 1,
		Payload:          body,
	})
}
