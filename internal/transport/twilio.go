// Package transport: Twilio Programmable Messaging (SMS) implementation.
package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultTwilioAPIBase = "https://api.twilio.com"

// TwilioTransport sends SMS messages via the Twilio REST API.
type TwilioTransport struct {
	AccountSID string
	AuthToken  string
	FromNumber string
	APIBase    string

	client *http.Client
}

// NewTwilioTransport builds a Twilio transport. apiBase may be empty to use
// the production Twilio API; tests can point it at an httptest server.
func NewTwilioTransport(accountSID, authToken, fromNumber, apiBase string) *TwilioTransport {
	if apiBase == "" {
		apiBase = defaultTwilioAPIBase
	}
	return &TwilioTransport{
		AccountSID: accountSID,
		AuthToken:  authToken,
		FromNumber: fromNumber,
		APIBase:    apiBase,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (t *TwilioTransport) Name() string { return "sms" }

type twilioMessageResponse struct {
	SID    string `json:"sid"`
	Status string `json:"status"`
	Code   int    `json:"code"`
	Msg    string `json:"message"`
}

func (t *TwilioTransport) Send(ctx context.Context, msg Message) (string, Status, error) {
	if t.AccountSID == "" || t.AuthToken == "" {
		return "", StatusFailed, fmt.Errorf("twilio: missing account credentials")
	}

	endpoint := fmt.Sprintf("%s/2010-04-01/Accounts/%s/Messages.json", t.APIBase, t.AccountSID)

	form := url.Values{}
	form.Set("To", msg.To)
	form.Set("From", t.FromNumber)
	form.Set("Body", msg.Text)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", StatusFailed, fmt.Errorf("twilio: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(t.AccountSID, t.AuthToken)

	resp, err := t.client.Do(req)
	if err != nil {
		return "", StatusFailed, fmt.Errorf("twilio: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", StatusFailed, fmt.Errorf("twilio: read response: %w", err)
	}

	var parsed twilioMessageResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", StatusFailed, fmt.Errorf("twilio: decode response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", StatusFailed, fmt.Errorf("twilio: api error (%d): %s", parsed.Code, parsed.Msg)
	}

	return parsed.SID, StatusSent, nil
}
