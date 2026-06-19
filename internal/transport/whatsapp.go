// Package transport: WhatsApp Cloud API implementation.
package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultWhatsAppAPIBase = "https://graph.facebook.com/v19.0"

// WhatsAppTransport sends text messages via the WhatsApp Cloud API.
type WhatsAppTransport struct {
	Token   string
	PhoneID string
	APIBase string

	client *http.Client
}

// NewWhatsAppTransport builds a WhatsApp Cloud API transport. apiBase may be
// empty to use the production Graph API; tests can point it at an httptest server.
func NewWhatsAppTransport(token, phoneID, apiBase string) *WhatsAppTransport {
	if apiBase == "" {
		apiBase = defaultWhatsAppAPIBase
	}
	return &WhatsAppTransport{
		Token:   token,
		PhoneID: phoneID,
		APIBase: apiBase,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (w *WhatsAppTransport) Name() string { return "whatsapp" }

type whatsAppTextBody struct {
	Body string `json:"body"`
}

type whatsAppRequest struct {
	MessagingProduct string           `json:"messaging_product"`
	To               string           `json:"to"`
	Type             string           `json:"type"`
	Text             whatsAppTextBody `json:"text"`
}

type whatsAppMessageRef struct {
	ID string `json:"id"`
}

type whatsAppResponse struct {
	Messages []whatsAppMessageRef `json:"messages"`
	Error    *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error,omitempty"`
}

func (w *WhatsAppTransport) Send(ctx context.Context, msg Message) (string, Status, error) {
	if w.Token == "" || w.PhoneID == "" {
		return "", StatusFailed, fmt.Errorf("whatsapp: missing API credentials")
	}

	endpoint := fmt.Sprintf("%s/%s/messages", w.APIBase, w.PhoneID)

	payload := whatsAppRequest{
		MessagingProduct: "whatsapp",
		To:               msg.To,
		Type:             "text",
		Text:             whatsAppTextBody{Body: msg.Text},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", StatusFailed, fmt.Errorf("whatsapp: encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", StatusFailed, fmt.Errorf("whatsapp: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+w.Token)

	resp, err := w.client.Do(req)
	if err != nil {
		return "", StatusFailed, fmt.Errorf("whatsapp: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", StatusFailed, fmt.Errorf("whatsapp: read response: %w", err)
	}

	var parsed whatsAppResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", StatusFailed, fmt.Errorf("whatsapp: decode response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if parsed.Error != nil {
			return "", StatusFailed, fmt.Errorf("whatsapp: api error (%d): %s", parsed.Error.Code, parsed.Error.Message)
		}
		return "", StatusFailed, fmt.Errorf("whatsapp: api error: status %d", resp.StatusCode)
	}

	if len(parsed.Messages) == 0 {
		return "", StatusFailed, fmt.Errorf("whatsapp: no message id in response")
	}

	return parsed.Messages[0].ID, StatusSent, nil
}
