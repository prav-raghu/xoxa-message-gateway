package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics exposed at /metrics for Prometheus scraping.
var (
	MessagesSent = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "messages_sent_total",
		Help: "Total number of messages successfully delivered, by channel.",
	}, []string{"channel"})

	MessagesFailed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "messages_failed_total",
		Help: "Total number of message delivery attempts that failed, by channel.",
	}, []string{"channel"})

	MessagesDeadLettered = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "messages_dead_lettered_total",
		Help: "Total number of messages that exhausted retries and were dead-lettered, by channel.",
	}, []string{"channel"})

	SendDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "message_send_duration_seconds",
		Help: "Time spent sending a message to a provider, by channel.",
	}, []string{"channel"})
)
