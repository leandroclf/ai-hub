package httpserver

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type spanContext struct{ trace, span, parent string }
type spanKey struct{}

func randomHex(size int) string {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func startSpan(ctx context.Context, header string) (context.Context, spanContext) {
	sc := spanContext{trace: randomHex(16), span: randomHex(8)}
	parts := strings.Split(header, "-")
	if len(parts) == 4 && parts[0] == "00" && len(parts[1]) == 32 && len(parts[2]) == 16 && len(parts[3]) == 2 {
		_, e1 := hex.DecodeString(parts[1])
		_, e2 := hex.DecodeString(parts[2])
		_, e3 := hex.DecodeString(parts[3])
		if e1 == nil && e2 == nil && e3 == nil && parts[1] != strings.Repeat("0", 32) && parts[2] != strings.Repeat("0", 16) {
			sc.trace = parts[1]
			sc.parent = parts[2]
		}
	}
	return context.WithValue(ctx, spanKey{}, sc), sc
}

// TraceTransport propagates W3C context without carrying identity or payload.
// Authentication is supplied by the domain's authenticated transport separately.
type TraceTransport struct{ Base http.RoundTripper }

func (t TraceTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	if sc, ok := r.Context().Value(spanKey{}).(spanContext); ok {
		r = r.Clone(r.Context())
		r.Header = r.Header.Clone()
		r.Header.Set("traceparent", "00-"+sc.trace+"-"+sc.span+"-01")
	}
	return base.RoundTrip(r)
}

type traceExporter struct {
	pending           chan []byte
	metrics           *Registry
	endpoint, service string
}

func newTraceExporter(metrics *Registry) *traceExporter {
	endpoint := strings.TrimRight(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"), "/")
	if endpoint == "" {
		return nil
	}
	e := &traceExporter{pending: make(chan []byte, 128), metrics: metrics, endpoint: endpoint + "/v1/traces", service: os.Getenv("OTEL_SERVICE_NAME")}
	go e.run()
	return e
}
func (e *traceExporter) run() {
	client := &http.Client{Timeout: time.Second}
	for payload := range e.pending {
		request, err := http.NewRequest(http.MethodPost, e.endpoint, bytes.NewReader(payload))
		if err == nil {
			request.Header.Set("Content-Type", "application/json")
			response, requestErr := client.Do(request)
			err = requestErr
			if response != nil {
				io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
				response.Body.Close()
				if response.StatusCode >= 300 {
					e.metrics.Inc("hub_telemetry_dropped_total", nil)
				}
			}
		}
		if err != nil {
			e.metrics.Inc("hub_telemetry_dropped_total", nil)
		}
	}
}
func (e *traceExporter) record(sc spanContext, route, method string, status int, start time.Time, end time.Time) {
	if e == nil {
		return
	}
	span := map[string]any{"traceId": sc.trace, "spanId": sc.span, "name": route, "kind": 2, "startTimeUnixNano": strconv.FormatInt(start.UnixNano(), 10), "endTimeUnixNano": strconv.FormatInt(end.UnixNano(), 10), "attributes": []any{map[string]any{"key": "http.request.method", "value": map[string]string{"stringValue": method}}, map[string]any{"key": "http.response.status_code", "value": map[string]int{"intValue": status}}}}
	if sc.parent != "" {
		span["parentSpanId"] = sc.parent
	}
	if status >= 500 {
		span["status"] = map[string]int{"code": 2}
	}
	payload, err := json.Marshal(map[string]any{"resourceSpans": []any{map[string]any{"resource": map[string]any{"attributes": []any{map[string]any{"key": "service.name", "value": map[string]string{"stringValue": e.service}}}}, "scopeSpans": []any{map[string]any{"scope": map[string]string{"name": "ai-hub/httpserver"}, "spans": []any{span}}}}}})
	if err != nil {
		e.metrics.Inc("hub_telemetry_dropped_total", nil)
		return
	}
	select {
	case e.pending <- payload:
	default:
		e.metrics.Inc("hub_telemetry_dropped_total", nil)
	}
}
