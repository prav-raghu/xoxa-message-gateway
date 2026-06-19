package transport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWhatsAppTransport_Send_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token123" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"messages":[{"id":"wamid.123"}]}`))
	}))
	defer srv.Close()

	wa := NewWhatsAppTransport("token123", "1234567890", srv.URL)
	providerID, status, err := wa.Send(context.Background(), Message{To: "+15550001", Text: "hi"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if providerID != "wamid.123" || status != StatusSent {
		t.Fatalf("unexpected result: %s %s", providerID, status)
	}
	if wa.Name() != "whatsapp" {
		t.Fatalf("unexpected name: %s", wa.Name())
	}
}

func TestWhatsAppTransport_Send_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":{"message":"invalid token","code":401}}`))
	}))
	defer srv.Close()

	wa := NewWhatsAppTransport("bad", "1234567890", srv.URL)
	_, status, err := wa.Send(context.Background(), Message{To: "+1", Text: "hi"})
	if err == nil || status != StatusFailed {
		t.Fatalf("expected failure, got status=%s err=%v", status, err)
	}
}

func TestWhatsAppTransport_Send_MissingCredentials(t *testing.T) {
	wa := NewWhatsAppTransport("", "", "")
	_, status, err := wa.Send(context.Background(), Message{To: "+1", Text: "hi"})
	if err == nil || status != StatusFailed {
		t.Fatalf("expected failure for missing credentials, got status=%s err=%v", status, err)
	}
}
