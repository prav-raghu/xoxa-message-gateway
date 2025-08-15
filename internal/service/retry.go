// Package service provides retry and DLQ logic
package service

import (
	"time"
)

func Backoff(attempt int) time.Duration {
	return time.Duration(attempt) * time.Second
}

func SendToDLQ(msg []byte) {
	// ...send to dead letter queue...
}
