package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"xoxa-message-gateway/internal/repo"
	"xoxa-message-gateway/internal/transport"
)

type stubTransport struct {
	name       string
	failUntil  int
	calls      int
	providerID string
}

func (s *stubTransport) Name() string { return s.name }

func (s *stubTransport) Send(ctx context.Context, msg transport.Message) (string, transport.Status, error) {
	s.calls++
	if s.calls <= s.failUntil {
		return "", transport.StatusFailed, errors.New("simulated failure")
	}
	return s.providerID, transport.StatusSent, nil
}

func mustCreateQueued(t *testing.T, r *repo.MessageRepository, externalID, channel string) {
	t.Helper()
	if err := r.Create(context.Background(), &repo.Message{
		ExternalID: externalID,
		Channel:    channel,
		Recipient:  "+1",
		Content:    "hi",
		Status:     repo.StatusQueued,
	}); err != nil {
		t.Fatalf("create message: %v", err)
	}
}

func TestProcessor_Process_SucceedsAfterRetries(t *testing.T) {
	stub := &stubTransport{name: "stub-retry-success", failUntil: 2, providerID: "p-1"}
	transport.Register(stub.name, stub)

	r := newTestRepo(t)
	mustCreateQueued(t, r, "msg_a", stub.name)

	proc := NewProcessor(r, newFakePublisher(), 5, time.Millisecond)
	job, _ := json.Marshal(SendJob{ExternalID: "msg_a", Channel: stub.name, To: "+1", Text: "hi"})

	if err := proc.Process(context.Background(), job); err != nil {
		t.Fatalf("process: %v", err)
	}

	got, err := r.GetByExternalID(context.Background(), "msg_a")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Status != repo.StatusSent || got.ProviderID != "p-1" || got.Attempts != 3 {
		t.Fatalf("unexpected message after process: %+v", got)
	}
}

func TestProcessor_Process_DeadLettersAfterExhaustingRetries(t *testing.T) {
	stub := &stubTransport{name: "stub-retry-fail", failUntil: 99}
	transport.Register(stub.name, stub)

	r := newTestRepo(t)
	mustCreateQueued(t, r, "msg_b", stub.name)

	pub := newFakePublisher()
	proc := NewProcessor(r, pub, 3, time.Millisecond)
	job, _ := json.Marshal(SendJob{ExternalID: "msg_b", Channel: stub.name, To: "+1", Text: "hi"})

	if err := proc.Process(context.Background(), job); err != nil {
		t.Fatalf("process: %v", err)
	}

	got, err := r.GetByExternalID(context.Background(), "msg_b")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Status != repo.StatusDeadLetter || got.Attempts != 3 {
		t.Fatalf("unexpected message after process: %+v", got)
	}
	if pub.count(SubjectDLQ) != 1 {
		t.Fatalf("expected 1 DLQ publish, got %d", pub.count(SubjectDLQ))
	}
}

func TestProcessor_Process_UnknownChannelGoesToDLQ(t *testing.T) {
	r := newTestRepo(t)
	mustCreateQueued(t, r, "msg_c", "unregistered-channel")

	pub := newFakePublisher()
	proc := NewProcessor(r, pub, 3, time.Millisecond)
	job, _ := json.Marshal(SendJob{ExternalID: "msg_c", Channel: "unregistered-channel", To: "+1", Text: "hi"})

	if err := proc.Process(context.Background(), job); err != nil {
		t.Fatalf("process: %v", err)
	}

	got, err := r.GetByExternalID(context.Background(), "msg_c")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Status != repo.StatusDeadLetter {
		t.Fatalf("expected dead_letter status, got %s", got.Status)
	}
}

func TestBackoff_Exponential(t *testing.T) {
	base := 100 * time.Millisecond
	if Backoff(1, base) != base {
		t.Fatalf("expected base delay for attempt 1")
	}
	if Backoff(2, base) != 2*base {
		t.Fatalf("expected 2x base delay for attempt 2")
	}
	if Backoff(3, base) != 4*base {
		t.Fatalf("expected 4x base delay for attempt 3")
	}
}
