// Comando orbita sobe a aplicacao de execucao: admissao, protocolo,
// idempotencia, despacho direto/em fila, deadline e consulta unificada
// (ARQ-01, EXE-*).
package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"ai-hub/hub/internal/atlasclient"
	"ai-hub/hub/internal/libraclient"
	"ai-hub/hub/internal/orbita"
	"ai-hub/hub/internal/outbox"
	"ai-hub/hub/internal/platform/config"
	"ai-hub/hub/internal/platform/httpserver"
	"ai-hub/hub/internal/platform/logging"
	"ai-hub/hub/internal/platform/pg"
	"ai-hub/hub/internal/queue"
)

func main() {
	log := logging.New("orbita")
	addr := config.Env("HTTP_ADDR", ":8080")
	dsn := config.Env("CORE_DSN", "postgres://hub:hub@localhost:5432/hub_core?sslmode=disable")
	atlasURL := config.Env("ATLAS_URL", "http://localhost:8081")
	libraURL := config.Env("LIBRA_URL", "http://localhost:8084")
	cometaURL := config.Env("COMETA_URL", "http://localhost:8082")
	queueEndpoint := config.Env("QUEUE_ENDPOINT", "http://localhost:4566")
	queueRegion := config.Env("QUEUE_REGION", "us-east-1")

	db, err := pg.WaitReady(dsn, 30_000_000_000)
	if err != nil {
		log.Error("nao foi possivel conectar a hub_core", "error", err)
		panic(err)
	}
	defer db.Close()

	store := orbita.NewStore(db)
	atlas := atlasclient.New(atlasURL, 30*time.Second)
	libra := libraclient.New(libraURL)
	finalizer := orbita.NewFinalizer(store, log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	q, err := queue.New(ctx, queueEndpoint, queueRegion)
	if err != nil {
		log.Error("falha ao conectar ao broker", "error", err)
		panic(err)
	}
	commandsQueueURL, err := q.EnsureQueue(ctx, "cometa-commands")
	if err != nil {
		log.Error("falha ao localizar fila de comandos do Cometa", "error", err)
		panic(err)
	}
	operationFactsTopicARN, err := q.EnsureTopic(ctx, "hub-operation-facts")
	if err != nil {
		log.Error("falha ao localizar topico de fatos de operacao", "error", err)
		panic(err)
	}
	operationFactsQueueURL, err := q.EnsureQueue(ctx, "orbita-operation-facts")
	if err != nil {
		log.Error("falha ao criar fila de fatos de operacao", "error", err)
		panic(err)
	}
	operationFactsQueueARN, err := q.QueueARN(ctx, operationFactsQueueURL)
	if err != nil {
		log.Error("falha ao obter ARN da fila de fatos de operacao", "error", err)
		panic(err)
	}
	_ = q.AllowSNSDelivery(ctx, operationFactsQueueURL, operationFactsQueueARN)
	_ = q.Subscribe(ctx, operationFactsTopicARN, operationFactsQueueARN)

	protocolFactsTopicARN, err := q.EnsureTopic(ctx, "hub-protocol-facts")
	if err != nil {
		log.Error("falha ao criar topico de fatos de protocolo", "error", err)
		panic(err)
	}

	dispatcher := orbita.NewDispatcher(cometaURL, q, commandsQueueURL)
	handlers := orbita.NewHandlers(store, atlas, libra, dispatcher, finalizer, log)

	// Relay da outbox (COM-03): publica "protocol.finalized" no SNS
	// para Pulsar (webhook) e Libra (receita), fora do caminho
	// obrigatorio de resposta ao cliente (DAD-09).
	go outbox.RunRelay(ctx, db, "protocol", func(ctx context.Context, row outbox.Row) error {
		var meta struct {
			TenantID string `json:"tenant_id"`
		}
		_ = json.Unmarshal(row.Payload, &meta)
		return q.PublishFact(ctx, protocolFactsTopicARN, queue.Envelope{
			EventID:          "protocol-fact-" + strconv.FormatInt(row.ID, 10),
			Type:             row.EventType,
			SchemaVersion:    1,
			Producer:         "orbita",
			TenantID:         meta.TenantID,
			ProtocolID:       row.AggregateID,
			OccurredAt:       time.Now().UTC(),
			RecordedAt:       time.Now().UTC(),
			AggregateVersion: 1,
			Payload:          row.Payload,
		})
	}, 1*time.Second, log)

	// Consumidor dos fatos de operacao (ASYNC/AUTO): finaliza o
	// protocolo quando Cometa observa o final externo.
	go orbita.RunOperationFactConsumer(ctx, q, operationFactsQueueURL, store, finalizer, log)

	// Temporizador de deadline (EXE-11).
	go orbita.RunDeadlineTimer(ctx, store, finalizer, 1*time.Second, log)

	readiness := func(ctx context.Context) error { return store.Ping(ctx) }
	srv := httpserver.New(log, readiness, readiness, nil)
	mux := http.NewServeMux()
	handlers.Register(mux)
	handlers.RegisterInternal(mux)
	srv.Handle("/v1/", mux)
	srv.Handle("/internal/", mux)

	if err := srv.ListenAndServe(addr); err != nil {
		log.Error("server stopped", "error", err)
	}
}
