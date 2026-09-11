package cometa

import (
	"time"

	"ai-hub/hub/internal/atlas"
)

func providerHTTPBudget(snapshot atlas.OfferSnapshot, fallback time.Duration) time.Duration {
	target, err := atlas.DecodeCatalogData(snapshot.Target)
	if err == nil && target.SyncHTTPBudgetSeconds > 0 {
		return time.Duration(target.SyncHTTPBudgetSeconds) * time.Second
	}
	return fallback
}

func minDuration(left, right time.Duration) time.Duration {
	if left <= 0 {
		return right
	}
	if right <= 0 || left < right {
		return left
	}
	return right
}
