// Package service contains the core business logic for sending messages.
package service

// Subjects used on the NATS message bus.
const (
	SubjectSend = "messages.send"
	SubjectDLQ  = "messages.dlq"
)

// Publisher abstracts the NATS connection so the service layer can be
// exercised in tests without a running broker. *nats.Conn satisfies this.
type Publisher interface {
	Publish(subject string, data []byte) error
}

// SendJob is the payload queued for asynchronous delivery by the worker.
type SendJob struct {
	ExternalID string `json:"external_id"`
	Channel    string `json:"channel"`
	To         string `json:"to"`
	Text       string `json:"text"`
}
