// Package dispatch define o contrato do despacho direto entre Orbita e
// Cometa (COM-06): o comando enviado pela Orbita e a resposta que
// Cometa devolve na mesma chamada HTTPS, em SYNC, ou publicada na fila
// (mesmo conteudo logico), em ASYNC/AUTO.
package dispatch

import "time"

// Mode e o modo de atendimento escolhido pelo cliente (EXE-02).
type Mode string

const (
	ModeSync  Mode = "SYNC"
	ModeAsync Mode = "ASYNC"
	ModeAuto  Mode = "AUTO"
)

// DispatchMode distingue a transicao de transporte do comando (EXE-15):
// DIRECT (SYNC, chamada HTTPS sincrona) ou QUEUED (ASYNC/AUTO, fila).
type DispatchMode string

const (
	DispatchDirect DispatchMode = "DIRECT"
	DispatchQueued DispatchMode = "QUEUED"
)

// Command e o comando de despacho (COM-06). Nunca inclui segredo em
// claro. Tem identidade estavel entre timeout, retry de transporte,
// evento de recuperacao e consulta interna.
type Command struct {
	TenantID          string       `json:"tenant_id"`
	ProtocolID        string       `json:"protocol_id"`
	StepID            string       `json:"step_id"`
	CommandID         string       `json:"command_id"`
	DispatchMode      DispatchMode `json:"dispatch_mode"`
	Epoch             int64        `json:"epoch"`
	ServiceCode       string       `json:"service_code"`
	ServiceVersion    int          `json:"service_version"`
	ProviderAccountID string       `json:"provider_account_id"`
	RequestBody       any          `json:"request_body"`
	// Deadline absoluto que o passo nao pode ultrapassar (EXE-14): nao
	// e prolongado por retry nem por reenvio de transporte.
	StepDeadline time.Time `json:"step_deadline"`
}

// FactKind classifica o que uma resposta direta pode representar
// (COM-06): apenas Succeeded/Failed permitem a Orbita finalizar sucesso
// no mesmo HTTP publico; Unknown e Rejected exigem tratamento distinto.
type FactKind string

const (
	FactSucceeded FactKind = "SUCCEEDED"
	FactFailed    FactKind = "FAILED"
	FactUnknown   FactKind = "UNKNOWN"  // operacao conhecida pendente/incerta
	FactRejected  FactKind = "REJECTED" // recusa anterior ao envio (validacao/credencial)
)

// Result e a resposta direta de Cometa (COM-06). Codigo 200 do
// adaptador sem fato conservado nao e final valido: por isso o
// resultado sempre carrega Kind e a evidencia de persistencia.
type Result struct {
	CommandID         string   `json:"command_id"`
	OperationID       string   `json:"operation_id"`
	ProviderRequestID string   `json:"provider_request_id,omitempty"`
	Kind              FactKind `json:"kind"`
	ResponseBody      any      `json:"response_body,omitempty"`
	ErrorCode         string   `json:"error_code,omitempty"`
	ErrorMessage      string   `json:"error_message,omitempty"`
}
