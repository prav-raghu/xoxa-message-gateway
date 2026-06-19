package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"xoxa-message-gateway/internal/repo"
	"xoxa-message-gateway/pkg/dto"
)

// Service implements the application's message-sending use cases. It is
// shared by both the REST and gRPC transports.
type Service struct {
	repo *repo.MessageRepository
	pub  Publisher
}

// New builds a Service backed by the given repository and queue publisher.
func New(repository *repo.MessageRepository, pub Publisher) *Service {
	return &Service{repo: repository, pub: pub}
}

// SendMessage persists a new message and enqueues it for async delivery.
func (s *Service) SendMessage(ctx context.Context, req dto.SendRequest) (dto.SendResponse, error) {
	externalID := "msg_" + uuid.NewString()

	msg := &repo.Message{
		ExternalID: externalID,
		Channel:    req.Channel,
		Recipient:  req.To,
		Content:    req.Text,
		Status:     repo.StatusQueued,
	}
	if err := s.repo.Create(ctx, msg); err != nil {
		return dto.SendResponse{}, fmt.Errorf("service: create message: %w", err)
	}

	job := SendJob{ExternalID: externalID, Channel: req.Channel, To: req.To, Text: req.Text}
	payload, err := json.Marshal(job)
	if err != nil {
		return dto.SendResponse{}, fmt.Errorf("service: encode job: %w", err)
	}
	if err := s.pub.Publish(SubjectSend, payload); err != nil {
		return dto.SendResponse{}, fmt.Errorf("service: publish job: %w", err)
	}

	return dto.SendResponse{ID: externalID, Status: repo.StatusQueued}, nil
}

// GetMessage returns the current state of a previously sent message.
func (s *Service) GetMessage(ctx context.Context, externalID string) (dto.MessageResponse, error) {
	msg, err := s.repo.GetByExternalID(ctx, externalID)
	if err != nil {
		return dto.MessageResponse{}, err
	}

	return dto.MessageResponse{
		ID:         msg.ExternalID,
		Channel:    msg.Channel,
		To:         msg.Recipient,
		Text:       msg.Content,
		Status:     msg.Status,
		ProviderID: msg.ProviderID,
		Error:      msg.Error,
		CreatedAt:  msg.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  msg.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
