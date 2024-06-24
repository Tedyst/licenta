package telemetry

import (
	"context"
	"fmt"
	"log/slog"

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

func initTracer(collectorEndpoint string, service string) error {
	var err error
	var exporter sdktrace.SpanExporter
	if collectorEndpoint == "" {
		slog.Info("Using stdout trace exporter")
		exporter, err = stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
	} else {
		slog.Info("Using otlp trace exporter with collector endpoint", "collectorEndpoint", collectorEndpoint)
		exporter, err = otlptrace.New(
			context.Background(),
			otlptracegrpc.NewClient(
				otlptracegrpc.WithInsecure(),
				otlptracegrpc.WithEndpoint(collectorEndpoint),
			),
		)
	}
	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(service),
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

func InitTelemetry(collectorEndpoint string, service string) error {
	if err := initTracer(collectorEndpoint, service); err != nil {
		return fmt.Errorf("failed to init tracer: %w", err)
	}
	if err := initMetric(collectorEndpoint); err != nil {
		return fmt.Errorf("failed to init metric: %w", err)
	}
	return nil
}

func ShutdownTelemetry() error {
	tp := otel.GetTracerProvider()
	if tp != nil {
		switch tp.(type) {
		case *sdktrace.TracerProvider:
			p, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider)
			if !ok {
				return fmt.Errorf("failed to cast TracerProvider to sdktrace.TracerProvider")
			}
			err := p.Shutdown(context.Background())
			if err != nil {
				return err
			}
		default:
		}
	}
	mp := otel.GetMeterProvider()
	if mp != nil {
		switch mp.(type) {
		case *sdkmetric.MeterProvider:
			p, ok := otel.GetMeterProvider().(*sdkmetric.MeterProvider)
			if !ok {
				return fmt.Errorf("failed to cast MeterProvider to sdkmetric.MeterProvider")
			}
			err := p.Shutdown(context.Background())
			if err != nil {
				return err
			}
		default:
		}
	}
	return nil
}
