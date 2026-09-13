// Package pg abre e configura o pool de conexoes PostgreSQL de cada
// aplicacao. Cada dominio (hub_control, hub_core, hub_finance) tem sua
// propria conexao/usuario, conforme DAD-01: nenhuma escrita ou join
// cruza dominios, mesmo compartilhando servidor em local/dev.
package pg

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// RuntimeDSN fixa o tenant no nível da sessão PostgreSQL. É usado por
// workloads dedicados a um único tenant; serviços multi-tenant devem abrir
// transações com contexto equivalente por request antes de adotar a role.
func RuntimeDSN(dsn, tenant string) (string, error) {
	if strings.TrimSpace(tenant) == "" || strings.ContainsAny(tenant, "\x00\r\n") {
		return "", fmt.Errorf("pg: tenant runtime inválido")
	}
	u, err := url.Parse(dsn)
	if err != nil || u.Scheme == "" {
		return "", fmt.Errorf("pg: DSN runtime inválido")
	}
	q := u.Query()
	q.Set("options", "-c app.tenant_id="+tenant)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// SetTenantScope aplica app.tenant_id LOCAL a uma transação já aberta pelo
// chamador. Use quando a transação é gerenciada manualmente (BeginTx/defer
// Rollback/Commit já existentes) e reestruturá-la em torno de um callback
// aumentaria o risco de alterar o fluxo; para uma transação nova e simples,
// prefira WithTenantTx.
func SetTenantScope(ctx context.Context, tx *sql.Tx, tenant string) error {
	if tx == nil || strings.TrimSpace(tenant) == "" || strings.ContainsAny(tenant, "\x00\r\n") {
		return fmt.Errorf("pg: transação runtime sem tenant válido")
	}
	if _, err := tx.ExecContext(ctx, `SELECT set_config('app.tenant_id',$1,true)`, tenant); err != nil {
		return fmt.Errorf("pg: aplicar tenant runtime: %w", err)
	}
	return nil
}

// SetAuditedScope aplica app.access_reason LOCAL a uma transação já aberta
// (ver WithAuditedScopeTx para a variante de transação nova).
func SetAuditedScope(ctx context.Context, tx *sql.Tx, reason string) error {
	if tx == nil || strings.TrimSpace(reason) == "" || strings.ContainsAny(reason, "\x00\r\n") {
		return fmt.Errorf("pg: acesso cruzado sem motivo auditavel valido")
	}
	if _, err := tx.ExecContext(ctx, `SELECT set_config('app.access_reason',$1,true)`, reason); err != nil {
		return fmt.Errorf("pg: aplicar motivo de acesso cruzado: %w", err)
	}
	return nil
}

// SetWorkerCellScope aplica app.worker_cell_id LOCAL a uma transação já
// aberta (ver WithWorkerCellTx para a variante de transação nova).
func SetWorkerCellScope(ctx context.Context, tx *sql.Tx, cell string) error {
	if tx == nil || strings.TrimSpace(cell) == "" || strings.ContainsAny(cell, "\x00\r\n") {
		return fmt.Errorf("pg: transação de worker sem célula válida")
	}
	if _, err := tx.ExecContext(ctx, `SELECT set_config('app.worker_cell_id',$1,true)`, cell); err != nil {
		return fmt.Errorf("pg: aplicar célula do worker: %w", err)
	}
	return nil
}

// WithTenantTx inicia uma transação cujo tenant é LOCAL à transação. O
// contexto não fica preso na conexão do pool e o callback só pode confirmar
// depois que o escopo RLS foi aplicado. Repositórios multi-tenant devem usar
// este primitivo em vez de SET global ou de confiar apenas em WHERE.
func WithTenantTx(ctx context.Context, db *sql.DB, tenant string, fn func(*sql.Tx) error) error {
	if db == nil {
		return fmt.Errorf("pg: transação runtime sem tenant válido")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = SetTenantScope(ctx, tx, tenant); err != nil {
		return err
	}
	if fn == nil {
		return fmt.Errorf("pg: callback da transação runtime ausente")
	}
	if err = fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

// WithAuditedScopeTx inicia uma transacao com escopo cruzado de tenant,
// autorizado apenas quando reason for uma justificativa nao vazia (R6-SEG-01:
// "workers globais e administrador nominal precisam claims limitados e
// auditados, nao acesso global implicito"). O motivo fica LOCAL a transacao,
// nunca preso na conexao do pool, e deve ser persistido pelo chamador na
// mesma transacao (ex.: protocol_access_audit) antes do commit. Sem reason,
// nenhuma linha de outro tenant fica visivel: a policy audited_scope exige
// current_setting('app.access_reason') != ”.
func WithAuditedScopeTx(ctx context.Context, db *sql.DB, reason string, fn func(*sql.Tx) error) error {
	if db == nil {
		return fmt.Errorf("pg: acesso cruzado sem motivo auditavel valido")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = SetAuditedScope(ctx, tx, reason); err != nil {
		return err
	}
	if fn == nil {
		return fmt.Errorf("pg: callback da transação de acesso cruzado ausente")
	}
	if err = fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

// WithWorkerCellTx inicia uma transacao escopada a uma unica celula
// operacional (app.worker_cell_id, LOCAL a transacao). Usado por workers
// globais que processam fila entre tenants por natureza (claim/complete de
// intents, varredura de deadlines): a celula e uma autoridade operacional
// especifica, ja particionada pela topologia (D-05/EXE-*), nao um tenant
// arbitrario. Cobre leitura e escrita, ao contrario de WithAuditedScopeTx
// (somente leitura, para diagnostico administrativo).
func WithWorkerCellTx(ctx context.Context, db *sql.DB, cell string, fn func(*sql.Tx) error) error {
	if db == nil {
		return fmt.Errorf("pg: transação de worker sem célula válida")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = SetWorkerCellScope(ctx, tx, cell); err != nil {
		return err
	}
	if fn == nil {
		return fmt.Errorf("pg: callback da transação de worker ausente")
	}
	if err = fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

// Config controla o orcamento de conexoes do pool, conforme OPE-02:
// concorrencia maxima por pod/servico deve caber no orcamento do banco.
type Config struct {
	DSN            string
	MaxOpenConns   int
	MaxIdleConns   int
	ConnMaxIdle    time.Duration
	ConnectTimeout time.Duration
}

// Open conecta ao PostgreSQL e valida a conexao com um ping limitado
// por ConnectTimeout, evitando transacao aberta em espera indefinida
// (EXE-13).
func Open(cfg Config) (*sql.DB, error) {
	if cfg.MaxOpenConns <= 0 {
		cfg.MaxOpenConns = 10
	}
	if cfg.MaxIdleConns <= 0 {
		cfg.MaxIdleConns = 5
	}
	if cfg.ConnMaxIdle <= 0 {
		cfg.ConnMaxIdle = 5 * time.Minute
	}
	if cfg.ConnectTimeout <= 0 {
		cfg.ConnectTimeout = 5 * time.Second
	}

	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("pg: abrir conexao: %w", err)
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdle)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("pg: ping: %w", err)
	}
	return db, nil
}

// WaitReady tenta conectar repetidamente ate o prazo informado — usado
// no bootstrap local/dev, onde o container do Postgres pode ainda
// estar subindo quando o servico inicia.
func WaitReady(dsn string, timeout time.Duration) (*sql.DB, error) {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		db, err := Open(Config{DSN: dsn, ConnectTimeout: 2 * time.Second})
		if err == nil {
			return db, nil
		}
		lastErr = err
		time.Sleep(500 * time.Millisecond)
	}
	return nil, fmt.Errorf("pg: banco nao ficou pronto em %s: %w", timeout, lastErr)
}
