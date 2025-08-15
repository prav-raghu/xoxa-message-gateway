// Package transport defines the transport interface and registry
package transport

type Transport interface {
	Send(to string, message string) error
}

var registry = map[string]Transport{}

func Register(name string, t Transport) {
	registry[name] = t
}

func Get(name string) Transport {
	return registry[name]
}
