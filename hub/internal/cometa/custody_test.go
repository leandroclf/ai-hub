package cometa

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"ai-hub/hub/internal/atlas"
	"ai-hub/hub/internal/atlasclient"
	"ai-hub/hub/internal/callbackauth"
	"ai-hub/hub/internal/dispatch"
	"ai-hub/hub/internal/platform/idgen"
	"ai-hub/hub/internal/platform/pg"
	"ai-hub/hub/internal/providerauth"
	"ai-hub/hub/internal/providersim"
	_ "github.com/lib/pq"
)

func TestPostgresSubmissionAndObservationCustody(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(12)
	s := NewStore(db)
	ctx := context.Background()
	id := idgen.New()
	defer func() {
		db.Exec("DELETE FROM outbox WHERE aggregate_id=$1", id)
		db.Exec("DELETE FROM operation_receipts WHERE operation_id=$1", id)
		db.Exec("DELETE FROM attempts WHERE operation_id=$1", id)
		db.Exec("DELETE FROM operations WHERE operation_id=$1", id)
	}()
	cmd := dispatch.Command{CommandID: id, ProtocolID: idgen.New(), TenantID: "custody-synthetic", ApplicationID: "app-a", CellID: "r2-cell-a", ProviderAccountID: "synthetic-account", RequestBody: map[string]any{"input": "different-from-output"}, StepDeadline: time.Now().Add(time.Minute), ConfigSnapshot: json.RawMessage(`{"version":1}`)}
	var winners atomic.Int32
	var wg sync.WaitGroup
	for n := 0; n < 24; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, owned, e := s.PrepareSubmission(ctx, cmd, "binding-a", "v1")
			if e != nil {
				t.Error(e)
			}
			if owned {
				winners.Add(1)
			}
		}()
	}
	wg.Wait()
	if winners.Load() != 1 {
		t.Fatalf("submit owners=%d", winners.Load())
	}
	var count int
	if e := db.QueryRow("SELECT count(*) FROM attempts WHERE operation_id=$1 AND prepared_at IS NOT NULL", id).Scan(&count); e != nil || count != 1 {
		t.Fatalf("attempt count %d %v", count, e)
	}
	response := dispatch.Result{Kind: dispatch.FactSucceeded, ResponseBody: map[string]string{"provider_marker": "actual-output"}, ProviderRequestID: "provider-id"}
	for n := 0; n < 12; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, e := s.ConserveObservation(ctx, cmd, response, "synthetic-race")
			if e != nil || !got.Durable || got.EvidenceID == "" {
				t.Errorf("custody %v %+v", e, got)
			}
		}()
	}
	wg.Wait()
	if e := db.QueryRow("SELECT count(*) FROM outbox WHERE aggregate_id=$1", id).Scan(&count); e != nil || count != 1 {
		t.Fatalf("terminal fact count %d %v", count, e)
	}
	got, e := s.DurableResult(ctx, cmd)
	if e != nil {
		t.Fatal(e)
	}
	body, _ := json.Marshal(got.ResponseBody)
	if string(body) != `{"provider_marker":"actual-output"}` {
		t.Fatalf("provider response lost: %s", body)
	}
	other := cmd
	other.TenantID = "other"
	if _, e = s.DurableResult(ctx, other); e == nil {
		t.Fatal("cross tenant result disclosed")
	}
	if _, _, e = s.PrepareSubmission(ctx, other, "binding-a", "v1"); e == nil {
		t.Fatal("conflicting command reused")
	}
	// UNKNOWN after an already committed final cannot overwrite the result.
	if _, e = s.ConserveObservation(ctx, cmd, dispatch.Result{Kind: dispatch.FactUnknown}, "late-timeout"); e != nil {
		t.Fatal(e)
	}
	again, e := s.DurableResult(ctx, cmd)
	if e != nil || again.EvidenceID != got.EvidenceID || again.Kind != dispatch.FactSucceeded {
		t.Fatalf("late observation changed final: %+v %v", again, e)
	}
	t.Log("24 submit contenders: one owner and prewritten attempt; 12 observations: one terminal fact; actual response retained; tenant conflict denied; late UNKNOWN preserved as receipt")
}

func TestPostgresSubmissionLeaseFencesStaleOwner(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewStore(db)
	ctx := context.Background()
	id := idgen.New()
	cmd := dispatch.Command{CommandID: id, ProtocolID: idgen.New(), TenantID: "submit-fence", ApplicationID: "app-a", CellID: "r2-cell-a", ProviderAccountID: "fixture", StepDeadline: time.Now().Add(time.Minute), ConfigSnapshot: json.RawMessage(`{"version":1}`)}
	t.Cleanup(func() {
		db.Exec("DELETE FROM outbox WHERE aggregate_id=$1", id)
		db.Exec("DELETE FROM operation_receipts WHERE operation_id=$1", id)
		db.Exec("DELETE FROM attempts WHERE operation_id=$1", id)
		db.Exec("DELETE FROM operations WHERE operation_id=$1", id)
	})
	claim, owned, err := store.PrepareSubmission(ctx, cmd, "binding", "v1")
	if err != nil || !owned {
		t.Fatalf("prepare: claim=%+v owned=%t err=%v", claim, owned, err)
	}
	if err = pg.WithTenantTx(ctx, db, cmd.TenantID, func(tx *sql.Tx) error {
		_, execErr := tx.ExecContext(ctx, "UPDATE operations SET submit_lease_until=clock_timestamp()-interval '1 second' WHERE operation_id=$1", id)
		return execErr
	}); err != nil {
		t.Fatal(err)
	}
	if err = store.ValidateSubmission(ctx, claim); !errors.Is(err, ErrSubmissionFence) {
		t.Fatalf("expired submission lease accepted: %v", err)
	}
	t.Log("expired submit lease fences the stale owner before external I/O")
}

func TestCallbackCapabilityIsRandomAndStoredOnlyAsHash(t *testing.T) {
	first, err := newCallbackToken()
	if err != nil {
		t.Fatal(err)
	}
	second, err := newCallbackToken()
	if err != nil {
		t.Fatal(err)
	}
	if first == second || len(first) < 40 {
		t.Fatalf("callback capability is not an independent high-entropy value")
	}
	hash := callbackTokenHash(first)
	if hash == first || len(hash) != 64 || hash == callbackTokenHash(second) {
		t.Fatalf("callback capability hashing is invalid")
	}
}

func TestPostgresOrphanCapabilityHashDoesNotShadowValidCallback(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewStore(db)
	operationID := idgen.New()
	body := []byte(`{"provider_request_id":"provider-correlation","status":"SUCCEEDED"}`)
	t.Cleanup(func() { _, _ = db.Exec("DELETE FROM callback_inbox WHERE operation_id=$1", operationID) })
	if err = store.StoreOrphanCallback(context.Background(), operationID, "invalid-capability", body); err != nil {
		t.Fatal(err)
	}
	if err = store.StoreOrphanCallback(context.Background(), operationID, "invalid-capability", body); err != nil {
		t.Fatal(err)
	}
	if err = store.StoreOrphanCallback(context.Background(), operationID, "valid-capability", body); err != nil {
		t.Fatal(err)
	}
	var rows, occurrences int
	if err = db.QueryRow("SELECT count(*),coalesce(sum(occurrences),0) FROM callback_inbox WHERE operation_id=$1", operationID).Scan(&rows, &occurrences); err != nil {
		t.Fatal(err)
	}
	if rows != 2 || occurrences != 3 {
		t.Fatalf("orphan identity was collapsed incorrectly: rows=%d occurrences=%d", rows, occurrences)
	}
	t.Log("same callback bytes with different capability hashes remain distinct; exact duplicate increments occurrences")
}

// R5-SEG-01: um callback de conta órfã autenticado (com provider_account_id)
// deve ser aceito, deduplicado por identidade estável (não pela contagem de
// linhas) e reenviado sem falha de SQL. Este é o caminho concreto que tinha
// um ON CONFLICT malformado antes da correção.
func TestPostgresOrphanAccountCallbackRequiresSignatureAndDeduplicates(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewStore(db)
	operationID := idgen.New()
	accountID := "orphan-account-" + operationID
	body := []byte(`{"provider_request_id":"orphan-correlation","status":"SUCCEEDED"}`)
	at := time.Now()
	t.Cleanup(func() { _, _ = db.Exec("DELETE FROM callback_inbox WHERE operation_id=$1", operationID) })

	t.Setenv("CALLBACK_INGRESS_KEY", "orphan-root-fixture")
	signature := callbackauth.Sign("orphan-root-fixture", accountID, operationID, at, body)
	if signature == "" {
		t.Fatal("fixture signature generation failed")
	}

	if err = store.StoreOrphanAccountCallback(context.Background(), operationID, accountID, at, signature, body); err != nil {
		t.Fatalf("authenticated orphan callback rejected: %v", err)
	}
	// A repeated delivery of the same event must not fail on the insert's
	// conflict target and must be counted as a single obligation.
	if err = store.StoreOrphanAccountCallback(context.Background(), operationID, accountID, at, signature, body); err != nil {
		t.Fatalf("duplicate authenticated orphan callback failed (regression: malformed ON CONFLICT): %v", err)
	}
	var rows, occurrences int
	if err = db.QueryRow("SELECT count(*),coalesce(sum(occurrences),0) FROM callback_inbox WHERE operation_id=$1 AND provider_account_id=$2", operationID, accountID).Scan(&rows, &occurrences); err != nil {
		t.Fatal(err)
	}
	if rows != 1 || occurrences != 2 {
		t.Fatalf("orphan account identity was not deduplicated correctly: rows=%d occurrences=%d", rows, occurrences)
	}

	if err = store.StoreOrphanAccountCallback(context.Background(), operationID, accountID, at, "v1=invalid", body); !errors.Is(err, ErrCallbackCapabilityInvalid) {
		t.Fatalf("forged signature was not rejected: %v", err)
	}
}

// TestPostgresCallbackSurvivesKeyRotationDuringReconciliation prova R6-SEG-02
// (S01): um callback aceito e custodiado com uma chave de ingresso continua
// reconciliavel depois que essa chave e retirada da lista "corrente" e uma
// chave nova assume — a reconciliacao usa o key_id persistido no momento do
// ingresso, nunca "a chave atual" (ver ReconcileCallbackInboxBatch).
func TestPostgresCallbackSurvivesKeyRotationDuringReconciliation(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	// Registered first so it runs LAST (t.Cleanup is LIFO): the DELETE
	// cleanups below must run while the connection is still open. A plain
	// `defer db.Close()` here would close it before t.Cleanup callbacks run,
	// silently discarding every cleanup delete.
	t.Cleanup(func() { db.Close() })
	store := NewStore(db)
	ctx := context.Background()

	accountID := "rotation-account-" + idgen.New()
	cmd := dispatch.Command{CommandID: idgen.New(), ProtocolID: idgen.New(), TenantID: "rotation-tenant", ApplicationID: "app", CellID: "cell-rotation", ProviderAccountID: accountID, StepDeadline: time.Now().Add(time.Minute)}
	claim, owned, err := store.PrepareSubmission(ctx, cmd, "binding", "v1")
	if err != nil || !owned {
		t.Fatalf("prepare rotation claim: owned=%t err=%v", owned, err)
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM callback_inbox WHERE operation_id=$1", claim.Command.CommandID)
		db.Exec("DELETE FROM attempts WHERE operation_id=$1", claim.Command.CommandID)
		db.Exec("DELETE FROM operations WHERE operation_id=$1", claim.Command.CommandID)
		db.Exec("DELETE FROM operation_receipts WHERE operation_id=$1", claim.Command.CommandID)
		db.Exec("DELETE FROM provider_receipts WHERE operation_id=$1", claim.Command.CommandID)
		db.Exec("DELETE FROM outbox WHERE aggregate_id=$1", claim.Command.CommandID)
	})

	// Ingress: only the OLD key is active.
	t.Setenv("CALLBACK_INGRESS_KEYS", `[{"id":"old","secret":"rotation-old-secret"}]`)
	t.Setenv("CALLBACK_INGRESS_CURRENT_KEY_ID", "old")
	body := []byte(`{"provider_request_id":"rotation-correlation","status":"SUCCEEDED"}`)
	at := time.Now()
	signature := callbackauth.Sign("rotation-old-secret", accountID, claim.Command.CommandID, at, body)
	if signature == "" {
		t.Fatal("fixture signature generation failed")
	}
	if err = store.StoreOrphanAccountCallback(ctx, claim.Command.CommandID, accountID, at, signature, body); err != nil {
		t.Fatalf("callback legítimo recusado antes da rotação: %v", err)
	}
	var storedKeyID string
	if err = db.QueryRow("SELECT callback_key_id FROM callback_inbox WHERE operation_id=$1", claim.Command.CommandID).Scan(&storedKeyID); err != nil {
		t.Fatal(err)
	}
	if storedKeyID != "old" {
		t.Fatalf("key_id não persistido no ingresso: got %q", storedKeyID)
	}

	// Rotation: a NEW key becomes current; the old key remains in the ring
	// only long enough to reconcile obligations already in flight — this
	// models the overlap window, not permanent coexistence.
	t.Setenv("CALLBACK_INGRESS_KEYS", `[{"id":"old","secret":"rotation-old-secret"},{"id":"new","secret":"rotation-new-secret"}]`)
	t.Setenv("CALLBACK_INGRESS_CURRENT_KEY_ID", "new")

	// O lote pode conter obrigações residuais de outros testes que
	// compartilham o mesmo banco real; a prova de rotação é o disposition da
	// própria linha, não a contagem agregada do lote.
	applied := 0
	for i := 0; i < 10; i++ {
		count, err := store.ReconcileCallbackInboxBatch(ctx, "rotation-owner", 5, func(context.Context, string, dispatch.Result) error {
			applied++
			return nil
		})
		if err != nil {
			t.Fatalf("reconciliation batch failed: %v", err)
		}
		if count == 0 {
			break
		}
	}
	var disposition string
	if err = db.QueryRow("SELECT disposition FROM callback_inbox WHERE operation_id=$1", claim.Command.CommandID).Scan(&disposition); err != nil {
		t.Fatal(err)
	}
	if disposition != "APPLIED" {
		t.Fatalf("obrigação não recebeu disposição aplicada após rotação: %s", disposition)
	}
}

// TestPostgresCallbackApplyFailureExhaustsIntoRecoverableQuarantine prova
// R6-SEG-02 (S02): tres falhas transitorias no apply de um callback ja
// autenticado colocam a obrigacao em QUARANTINED (nunca REJECTED, nunca
// elegivel a prune automatico) e um replay autorizado explicito a devolve a
// processamento — o resultado chega sem que o provedor reenvie nada.
func TestPostgresCallbackApplyFailureExhaustsIntoRecoverableQuarantine(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	// Registered first so it runs LAST (t.Cleanup is LIFO) — see the
	// analogous comment in TestPostgresCallbackSurvivesKeyRotationDuringReconciliation.
	t.Cleanup(func() { db.Close() })
	store := NewStore(db)
	ctx := context.Background()

	accountID := "quarantine-account-" + idgen.New()
	cmd := dispatch.Command{CommandID: idgen.New(), ProtocolID: idgen.New(), TenantID: "quarantine-tenant", ApplicationID: "app", CellID: "cell-quarantine", ProviderAccountID: accountID, StepDeadline: time.Now().Add(time.Minute)}
	claim, owned, err := store.PrepareSubmission(ctx, cmd, "binding", "v1")
	if err != nil || !owned {
		t.Fatalf("prepare quarantine claim: owned=%t err=%v", owned, err)
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM callback_inbox WHERE operation_id=$1", claim.Command.CommandID)
		db.Exec("DELETE FROM attempts WHERE operation_id=$1", claim.Command.CommandID)
		db.Exec("DELETE FROM operations WHERE operation_id=$1", claim.Command.CommandID)
	})

	t.Setenv("CALLBACK_INGRESS_KEY", "quarantine-secret")
	body := []byte(`{"provider_request_id":"quarantine-correlation","status":"SUCCEEDED"}`)
	at := time.Now()
	signature := callbackauth.Sign("quarantine-secret", accountID, claim.Command.CommandID, at, body)
	if signature == "" {
		t.Fatal("fixture signature generation failed")
	}
	if err = store.StoreOrphanAccountCallback(ctx, claim.Command.CommandID, accountID, at, signature, body); err != nil {
		t.Fatalf("callback legítimo recusado: %v", err)
	}

	// O lote pode conter obrigações residuais de outros testes que
	// compartilham o mesmo banco real (received_at mais antigo, reclamadas
	// primeiro): repete até a PRÓPRIA linha acumular 3 tentativas, em vez de
	// assumir que cada chamada ao lote necessariamente reclama esta linha.
	transientErr := errors.New("simulated transient dependency outage")
	var disposition string
	var inboxID string
	var attemptsSoFar int
	for i := 0; i < 500; i++ {
		if _, err = store.ReconcileCallbackInboxBatch(ctx, "quarantine-owner", 200, func(context.Context, string, dispatch.Result) error {
			return transientErr
		}); err != nil {
			t.Fatalf("reconciliation attempt %d failed: %v", i+1, err)
		}
		if err = db.QueryRow("SELECT inbox_id,disposition,processing_attempts FROM callback_inbox WHERE operation_id=$1", claim.Command.CommandID).Scan(&inboxID, &disposition, &attemptsSoFar); err != nil {
			t.Fatal(err)
		}
		if disposition != "RECEIVED" {
			break
		}
	}
	if disposition != "QUARANTINED" {
		t.Fatalf("esgotamento de retry não preservou a obrigação em quarentena: disposition=%s attempts=%d", disposition, attemptsSoFar)
	}

	// Retention must never sweep a QUARANTINED obligation away.
	if _, err = store.PruneCallbackInbox(ctx, 24*time.Hour, 100); err != nil {
		t.Fatal(err)
	}
	if err = db.QueryRow("SELECT disposition FROM callback_inbox WHERE operation_id=$1", claim.Command.CommandID).Scan(&disposition); err != nil {
		t.Fatalf("retenção descartou obrigação em quarentena: %v", err)
	}
	if disposition != "QUARANTINED" {
		t.Fatalf("retenção alterou disposição indevidamente: %s", disposition)
	}

	// An unauthorized/incorrect replay target must fail closed.
	if err = store.ReplayCallbackInbox(ctx, "00000000-0000-0000-0000-000000000000", "operator-x:test"); !errors.Is(err, ErrCallbackNotQuarantined) {
		t.Fatalf("replay de id inexistente não foi recusado: %v", err)
	}

	// Authorized replay: the SAME stored body/signature reprocesses — the
	// provider is never asked to resend anything.
	if err = store.ReplayCallbackInbox(ctx, inboxID, "operator-x:dependency restored"); err != nil {
		t.Fatalf("replay autorizado falhou: %v", err)
	}
	var attempts int
	if err = db.QueryRow("SELECT disposition,processing_attempts FROM callback_inbox WHERE operation_id=$1", claim.Command.CommandID).Scan(&disposition, &attempts); err != nil {
		t.Fatal(err)
	}
	if disposition != "RECEIVED" || attempts != 0 {
		t.Fatalf("replay não restaurou o orçamento de retentativa: disposition=%s attempts=%d", disposition, attempts)
	}

	for i := 0; i < 10; i++ {
		count, err := store.ReconcileCallbackInboxBatch(ctx, "quarantine-owner-2", 5, func(context.Context, string, dispatch.Result) error {
			return nil
		})
		if err != nil {
			t.Fatalf("reconciliation after replay failed: %v", err)
		}
		if count == 0 {
			break
		}
	}
	if err = db.QueryRow("SELECT disposition FROM callback_inbox WHERE operation_id=$1", claim.Command.CommandID).Scan(&disposition); err != nil {
		t.Fatal(err)
	}
	if disposition != "APPLIED" {
		t.Fatalf("obrigação reprocessada não chegou a disposição terminal aplicada: %s", disposition)
	}
}

func TestPostgresCallbackInboxBatchClaimsBoundedAndDisposesInvalidCapability(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := NewStore(db)
	ctx := context.Background()
	body := []byte(`{"provider_request_id":"provider-correlation","status":"SUCCEEDED","detail":"orphan"}`)
	claims := make([]Submission, 0, 2)
	for _, tenant := range []string{"callback-batch-a", "callback-batch-b"} {
		cmd := dispatch.Command{CommandID: idgen.New(), ProtocolID: idgen.New(), TenantID: tenant, ApplicationID: "app", CellID: "cell-" + tenant, ProviderAccountID: "account", StepDeadline: time.Now().Add(time.Minute)}
		claim, owned, err := store.PrepareSubmission(ctx, cmd, "binding", "v1")
		if err != nil || !owned {
			t.Fatalf("prepare callback batch claim: owned=%t err=%v", owned, err)
		}
		claims = append(claims, claim)
	}
	t.Cleanup(func() {
		for _, claim := range claims {
			db.Exec("DELETE FROM callback_inbox WHERE operation_id=$1", claim.Command.CommandID)
			db.Exec("DELETE FROM attempts WHERE operation_id=$1", claim.Command.CommandID)
			db.Exec("DELETE FROM operations WHERE operation_id=$1", claim.Command.CommandID)
		}
	})
	if err := store.StoreOrphanCallback(ctx, claims[0].Command.CommandID, "wrong-capability", body); err != nil {
		t.Fatal(err)
	}
	if err := store.StoreOrphanCallback(ctx, claims[1].Command.CommandID, claims[1].CallbackToken, body); err != nil {
		t.Fatal(err)
	}

	applied := 0
	count, err := store.ReconcileCallbackInboxBatch(ctx, "batch-owner-a", 1, func(context.Context, string, dispatch.Result) error {
		applied++
		return nil
	})
	if err != nil || count != 0 || applied != 0 {
		t.Fatalf("lote limitado processou capability inválida: count=%d applied=%d err=%v", count, applied, err)
	}
	var disposition string
	if err := db.QueryRow("SELECT disposition FROM callback_inbox WHERE operation_id=$1", claims[0].Command.CommandID).Scan(&disposition); err != nil {
		t.Fatal(err)
	}
	if disposition != "REJECTED" {
		t.Fatalf("capability inválida não recebeu disposição terminal: %s", disposition)
	}

	count, err = store.ReconcileCallbackInboxBatch(ctx, "batch-owner-b", 1, func(context.Context, string, dispatch.Result) error {
		applied++
		return nil
	})
	if err != nil || count != 1 || applied != 1 {
		t.Fatalf("callback válido não foi aplicado no lote seguinte: count=%d applied=%d err=%v", count, applied, err)
	}
	if err := db.QueryRow("SELECT disposition FROM callback_inbox WHERE operation_id=$1", claims[1].Command.CommandID).Scan(&disposition); err != nil {
		t.Fatal(err)
	}
	if disposition != "APPLIED" {
		t.Fatalf("callback válido não recebeu disposição aplicada: %s", disposition)
	}
	t.Log("callback inbox real: lote limitado a um item, capability inválida rejeitada sem callback de aplicação e callback válido aplicado no lote seguinte")
}

func TestPostgresExecutorDropAfterEffectIsUnknownAndNotReexecuted(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL")
	}
	t.Setenv("CELL_ID", "r2-cell-a")
	t.Setenv("ENVIRONMENT", "local")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(8)
	store := NewStore(db)
	id := idgen.New()
	capacity := NewCapacityController(db)
	capacityPolicy := CapacityPolicy{Domain: "synthetic-executor-capacity-" + id, Version: "fixture-v1", EvidenceRef: "synthetic-executor-capacity", ValidUntil: time.Now().Add(time.Hour), MaxConcurrent: 8, MinConcurrent: 5, ReconciliationReserve: 1, MaxPending: 8, RatePerWindow: 200, WindowMillis: 1000, LeaseMillis: 5000, StableMillis: 100, LatencyThresholdMillis: 100, TenantLimits: map[string]int{"drop-synthetic": 2}, TenantPendingLimits: map[string]int{"drop-synthetic": 4}, TenantRateLimits: map[string]int{"drop-synthetic": 90}}
	if err := capacity.InstallPolicy(context.Background(), capacityPolicy); err != nil {
		t.Fatal(err)
	}
	cmd := dispatch.Command{CommandID: id, ProtocolID: idgen.New(), TenantID: "drop-synthetic", ApplicationID: "app-a", CellID: "r2-cell-a", ProviderAccountID: "account", StepDeadline: time.Now().Add(time.Minute), RequestBody: map[string]any{"force_drop_after_effect": true}}
	t.Cleanup(func() {
		db.Exec("DELETE FROM capacity_feedback WHERE domain_id=$1", capacityPolicy.Domain)
		db.Exec("DELETE FROM capacity_permits WHERE domain_id=$1", capacityPolicy.Domain)
		db.Exec("DELETE FROM capacity_domains WHERE domain_id=$1", capacityPolicy.Domain)
		db.Exec("DELETE FROM outbox WHERE aggregate_id=$1", id)
		db.Exec("DELETE FROM operation_receipts WHERE operation_id=$1", id)
		db.Exec("DELETE FROM attempts WHERE operation_id=$1", id)
		db.Exec("DELETE FROM operations WHERE operation_id=$1", id)
	})
	provider := providersim.NewServer()
	providerMux := http.NewServeMux()
	provider.Routes(providerMux)
	providerHTTP := httptest.NewServer(providerMux)
	defer providerHTTP.Close()
	providerURL, _ := url.Parse(providerHTTP.URL)
	t.Setenv("EGRESS_HTTP_ORIGINS", providerHTTP.URL)
	t.Setenv("EGRESS_PRIVATE_RULES", providerURL.Host+"=127.0.0.1/32")
	catalog := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "fixture-workload", "expires_in": 60, "token_type": "Bearer"})
			return
		}
		if r.Header.Get("Authorization") != "Bearer fixture-workload" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		_ = json.NewEncoder(w).Encode(atlasclient.CredentialBinding{BindingID: "binding", SecretRef: "fixture-secret", SecretVersion: "v1", CredentialMode: "SHARED_HUB"})
	}))
	defer catalog.Close()
	account, _ := json.Marshal(map[string]any{"base_url": providerHTTP.URL, "provider_mode": "sync", "auth_type": "NONE"})
	snapshot := atlas.OfferSnapshot{Account: atlas.Resource{ID: "account", Data: account}, Binding: atlas.Resource{ID: "binding", Version: 1, Data: json.RawMessage(`{"secret_version":"v1","provider_account_id":"account"}`)}, Target: atlas.Resource{Data: json.RawMessage(`{"adapter_id":"synthetic-provider","output_schema":{"type":"object","properties":{}}}`)}, SelectedRoute: atlas.Route{ProviderAccountID: "account", CapacityDomain: capacityPolicy.Domain}}
	cmd.ConfigSnapshot, _ = json.Marshal(snapshot)
	tokenFile := t.TempDir() + "/secret"
	if err := os.WriteFile(tokenFile, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OIDC_TOKEN_URL", catalog.URL+"/token")
	t.Setenv("WORKLOAD_CLIENT_ID", "fixture")
	t.Setenv("WORKLOAD_CLIENT_SECRET_FILE", tokenFile)
	exec := NewExecutor(store, atlasclient.New(catalog.URL, time.Second), slog.New(slog.NewTextHandler(io.Discard, nil)), "", &providerauth.TokenCache{Resolver: dropFixtureVault{}})
	exec.SetCapacityController(capacity)
	first := exec.Execute(context.Background(), cmd)
	t.Logf("first execution: %+v", first)
	if first.Kind != dispatch.FactUnknown || !first.Durable {
		t.Fatalf("first execution should be durable UNKNOWN: %+v", first)
	}
	second := exec.Execute(context.Background(), cmd)
	if second.Kind != dispatch.FactUnknown || !second.Durable || second.EvidenceID != first.EvidenceID {
		t.Fatalf("recovery changed durable UNKNOWN: first=%+v second=%+v", first, second)
	}
	response, err := providerHTTP.Client().Get(providerHTTP.URL + "/__qualification/effects")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var effects struct {
		Effects int `json:"effects"`
	}
	if err = json.NewDecoder(response.Body).Decode(&effects); err != nil {
		t.Fatal(err)
	}
	if effects.Effects != 1 {
		t.Fatalf("external effect was reexecuted: %d", effects.Effects)
	}
	state, err := capacity.State(context.Background(), capacityPolicy.Domain)
	if err != nil || state.TransportOpen != 0 || state.PendingExternal != 1 {
		t.Fatalf("capacity did not retain uncertain effect: %+v %v", state, err)
	}
	var permits int
	if err = pg.WithAuditedScopeTx(context.Background(), db, "test-fixture:count-permits", func(tx *sql.Tx) error {
		return tx.QueryRowContext(context.Background(), "SELECT count(*) FROM capacity_permits WHERE domain_id=$1", capacityPolicy.Domain).Scan(&permits)
	}); err != nil || permits != 1 {
		t.Fatalf("capacity permit count=%d error=%v", permits, err)
	}
}

func TestPostgresValidResponseCommitFailureLeavesRecoverableObligation(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL")
	}
	t.Setenv("CELL_ID", "r2-cell-a")
	t.Setenv("ENVIRONMENT", "local")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	db.SetMaxOpenConns(8)
	ctx := context.Background()
	store := NewStore(db)
	// This fixture needs a synthetic trigger to inject a terminal commit
	// failure — genuine DDL rights that a non-superuser runtime role
	// (hub_runtime, R6-SEG-01) correctly does not have. R2_CORE_ADMIN_DSN lets
	// the qualification harness supply a migrator-privileged connection for
	// just the DDL setup/teardown; it defaults to R2_CORE_TEST_DSN so this
	// test is unaffected when run with an owner/superuser DSN, as before.
	adminDSN := os.Getenv("R2_CORE_ADMIN_DSN")
	if adminDSN == "" {
		adminDSN = dsn
	}
	adminDB, err := sql.Open("postgres", adminDSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = adminDB.Close() })
	operationID := idgen.New()
	triggerName := "qualification_fail_commit_" + strings.ReplaceAll(operationID, "-", "")
	functionName := triggerName + "_fn"
	targetsTable := triggerName + "_targets"
	_, err = adminDB.Exec(fmt.Sprintf(`CREATE TABLE %s (operation_id uuid PRIMARY KEY);
CREATE FUNCTION %s() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.state = 'SUCCEEDED' AND EXISTS (SELECT 1 FROM %s WHERE operation_id = NEW.operation_id) THEN
        RAISE EXCEPTION 'qualification: terminal commit failure';
    END IF;
    RETURN NEW;
END
$$;
CREATE CONSTRAINT TRIGGER %s AFTER UPDATE ON operations
DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION %s();`, targetsTable, functionName, targetsTable, triggerName, functionName))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = adminDB.Exec("INSERT INTO "+targetsTable+" (operation_id) VALUES ($1)", operationID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = adminDB.Exec("DROP TRIGGER IF EXISTS " + triggerName + " ON operations")
		_, _ = adminDB.Exec("DROP FUNCTION IF EXISTS " + functionName + "()")
		_, _ = adminDB.Exec("DROP TABLE IF EXISTS " + targetsTable)
		_, _ = db.Exec("DELETE FROM outbox WHERE aggregate_id=$1", operationID)
		_, _ = db.Exec("DELETE FROM operation_receipts WHERE operation_id=$1", operationID)
		_, _ = db.Exec("DELETE FROM provider_receipts WHERE operation_id=$1", operationID)
		_, _ = db.Exec("DELETE FROM attempts WHERE operation_id=$1", operationID)
		_, _ = db.Exec("DELETE FROM operations WHERE operation_id=$1", operationID)
	})

	provider := providersim.NewServer()
	providerMux := http.NewServeMux()
	provider.Routes(providerMux)
	providerHTTP := httptest.NewServer(providerMux)
	defer providerHTTP.Close()
	providerURL, _ := url.Parse(providerHTTP.URL)
	t.Setenv("EGRESS_HTTP_ORIGINS", providerHTTP.URL)
	t.Setenv("EGRESS_PRIVATE_RULES", providerURL.Host+"=127.0.0.1/32")

	tenant := "commit-recovery-" + idgen.New()
	bindingID := "binding-" + idgen.New()
	catalog := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "fixture-workload", "expires_in": 60, "token_type": "Bearer"})
			return
		}
		_ = json.NewEncoder(w).Encode(atlasclient.CredentialBinding{BindingID: bindingID, TenantID: tenant, SecretRef: "vault://fixture", SecretVersion: "v1", CredentialMode: "SHARED_HUB"})
	}))
	defer catalog.Close()

	capacity := NewCapacityController(db)
	capacityPolicy := CapacityPolicy{Domain: "commit-recovery-capacity-" + operationID, Version: "fixture-v1", EvidenceRef: "commit-recovery-capacity", ValidUntil: time.Now().Add(time.Hour), MaxConcurrent: 8, MinConcurrent: 5, ReconciliationReserve: 1, MaxPending: 8, RatePerWindow: 200, WindowMillis: 1000, LeaseMillis: 5000, StableMillis: 100, LatencyThresholdMillis: 100, TenantLimits: map[string]int{tenant: 2}, TenantPendingLimits: map[string]int{tenant: 4}, TenantRateLimits: map[string]int{tenant: 90}}
	if err = capacity.InstallPolicy(context.Background(), capacityPolicy); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM capacity_feedback WHERE domain_id=$1", capacityPolicy.Domain)
		db.Exec("DELETE FROM capacity_permits WHERE domain_id=$1", capacityPolicy.Domain)
		db.Exec("DELETE FROM capacity_domains WHERE domain_id=$1", capacityPolicy.Domain)
	})
	account, _ := json.Marshal(map[string]any{"base_url": providerHTTP.URL, "provider_mode": "sync", "auth_type": "NONE"})
	snapshot := atlas.OfferSnapshot{
		Account:       atlas.Resource{ID: "account", Data: account},
		Binding:       atlas.Resource{ID: bindingID, Version: 1, Data: json.RawMessage(`{"secret_version":"v1","provider_account_id":"account"}`)},
		Target:        atlas.Resource{Data: json.RawMessage(`{"adapter_id":"synthetic-provider","output_schema":{"type":"object"}}`)},
		SelectedRoute: atlas.Route{ProviderAccountID: "account", CapacityDomain: capacityPolicy.Domain},
	}
	config, _ := json.Marshal(snapshot)
	cmd := dispatch.Command{
		CommandID: operationID, ProtocolID: idgen.New(), TenantID: tenant, ApplicationID: "app-a", CellID: "r2-cell-a",
		ProviderAccountID: "account", DispatchMode: dispatch.DispatchDirect, ConfigSnapshot: config,
		RequestBody: map[string]any{"marker": "valid-external-response"}, StepDeadline: time.Now().Add(time.Minute),
	}
	secretPath := t.TempDir() + "/secret"
	if err = os.WriteFile(secretPath, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OIDC_TOKEN_URL", catalog.URL+"/token")
	t.Setenv("WORKLOAD_CLIENT_ID", "fixture")
	t.Setenv("WORKLOAD_CLIENT_SECRET_FILE", secretPath)

	executor := NewExecutor(store, atlasclient.New(catalog.URL, time.Second), slog.New(slog.NewTextHandler(io.Discard, nil)), "", &providerauth.TokenCache{Resolver: dropFixtureVault{}})
	executor.SetCapacityController(capacity)
	first := executor.Execute(ctx, cmd)
	if first.Kind != dispatch.FactUnknown || !first.Durable || first.ProviderRequestID == "" || first.EvidenceID == "" {
		t.Fatalf("falha de commit não virou obrigação UNKNOWN durável: %+v", first)
	}
	var state, providerRequestID, source string
	var storedResult []byte
	var receipts int
	if err = pg.WithTenantTx(ctx, db, tenant, func(tx *sql.Tx) error {
		if err := tx.QueryRow("SELECT state,COALESCE(provider_request_id,''),COALESCE(result,'null') FROM operations WHERE operation_id=$1", operationID).Scan(&state, &providerRequestID, &storedResult); err != nil {
			return err
		}
		if err := tx.QueryRow("SELECT source FROM operation_receipts WHERE evidence_id=$1", first.EvidenceID).Scan(&source); err != nil {
			return err
		}
		return tx.QueryRow("SELECT count(*) FROM provider_receipts WHERE operation_id=$1", operationID).Scan(&receipts)
	}); err != nil {
		t.Fatal(err)
	}
	if state != string(StateUnknown) || providerRequestID != first.ProviderRequestID {
		t.Fatalf("obrigação não reteve correlação externa: state=%s provider_request_id=%s result=%s", state, providerRequestID, storedResult)
	}
	if source != "PROVIDER_RECOVERY" {
		t.Fatalf("origem da obrigação inesperada: %s", source)
	}
	var outbox int
	if err = db.QueryRow("SELECT count(*) FROM outbox WHERE aggregate_id=$1", operationID).Scan(&outbox); err != nil {
		t.Fatal(err)
	}
	if receipts != 1 || outbox != 0 {
		t.Fatalf("custódia parcial indevida: provider_receipts=%d outbox=%d", receipts, outbox)
	}

	var effectsBefore struct {
		Effects int `json:"effects"`
	}
	response, err := providerHTTP.Client().Get(providerHTTP.URL + "/__qualification/effects")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.NewDecoder(response.Body).Decode(&effectsBefore); err != nil {
		response.Body.Close()
		t.Fatal(err)
	}
	response.Body.Close()
	second := executor.Execute(ctx, cmd)
	if second.Kind != dispatch.FactUnknown || !second.Durable || second.ProviderRequestID != first.ProviderRequestID || second.EvidenceID != first.EvidenceID {
		t.Fatalf("reentrega alterou a obrigação ou sua evidência: first=%+v second=%+v", first, second)
	}
	var effectsAfter struct {
		Effects int `json:"effects"`
	}
	response, err = providerHTTP.Client().Get(providerHTTP.URL + "/__qualification/effects")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.NewDecoder(response.Body).Decode(&effectsAfter); err != nil {
		response.Body.Close()
		t.Fatal(err)
	}
	response.Body.Close()
	if effectsBefore.Effects != 1 || effectsAfter.Effects != effectsBefore.Effects {
		t.Fatalf("reentrega disparou novo submit: antes=%d depois=%d", effectsBefore.Effects, effectsAfter.Effects)
	}
	t.Logf("resposta externa válida + commit terminal falho: UNKNOWN reconciliável, correlação/recibo preservados, zero outbox e zero novo POST (provider_request_id=%s)", providerRequestID)
}

func TestPostgresInvalidProviderResponseDoesNotBecomeSuccess(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL")
	}
	t.Setenv("CELL_ID", "r2-cell-a")
	t.Setenv("ENVIRONMENT", "local")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	db.SetMaxOpenConns(8)
	ctx := context.Background()
	store := NewStore(db)
	operationID := idgen.New()
	tenant := "invalid-provider-" + idgen.New()
	bindingID := "binding-" + idgen.New()
	providerCalls := atomic.Int32{}
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		providerCalls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{}`)
	}))
	defer provider.Close()
	providerURL, _ := url.Parse(provider.URL)
	t.Setenv("EGRESS_HTTP_ORIGINS", provider.URL)
	t.Setenv("EGRESS_PRIVATE_RULES", providerURL.Host+"=127.0.0.1/32")
	catalog := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "fixture-workload", "expires_in": 60, "token_type": "Bearer"})
			return
		}
		_ = json.NewEncoder(w).Encode(atlasclient.CredentialBinding{BindingID: bindingID, TenantID: tenant, SecretRef: "vault://fixture", SecretVersion: "v1", CredentialMode: "SHARED_HUB"})
	}))
	defer catalog.Close()
	capacity := NewCapacityController(db)
	capacityPolicy := CapacityPolicy{Domain: "invalid-provider-capacity-" + operationID, Version: "fixture-v1", EvidenceRef: "invalid-provider-capacity", ValidUntil: time.Now().Add(time.Hour), MaxConcurrent: 8, MinConcurrent: 5, ReconciliationReserve: 1, MaxPending: 8, RatePerWindow: 200, WindowMillis: 1000, LeaseMillis: 5000, StableMillis: 100, LatencyThresholdMillis: 100, TenantLimits: map[string]int{tenant: 2}, TenantPendingLimits: map[string]int{tenant: 4}, TenantRateLimits: map[string]int{tenant: 90}}
	if err = capacity.InstallPolicy(context.Background(), capacityPolicy); err != nil {
		t.Fatal(err)
	}
	account, _ := json.Marshal(map[string]any{"base_url": provider.URL, "provider_mode": "sync", "auth_type": "NONE"})
	snapshot := atlas.OfferSnapshot{
		Account:       atlas.Resource{ID: "account", Data: account},
		Binding:       atlas.Resource{ID: bindingID, Version: 1, Data: json.RawMessage(`{"secret_version":"v1","provider_account_id":"account"}`)},
		Target:        atlas.Resource{Data: json.RawMessage(`{"adapter_id":"synthetic-provider","output_schema":{"type":"object"}}`)},
		SelectedRoute: atlas.Route{ProviderAccountID: "account", CapacityDomain: capacityPolicy.Domain},
	}
	config, _ := json.Marshal(snapshot)
	cmd := dispatch.Command{CommandID: operationID, ProtocolID: idgen.New(), TenantID: tenant, ApplicationID: "app-a", CellID: "r2-cell-a", ProviderAccountID: "account", DispatchMode: dispatch.DispatchDirect, ConfigSnapshot: config, RequestBody: map[string]any{"marker": "invalid-response"}, StepDeadline: time.Now().Add(time.Minute)}
	secretPath := t.TempDir() + "/secret"
	if err = os.WriteFile(secretPath, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OIDC_TOKEN_URL", catalog.URL+"/token")
	t.Setenv("WORKLOAD_CLIENT_ID", "fixture")
	t.Setenv("WORKLOAD_CLIENT_SECRET_FILE", secretPath)
	t.Cleanup(func() {
		db.Exec("DELETE FROM capacity_feedback WHERE domain_id=$1", capacityPolicy.Domain)
		db.Exec("DELETE FROM capacity_permits WHERE domain_id=$1", capacityPolicy.Domain)
		db.Exec("DELETE FROM capacity_domains WHERE domain_id=$1", capacityPolicy.Domain)
		_, _ = db.Exec("DELETE FROM outbox WHERE aggregate_id=$1", operationID)
		_, _ = db.Exec("DELETE FROM operation_receipts WHERE operation_id=$1", operationID)
		_, _ = db.Exec("DELETE FROM provider_receipts WHERE operation_id=$1", operationID)
		_, _ = db.Exec("DELETE FROM attempts WHERE operation_id=$1", operationID)
		_, _ = db.Exec("DELETE FROM operations WHERE operation_id=$1", operationID)
	})

	executor := NewExecutor(store, atlasclient.New(catalog.URL, time.Second), slog.New(slog.NewTextHandler(io.Discard, nil)), "", &providerauth.TokenCache{Resolver: dropFixtureVault{}})
	executor.SetCapacityController(capacity)
	first := executor.Execute(ctx, cmd)
	if first.Kind != dispatch.FactUnknown || !first.Durable || first.ErrorCode != "invalid_provider_response" {
		t.Fatalf("resposta inválida não virou UNKNOWN durável: %+v", first)
	}
	var state string
	var succeeded, receipts, facts int
	if err = pg.WithTenantTx(ctx, db, tenant, func(tx *sql.Tx) error {
		if scanErr := tx.QueryRowContext(ctx, "SELECT state FROM operations WHERE operation_id=$1", operationID).Scan(&state); scanErr != nil {
			return scanErr
		}
		if scanErr := tx.QueryRowContext(ctx, "SELECT count(*) FROM operations WHERE operation_id=$1 AND state='SUCCEEDED'", operationID).Scan(&succeeded); scanErr != nil {
			return scanErr
		}
		return tx.QueryRowContext(ctx, "SELECT count(*) FROM provider_receipts WHERE operation_id=$1", operationID).Scan(&receipts)
	}); err != nil {
		t.Fatal(err)
	}
	if err = db.QueryRow("SELECT count(*) FROM outbox WHERE aggregate_id=$1 AND event_type='operation.observed'", operationID).Scan(&facts); err != nil {
		t.Fatal(err)
	}
	if state != string(StateUnknown) || succeeded != 0 || receipts != 0 || facts != 1 {
		t.Fatalf("custódia de resposta inválida incorreta: state=%s succeeded=%d receipts=%d facts=%d", state, succeeded, receipts, facts)
	}
	second := executor.Execute(ctx, cmd)
	if second.EvidenceID != first.EvidenceID || second.Kind != dispatch.FactUnknown || providerCalls.Load() != 1 {
		t.Fatalf("reentrega alterou a decisão ou repetiu efeito: first=%+v second=%+v provider_calls=%d", first, second, providerCalls.Load())
	}
	t.Logf("HTTP retornou JSON inválido: Cometa preservou UNKNOWN com erro de contrato, sem SUCCEEDED/recibo bruto e sem novo POST (calls=%d)", providerCalls.Load())
}

type dropFixtureVault struct{}

func (dropFixtureVault) Resolve(context.Context, string, string) (providerauth.Secret, error) {
	return providerauth.Secret{Value: "fixture", Version: "v1"}, nil
}

func TestPostgresPendingCustodyAtomic(t *testing.T) {
	t.Setenv("ENVIRONMENT", "local")
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("requires isolated migrated PostgreSQL")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewStore(db)
	ctx := context.Background()
	id := idgen.New()
	defer func() {
		db.Exec("DELETE FROM polling_schedule WHERE operation_id=$1", id)
		db.Exec("DELETE FROM outbox WHERE aggregate_id=$1", id)
		db.Exec("DELETE FROM operation_receipts WHERE operation_id=$1", id)
		db.Exec("DELETE FROM attempts WHERE operation_id=$1", id)
		db.Exec("DELETE FROM operations WHERE operation_id=$1", id)
	}()
	// Missing durable polling budget is detected after tentative state/receipt
	// writes; rollback must leave neither a receipt nor a false acceptance.
	cmd := dispatch.Command{CommandID: id, ProtocolID: idgen.New(), TenantID: "pending-test", ApplicationID: "app-a", CellID: "r2-cell-a", ProviderAccountID: "fixture", ConfigSnapshot: json.RawMessage(`{}`)}
	if _, owned, err := s.PrepareSubmission(ctx, cmd, "binding", "version"); err != nil || !owned {
		t.Fatalf("prepare: %v", err)
	}
	if r, err := s.ConserveAcceptance(ctx, cmd, "provider-pending", true); err == nil || r.Durable {
		t.Fatalf("false custody: %+v %v", r, err)
	}
	var state string
	var count int
	if err = pg.WithTenantTx(ctx, db, cmd.TenantID, func(tx *sql.Tx) error {
		if scanErr := tx.QueryRowContext(ctx, "SELECT state FROM operations WHERE operation_id=$1", id).Scan(&state); scanErr != nil {
			return scanErr
		}
		return tx.QueryRowContext(ctx, "SELECT count(*) FROM operation_receipts WHERE operation_id=$1", id).Scan(&count)
	}); err != nil || state != "SUBMITTING" {
		t.Fatalf("partial state: %s %v", state, err)
	}
	if count != 0 {
		t.Fatalf("partial receipt: %d %v", count, err)
	}
	// Fixture adjustment of its persisted retry horizon, never a runtime backfill.
	cmd.RetryDeadline = time.Now().Add(time.Minute)
	raw, _ := json.Marshal(cmd)
	if err = pg.WithTenantTx(ctx, db, cmd.TenantID, func(tx *sql.Tx) error {
		_, execErr := tx.ExecContext(ctx, "UPDATE operations SET command=$2 WHERE operation_id=$1", id, raw)
		return execErr
	}); err != nil {
		t.Fatal(err)
	}
	attemptID := idgen.New()
	got, err := s.ConserveAcceptance(ctx, cmd, "provider-pending", true, attemptID)
	if err != nil || !got.Durable || got.EvidenceID == "" {
		t.Fatalf("acceptance: %+v %v", got, err)
	}
	if err = db.QueryRow("SELECT count(*) FROM polling_schedule WHERE operation_id=$1 AND deadline_at=$2", id, cmd.RetryDeadline).Scan(&count); err != nil || count != 1 {
		t.Fatalf("polling obligation: %d %v", count, err)
	}
	if err = db.QueryRow("SELECT count(*) FROM outbox WHERE aggregate_id=$1", id).Scan(&count); err != nil || count != 1 {
		t.Fatalf("outbox: %d %v", count, err)
	}
	var incidence, payloadAttempt string
	if err = db.QueryRow("SELECT payload->>'economic_kind',payload->>'attempt_id' FROM outbox WHERE aggregate_id=$1", id).Scan(&incidence, &payloadAttempt); err != nil || incidence != "SUBMITTED" || payloadAttempt != attemptID {
		t.Fatalf("economic acceptance: incidence=%q attempt=%q err=%v", incidence, payloadAttempt, err)
	}
	replay, err := s.DurableResult(ctx, cmd)
	if err != nil || !replay.Durable || replay.ProviderRequestID != "provider-pending" {
		t.Fatalf("replay: %+v %v", replay, err)
	}
	t.Log("real PostgreSQL: failed polling budget rolls back state and receipt; valid acceptance conserves correlation, receipt, schedule, result and outbox")
}
