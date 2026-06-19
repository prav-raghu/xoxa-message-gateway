// Package transport defines the pluggable provider interface and registry
// used to deliver messages over channels such as Twilio SMS or WhatsApp Cloud.
package transport

import (
	"context"
	"fmt"
	"sync"
)

// Status represents the outcome of a send attempt at the provider.
type Status string

const (
	StatusSent   Status = "sent"
	StatusFailed Status = "failed"
)

// Message is the channel-agnostic payload handed to a Transport.
type Message struct {
	To   string
	Text string
}

// Transport is implemented by every channel provider (Twilio, WhatsApp, ...).
type Transport interface {
	Name() string
	Send(ctx context.Context, msg Message) (providerID string, finalStatus Status, err error)
}

var (
	mu       sync.RWMutex
	registry = map[string]Transport{}
)

// Register adds a transport implementation under its channel name.
// It is typically called from the implementation's init() function.
func Register(name string, t Transport) {
	mu.Lock()
	defer mu.Unlock()
	registry[name] = t
}

// Get returns the transport registered for the given channel name.
func Get(name string) (Transport, error) {
	mu.RLock()
	defer mu.RUnlock()
	t, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("transport: no provider registered for channel %q", name)
	}
	return t, nil
}
