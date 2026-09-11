// Package orbita implementa admissao duravel, a maquina de estados do
// protocolo, o despacho direto/em fila e a consulta unificada
// (EXE-01/02/03/07/11/14/15/16, COM-05/06).
package orbita

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// ErrIdempotencyConflict e retornado quando a mesma Idempotency-Key e
// reenviada com um hash de pedido diferente (EXE-01).
var ErrIdempotencyConflict = errors.New("orbita: idempotency-key reutilizada com payload diferente")

// ErrProtocolNotFound e retornado quando o protocolo nao existe ou
// pertence a outro tenant (EXE-16: GET valida o tenant).
var ErrProtocolNotFound = errors.New("orbita: protocolo nao encontrado")

var ErrResultLate = errors.New("orbita: result reached authority after deadline")

// Status e o estado normativo do protocolo (EXE-03).
type Status string

const (
	StatusAccepted           Status = "ACCEPTED"
	StatusRunning            Status = "RUNNING"
	StatusWaitingProvider    Status = "WAITING_PROVIDER"
	StatusReconciling        Status = "RECONCILING"
	StatusSucceeded          Status = "SUCCEEDED"
	StatusPartiallySucceeded Status = "PARTIALLY_SUCCEEDED"
	StatusFailed             Status = "FAILED"
	StatusExpired            Status = "EXPIRED"
	StatusCancelled          Status = "CANCELLED"
)

var terminalStatuses = map[Status]bool{
	StatusSucceeded: true, StatusPartiallySucceeded: true, StatusFailed: true,
	StatusExpired: true, StatusCancelled: true,
}

// IsTerminal indica se o status encerra o protocolo (EXE-03/EXE-11).
func IsTerminal(s Status) bool { return terminalStatuses[s] }

// Protocol e a projecao persistida do protocolo (DAD-02).
type Protocol struct {
	ApplicationID       string
	CellID              string
	ProtocolID          string
	TenantID            string
	IdempotencyKey      string
	RequestHash         string
	RequestBody         json.RawMessage
	Mode                string
	DispatchMode        string
	CommandID           string
	Status              Status
	ResultVersion       int
	FinalBody           json.RawMessage
	FinalRepresentation []byte
	FinalEventID        sql.NullString
	TerminalReason      sql.NullString
	AcceptedAt          time.Time
	ClientDeadlineAt    time.Time
	FinalizedAt         sql.NullTime
	Version             int
}

// Store encapsula o acesso a tabela protocols em hub_core.
type Store struct {
	db *sql.DB
}

// NewStore cria um Store a partir de uma conexao ja aberta com hub_core.
func NewStore(db *sql.DB) *Store { return &Store{db: db} }

// DB expoe a conexao subjacente para o relay de outbox.
func (s *Store) DB() *sql.DB { return s.db }

// FindByIdempotencyKey busca um protocolo existente pela chave de
// idempotencia do tenant (EXE-01): repeticao com o mesmo hash recupera
// o mesmo protocolo; hash diferente e conflito.
func (s *Store) FindByIdempotencyKey(ctx context.Context, tenantID, key, requestHash string) (Protocol, error) {
	p, err := s.getByIdempotencyKey(ctx, tenantID, key)
	if errors.Is(err, sql.ErrNoRows) {
		return Protocol{}, sql.ErrNoRows
	}
	if err != nil {
		return Protocol{}, err
	}
	if p.RequestHash != requestHash {
		return Protocol{}, ErrIdempotencyConflict
	}
	return p, nil
}

func (s *Store) getByIdempotencyKey(ctx context.Context, tenantID, key string) (Protocol, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT protocol_id, tenant_id, idempotency_key, request_hash, request_body, mode, dispatch_mode,
		       command_id, status, result_version, final_body, final_event_id, terminal_reason, application_id, cell_id, final_representation,
		       accepted_at, client_deadline_at, finalized_at, version
		FROM protocols WHERE tenant_id = $1 AND idempotency_key = $2 AND application_id = $3
	`, tenantID, key, applicationFromContext(ctx))
	return scanProtocol(row)
}

// Create persiste um novo protocolo atomicamente (EXE-01, DAD-03):
// UUIDv7, hash/idempotencia, pedido canonico, versoes aplicaveis e
// intencao de despacho, antes de qualquer efeito externo.
func (s *Store) Create(ctx context.Context, p Protocol) error {
	body, err := json.Marshal(p.RequestBody)
	if err != nil {
		return fmt.Errorf("orbita: serializar pedido: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO protocols (protocol_id, tenant_id, idempotency_key, request_hash, request_body, mode,
		                        dispatch_mode, command_id, status, accepted_at, client_deadline_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, p.ProtocolID, p.TenantID, p.IdempotencyKey, p.RequestHash, body, p.Mode, p.DispatchMode,
		p.CommandID, string(p.Status), p.AcceptedAt, p.ClientDeadlineAt)
	if err != nil {
		return fmt.Errorf("orbita: criar protocolo: %w", err)
	}
	return nil
}

// Get le um protocolo validando o tenant (EXE-16: "nenhuma credencial
// de cliente consulta outro tenant por conhecer o UUID").
func (s *Store) Get(ctx context.Context, tenantID, protocolID string) (Protocol, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT protocol_id, tenant_id, idempotency_key, request_hash, request_body, mode, dispatch_mode,
		       command_id, status, result_version, final_body, final_event_id, terminal_reason, application_id, cell_id, final_representation,
		       accepted_at, client_deadline_at, finalized_at, version
		FROM protocols WHERE protocol_id = $1 AND tenant_id = $2
	`, protocolID, tenantID)
	p, err := scanProtocol(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Protocol{}, ErrProtocolNotFound
	}
	return p, err
}

// GetByID le um protocolo por ID, sem verificar tenant — usado apenas
// internamente pelo consumidor de fatos (que ja confia na origem
// interna do evento) e pelo perfil administrativo auditado.
func (s *Store) GetByID(ctx context.Context, protocolID string) (Protocol, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT protocol_id, tenant_id, idempotency_key, request_hash, request_body, mode, dispatch_mode,
		       command_id, status, result_version, final_body, final_event_id, terminal_reason, application_id, cell_id, final_representation,
		       accepted_at, client_deadline_at, finalized_at, version
		FROM protocols WHERE protocol_id = $1
	`, protocolID)
	p, err := scanProtocol(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Protocol{}, ErrProtocolNotFound
	}
	return p, err
}

func scanProtocol(row *sql.Row) (Protocol, error) {
	var p Protocol
	var status string
	var finalBody []byte
	err := row.Scan(&p.ProtocolID, &p.TenantID, &p.IdempotencyKey, &p.RequestHash, &p.RequestBody, &p.Mode,
		&p.DispatchMode, &p.CommandID, &status, &p.ResultVersion, &finalBody, &p.FinalEventID, &p.TerminalReason,
		&p.ApplicationID, &p.CellID, &p.FinalRepresentation, &p.AcceptedAt, &p.ClientDeadlineAt, &p.FinalizedAt, &p.Version)
	if err != nil {
		return Protocol{}, err
	}
	p.Status = Status(status)
	if finalBody != nil {
		p.FinalBody = json.RawMessage(finalBody)
	}
	return p, nil
}

// FinalizeParams agrupa os campos gravados na conclusao (DAD-03: "na
// conclusao, preparar e validar a representacao do cliente e suas
// referencias; persistir atomicamente estado, versao final, decisao
// de prazo, historico e evento final").
type FinalizeParams struct {
	ProtocolID      string
	ExpectedVersion int
	Status          Status
	FinalBody       any
	TerminalReason  string
	FinalEventID    string
}

// Finalize aplica uma unica transicao terminal serializavel (EXE-11):
// so ha sucesso se ExpectedVersion ainda bater (protocolo nao foi
// finalizado por outro caminho, ex.: temporizador de deadline
// concorrente). Publica o fato "protocol.finalized" na mesma
// transacao (COM-03).
func (s *Store) Finalize(ctx context.Context, p FinalizeParams, publish func(tx *sqlTx) error) (bool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var mode string
	var deadline, accepted, observed time.Time
	var snapshotRaw []byte
	if err = tx.QueryRowContext(ctx, "SELECT mode,client_deadline_at,accepted_at,config_snapshot FROM protocols WHERE protocol_id=$1 FOR UPDATE", p.ProtocolID).Scan(&mode, &deadline, &accepted, &snapshotRaw); err != nil {
		return false, err
	}
	if err = tx.QueryRowContext(ctx, "SELECT clock_timestamp()").Scan(&observed); err != nil {
		return false, err
	}
	if p.Status == StatusSucceeded || p.Status == StatusPartiallySucceeded {
		var snapshot struct {
			Target struct {
				Data json.RawMessage `json:"data"`
			} `json:"target"`
		}
		var target struct {
			ProviderSLASeconds int    `json:"provider_sla_seconds"`
			ProviderSLAPolicy  string `json:"provider_sla_policy"`
		}
		if json.Unmarshal(snapshotRaw, &snapshot) == nil {
			_ = json.Unmarshal(snapshot.Target.Data, &target)
		}
		policy := strings.ToUpper(strings.TrimSpace(target.ProviderSLAPolicy))
		if policy == "REJECT_LATE" && target.ProviderSLASeconds > 0 && !observed.Before(accepted.Add(time.Duration(target.ProviderSLASeconds)*time.Second)) {
			p.Status = StatusFailed
			p.TerminalReason = "REJECTED_PROVIDER_SLA"
			p.FinalBody = providerSLAFailureBody(p.ProtocolID)
		}
	}
	body, err := json.Marshal(p.FinalBody)
	if err != nil {
		return false, fmt.Errorf("orbita: serializar corpo final: %w", err)
	}
	// This rejects already late results independently of the timer. It is NOT
	// proof that a subsequent commit cannot cross the deadline (T-R2-01).
	if (p.Status == StatusSucceeded || p.Status == StatusPartiallySucceeded) && !observed.Before(deadline) {
		return false, ErrResultLate
	}
	representation, err := json.Marshal(struct {
		ProtocolID    string          `json:"protocol_id"`
		Status        Status          `json:"status"`
		ResultVersion int             `json:"result_version"`
		Mode          string          `json:"mode"`
		FinalBody     json.RawMessage `json:"final_body"`
	}{p.ProtocolID, p.Status, 1, mode, body})
	if err != nil {
		return false, err
	}
	sum := sha256.Sum256(representation)

	res, err := tx.ExecContext(ctx, `
		UPDATE protocols
		SET status = $1, final_body = $2, terminal_reason = NULLIF($3, ''), final_event_id = $4,
		    final_representation=$7, final_sha256=$8, final_media_type='application/json', finalized_at = clock_timestamp(), result_version = result_version + 1, version = version + 1, updated_at = clock_timestamp()
		WHERE protocol_id = $5 AND version = $6 AND status NOT IN ('SUCCEEDED','PARTIALLY_SUCCEEDED','FAILED','EXPIRED','CANCELLED')
		  AND ($1 NOT IN ('SUCCEEDED','PARTIALLY_SUCCEEDED') OR clock_timestamp() < client_deadline_at)
	`, string(p.Status), body, p.TerminalReason, p.FinalEventID, p.ProtocolID, p.ExpectedVersion, representation, hex.EncodeToString(sum[:]))
	if err != nil {
		return false, fmt.Errorf("orbita: finalizar protocolo: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	if rows == 0 {
		// A condição de prazo pode perder a corrida entre a leitura acima e a
		// atualização. Diferencie esse caso de uma finalização concorrente para
		// que o chamador materialize EXPIRED, sem publicar sucesso tardio.
		if p.Status == StatusSucceeded || p.Status == StatusPartiallySucceeded {
			var currentStatus string
			var currentVersion int
			var currentDeadline, currentNow time.Time
			if scanErr := tx.QueryRowContext(ctx, "SELECT status,version,client_deadline_at,clock_timestamp() FROM protocols WHERE protocol_id=$1", p.ProtocolID).Scan(&currentStatus, &currentVersion, &currentDeadline, &currentNow); scanErr == nil && currentVersion == p.ExpectedVersion && currentStatus != string(StatusSucceeded) && currentStatus != string(StatusPartiallySucceeded) && currentStatus != string(StatusFailed) && currentStatus != string(StatusExpired) && currentStatus != string(StatusCancelled) && !currentNow.Before(currentDeadline) {
				return false, ErrResultLate
			}
		}
		// Ja finalizado por outro caminho (ex.: deadline concorrente) —
		// nao ha segunda transicao terminal (EXE-11).
		return false, nil
	}
	if publish != nil {
		if err := publish(&sqlTx{tx: tx, Representation: representation}); err != nil {
			return false, err
		}
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}

func providerSLAFailureBody(protocolID string) FinalBody {
	return FinalBody{
		ProtocolID:   protocolID,
		Status:       string(StatusFailed),
		ErrorCode:    "PROVIDER_SLA_EXCEEDED",
		ErrorMessage: "o prazo contratado do provedor foi excedido; o resultado não é elegível",
	}
}

// sqlTx encapsula *sql.Tx para o chamador de Finalize sem expor o
// pacote database/sql diretamente na assinatura publica (mantendo o
// pacote outbox como unico ponto de insercao na outbox).
type sqlTx struct {
	tx             *sql.Tx
	Representation []byte
}

// Tx expoe a transacao subjacente para o pacote outbox.
func (t *sqlTx) Tx() *sql.Tx { return t.tx }

// DueDeadlines retorna protocolos abertos cujo client_deadline_at ja
// foi atingido (EXE-11).
type OpenProtocol struct {
	ProtocolID string
	TenantID   string
	Version    int
}

func (s *Store) DueDeadlines(ctx context.Context, now time.Time, limit int) ([]OpenProtocol, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT protocol_id, tenant_id, version FROM protocols
		WHERE cell_id=$3 AND client_deadline_at <= $1 AND status NOT IN ('SUCCEEDED','PARTIALLY_SUCCEEDED','FAILED','EXPIRED','CANCELLED')
		ORDER BY client_deadline_at
		LIMIT $2
	`, now, limit, os.Getenv("CELL_ID"))
	if err != nil {
		return nil, fmt.Errorf("orbita: buscar deadlines vencidos: %w", err)
	}
	defer rows.Close()
	var out []OpenProtocol
	for rows.Next() {
		var o OpenProtocol
		if err := rows.Scan(&o.ProtocolID, &o.TenantID, &o.Version); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// Ping verifica a conectividade com hub_core (readiness).
func (s *Store) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }
