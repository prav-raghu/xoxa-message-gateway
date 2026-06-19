// Command api starts the REST + gRPC servers for the message gateway.
package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"

	"xoxa-message-gateway/internal/config"
	grpcserver "xoxa-message-gateway/internal/grpc"
	httpapi "xoxa-message-gateway/internal/http"
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

	shutdownTracing, err := telemetry.InitTracing(ctx, cfg.ServiceName, cfg.OTELExporterEndpoint)
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

	svc := service.New(repo.NewMessageRepository(db), nc)

	r := gin.Default()
	httpapi.RegisterHandlers(r, svc, cfg)
	srv := &http.Server{Addr: ":" + cfg.Port, Handler: r}

	go func() {
		if err := grpcserver.Serve(ctx, ":"+cfg.GRPCPort, svc); err != nil {
			log.Printf("grpc: server error: %v", err)
		}
	}()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("http: shutdown error: %v", err)
		}
	}()

	log.Printf("api: listening on :%s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("http: server error: %v", err)
	}
}
