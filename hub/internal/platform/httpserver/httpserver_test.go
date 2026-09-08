package httpserver

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAuthDoesNotProtectProbesAndMetricsHaveBoundedRoutes(t *testing.T) {
	s := New(slog.New(slog.NewTextHandler(io.Discard, nil)), nil, nil, nil)
	s.AuthMiddleware = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") == "" {
				http.Error(w, "unauthorized", 401)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
	s.HandleFunc("GET /objects/{id}", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("result")) })
	for _, path := range []string{"/objects/first-secret-id", "/objects/second-secret-id"} {
		rec := httptest.NewRecorder()
		s.withLogging(s.mux).ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		if rec.Code != 401 {
			t.Fatal(rec.Code)
		}
	}
	probe := httptest.NewRecorder()
	s.mux.ServeHTTP(probe, httptest.NewRequest("GET", "/healthz/ready", nil))
	if probe.Code != 200 {
		t.Fatal(probe.Code)
	}
	metrics := httptest.NewRecorder()
	s.metrics.ServeHTTP(metrics, nil)
	if strings.Contains(metrics.Body.String(), "secret-id") {
		t.Fatal("unbounded metric labels")
	}
	if !strings.Contains(metrics.Body.String(), `route="GET /objects/{id}"`) || !strings.Contains(metrics.Body.String(), `_count{method="GET",route="GET /objects/{id}",status="401"} 2`) {
		t.Fatal(metrics.Body.String())
	}
	s.draining.Store(true)
	probe = httptest.NewRecorder()
	s.mux.ServeHTTP(probe, httptest.NewRequest("GET", "/healthz/ready", nil))
	if probe.Code != 503 {
		t.Fatal("draining workload remained ready")
	}
}

func TestShutdownWaitsForAcceptedRequest(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := listener.Addr().String()
	listener.Close()
	s := New(slog.New(slog.NewTextHandler(io.Discard, nil)), nil, nil, nil)
	entered, release := make(chan struct{}), make(chan struct{})
	s.HandleFunc("GET /slow", func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-release
		w.Write([]byte("durable-final"))
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- s.ListenAndServeContext(ctx, addr) }()
	client := &http.Client{Timeout: 3 * time.Second}
	deadline := time.Now().Add(3 * time.Second)
	for {
		response, err := client.Get("http://" + addr + "/healthz/ready")
		if err == nil {
			response.Body.Close()
			break
		}
		if time.Now().After(deadline) {
			t.Fatal(err)
		}
		time.Sleep(time.Millisecond)
	}
	result := make(chan string, 1)
	go func() {
		response, err := client.Get("http://" + addr + "/slow")
		if err != nil {
			result <- err.Error()
			return
		}
		defer response.Body.Close()
		body, _ := io.ReadAll(response.Body)
		result <- string(body)
	}()
	<-entered
	cancel()
	select {
	case err := <-done:
		t.Fatalf("server returned before accepted request drained: %v", err)
	case <-time.After(30 * time.Millisecond):
	}
	close(release)
	if body := <-result; body != "durable-final" {
		t.Fatal(body)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("shutdown did not finish")
	}
}

func TestTelemetryDropsWithoutBlockingBusiness(t *testing.T) {
	e := &traceExporter{pending: make(chan []byte, 1), metrics: NewRegistry(), service: "fixture"}
	now := time.Now()
	sc := spanContext{trace: strings.Repeat("a", 32), span: strings.Repeat("b", 16)}
	for range 1000 {
		e.record(sc, "GET /objects/{id}", "GET", 200, now, now)
	}
	if len(e.pending) != 1 {
		t.Fatal("unbounded queue")
	}
	metrics := httptest.NewRecorder()
	e.metrics.ServeHTTP(metrics, nil)
	if !strings.Contains(metrics.Body.String(), "hub_telemetry_dropped_total 999") {
		t.Fatal(metrics.Body.String())
	}
}

func TestW3CTracePropagationAndInvalidParent(t *testing.T) {
	ctx, span := startSpan(context.Background(), "00-"+strings.Repeat("a", 32)+"-"+strings.Repeat("b", 16)+"-01")
	if span.trace != strings.Repeat("a", 32) || span.parent != strings.Repeat("b", 16) {
		t.Fatal(span)
	}
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("traceparent") != "00-"+span.trace+"-"+span.span+"-01" {
			t.Error("context lost")
		}
	}))
	defer target.Close()
	r, _ := http.NewRequestWithContext(ctx, "GET", target.URL, nil)
	response, err := (&http.Client{Transport: TraceTransport{}}).Do(r)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	_, invalid := startSpan(context.Background(), "00-"+strings.Repeat("0", 32)+"-"+strings.Repeat("b", 16)+"-01")
	if invalid.parent != "" {
		t.Fatal("zero trace accepted")
	}
}
