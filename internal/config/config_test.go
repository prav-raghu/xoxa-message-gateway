package config

import (
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "8080" {
		t.Fatalf("expected default port 8080, got %s", cfg.Port)
	}
	if cfg.MaxSendAttempts != 5 {
		t.Fatalf("expected default max attempts 5, got %d", cfg.MaxSendAttempts)
	}
}

func TestLoad_OverridesFromEnv(t *testing.T) {
	t.Setenv("PORT", "9999")
	t.Setenv("MAX_SEND_ATTEMPTS", "7")
	t.Setenv("BASE_RETRY_DELAY", "2s")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "9999" {
		t.Fatalf("expected overridden port 9999, got %s", cfg.Port)
	}
	if cfg.MaxSendAttempts != 7 {
		t.Fatalf("expected overridden max attempts 7, got %d", cfg.MaxSendAttempts)
	}
	if cfg.BaseRetryDelay != 2*time.Second {
		t.Fatalf("expected overridden base delay 2s, got %s", cfg.BaseRetryDelay)
	}
}
