package repo

import (
	"context"
	"errors"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&Message{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return db
}

func TestMessageRepository_CreateAndGet(t *testing.T) {
	repo := NewMessageRepository(newTestDB(t))
	ctx := context.Background()

	msg := &Message{
		ExternalID: "msg_1",
		Channel:    "sms",
		Recipient:  "+15550001",
		Content:    "hello",
		Status:     StatusQueued,
	}
	if err := repo.Create(ctx, msg); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.GetByExternalID(ctx, "msg_1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Status != StatusQueued || got.Recipient != "+15550001" {
		t.Fatalf("unexpected message: %+v", got)
	}
}

func TestMessageRepository_GetByExternalID_NotFound(t *testing.T) {
	repo := NewMessageRepository(newTestDB(t))

	_, err := repo.GetByExternalID(context.Background(), "missing")
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestMessageRepository_UpdateResult(t *testing.T) {
	repo := NewMessageRepository(newTestDB(t))
	ctx := context.Background()

	msg := &Message{ExternalID: "msg_2", Channel: "whatsapp", Recipient: "+1", Content: "hi", Status: StatusQueued}
	if err := repo.Create(ctx, msg); err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := repo.UpdateResult(ctx, "msg_2", StatusSent, "provider-123", "", 1); err != nil {
		t.Fatalf("update: %v", err)
	}

	got, err := repo.GetByExternalID(ctx, "msg_2")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Status != StatusSent || got.ProviderID != "provider-123" || got.Attempts != 1 {
		t.Fatalf("unexpected message after update: %+v", got)
	}
}
