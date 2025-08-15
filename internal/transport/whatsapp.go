// Package transport implements WhatsApp transport
package transport

import "fmt"

type WhatsAppTransport struct{}

func (w *WhatsAppTransport) Send(to string, message string) error {
	fmt.Printf("Sending via WhatsApp to %s: %s\n", to, message)
	return nil
}

func init() {
	Register("whatsapp", &WhatsAppTransport{})
}
