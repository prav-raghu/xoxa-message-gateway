// Package dto defines request DTOs shared by the REST and gRPC APIs.
package dto

// SendRequest is the payload for POST /api/v1/messages.
type SendRequest struct {
	Channel string `json:"channel" binding:"required"`
	To      string `json:"to" binding:"required"`
	Text    string `json:"text" binding:"required"`
}
