package telemetry

import (
	"context"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
)

func initTracer(collectorEndpoint string) error {
	var err error
	var exporter sdktrace.SpanExporter
	if collectorEndpoint == "" {
		exporter, err = stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
	} else {
		exporter, err = otlptrace.New(
			context.Background(),
			otlptracegrpc.NewClient(
				otlptracegrpc.WithInsecure(),
				otlptracegrpc.WithEndpoint(collectorEndpoint),
			),
		)
	}
	if err != nil {
		return errors.Wrap(err, "failed to create trace exporter")
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String("backend"),
			)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return nil
}

func initMetric(collectorEndpoint string) error {
	if collectorEndpoint == "" {
		return nil
	}
	exporter, err := otlpmetricgrpc.New(
		context.Background(),
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(collectorEndpoint),
	)
	if err != nil {
		return err
	}
	reader := sdkmetric.NewPeriodicReader(exporter)
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(provider)

	return nil
}

func InitTelemetry(collectorEndpoint string) error {
	if err := initTracer(collectorEndpoint); err != nil {
		return errors.Wrap(err, "failed to init tracer")
	}
	if err := initMetric(collectorEndpoint); err != nil {
		return errors.Wrap(err, "failed to init metric")
	}
	return nil
}
