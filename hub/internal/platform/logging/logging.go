// Package logging padroniza logs estruturados por componente
// (Portal/Atlas/Orbita/Cometa/Pulsar/Libra), conforme OPE-06: logs
// estruturados, sem protocol_id/event_id/dado pessoal como label de
// metrica (isso e responsabilidade do pacote metrics, nao deste).
//
// Nivel configuravel via LOG_LEVEL (debug|info|warn|error, default
// info) — permite capturar evidencia em nivel DEBUG sem recompilar.
// "Trace" nesta referencia significa correlacao por trace_id
// propagado em toda a cadeia (Orbita -> Cometa -> fatos -> Pulsar/
// Libra), logado como campo estruturado; nao e OpenTelemetry/spans
// distribuidos reais (ver IMPLEMENTATION_AUDIT.md).
package logging

import (
	"log/slog"
	"os"
	"strings"
)

func levelFromEnv() slog.Level {
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// New cria um logger JSON estruturado identificando o componente, com
// nivel controlado por LOG_LEVEL.
func New(component string) *slog.Logger {
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: levelFromEnv()})
	return slog.New(h).With("component", component)
}

// WithTrace anexa trace_id (correlacao ponta a ponta) e demais
// identificadores de negocio relevantes a um logger, sem que
// protocol_id/event_id virem label de metrica (isso continua restrito
// ao pacote metrics).
func WithTrace(log *slog.Logger, traceID string, kv ...any) *slog.Logger {
	args := append([]any{"trace_id", traceID}, kv...)
	return log.With(args...)
}
