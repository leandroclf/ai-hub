// Comando orbita sobe a aplicacao de execucao: admissao, protocolo,
// idempotencia, despacho direto/em fila, deadline e consulta unificada
// (ARQ-01, EXE-*).
package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai-hub/hub/internal/atlasclient"
	"ai-hub/hub/internal/libraclient"
	"ai-hub/hub/internal/objectstore"
	"ai-hub/hub/internal/orbita"
	"ai-hub/hub/internal/outbox"
	"ai-hub/hub/internal/platform/auth"
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
	if tenant := config.Env("RUNTIME_TENANT_ID", ""); tenant != "" {
		var dsnErr error
		dsn, dsnErr = pg.RuntimeDSN(dsn, tenant)
		if dsnErr != nil {
			panic(dsnErr)
		}
	}
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

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	q, err := queue.New(ctx, queueEndpoint, queueRegion)
	if err != nil {
		log.Error("falha ao conectar ao broker", "error", err)
		panic(err)
	}
	dispatcher := orbita.NewDispatcher(cometaURL, q, "")
	handlers := orbita.NewHandlers(store, atlas, libra, dispatcher, finalizer, log)
	go queue.RunBootstrap(ctx, func() error {
		commandsQueueURL, err := q.EnsureQueue(ctx, "cometa-commands")
		if err != nil {
			log.Error("falha ao localizar fila de comandos do Cometa", "error", err)
			return err
		}
		operationFactsTopicARN, err := q.EnsureTopic(ctx, "hub-operation-facts")
		if err != nil {
			log.Error("falha ao localizar topico de fatos de operacao", "error", err)
			return err
		}
		operationFactsQueueURL, err := q.EnsureQueue(ctx, "orbita-operation-facts")
		if err != nil {
			log.Error("falha ao criar fila de fatos de operacao", "error", err)
			return err
		}
		operationFactsQueueARN, err := q.QueueARN(ctx, operationFactsQueueURL)
		if err != nil {
			log.Error("falha ao obter ARN da fila de fatos de operacao", "error", err)
			return err
		}
		if err := q.AllowSNSDelivery(ctx, operationFactsQueueURL, operationFactsQueueARN, operationFactsTopicARN); err != nil {
			return err
		}
		if err := q.Subscribe(ctx, operationFactsTopicARN, operationFactsQueueARN); err != nil {
			return err
		}

		protocolFactsTopicARN, err := q.EnsureTopic(ctx, "hub-protocol-facts")
		if err != nil {
			log.Error("falha ao criar topico de fatos de protocolo", "error", err)
			return err
		}

		dispatcher.SetQueueURL(commandsQueueURL)

		// Relay da outbox (COM-03): publica "protocol.finalized" no SNS
		// para Pulsar (webhook) e Libra (receita), fora do caminho
		// obrigatorio de resposta ao cliente (DAD-09).
		go outbox.RunRelay(ctx, db, "protocol", func(ctx context.Context, row outbox.Row) error {
			var meta struct {
				TenantID string `json:"tenant_id"`
			}
			_ = json.Unmarshal(row.Payload, &meta)
			return q.PublishFact(ctx, protocolFactsTopicARN, queue.Envelope{
				EventID:          row.EventID,
				Type:             row.EventType,
				SchemaVersion:    1,
				Producer:         "orbita",
				TenantID:         meta.TenantID,
				ProtocolID:       row.AggregateID,
				OccurredAt:       row.OccurredAt,
				RecordedAt:       row.RecordedAt,
				AggregateVersion: 1,
				Payload:          row.Payload,
			})
		}, 1*time.Second, log)

		// Consumidor dos fatos de operacao (ASYNC/AUTO): finaliza o
		// protocolo quando Cometa observa o final externo.
		go orbita.RunOperationFactConsumer(ctx, q, operationFactsQueueURL, store, finalizer, log)

		return nil
	}, log)

	// Temporizador de deadline (EXE-11).
	go orbita.RunDeadlineTimer(ctx, store, finalizer, 1*time.Second, log)
	go orbita.RunIntentPublisher(ctx, store, dispatcher, config.Env("CELL_ID", ""), log)
	go orbita.RunDirectRecovery(ctx, store, dispatcher, finalizer, config.Env("CELL_ID", ""), log)
	go orbita.RunReservationRecovery(ctx, store, libra, config.Env("CELL_ID", ""), log)

	readiness := func(ctx context.Context) error { return store.Ping(ctx) }
	srv := httpserver.New(log, readiness, readiness, nil)
	srv.AuthMiddleware = auth.FromEnv().Middleware
	mux := http.NewServeMux()
	handlers.Register(mux)
	objects, objectErr := objectstore.New(ctx, config.Env("S3_ENDPOINT", ""), queueRegion, config.Env("S3_BUCKET", "r2-custody"))
	if objectErr != nil {
		log.Error("object client unavailable")
	} else {
		fileCatalog := objectstore.NewCatalog(db, objects)
		handlers.SetFileCatalog(fileCatalog)
		objectstore.NewHandlers(fileCatalog).Register(mux)
		// Expurgo é opt-in e exige tenants explícitos. Isso evita que um
		// processo de runtime obtenha escopo global por configuração implícita.
		retentionTenants := objectstore.TenantsFromEnv(config.Env("RETENTION_TENANTS", ""))
		if len(retentionTenants) > 0 {
			retentionActor := config.Env("RETENTION_ACTOR", "retention-worker")
			go objectstore.RunRetentionWorker(ctx, fileCatalog, retentionTenants, retentionActor, config.EnvDurationSeconds("RETENTION_INTERVAL_SECONDS", 60), log)
		}
		if config.Env("ENVIRONMENT", "") == "local" {
			go queue.RunBootstrap(ctx, func() error { return objects.EnsureBucket(ctx) }, log)
		}
	}
	handlers.RegisterInternal(mux)
	handlers.RegisterAdmin(mux)
	srv.Handle("/v1/", mux)
	srv.Handle("/internal/", mux)
	srv.Handle("/admin/v1/", mux)

	if err := srv.ListenAndServe(addr); err != nil {
		log.Error("server stopped", "error", err)
	}
}
