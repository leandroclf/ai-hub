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
