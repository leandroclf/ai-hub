// Package httpserver fornece o esqueleto HTTP comum a todos os
// servicos: probes diferenciadas por dependencia (OPE-05: startup,
// readiness, liveness) e um endpoint /metrics de baixa cardinalidade
// (OPE-06). Nao expoe credenciais nem topologia interna nos probes.
package httpserver

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
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
	s.mux.HandleFunc("/healthz/startup", s.probeHandler(startup))
	s.mux.HandleFunc("/healthz/ready", s.probeHandler(readiness))
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
	s.mux.Handle(pattern, handler)
}

// HandleFunc registra uma rota de negocio no mux subjacente.
func (s *Server) HandleFunc(pattern string, handler http.HandlerFunc) {
	s.mux.HandleFunc(pattern, handler)
}

// ListenAndServe sobe o servidor HTTP com timeouts explicitos (nenhuma
// espera bloqueante indefinida no caminho critico, conforme EXE-13).
func (s *Server) ListenAndServe(addr string) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           s.withLogging(s.mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	s.log.Info("http server starting", "addr", addr)
	return srv.ListenAndServe()
}

func (s *Server) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		// Dimensoes permitidas em log (nao em metrica): metodo, rota,
		// status, duracao. Sem protocol_id/tenant/dado pessoal aqui.
		s.log.Info("http_request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
		// Métricas deliberadamente sem path, tenant, protocol_id ou
		// qualquer dado de negócio: URLs com IDs gerariam cardinalidade
		// ilimitada. Método e status são suficientes para saúde operacional.
		s.metrics.Inc("http_requests_total", map[string]string{
			"method": r.Method,
			"status": http.StatusText(rec.status),
		})
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}
