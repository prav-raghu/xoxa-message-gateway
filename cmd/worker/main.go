// Package main starts the NATS consumer(s)
package main

import (
	"log"
	"os"
	"github.com/nats-io/nats.go"
	"xoxa-message-gateway/internal/service"
)

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf("Error connecting to NATS: %v", err)
	}
	defer nc.Close()

	nc.Subscribe("messages.send", func(m *nats.Msg) {
		service.HandleSendMessage(m.Data)
	})

	select {} // block forever
}
