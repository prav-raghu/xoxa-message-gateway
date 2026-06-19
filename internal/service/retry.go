package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"xoxa-message-gateway/internal/repo"
	"xoxa-message-gateway/internal/telemetry"
	"xoxa-message-gateway/internal/transport"
)

// Processor consumes queued SendJobs, dispatches them to the appropriate
// transport with retry/backoff, and falls back to the dead-letter subject
// once attempts are exhausted.
type Processor struct {
	repo        *repo.MessageRepository
	pub         Publisher
	maxAttempts int
	baseDelay   time.Duration
}

// NewProcessor builds a Processor.
func NewProcessor(repository *repo.MessageRepository, pub Publisher, maxAttempts int, baseDelay time.Duration) *Processor {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	return &Processor{repo: repository, pub: pub, maxAttempts: maxAttempts, baseDelay: baseDelay}
}

// Backoff returns the delay before retry attempt n (1-indexed), growing
// exponentially from base.
func Backoff(attempt int, base time.Duration) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	return base * time.Duration(1<<uint(attempt-1))
}

// Process handles one queued job, retrying delivery before giving up.
func (p *Processor) Process(ctx context.Context, data []byte) error {
	var job SendJob
	if err := json.Unmarshal(data, &job); err != nil {
		return fmt.Errorf("processor: decode job: %w", err)
	}

	t, err := transport.Get(job.Channel)
	if err != nil {
		return p.sendToDLQ(ctx, job, 0, err)
	}

	var lastErr error
	for attempt := 1; attempt <= p.maxAttempts; attempt++ {
		start := time.Now()
		providerID, status, sendErr := t.Send(ctx, transport.Message{To: job.To, Text: job.Text})
		telemetry.SendDuration.WithLabelValues(job.Channel).Observe(time.Since(start).Seconds())

		if sendErr == nil && status == transport.StatusSent {
			telemetry.MessagesSent.WithLabelValues(job.Channel).Inc()
			return p.repo.UpdateResult(ctx, job.ExternalID, repo.StatusSent, providerID, "", attempt)
		}
		telemetry.MessagesFailed.WithLabelValues(job.Channel).Inc()
		lastErr = sendErr

		if attempt < p.maxAttempts {
			if err := sleep(ctx, Backoff(attempt, p.baseDelay)); err != nil {
				return err
			}
		}
	}

	telemetry.MessagesDeadLettered.WithLabelValues(job.Channel).Inc()
	return p.sendToDLQ(ctx, job, p.maxAttempts, lastErr)
}

func (p *Processor) sendToDLQ(ctx context.Context, job SendJob, attempts int, cause error) error {
	errMsg := ""
	if cause != nil {
		errMsg = cause.Error()
	}
	if err := p.repo.UpdateResult(ctx, job.ExternalID, repo.StatusDeadLetter, "", errMsg, attempts); err != nil {
		return fmt.Errorf("processor: record dead letter: %w", err)
	}

	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("processor: encode dlq job: %w", err)
	}
	return p.pub.Publish(SubjectDLQ, payload)
}

func sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
