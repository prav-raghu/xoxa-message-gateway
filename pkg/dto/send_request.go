// Package dto defines request DTOs
package dto

type SendRequest struct {
	To      string `json:"to"`
	Message string `json:"message"`
	Channel string `json:"channel"`
}
