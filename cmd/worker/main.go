// Command worker consumes queued sends from NATS and dispatches them to providers.
package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"

	"xoxa-message-gateway/internal/config"
	"xoxa-message-gateway/internal/repo"
	"xoxa-message-gateway/internal/service"
	"xoxa-message-gateway/internal/telemetry"
	"xoxa-message-gateway/internal/transport"
	"xoxa-message-gateway/migrations"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	shutdownTracing, err := telemetry.InitTracing(ctx, cfg.ServiceName+"-worker", cfg.OTELExporterEndpoint)
	if err != nil {
		log.Fatalf("telemetry: %v", err)
	}
	defer shutdownTracing(context.Background())

	db, err := repo.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	if err := repo.Migrate(ctx, db, migrations.Files); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		log.Fatalf("nats: %v", err)
	}
	defer nc.Close()

	transport.Register("sms", transport.NewTwilioTransport(cfg.TwilioAccountSID, cfg.TwilioAuthToken, cfg.TwilioFromNumber, ""))
	transport.Register("whatsapp", transport.NewWhatsAppTransport(cfg.WhatsAppToken, cfg.WhatsAppPhoneID, cfg.WhatsAppAPIBase))

	go func() {
		if err := telemetry.ServeMetrics(ctx, ":"+cfg.MetricsPort); err != nil {
			log.Printf("worker: metrics server error: %v", err)
		}
	}()

	processor := service.NewProcessor(repo.NewMessageRepository(db), nc, cfg.MaxSendAttempts, cfg.BaseRetryDelay)

	sub, err := nc.Subscribe(service.SubjectSend, func(m *nats.Msg) {
		if err := processor.Process(ctx, m.Data); err != nil {
			log.Printf("worker: process job failed: %v", err)
		}
	})
	if err != nil {
		log.Fatalf("nats: subscribe: %v", err)
	}
	defer sub.Unsubscribe()

	log.Printf("worker: subscribed to %s", service.SubjectSend)
	<-ctx.Done()
	log.Println("worker: shutting down")
}
