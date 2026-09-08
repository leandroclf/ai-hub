// Package cometa implementa a integracao com provedores externos:
// operacoes, tentativas, resolucao de credencial, polling e o
// controlador de despacho direto/fila (EXE-04, EXE-05, SEG-05).
package cometa

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"
)

// ErrOperationNotFound e retornado quando uma operacao/command_id nao
// e conhecida por este Cometa.
var ErrOperationNotFound = errors.New("cometa: operacao nao encontrada")

// State e o estado normativo da operacao externa (EXE-03).
type State string

const (
	StatePrepared         State = "PREPARED"
	StateSubmitting       State = "SUBMITTING"
	StateAcceptedExternal State = "ACCEPTED_EXTERNAL"
	StateWaitingFinal     State = "WAITING_FINAL"
	StateUnknown          State = "UNKNOWN"
	StateSucceeded        State = "SUCCEEDED"
	StateFailed           State = "FAILED"
	StateCancelled        State = "CANCELLED"
)

// Operation e a projecao persistida de uma operacao externa (DAD-02).
type Operation struct {
	OperationID         string
	ProtocolID          string
	ProviderAccountID   string
	ProviderRequestID   sql.NullString
	CredentialBindingID sql.NullString
	SecretVersionID     sql.NullString
	State               State
}

// Store encapsula o acesso as tabelas de dominio do Cometa em hub_core:
// operations, attempts e polling_schedule.
type Store struct {
	db *sql.DB
}

// NewStore cria um Store a partir de uma conexao ja aberta com hub_core.
func NewStore(db *sql.DB) *Store { return &Store{db: db} }

// CreateOperation persiste uma nova operacao externa em estado
// PREPARED, antes de qualquer efeito externo (DAD-09: "Preparacao
// Cometa" e uma fronteira duravel que antecede o envio).
func (s *Store) CreateOperation(ctx context.Context, op Operation) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO operations (operation_id, protocol_id, provider_account_id, credential_binding_id, state)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (operation_id) DO NOTHING
	`, op.OperationID, op.ProtocolID, op.ProviderAccountID, nullOrStr(op.CredentialBindingID), string(op.State))
	if err != nil {
		return fmt.Errorf("cometa: criar operacao: %w", err)
	}
	return nil
}

func nullOrStr(n sql.NullString) any {
	if n.Valid {
		return n.String
	}
	return nil
}

// UpdateState atualiza o estado e, quando conhecido, o
// provider_request_id da operacao.
func (s *Store) UpdateState(ctx context.Context, operationID string, state State, providerRequestID string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE operations
		SET state = $2, provider_request_id = COALESCE(NULLIF($3, ''), provider_request_id), updated_at = now()
		WHERE operation_id = $1
	`, operationID, string(state), providerRequestID)
	if err != nil {
		return fmt.Errorf("cometa: atualizar estado da operacao: %w", err)
	}
	return nil
}

// Get le uma operacao pelo ID — usada pela consulta interna de EXE-04
// ("consulta interna por command_id/operation_id le o estado de
// Cometa e nao consulta automaticamente o provedor").
func (s *Store) Get(ctx context.Context, operationID string) (Operation, error) {
	var op Operation
	var state string
	row := s.db.QueryRowContext(ctx, `
		SELECT operation_id, protocol_id, provider_account_id, provider_request_id, credential_binding_id, state
		FROM operations WHERE operation_id = $1
	`, operationID)
	err := row.Scan(&op.OperationID, &op.ProtocolID, &op.ProviderAccountID, &op.ProviderRequestID, &op.CredentialBindingID, &state)
	if errors.Is(err, sql.ErrNoRows) {
		return Operation{}, ErrOperationNotFound
	}
	if err != nil {
		return Operation{}, fmt.Errorf("cometa: ler operacao: %w", err)
	}
	op.State = State(state)
	return op, nil
}

// FindByProviderRequestID localiza a operacao correlacionada a um
// provider_request_id numa conta especifica (EXE-04: correlacao no
// escopo documentado do provedor/conta).
func (s *Store) FindByProviderRequestID(ctx context.Context, providerAccountID, providerRequestID string) (Operation, error) {
	var op Operation
	var state string
	row := s.db.QueryRowContext(ctx, `
		SELECT operation_id, protocol_id, provider_account_id, provider_request_id, credential_binding_id, state
		FROM operations WHERE provider_account_id = $1 AND provider_request_id = $2
	`, providerAccountID, providerRequestID)
	err := row.Scan(&op.OperationID, &op.ProtocolID, &op.ProviderAccountID, &op.ProviderRequestID, &op.CredentialBindingID, &state)
	if errors.Is(err, sql.ErrNoRows) {
		return Operation{}, ErrOperationNotFound
	}
	if err != nil {
		return Operation{}, fmt.Errorf("cometa: localizar operacao por provider_request_id: %w", err)
	}
	op.State = State(state)
	return op, nil
}

// RecordAttempt persiste uma tentativa fisica (DAD-02): nova tentativa
// nao cria nova operacao logica por si so.
func (s *Store) RecordAttempt(ctx context.Context, attemptID, operationID, attemptType string, sentAt, receivedAt *time.Time, errorCode string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO attempts (attempt_id, operation_id, attempt_type, sent_at, received_at, error_code)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''))
	`, attemptID, operationID, attemptType, sentAt, receivedAt, errorCode)
	if err != nil {
		return fmt.Errorf("cometa: registrar tentativa: %w", err)
	}
	return nil
}

// SchedulePolling cria/atualiza a agenda durável de polling de uma
// operacao (EXE-05, DAD-02).
func (s *Store) SchedulePolling(ctx context.Context, operationID string, nextRunAt, deadline time.Time, intervalSeconds int) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO polling_schedule (operation_id, next_run_at, interval_seconds, deadline_at, attempts_count)
		VALUES ($1, $2, $3, $4, 0)
		ON CONFLICT (operation_id) DO UPDATE SET
			next_run_at = EXCLUDED.next_run_at,
			interval_seconds = EXCLUDED.interval_seconds,
			deadline_at = EXCLUDED.deadline_at
	`, operationID, nextRunAt, intervalSeconds, deadline)
	if err != nil {
		return fmt.Errorf("cometa: agendar polling: %w", err)
	}
	return nil
}

// DuePolling retorna operacoes cuja proxima consulta ja venceu e o
// prazo de polling ainda nao expirou.
type DuePoll struct {
	OperationID        string
	ProtocolID         string
	ProviderAccountID  string
	ProviderRequestID  string
	IntervalSeconds    int
	MaxIntervalSeconds int
	AttemptsCount      int
}

func (s *Store) DuePolling(ctx context.Context, now time.Time, limit int) ([]DuePoll, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT ps.operation_id, o.protocol_id, o.provider_account_id, COALESCE(o.provider_request_id, ''),
		       ps.interval_seconds, ps.max_interval_seconds, ps.attempts_count
		FROM polling_schedule ps
		JOIN operations o ON o.operation_id = ps.operation_id
		WHERE o.cell_id=$3 AND ps.next_run_at <= $1 AND ps.deadline_at > $1 AND o.state = 'ACCEPTED_EXTERNAL'
		ORDER BY ps.next_run_at
		LIMIT $2
	`, now, limit, os.Getenv("CELL_ID"))
	if err != nil {
		return nil, fmt.Errorf("cometa: buscar polling devido: %w", err)
	}
	defer rows.Close()
	var out []DuePoll
	for rows.Next() {
		var d DuePoll
		if err := rows.Scan(&d.OperationID, &d.ProtocolID, &d.ProviderAccountID, &d.ProviderRequestID, &d.IntervalSeconds, &d.MaxIntervalSeconds, &d.AttemptsCount); err != nil {
			return nil, fmt.Errorf("cometa: ler linha de polling devido: %w", err)
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// AdvancePolling reagenda a proxima consulta com backoff simples,
// preservando a agenda existente (EXE-05: um worker reclama trabalho
// de forma exclusiva; aqui, simplificado, sem lease multi-instancia).
func (s *Store) AdvancePolling(ctx context.Context, operationID string, nextRunAt time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE polling_schedule
		SET next_run_at = $2, attempts_count = attempts_count + 1
		WHERE operation_id = $1
	`, operationID, nextRunAt)
	if err != nil {
		return fmt.Errorf("cometa: avancar polling: %w", err)
	}
	return nil
}

// ClearPolling remove a agenda de polling de uma operacao concluida
// (EXE-05: final persistido encerra agendamentos produtivos futuros).
func (s *Store) ClearPolling(ctx context.Context, operationID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM polling_schedule WHERE operation_id = $1`, operationID)
	if err != nil {
		return fmt.Errorf("cometa: limpar polling: %w", err)
	}
	return nil
}

// Ping verifica a conectividade com hub_core (readiness).
func (s *Store) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }

// DB expoe a conexao subjacente para o relay de outbox e transacoes
// que o executor precisa controlar diretamente.
func (s *Store) DB() *sql.DB { return s.db }
