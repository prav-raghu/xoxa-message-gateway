package service

import (
	"context"
	"sync"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"xoxa-message-gateway/internal/repo"
	"xoxa-message-gateway/pkg/dto"
)

type fakePublisher struct {
	mu        sync.Mutex
	published map[string][][]byte
}

func newFakePublisher() *fakePublisher {
	return &fakePublisher{published: map[string][][]byte{}}
}

func (f *fakePublisher) Publish(subject string, data []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.published[subject] = append(f.published[subject], data)
	return nil
}

func (f *fakePublisher) count(subject string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.published[subject])
}

func newTestRepo(t *testing.T) *repo.MessageRepository {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&repo.Message{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return repo.NewMessageRepository(db)
}

func TestService_SendMessage(t *testing.T) {
	pub := newFakePublisher()
	svc := New(newTestRepo(t), pub)

	resp, err := svc.SendMessage(context.Background(), dto.SendRequest{Channel: "sms", To: "+1", Text: "hi"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID == "" || resp.Status != repo.StatusQueued {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if pub.count(SubjectSend) != 1 {
		t.Fatalf("expected 1 published job, got %d", pub.count(SubjectSend))
	}
}

func TestService_GetMessage(t *testing.T) {
	pub := newFakePublisher()
	svc := New(newTestRepo(t), pub)

	created, err := svc.SendMessage(context.Background(), dto.SendRequest{Channel: "whatsapp", To: "+2", Text: "yo"})
	if err != nil {
		t.Fatalf("send: %v", err)
	}

	got, err := svc.GetMessage(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ID != created.ID || got.Channel != "whatsapp" || got.To != "+2" || got.Text != "yo" {
		t.Fatalf("unexpected message: %+v", got)
	}
}

func TestService_GetMessage_NotFound(t *testing.T) {
	svc := New(newTestRepo(t), newFakePublisher())
	if _, err := svc.GetMessage(context.Background(), "msg_missing"); err == nil {
		t.Fatal("expected error for missing message")
	}
}
