package cometa

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

var ErrCapacityDenied = errors.New("cometa: aggregate capacity unavailable")
var ErrCapacityFence = errors.New("cometa: capacity permit fenced or expired")
var ErrCapacityPolicy = errors.New("cometa: capacity policy missing, invalid or changed")

// CapacityPolicy is a qualified, immutable projection, never a caller supplied
// quota. Domain excludes cell and secret version. TenantLimits reserve isolation
// by bounding each tenant; their sum cannot exceed the submission floor.
// ValidUntil and EvidenceRef do not by themselves constitute homologation:
// the caller must obtain this projection from the authorized catalog.
type CapacityPolicy struct {
	Domain                 string         `json:"domain"`
	Version                string         `json:"version"`
	EvidenceRef            string         `json:"evidence_ref"`
	ValidUntil             time.Time      `json:"valid_until"`
	MaxConcurrent          int            `json:"max_concurrent"`
	MinConcurrent          int            `json:"min_concurrent"`
	ReconciliationReserve  int            `json:"reconciliation_reserve"`
	MaxPending             int            `json:"max_pending"`
	RatePerWindow          int            `json:"rate_per_window"`
	WindowMillis           int64          `json:"window_millis"`
	LeaseMillis            int64          `json:"lease_millis"`
	StableMillis           int64          `json:"stable_millis"`
	LatencyThresholdMillis int64          `json:"latency_threshold_millis"`
	TenantLimits           map[string]int `json:"tenant_limits"`
	TenantPendingLimits    map[string]int `json:"tenant_pending_limits"`
	TenantRateLimits       map[string]int `json:"tenant_rate_limits"`
}

func (p CapacityPolicy) validate() error {
	if p.Domain == "" || p.Version == "" || p.EvidenceRef == "" || p.ValidUntil.IsZero() || p.MaxConcurrent < 2 || p.MinConcurrent < 2 || p.MinConcurrent > p.MaxConcurrent || p.ReconciliationReserve < 1 || p.ReconciliationReserve >= p.MinConcurrent || p.MaxPending < 1 || p.RatePerWindow < 2 || p.WindowMillis < 1 || p.LeaseMillis < 1 || p.StableMillis < 1 || p.LatencyThresholdMillis < 1 || len(p.TenantLimits) == 0 {
		return ErrCapacityPolicy
	}
	total, totalPending, totalRate := 0, 0, 0
	for tenant, limit := range p.TenantLimits {
		if tenant == "" || limit < 1 || p.TenantPendingLimits[tenant] < 1 || p.TenantRateLimits[tenant] < 1 {
			return ErrCapacityPolicy
		}
		total += limit
		totalPending += p.TenantPendingLimits[tenant]
		totalRate += p.TenantRateLimits[tenant]
	}
	if total > p.MinConcurrent-p.ReconciliationReserve || totalPending > p.MaxPending || totalRate > p.RatePerWindow-1 || len(p.TenantPendingLimits) != len(p.TenantLimits) || len(p.TenantRateLimits) != len(p.TenantLimits) {
		return ErrCapacityPolicy
	}
	return nil
}

type CapacityController struct{ db *sql.DB }

func NewCapacityController(db *sql.DB) *CapacityController { return &CapacityController{db: db} }

// InstallPoliciesFromEnv carrega somente políticas previamente aprovadas pelo
// operador. O manifesto não contém limites comerciais implícitos: quando a
// configuração não existe, nenhum domínio é criado e a chamada que escolher
// um domínio falha fechada em Acquire.
func InstallPoliciesFromEnv(ctx context.Context, c *CapacityController) error {
	raw := strings.TrimSpace(os.Getenv("CAPACITY_POLICY_JSON"))
	domains := strings.Split(strings.TrimSpace(os.Getenv("CAPACITY_DOMAINS")), ",")
	if raw == "" && (len(domains) == 0 || domains[0] == "") {
		return nil
	}
	if c == nil || raw == "" || len(raw) > 64*1024 {
		return ErrCapacityPolicy
	}
	var policy CapacityPolicy
	if json.Unmarshal([]byte(raw), &policy) != nil {
		return ErrCapacityPolicy
	}
	seen := map[string]bool{}
	for _, domain := range domains {
		domain = strings.TrimSpace(domain)
		if domain == "" || seen[domain] {
			continue
		}
		seen[domain] = true
		policy.Domain = domain
		if err := c.InstallPolicy(ctx, policy); err != nil {
			return fmt.Errorf("install capacity policy %s: %w", domain, err)
		}
	}
	if len(seen) == 0 {
		return ErrCapacityPolicy
	}
	return nil
}

// InstallPolicy only accepts a new domain or the identical frozen projection.
// Changing a budget requires an explicit migration/reconciliation workflow;
// publishing another secret/account version never replenishes this budget.
func (c *CapacityController) InstallPolicy(ctx context.Context, p CapacityPolicy) error {
	if err := p.validate(); err != nil {
		return err
	}
	b, _ := json.Marshal(p)
	h := sha256.Sum256(b)
	hash := hex.EncodeToString(h[:])
	_, err := c.db.ExecContext(ctx, `INSERT INTO capacity_domains(domain_id,policy,policy_hash,current_limit) VALUES($1,$2,$3,$4) ON CONFLICT DO NOTHING`, p.Domain, b, hash, p.MaxConcurrent)
	if err != nil {
		return err
	}
	var existing string
	err = c.db.QueryRowContext(ctx, `SELECT policy_hash FROM capacity_domains WHERE domain_id=$1`, p.Domain).Scan(&existing)
	if err == nil && existing != hash {
		return ErrCapacityPolicy
	}
	return err
}

type CapacityPermit struct {
	Domain, ID, Tenant, Cell, Owner, Action string
	Epoch                                   int64
	LeaseUntil                              time.Time
}
type CapacityState struct {
	Limit, TransportOpen, PendingExternal, RateUsed int
	LastFeedback                                    string
}

const capacitySafetyMargin = time.Second

// effectiveHTTPBudget keeps a transport inside both the configured operation
// budget and the currently owned capacity lease. A lease is never renewed as a
// side effect of an HTTP call; when no positive window remains, the caller must
// release the permit before starting external I/O.
func effectiveHTTPBudget(ctx context.Context, requested time.Duration, permit CapacityPermit, enabled bool) time.Duration {
	if requested <= 0 {
		requested = 15 * time.Second
	}
	if enabled {
		remaining := time.Until(permit.LeaseUntil) - capacitySafetyMargin
		if remaining <= 0 {
			return 0
		}
		if remaining < requested {
			requested = remaining
		}
	}
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return 0
		}
		if remaining < requested {
			requested = remaining
		}
	}
	if requested < time.Millisecond {
		return 0
	}
	return requested
}

func (c *CapacityController) Acquire(ctx context.Context, domain, id, tenant, cell, owner, action string) (CapacityPermit, error) {
	permit := CapacityPermit{Domain: domain, ID: id, Tenant: tenant, Cell: cell, Owner: owner, Action: action}
	if id == "" || tenant == "" || cell == "" || owner == "" || (action != "SUBMIT" && action != "STATUS" && action != "FETCH" && action != "CANCEL") {
		return permit, ErrCapacityPolicy
	}
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return permit, err
	}
	defer tx.Rollback()
	var raw []byte
	var limit, used int
	var epoch int64
	var started, now time.Time
	err = tx.QueryRowContext(ctx, `SELECT policy,current_limit,epoch,rate_started_at,rate_used FROM capacity_domains WHERE domain_id=$1 FOR UPDATE`, domain).Scan(&raw, &limit, &epoch, &started, &used)
	if errors.Is(err, sql.ErrNoRows) {
		return permit, ErrCapacityPolicy
	}
	if err != nil {
		return permit, err
	}
	if err = tx.QueryRowContext(ctx, `SELECT clock_timestamp()`).Scan(&now); err != nil {
		return permit, err
	}
	var p CapacityPolicy
	if json.Unmarshal(raw, &p) != nil || p.validate() != nil || !now.Before(p.ValidUntil) {
		return permit, ErrCapacityPolicy
	}
	tenantLimit, ok := p.TenantLimits[tenant]
	if !ok {
		return permit, ErrCapacityDenied
	}
	var open, pending, tenantOpen, submitted, tenantPending, tenantRate, tenantTransports int
	started = now.Add(-time.Duration(p.WindowMillis) * time.Millisecond)
	err = tx.QueryRowContext(ctx, `SELECT count(*) FILTER(WHERE transport_open),count(*) FILTER(WHERE pending_external),count(*) FILTER(WHERE transport_open AND tenant_id=$2 AND action='SUBMIT'),count(*) FILTER(WHERE transport_open AND action='SUBMIT'),count(*) FILTER(WHERE pending_external AND tenant_id=$2),count(*) FILTER(WHERE tenant_id=$2 AND created_at >= $3),count(*) FILTER(WHERE transport_open AND tenant_id=$2),count(*) FILTER(WHERE created_at >= $3) FROM capacity_permits WHERE domain_id=$1`, domain, tenant, started).Scan(&open, &pending, &tenantOpen, &submitted, &tenantPending, &tenantRate, &tenantTransports, &used)
	if err != nil {
		return permit, err
	}

	// Reserve one rate credit for reconciliation as well as transport slots.
	rateLimit := p.RatePerWindow
	if action == "SUBMIT" {
		rateLimit--
	}
	if open >= limit || used >= rateLimit || tenantRate >= p.TenantRateLimits[tenant] || tenantTransports >= tenantLimit+p.ReconciliationReserve || (action == "SUBMIT" && (submitted >= limit-p.ReconciliationReserve || tenantOpen >= tenantLimit || pending >= p.MaxPending || tenantPending >= p.TenantPendingLimits[tenant])) {
		return permit, ErrCapacityDenied
	}
	permit.Epoch = epoch + 1
	permit.LeaseUntil = now.Add(time.Duration(p.LeaseMillis) * time.Millisecond)
	result, err := tx.ExecContext(ctx, `INSERT INTO capacity_permits(domain_id,permit_id,tenant_id,cell_id,owner_id,action,epoch,lease_until,pending_external) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT DO NOTHING`, domain, id, tenant, cell, owner, action, permit.Epoch, permit.LeaseUntil, action == "SUBMIT")
	if err != nil {
		return permit, err
	}
	n, _ := result.RowsAffected()
	if n != 1 {
		return permit, ErrCapacityFence
	}
	_, err = tx.ExecContext(ctx, `UPDATE capacity_domains SET epoch=$2,rate_started_at=$3,rate_used=$4,updated_at=clock_timestamp() WHERE domain_id=$1`, domain, permit.Epoch, started, used+1)
	if err != nil {
		return permit, err
	}
	return permit, tx.Commit()
}

// Validate must run immediately before transport. There is no offline permit
// acquisition and no automatic expiry recycling. Operation submission still
// requires its separate durable operation fence before any external effect.
func (c *CapacityController) Validate(ctx context.Context, p CapacityPermit) error {
	var ok bool
	err := c.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM capacity_permits WHERE domain_id=$1 AND permit_id=$2 AND owner_id=$3 AND epoch=$4 AND tenant_id=$5 AND cell_id=$6 AND action=$7 AND transport_open AND lease_until>clock_timestamp())`, p.Domain, p.ID, p.Owner, p.Epoch, p.Tenant, p.Cell, p.Action).Scan(&ok)
	if err != nil {
		return err
	}
	if !ok {
		return ErrCapacityFence
	}
	return nil
}

// CompleteTransport requires receipt evidence. UNKNOWN keeps the separate
// pending_external obligation. An expired lease cannot release capacity here;
// reconciliation must provide evidence through ResolvePending instead.
func (c *CapacityController) CompleteTransport(ctx context.Context, p CapacityPermit, signal string, latency time.Duration, externalPending bool, evidence string) error {
	if evidence == "" || latency < 0 || (signal != "SUCCESS" && signal != "TIMEOUT" && signal != "THROTTLED" && signal != "UNAVAILABLE") {
		return ErrCapacityPolicy
	}
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var raw []byte
	var limit int
	var stable, now time.Time
	err = tx.QueryRowContext(ctx, `SELECT policy,current_limit,stable_since FROM capacity_domains WHERE domain_id=$1 FOR UPDATE`, p.Domain).Scan(&raw, &limit, &stable)
	if err != nil {
		return err
	}
	if err = tx.QueryRowContext(ctx, `SELECT clock_timestamp()`).Scan(&now); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `UPDATE capacity_permits SET transport_open=false,pending_external=$8,completed_at=clock_timestamp(),evidence_ref=$9 WHERE domain_id=$1 AND permit_id=$2 AND owner_id=$3 AND epoch=$4 AND tenant_id=$5 AND cell_id=$6 AND action=$7 AND transport_open AND lease_until>clock_timestamp()`, p.Domain, p.ID, p.Owner, p.Epoch, p.Tenant, p.Cell, p.Action, externalPending, evidence)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n != 1 {
		return ErrCapacityFence
	}
	var policy CapacityPolicy
	if err = json.Unmarshal(raw, &policy); err != nil {
		return err
	}
	reason := signal
	if signal != "SUCCESS" || latency.Milliseconds() > policy.LatencyThresholdMillis {
		if signal == "SUCCESS" {
			reason = "LATENCY"
		}
		limit = max(policy.MinConcurrent, limit/2)
		stable = now
	} else if now.Sub(stable) >= time.Duration(policy.StableMillis)*time.Millisecond {
		limit = min(policy.MaxConcurrent, limit+1)
		stable = now
	}
	_, err = tx.ExecContext(ctx, `UPDATE capacity_domains SET current_limit=$2,stable_since=$3,last_feedback=$4,updated_at=clock_timestamp() WHERE domain_id=$1`, p.Domain, limit, stable, reason)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO capacity_feedback(domain_id,permit_id,signal,observed_latency_ms,resulting_limit,evidence_ref) VALUES($1,$2,$3,$4,$5,$6)`, p.Domain, p.ID, reason, latency.Milliseconds(), limit, evidence)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// Release cancela uma concessão antes do transporte começar. É usado quando
// a custódia concorrente identifica uma operação já existente ou quando a
// preparação local falha; não altera o feedback adaptativo nem libera uma
// obrigação externa pendente.
func (c *CapacityController) Release(ctx context.Context, p CapacityPermit, evidence string) error {
	if evidence == "" {
		return ErrCapacityPolicy
	}
	result, err := c.db.ExecContext(ctx, `UPDATE capacity_permits
		SET transport_open=false,pending_external=false,completed_at=clock_timestamp(),evidence_ref=$5
		WHERE domain_id=$1 AND permit_id=$2 AND owner_id=$3 AND epoch=$4 AND transport_open`, p.Domain, p.ID, p.Owner, p.Epoch, evidence)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n != 1 {
		return ErrCapacityFence
	}
	return nil
}

// ResolvePending is for a separately authorized reconciler with positive proof
// that the external operation/transport ended. Time passing is never evidence.
// The original epoch identifies the obligation; it cannot authorize a new send.
func (c *CapacityController) ResolvePending(ctx context.Context, domain, id string, epoch int64, evidence string) error {
	if evidence == "" {
		return ErrCapacityPolicy
	}
	result, err := c.db.ExecContext(ctx, `UPDATE capacity_permits SET transport_open=false,pending_external=false,completed_at=clock_timestamp(),evidence_ref=$4 WHERE domain_id=$1 AND permit_id=$2 AND epoch=$3 AND (transport_open OR pending_external)`, domain, id, epoch, evidence)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n != 1 {
		return ErrCapacityFence
	}
	return nil
}

// ResolvePendingByID fecha a obrigação externa criada pela submissão quando a
// observação terminal chega por polling, callback ou reconciliação. O domínio
// vem do snapshot imutável da rota e o permit_id é o operation_id; não há
// criação de uma nova concessão nem reciclagem por expiração.
func (c *CapacityController) ResolvePendingByID(ctx context.Context, domain, id, evidence string) error {
	if domain == "" || id == "" || evidence == "" {
		return ErrCapacityPolicy
	}
	result, err := c.db.ExecContext(ctx, `UPDATE capacity_permits SET transport_open=false,pending_external=false,completed_at=clock_timestamp(),evidence_ref=$3 WHERE domain_id=$1 AND permit_id=$2 AND pending_external`, domain, id, evidence)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n != 1 {
		return ErrCapacityFence
	}
	return nil
}
func (c *CapacityController) State(ctx context.Context, domain string) (CapacityState, error) {
	var s CapacityState
	err := c.db.QueryRowContext(ctx, `SELECT d.current_limit,d.rate_used,d.last_feedback,(SELECT count(*) FROM capacity_permits p WHERE p.domain_id=d.domain_id AND transport_open),(SELECT count(*) FROM capacity_permits p WHERE p.domain_id=d.domain_id AND pending_external) FROM capacity_domains d WHERE domain_id=$1`, domain).Scan(&s.Limit, &s.RateUsed, &s.LastFeedback, &s.TransportOpen, &s.PendingExternal)
	return s, err
}
func (p CapacityPermit) String() string {
	return fmt.Sprintf("capacity permit domain=%s epoch=%d", p.Domain, p.Epoch)
}
