package outbox

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"ai-hub/hub/internal/platform/httpserver"
	_ "github.com/lib/pq"
)

func TestPendingAgeMetricPublishesBoundedGauge(t *testing.T) {
	dsn := os.Getenv("R2_CORE_TEST_DSN")
	if dsn == "" {
		t.Skip("R2_CORE_TEST_DSN não definido")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err = db.ExecContext(ctx, `INSERT INTO outbox(aggregate_type,aggregate_id,event_type,payload,published,created_at) VALUES('metric-test','metric-test','metric.test','{}',FALSE,clock_timestamp()-interval '42 seconds')`); err != nil {
		t.Fatal(err)
	}
	defer db.ExecContext(context.Background(), `DELETE FROM outbox WHERE aggregate_type='metric-test'`)
	registry := httpserver.NewRegistry()
	metricCtx, stop := context.WithCancel(context.Background())
	defer stop()
	go RunPendingAgeMetric(metricCtx, db, "metric-test", "test", time.Hour, registry, slog.Default())
	time.Sleep(50 * time.Millisecond)
	recorder := httptest.NewRecorder()
	registry.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if !strings.Contains(recorder.Body.String(), `hub_obligation_age_seconds{component="test",kind="outbox"}`) {
		t.Fatalf("métrica ausente: %s", recorder.Body.String())
	}
}
