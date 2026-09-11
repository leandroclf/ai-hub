package cometa

import (
	"context"
	"testing"
	"time"

	"ai-hub/hub/internal/atlas"
)

func TestProviderHTTPBudgetUsesFrozenTargetBudget(t *testing.T) {
	snapshot := atlas.OfferSnapshot{Target: atlas.Resource{Data: []byte(`{"sync_http_budget_seconds":2}`)}}
	if got := providerHTTPBudget(snapshot, 15*time.Second); got != 2*time.Second {
		t.Fatalf("target budget = %s, want 2s", got)
	}
	if got := providerHTTPBudget(atlas.OfferSnapshot{}, 15*time.Second); got != 15*time.Second {
		t.Fatalf("fallback budget = %s, want 15s", got)
	}
}

func TestEffectiveHTTPBudgetNeverExceedsLeaseOrContext(t *testing.T) {
	now := time.Now()
	ctx, cancel := context.WithDeadline(context.Background(), now.Add(20*time.Second))
	defer cancel()
	permit := CapacityPermit{LeaseUntil: now.Add(5 * time.Second)}
	got := effectiveHTTPBudget(ctx, 15*time.Second, permit, true)
	if got < 3*time.Second || got > 5*time.Second {
		t.Fatalf("effective lease budget = %s, want roughly 4s", got)
	}

	expired := CapacityPermit{LeaseUntil: now.Add(-time.Second)}
	if got = effectiveHTTPBudget(ctx, time.Second, expired, true); got != 0 {
		t.Fatalf("expired lease budget = %s, want zero", got)
	}

	shortCtx, shortCancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer shortCancel()
	if got = effectiveHTTPBudget(shortCtx, 15*time.Second, CapacityPermit{}, false); got <= 0 || got > 200*time.Millisecond {
		t.Fatalf("context budget = %s, want <=200ms", got)
	}
}
