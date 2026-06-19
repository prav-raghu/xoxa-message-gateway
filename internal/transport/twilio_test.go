package transport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTwilioTransport_Send_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		user, pass, ok := r.BasicAuth()
		if !ok || user != "AC123" || pass != "secret" {
			t.Errorf("unexpected basic auth: %s/%s", user, pass)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"sid":"SM123","status":"queued"}`))
	}))
	defer srv.Close()

	tr := NewTwilioTransport("AC123", "secret", "+15550000", srv.URL)
	providerID, status, err := tr.Send(context.Background(), Message{To: "+15550001", Text: "hi"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if providerID != "SM123" || status != StatusSent {
		t.Fatalf("unexpected result: %s %s", providerID, status)
	}
	if tr.Name() != "sms" {
		t.Fatalf("unexpected name: %s", tr.Name())
	}
}

func TestTwilioTransport_Send_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":21211,"message":"invalid number"}`))
	}))
	defer srv.Close()

	tr := NewTwilioTransport("AC123", "secret", "+15550000", srv.URL)
	_, status, err := tr.Send(context.Background(), Message{To: "bad", Text: "hi"})
	if err == nil {
		t.Fatal("expected error")
	}
	if status != StatusFailed {
		t.Fatalf("expected StatusFailed, got %s", status)
	}
}

func TestTwilioTransport_Send_MissingCredentials(t *testing.T) {
	tr := NewTwilioTransport("", "", "", "")
	_, status, err := tr.Send(context.Background(), Message{To: "+1", Text: "hi"})
	if err == nil || status != StatusFailed {
		t.Fatalf("expected failure for missing credentials, got status=%s err=%v", status, err)
	}
}
