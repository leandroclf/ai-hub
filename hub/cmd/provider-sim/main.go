// Comando provider-sim sobe o provedor externo simulado usado pelos
// ensaios locais de Cometa (QUA-01: "Simulador local reproduz
// comportamento deterministico e falhas injetadas").
package main

import (
	"net/http"

	"ai-hub/hub/internal/platform/config"
	"ai-hub/hub/internal/platform/httpserver"
	"ai-hub/hub/internal/platform/logging"
	"ai-hub/hub/internal/providersim"
)

func main() {
	log := logging.New("provider-sim")
	addr := config.Env("HTTP_ADDR", ":8090")

	srv := httpserver.New(log, nil, nil, nil)
	sim := providersim.NewServer()
	mux := http.NewServeMux()
	sim.Routes(mux)
	srv.Handle("/v1/", mux)

	if err := srv.ListenAndServe(addr); err != nil {
		log.Error("server stopped", "error", err)
	}
}
