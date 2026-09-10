// Comando cometa sobe a aplicacao de integracao com provedores
// externos: despacho direto (SYNC), worker de fila (ASYNC/AUTO) e
// scheduler de polling (EXE-04/EXE-05, COM-06).
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
	"ai-hub/hub/internal/cometa"
	"ai-hub/hub/internal/outbox"
	"ai-hub/hub/internal/platform/auth"
	"ai-hub/hub/internal/platform/config"
	"ai-hub/hub/internal/platform/httpserver"
	"ai-hub/hub/internal/platform/logging"
	"ai-hub/hub/internal/platform/pg"
	"ai-hub/hub/internal/providerauth"
	"ai-hub/hub/internal/queue"
)

func main() {
	log := logging.New("cometa")
	addr := config.Env("HTTP_ADDR", ":8082")
	selfURL := config.Env("SELF_URL", "http://localhost:8082")
	dsn := config.Env("CORE_DSN", "postgres://hub:hub@localhost:5432/hub_core?sslmode=disable")
	if tenant := config.Env("RUNTIME_TENANT_ID", ""); tenant != "" {
		var dsnErr error
		dsn, dsnErr = pg.RuntimeDSN(dsn, tenant)
		if dsnErr != nil {
			panic(dsnErr)
		}
	}
	atlasURL := config.Env("ATLAS_URL", "http://localhost:8081")
	queueEndpoint := config.Env("QUEUE_ENDPOINT", "http://localhost:4566")
	queueRegion := config.Env("QUEUE_REGION", "us-east-1")
	redisAddr := config.Env("REDIS_ADDR", "localhost:6379")

	db, err := pg.WaitReady(dsn, 30_000_000_000)
	if err != nil {
		log.Error("nao foi possivel conectar a hub_core", "error", err)
		panic(err)
	}
	defer db.Close()

	store := cometa.NewStore(db)
	atlas := atlasclient.New(atlasURL, 30*time.Second)
	tokenCache := providerauth.NewTokenCache(redisAddr)
	defer tokenCache.Close()
	exec := cometa.NewExecutor(store, atlas, log, selfURL, tokenCache)
	handlers := cometa.NewHandlers(exec, store)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	q, err := queue.New(ctx, queueEndpoint, queueRegion)
	if err != nil {
		log.Error("falha ao conectar ao broker", "error", err)
		panic(err)
	}
	go queue.RunBootstrap(ctx, func() error {
		commandsQueueURL, err := q.EnsureQueue(ctx, "cometa-commands")
		if err != nil {
			log.Error("falha ao criar fila de comandos", "error", err)
			return err
		}
		factsTopicARN, err := q.EnsureTopic(ctx, "hub-operation-facts")
		if err != nil {
			log.Error("falha ao criar topico de fatos", "error", err)
			return err
		}

		// Relay da outbox (COM-03): publica fatos de operacao no SNS
		// somente apos o commit local ja realizado pelo Executor.
		go outbox.RunRelay(ctx, db, "operation", func(ctx context.Context, row outbox.Row) error {
			var meta struct {
				ProtocolID string `json:"protocol_id"`
				TenantID   string `json:"tenant_id"`
			}
			_ = json.Unmarshal(row.Payload, &meta)
			return q.PublishFact(ctx, factsTopicARN, queue.Envelope{
				EventID:          row.EventID,
				Type:             row.EventType,
				SchemaVersion:    1,
				Producer:         "cometa",
				ProtocolID:       meta.ProtocolID,
				TenantID:         meta.TenantID,
				OccurredAt:       row.OccurredAt,
				RecordedAt:       row.RecordedAt,
				AggregateVersion: 1,
				Payload:          row.Payload,
			})
		}, 1*time.Second, log)

		// Worker ASYNC/AUTO: consome comandos QUEUED da fila dedicada.
		go cometa.RunCommandWorker(ctx, q, commandsQueueURL, exec, log)

		return nil
	}, log)

	// Scheduler de polling (EXE-05).
	go cometa.RunPoller(ctx, store, exec, atlas, 2*time.Second, log)
	go cometa.RunCallbackInboxWorker(ctx, store, exec, 2*time.Second, log)
	// Reconciliação administrativa consulta somente o status já correlacionado
	// no provedor; nunca reenvia a submissão original.
	go cometa.RunReconciliationWorker(ctx, store, exec, config.Env("CELL_ID", ""), 2*time.Second, log)

	readiness := func(ctx context.Context) error { return store.Ping(ctx) }
	srv := httpserver.New(log, readiness, readiness, nil)
	srv.AuthMiddleware = auth.FromEnv().Middleware
	mux := http.NewServeMux()
	handlers.RegisterInternal(mux)
	srv.Handle("/internal/", mux)
	callbackMux := http.NewServeMux()
	handlers.RegisterCallback(callbackMux)
	srv.HandlePublic("/callbacks/", callbackMux)

	if err := srv.ListenAndServe(addr); err != nil {
		log.Error("server stopped", "error", err)
	}
}
