// Package config loads application configuration from environment variables.
package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration for the API and worker processes.
type Config struct {
	Port        string
	GRPCPort    string
	MetricsPort string

	DatabaseURL string
	NatsURL     string

	JWTIssuer    string
	JWTAudience  string
	JWTSecret    string // HS256 shared secret, used if set
	JWTPublicKey string // path to RSA public key, used if JWTSecret is empty

	OTELExporterEndpoint string
	ServiceName          string

	TwilioAccountSID string
	TwilioAuthToken  string
	TwilioFromNumber string
	WhatsAppToken    string
	WhatsAppPhoneID  string
	WhatsAppAPIBase  string

	MaxSendAttempts int
	BaseRetryDelay  time.Duration
}

// Load reads configuration from a local .env file (if present) and the
// process environment, with environment variables taking precedence.
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		GRPCPort:    getEnv("GRPC_PORT", "9090"),
		MetricsPort: getEnv("METRICS_PORT", "9091"),

		DatabaseURL: getEnv("DATABASE_URL", "postgres://xoxa:xoxa@localhost:5432/xoxa?sslmode=disable"),
		NatsURL:     getEnv("NATS_URL", "nats://localhost:4222"),

		JWTIssuer:    getEnv("JWT_ISSUER", "xoxa-gateway"),
		JWTAudience:  getEnv("JWT_AUDIENCE", "internal"),
		JWTSecret:    getEnv("JWT_SECRET", ""),
		JWTPublicKey: getEnv("JWT_PUBLIC_KEY_PATH", ""),

		OTELExporterEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		ServiceName:          getEnv("OTEL_SERVICE_NAME", "xoxa-message-gateway"),

		TwilioAccountSID: getEnv("TWILIO_ACCOUNT_SID", ""),
		TwilioAuthToken:  getEnv("TWILIO_AUTH_TOKEN", ""),
		TwilioFromNumber: getEnv("TWILIO_FROM_NUMBER", ""),
		WhatsAppToken:    getEnv("WHATSAPP_TOKEN", ""),
		WhatsAppPhoneID:  getEnv("WHATSAPP_PHONE_NUMBER_ID", ""),
		WhatsAppAPIBase:  getEnv("WHATSAPP_API_BASE", "https://graph.facebook.com/v19.0"),

		MaxSendAttempts: getEnvInt("MAX_SEND_ATTEMPTS", 5),
		BaseRetryDelay:  getEnvDuration("BASE_RETRY_DELAY", time.Second),
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
