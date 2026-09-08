// Comando atlas sobe o plano de controle: catalogo, contas de
// provedor, vinculos de credencial e contratos (ARQ-01).
package main

import (
	"context"
	"net/http"

	"ai-hub/hub/internal/atlas"
	"ai-hub/hub/internal/platform/auth"
	"ai-hub/hub/internal/platform/config"
	"ai-hub/hub/internal/platform/httpserver"
	"ai-hub/hub/internal/platform/logging"
	"ai-hub/hub/internal/platform/pg"
)

func main() {
	log := logging.New("atlas")
	addr := config.Env("HTTP_ADDR", ":8081")
	dsn := config.Env("CONTROL_DSN", "postgres://hub:hub@localhost:5432/hub_control?sslmode=disable")

	db, err := pg.WaitReady(dsn, 30_000_000_000) // 30s
	if err != nil {
		log.Error("nao foi possivel conectar a hub_control", "error", err)
		panic(err)
	}
	defer db.Close()

	store := atlas.NewStore(db)
	handlers := atlas.NewHandlers(store)

	readiness := func(ctx context.Context) error { return store.Ping(ctx) }
	srv := httpserver.New(log, readiness, readiness, nil)
	srv.AuthMiddleware = auth.FromEnv().Middleware
	mux := http.NewServeMux()
	handlers.Register(mux)
	srv.Handle("/v1/", mux)
	srv.Handle("/admin/v1/", mux)
	srv.HandleFunc("/admin/v1/me", auth.Me)

	if err := srv.ListenAndServe(addr); err != nil {
		log.Error("server stopped", "error", err)
	}
}
