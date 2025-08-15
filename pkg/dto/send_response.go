// Package dto defines response DTOs
package dto

type SendResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}
