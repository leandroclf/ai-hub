// Package logging padroniza logs estruturados por componente
// (Portal/Atlas/Orbita/Cometa/Pulsar/Libra), conforme OPE-06: logs
// estruturados, sem protocol_id/event_id/dado pessoal como label de
// metrica (isso e responsabilidade do pacote metrics, nao deste).
package logging

import (
	"log/slog"
	"os"
)

// New cria um logger JSON estruturado identificando o componente.
func New(component string) *slog.Logger {
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	return slog.New(h).With("component", component)
}
