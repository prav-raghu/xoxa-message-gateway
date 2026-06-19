package transport

import (
	"context"
	"testing"
)

type fakeTransport struct{ name string }

func (f *fakeTransport) Name() string { return f.name }
func (f *fakeTransport) Send(ctx context.Context, msg Message) (string, Status, error) {
	return "fake-id", StatusSent, nil
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	Register("fake-channel", &fakeTransport{name: "fake-channel"})

	tr, err := Get("fake-channel")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr.Name() != "fake-channel" {
		t.Fatalf("unexpected transport: %s", tr.Name())
	}
}

func TestRegistry_GetUnknown(t *testing.T) {
	if _, err := Get("does-not-exist"); err == nil {
		t.Fatal("expected error for unregistered channel")
	}
}
