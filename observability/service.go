package observability

import (
	"context"
	"errors"
	"fmt"
	"github.com/davfer/goforarun/logger"
	"log/slog"
	"os"

	slogmulti "github.com/samber/slog-multi"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type Cfg struct {
	loggerLevel    slog.Leveler
	loggerChannels map[string]string
	loggerStdout   bool
	serviceVersion string
	serviceName    string
}

type Customizer func(*Cfg)

func WithServiceVersion(version string) Customizer {
	return func(c *Cfg) {
		c.serviceVersion = version
	}
}
func WithLoggerChannels(channels map[string]string) Customizer {
	return func(c *Cfg) {
		c.loggerChannels = channels
	}
}
func WithLoggerLevel(level slog.Leveler) Customizer {
	return func(c *Cfg) {
		c.loggerLevel = level
	}
}
func WithServiceName(serviceName string) Customizer {
	return func(c *Cfg) {
		c.serviceName = serviceName
	}
}
func WithLoggerStdout(stdout bool) Customizer {
	return func(c *Cfg) {
		c.loggerStdout = stdout
	}
}

func StartObservability(ctx context.Context, opts ...Customizer) error {
	c := Cfg{}
	for _, opt := range opts {
		opt(&c)
	}

	var attrs []attribute.KeyValue
	if hostName, err := os.Hostname(); err == nil {
		attrs = append(attrs, semconv.HostName(hostName))
	}
	if c.serviceName != "" {
		attrs = append(attrs, semconv.ServiceNameKey.String(c.serviceName))
	}
	if c.serviceVersion != "" {
		attrs = append(attrs, semconv.ServiceVersionKey.String(c.serviceVersion))
	}

	res, err := resource.New(ctx, resource.WithSchemaURL(semconv.SchemaURL), resource.WithTelemetrySDK(), resource.WithAttributes(attrs...))
	if err != nil {
		return err
	}

	// LOG PARTY
	logExporter, err := otlploggrpc.New(ctx)
	if err != nil {
		return err
	}

	provider := log.NewLoggerProvider(log.WithProcessor(log.NewBatchProcessor(logExporter)), log.WithResource(res))
	global.SetLoggerProvider(provider)

	var slogHandler slog.Handler
	if val, ok := os.LookupEnv("OTEL_SDK_DISABLED"); !ok || val != "true" {
		slogHandler = otelslog.NewHandler(c.serviceName, otelslog.WithLoggerProvider(global.GetLoggerProvider()))
	}
	if c.loggerStdout {
		l := (slog.Leveler)(slog.LevelDebug)
		if c.loggerLevel.Level() >= 0 {
			l = c.loggerLevel
		}
		textHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: l})
		if slogHandler != nil {
			slogHandler = slogmulti.Fanout(slogHandler, textHandler)
		} else {
			slogHandler = textHandler
		}
	}
	if len(c.loggerChannels) > 0 {
		var m map[string]slog.Leveler
		m, err = mapToLeveler(c.loggerChannels)
		if err != nil {
			return err
		}

		slogHandler = logger.NewChanneledHandler(slogHandler, m)
	}

	slog.SetDefault(slog.New(slogHandler))

	// TRACE PARTY
	traceExporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return err
	}

	bsp := trace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := trace.NewTracerProvider(trace.WithSampler(trace.AlwaysSample()), trace.WithResource(res), trace.WithSpanProcessor(bsp))
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// METER PARTY
	metricExporter, err := otlpmetricgrpc.New(ctx)
	if err != nil {
		return err
	}
	meterProvider := metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(metricExporter)), metric.WithResource(res))
	otel.SetMeterProvider(meterProvider)

	return nil
}

func StopObservability(ctx context.Context) (err error) {
	if t, ok := otel.GetTracerProvider().(*trace.TracerProvider); ok {
		err = errors.Join(t.Shutdown(ctx))
	}
	if m, ok := otel.GetMeterProvider().(*metric.MeterProvider); ok {
		err = errors.Join(m.Shutdown(ctx))
	}
	if l, ok := global.GetLoggerProvider().(*log.LoggerProvider); ok {
		err = errors.Join(l.Shutdown(ctx))
	}

	return
}

func mapToLeveler(m map[string]string) (res map[string]slog.Leveler, err error) {
	res = make(map[string]slog.Leveler, len(m))
	var l slog.Leveler
	for k, v := range m {
		l, err = ParseLevel(v)
		if err != nil {
			return
		}

		res[k] = l
	}
	return
}

func ParseLevel(s string) (l slog.Leveler, err error) {
	switch s {
	case "debug":
		l = slog.LevelDebug
	case "info":
		l = slog.LevelInfo
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		err = fmt.Errorf("level %s not supported", s)
	}
	return
}
