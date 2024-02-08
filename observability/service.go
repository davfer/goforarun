package observability

import (
	"context"
	"fmt"
	"github.com/davfer/goforarun/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/uptrace/opentelemetry-go-extra/otellogrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	"io"
	"os"
)

var serviceName string
var traceProvider *sdkTrace.TracerProvider

var observabilityMode string
var fileResource *os.File

var baseLogger = *logrus.New()
var baseFields = logrus.Fields{}

// INITIALIZERS
func SetObservabilityConfig(c *config.BaseAppConfig) error {
	serviceName = c.ServiceName
	observabilityMode = c.ObservabilityMode

	// LOG PARTY
	fields := logrus.Fields{
		"service": serviceName,
	}
	if c.BuildInfo != nil {
		fields["version"] = c.BuildInfo.Version
		fields["commit"] = c.BuildInfo.Commit
		fields["date"] = c.BuildInfo.Date
	}

	err := setLoggerConfig(
		c.LoggingLevel,
		os.Stdout,
		c.LoggingFormat,
		fields,
	)
	if err != nil {
		return errors.Wrap(err, "couldn't set logger config")
	}

	// TRACE PARTY
	var exp sdkTrace.SpanExporter
	var res *resource.Resource

	// exp selection
	if c.ObservabilityMode == "otpl" {
		// TODO
	} else if c.ObservabilityMode == "file" || c.ObservabilityMode == "stdout" {
		if c.ObservabilityMode == "file" {
			fileResource, err = os.Create("traces.txt")
			if err != nil {
				return errors.Wrap(err, "failed to create trace file")
			}
		} else {
			fileResource = os.Stdout
		}

		exp, err = stdouttrace.New(
			stdouttrace.WithWriter(fileResource),
			stdouttrace.WithPrettyPrint(),
			stdouttrace.WithoutTimestamps(),
		)

		if err != nil {
			return errors.Wrap(err, "failed to create stdout trace exporter")
		}
	} else {
		observabilityMode = "disabled"
		return nil
	}

	// res decoration
	res = getResource(nil)

	traceProvider = sdkTrace.NewTracerProvider(
		sdkTrace.WithBatcher(exp),
		sdkTrace.WithResource(res),
	)

	// set trace provider
	otel.SetTracerProvider(traceProvider)

	// logger hooks
	baseLogger.AddHook(getLogrusHook())

	return nil
}
func CloseObservability(ctx context.Context) error {
	if observabilityMode == "disabled" {
		baseLogger.Info("observability disabled, skipping close")
		return nil
	}
	if err := traceProvider.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "failed to shutdown trace provider")
	}
	if observabilityMode == "file" {
		err := fileResource.Close()
		if err != nil {
			return errors.Wrap(err, "failed to close trace file")
		}
	}

	return nil
}

func NewTracer(name string) trace.Tracer {
	if name == "" {
		return otel.Tracer(fmt.Sprintf("has/%s", serviceName))
	}

	return otel.Tracer(fmt.Sprintf("has/%s/%s", serviceName, name))
}
func NewLogger(channel string) *logrus.Entry {
	return baseLogger.WithFields(baseFields).WithField("channel", channel)
}

func NilLogger() *logrus.Entry {
	nilLogger := *logrus.New()
	nilLogger.SetOutput(io.Discard)

	return nilLogger.WithFields(baseFields)
}

type ObservableStruct struct {
	Tracer trace.Tracer
	Logger *logrus.Entry
}

func (o *ObservableStruct) InitObservableStruct(name string) {
	o.Tracer = NewTracer(name)
	o.Logger = NewLogger(name)
}

func getResource(tags []attribute.KeyValue) *resource.Resource {
	attrs := []attribute.KeyValue{
		semconv.ServiceNameKey.String(serviceName),
		semconv.ServiceVersionKey.String("v0.0.1"),
	}
	attrs = append(attrs, tags...)

	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			attrs...,
		),
	)
	return r
}

func getLogrusHook() logrus.Hook {
	return otellogrus.NewHook(
		otellogrus.WithLevels(
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
			logrus.InfoLevel,
		),
	)
}

func setLoggerConfig(levelName string, out io.Writer, formatterName string, fields map[string]interface{}) error {
	var level logrus.Level
	switch levelName {
	case "debug":
		level = logrus.DebugLevel
		break
	case "info":
		level = logrus.InfoLevel
		break
	case "warn":
		level = logrus.WarnLevel
		break
	case "error":
		level = logrus.ErrorLevel
		break
	case "fatal":
		level = logrus.FatalLevel
		break
	case "panic":
		level = logrus.PanicLevel
		break
	default:
		return errors.New(fmt.Sprintf("invalid log level: %s", levelName))
	}
	baseLogger.SetLevel(level)
	baseLogger.SetOutput(out)

	var formatter logrus.Formatter
	switch formatterName {
	case "text":
		formatter = new(logrus.TextFormatter)
		break
	case "json":
		formatter = new(logrus.JSONFormatter)
		break
	default:
		return errors.New("invalid log formatter")
	}
	baseLogger.SetFormatter(formatter)
	baseFields = fields

	return nil
}
