package observability

import (
	"context"
	"github.com/davfer/goforarun/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"os"
)

var serviceName string
var observabilityMode string
var fileResource *os.File

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

	err := InitLogger(c.LoggingConfig, fields)
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

	// logger hooks TODO
	logger.baseLogger.AddHook(getLogrusHook())

	return nil
}
func CloseObservability(ctx context.Context) error {
	if observabilityMode == "disabled" {
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
