// Comando libra sobe a aplicacao financeira: medicao economica,
// reservas estritas e ledger gerencial (FIN-04/06/07).
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai-hub/hub/internal/atlasclient"
	"ai-hub/hub/internal/libra"
	"ai-hub/hub/internal/platform/auth"
	"ai-hub/hub/internal/platform/config"
	"ai-hub/hub/internal/platform/httpserver"
	"ai-hub/hub/internal/platform/logging"
	"ai-hub/hub/internal/platform/pg"
	"ai-hub/hub/internal/queue"
)

func main() {
	log := logging.New("libra")
	addr := config.Env("HTTP_ADDR", ":8084")
	dsn := config.Env("FINANCE_DSN", "postgres://hub:hub@localhost:5432/hub_finance?sslmode=disable")
	atlasURL := config.Env("ATLAS_URL", "http://localhost:8081")
	queueEndpoint := config.Env("QUEUE_ENDPOINT", "http://localhost:4566")
	queueRegion := config.Env("QUEUE_REGION", "us-east-1")

	db, err := pg.WaitReady(dsn, 30_000_000_000)
	if err != nil {
		log.Error("nao foi possivel conectar a hub_finance", "error", err)
		panic(err)
	}
	defer db.Close()

	store := libra.NewStore(db)
	atlas := atlasclient.New(atlasURL, 30*time.Second)
	handlers := libra.NewHandlers(store, log)

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
		revenueQueueURL, err := q.EnsureQueue(ctx, "libra-revenue-facts")
		if err != nil {
			log.Error("falha ao criar fila de receita", "error", err)
			return err
		}
		revenueQueueARN, err := q.QueueARN(ctx, revenueQueueURL)
		if err != nil {
			log.Error("falha ao obter ARN da fila de receita", "error", err)
			return err
		}
		if err := q.AllowSNSDelivery(ctx, revenueQueueURL, revenueQueueARN, protocolFactsTopicARN); err != nil {
			return err
		}
		if err := q.Subscribe(ctx, protocolFactsTopicARN, revenueQueueARN); err != nil {
			return err
		}

		operationFactsTopicARN, err := q.EnsureTopic(ctx, "hub-operation-facts")
		if err != nil {
			log.Error("falha ao localizar topico de fatos de operacao", "error", err)
			return err
		}
		costQueueURL, err := q.EnsureQueue(ctx, "libra-cost-facts")
		if err != nil {
			log.Error("falha ao criar fila de custo", "error", err)
			return err
		}
		costQueueARN, err := q.QueueARN(ctx, costQueueURL)
		if err != nil {
			log.Error("falha ao obter ARN da fila de custo", "error", err)
			return err
		}
		if err := q.AllowSNSDelivery(ctx, costQueueURL, costQueueARN, operationFactsTopicARN); err != nil {
			return err
		}
		if err := q.Subscribe(ctx, operationFactsTopicARN, costQueueARN); err != nil {
			return err
		}

		go libra.RunRevenueConsumer(ctx, q, revenueQueueURL, store, atlas, log)
		go libra.RunCostConsumer(ctx, q, costQueueURL, store, log)

		return nil
	}, log)

	readiness := func(ctx context.Context) error { return store.Ping(ctx) }
	srv := httpserver.New(log, readiness, readiness, nil)
	srv.AuthMiddleware = auth.FromEnv().Middleware
	mux := http.NewServeMux()
	handlers.Register(mux)
	srv.Handle("/internal/", mux)
	srv.Handle("/admin/v1/finance/", mux)

	if err := srv.ListenAndServe(addr); err != nil {
		log.Error("server stopped", "error", err)
	}
}
