package queue

import (
	"context"
	"log/slog"
	"time"
)

// RunBootstrap retries broker resource discovery without preventing independent
// HTTP and durable database obligations from starting. Setup must start workers
// only after all subscriptions have been confirmed.
func RunBootstrap(ctx context.Context, setup func() error, log *slog.Logger) {
	for {
		if ctx.Err() != nil {
			return
		}
		if err := setup(); err == nil {
			return
		} else {
			log.Warn("broker bootstrap unavailable; durable obligations retained", "error", err)
		}
		timer := time.NewTimer(2 * time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}
