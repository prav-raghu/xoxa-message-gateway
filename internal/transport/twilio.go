// Package transport implements Twilio transport
package transport

import "fmt"

type TwilioTransport struct{}

func (t *TwilioTransport) Send(to string, message string) error {
	fmt.Printf("Sending via Twilio to %s: %s\n", to, message)
	return nil
}

func init() {
	Register("twilio", &TwilioTransport{})
}
