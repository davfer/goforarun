package observability

import (
	"fmt"
	"go.opentelemetry.io/otel"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var traceProvider *sdkTrace.TracerProvider

func NewTracer(name string) trace.Tracer {
	if name == "" {
		return otel.Tracer(fmt.Sprintf("has/%s", serviceName))
	}

	return otel.Tracer(fmt.Sprintf("has/%s/%s", serviceName, name))
}
