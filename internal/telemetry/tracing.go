// Package telemetry provides tracing setup
package telemetry

import (
	"go.opentelemetry.io/otel"
)

func InitTracing() {
	// ...initialize OpenTelemetry...
	otel.SetTracerProvider(nil)
}
