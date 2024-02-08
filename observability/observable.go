package observability

import (
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

type ObservableStruct struct {
	Tracer trace.Tracer
	Logger *logrus.Entry
}

func (o *ObservableStruct) InitObservableStruct(name string) {
	o.Tracer = NewTracer(name)
	o.Logger = NewLogger(name)
}
