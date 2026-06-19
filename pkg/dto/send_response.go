// Package dto defines response DTOs shared by the REST and gRPC APIs.
package dto

// SendResponse is returned from POST /api/v1/messages.
type SendResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// MessageResponse is returned from GET /api/v1/messages/:id.
type MessageResponse struct {
	ID         string `json:"id"`
	Channel    string `json:"channel"`
	To         string `json:"to"`
	Text       string `json:"text"`
	Status     string `json:"status"`
	ProviderID string `json:"provider_id,omitempty"`
	Error      string `json:"error,omitempty"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}
