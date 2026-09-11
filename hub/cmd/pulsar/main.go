// Comando pulsar sobe a aplicacao de entrega: obrigacoes de webhook,
// agenda, assinatura e tentativas (EXE-08).
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai-hub/hub/internal/cometa"
	"ai-hub/hub/internal/platform/auth"
	"ai-hub/hub/internal/platform/config"
	"ai-hub/hub/internal/platform/httpserver"
	"ai-hub/hub/internal/platform/logging"
	"ai-hub/hub/internal/platform/pg"
	"ai-hub/hub/internal/pulsar"
	"ai-hub/hub/internal/queue"
)

func main() {
	log := logging.New("pulsar")
	addr := config.Env("HTTP_ADDR", ":8083")
	dsn := config.Env("CORE_DSN", "postgres://hub:hub@localhost:5432/hub_core?sslmode=disable")
	if tenant := config.Env("RUNTIME_TENANT_ID", ""); tenant != "" {
		var dsnErr error
		dsn, dsnErr = pg.RuntimeDSN(dsn, tenant)
		if dsnErr != nil {
			panic(dsnErr)
		}
	}
	orbitaURL := config.Env("ORBITA_URL", "http://localhost:8080")
	queueEndpoint := config.Env("QUEUE_ENDPOINT", "http://localhost:4566")
	queueRegion := config.Env("QUEUE_REGION", "us-east-1")

	db, err := pg.WaitReady(dsn, 30_000_000_000)
	if err != nil {
		log.Error("nao foi possivel conectar a hub_core", "error", err)
		panic(err)
	}
	defer db.Close()

	store := pulsar.NewStore(db)
	handlers := pulsar.NewHandlers(store)
	capacity := cometa.NewCapacityController(db)
	if err := cometa.InstallPoliciesFromEnv(context.Background(), capacity); err != nil {
		log.Error("falha ao instalar politica de capacidade do pulsar", "error", err)
		panic(err)
	}
	worker := pulsar.NewDeliveryWorker(store, orbitaURL, log)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	q, err := queue.New(ctx, queueEndpoint, queueRegion)
	if err != nil {
		log.Error("falha ao conectar ao broker", "error", err)
		panic(err)
	}
	go queue.RunBootstrap(ctx, func() error {
		protocolFactsTopicARN, err := q.EnsureTopic(ctx, "hub-protocol-facts")
		if err != nil {
			log.Error("falha ao localizar topico de fatos de protocolo", "error", err)
			return err
		}
		factsQueueURL, err := q.EnsureQueue(ctx, "pulsar-protocol-facts")
		if err != nil {
			log.Error("falha ao criar fila de fatos de protocolo", "error", err)
			return err
		}
		factsQueueARN, err := q.QueueARN(ctx, factsQueueURL)
		if err != nil {
			log.Error("falha ao obter ARN da fila", "error", err)
			return err
		}
		if err := q.AllowSNSDelivery(ctx, factsQueueURL, factsQueueARN, protocolFactsTopicARN); err != nil {
			return err
		}
		if err := q.Subscribe(ctx, protocolFactsTopicARN, factsQueueARN); err != nil {
			return err
		}

		go pulsar.RunFactConsumer(ctx, q, factsQueueURL, store, log)
		return nil
	}, log)

	go worker.Run(ctx, 1*time.Second)

	readiness := func(ctx context.Context) error { return store.Ping(ctx) }
	srv := httpserver.New(log, readiness, readiness, nil)
	srv.AuthMiddleware = auth.FromEnv().Middleware
	mux := http.NewServeMux()
	handlers.Register(mux)
	srv.Handle("/internal/", mux)
	srv.Handle("/admin/v1/", mux)

	if err := srv.ListenAndServe(addr); err != nil {
		log.Error("server stopped", "error", err)
	}
}
