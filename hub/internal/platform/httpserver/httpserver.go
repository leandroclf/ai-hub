// Package httpserver fornece o esqueleto HTTP comum a todos os
// servicos: probes diferenciadas por dependencia (OPE-05: startup,
// readiness, liveness) e um endpoint /metrics de baixa cardinalidade
// (OPE-06). Nao expoe credenciais nem topologia interna nos probes.
package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"
)

// CheckFunc verifica uma dependencia especifica (ex.: conseguir
// persistir/consultar o banco). Retornar erro marca o probe como
// indisponivel sem derrubar o processo inteiro por si so.
type CheckFunc func(ctx context.Context) error

// Server encapsula um *http.ServeMux com probes e metricas jah
// registrados, deixando o servico livre para registrar suas proprias
// rotas de negocio via Handle/HandleFunc.
type Server struct {
	mux       *http.ServeMux
	log       *slog.Logger
	startup   CheckFunc
	readiness CheckFunc
	liveness  CheckFunc
	metrics   *Registry
	traces    *traceExporter
	draining  atomic.Bool
	// AuthMiddleware is applied to business routes only. Probes remain local.
	AuthMiddleware func(http.Handler) http.Handler
}

// New cria um Server com as tres probes (OPE-05). Qualquer CheckFunc
// pode ser nil, o que a torna sempre bem-sucedida (ex.: liveness nao
// deve depender de provedor/broker externo).
func New(log *slog.Logger, startup, readiness, liveness CheckFunc) *Server {
	s := &Server{
		mux:       http.NewServeMux(),
		log:       log,
		startup:   startup,
		readiness: readiness,
		liveness:  liveness,
		metrics:   NewRegistry(),
	}
	s.traces = newTraceExporter(s.metrics)
	s.mux.HandleFunc("/healthz/startup", s.probeHandler(startup))
	s.mux.HandleFunc("/healthz/ready", s.probeHandler(func(ctx context.Context) error {
		if s.draining.Load() {
			return errors.New("draining")
		}
		if readiness != nil {
			return readiness(ctx)
		}
		return nil
	}))
	s.mux.HandleFunc("/healthz/live", s.probeHandler(liveness))
	s.mux.HandleFunc("/metrics", s.metrics.ServeHTTP)
	return s
}

func (s *Server) probeHandler(check CheckFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if check == nil {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok\n"))
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := check(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "unavailable"})
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	}
}

// Metrics retorna o registry para os servicos incrementarem contadores
// de negocio (admissoes, recusas por causa, operacoes UNKNOWN, etc.).
func (s *Server) Metrics() *Registry { return s.metrics }

// Handle registra uma rota de negocio no mux subjacente.
func (s *Server) Handle(pattern string, handler http.Handler) {
	if s.AuthMiddleware != nil {
		handler = s.AuthMiddleware(handler)
	}
	s.mux.Handle(pattern, handler)
}

// HandleFunc registra uma rota de negocio no mux subjacente.
func (s *Server) HandleFunc(pattern string, handler http.HandlerFunc) {
	s.Handle(pattern, handler)
}

// ListenAndServe sobe o servidor HTTP com timeouts explicitos (nenhuma
// espera bloqueante indefinida no caminho critico, conforme EXE-13).
func (s *Server) ListenAndServe(addr string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return s.ListenAndServeContext(ctx, addr)
}

// ListenAndServeContext drains accepted HTTP requests before returning on shutdown.
// Workers must stop acquiring leases on their own lifecycle context.
func (s *Server) ListenAndServeContext(ctx context.Context, addr string) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           s.withLogging(s.mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	done := make(chan struct{})
	drained := make(chan struct{})
	go func() {
		defer close(drained)
		select {
		case <-ctx.Done():
		case <-done:
			return
		}
		s.draining.Store(true)
		shutdown, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdown); err != nil {
			s.log.Error("http drain exceeded grace", "error", err)
			_ = srv.Close()
		}
	}()
	s.log.Info("http server starting", "addr", addr)
	err := srv.ListenAndServe()
	close(done)
	<-drained
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx, span := startSpan(r.Context(), r.Header.Get("traceparent"))
		r = r.WithContext(ctx)
		w.Header().Set("traceparent", "00-"+span.trace+"-"+span.span+"-01")
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		route := r.Pattern
		if route == "" {
			route = "unmatched"
		}
		s.metrics.Observe("hub_http_duration_seconds", map[string]string{"method": safeMethod(r.Method), "route": route, "status": strconv.Itoa(rec.status)}, time.Since(start).Seconds())
		s.traces.record(span, route, safeMethod(r.Method), rec.status, start, time.Now())
		// Dimensoes permitidas em log (nao em metrica): metodo, rota,
		// status, duracao. Sem protocol_id/tenant/dado pessoal aqui.
		s.log.Info("http_request",
			"method", r.Method,
			"route", route,
			"trace_id", span.trace,
			"status", rec.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
		// Métricas deliberadamente sem path, tenant, protocol_id ou
		// qualquer dado de negócio: URLs com IDs gerariam cardinalidade
		// ilimitada. Método e status são suficientes para saúde operacional.
		s.metrics.Inc("http_requests_total", map[string]string{
			"method": safeMethod(r.Method),
			"status": http.StatusText(rec.status),
		})
	})
}

func safeMethod(method string) string {
	switch method {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD":
		return method
	}
	return "OTHER"
}

type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(code int) {
	if r.wroteHeader {
		return
	}
	r.wroteHeader = true
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(p []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	return r.ResponseWriter.Write(p)
}

func (r *statusRecorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }
