// Comando libra sobe a aplicacao financeira: medicao economica,
// reservas estritas e ledger gerencial (FIN-04/06/07).
package main

import (
	"context"
	"net/http"
	"time"

	"ai-hub/hub/internal/atlasclient"
	"ai-hub/hub/internal/libra"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	q, err := queue.New(ctx, queueEndpoint, queueRegion)
	if err != nil {
		log.Error("falha ao conectar ao broker", "error", err)
		panic(err)
	}

	protocolFactsTopicARN, err := q.EnsureTopic(ctx, "hub-protocol-facts")
	if err != nil {
		log.Error("falha ao localizar topico de fatos de protocolo", "error", err)
		panic(err)
	}
	revenueQueueURL, err := q.EnsureQueue(ctx, "libra-revenue-facts")
	if err != nil {
		log.Error("falha ao criar fila de receita", "error", err)
		panic(err)
	}
	revenueQueueARN, err := q.QueueARN(ctx, revenueQueueURL)
	if err != nil {
		log.Error("falha ao obter ARN da fila de receita", "error", err)
		panic(err)
	}
	_ = q.AllowSNSDelivery(ctx, revenueQueueURL, revenueQueueARN)
	_ = q.Subscribe(ctx, protocolFactsTopicARN, revenueQueueARN)

	operationFactsTopicARN, err := q.EnsureTopic(ctx, "hub-operation-facts")
	if err != nil {
		log.Error("falha ao localizar topico de fatos de operacao", "error", err)
		panic(err)
	}
	costQueueURL, err := q.EnsureQueue(ctx, "libra-cost-facts")
	if err != nil {
		log.Error("falha ao criar fila de custo", "error", err)
		panic(err)
	}
	costQueueARN, err := q.QueueARN(ctx, costQueueURL)
	if err != nil {
		log.Error("falha ao obter ARN da fila de custo", "error", err)
		panic(err)
	}
	_ = q.AllowSNSDelivery(ctx, costQueueURL, costQueueARN)
	_ = q.Subscribe(ctx, operationFactsTopicARN, costQueueARN)

	go libra.RunRevenueConsumer(ctx, q, revenueQueueURL, store, atlas, log)
	go libra.RunCostConsumer(ctx, q, costQueueURL, store, log)

	readiness := func(ctx context.Context) error { return store.Ping(ctx) }
	srv := httpserver.New(log, readiness, readiness, nil)
	mux := http.NewServeMux()
	handlers.Register(mux)
	srv.Handle("/internal/", mux)

	if err := srv.ListenAndServe(addr); err != nil {
		log.Error("server stopped", "error", err)
	}
}
