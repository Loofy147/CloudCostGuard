package tracing

import (
	"context"
	"io"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// NewTracerProvider creates a new tracer provider.
func NewTracerProvider(w io.Writer) (*trace.TracerProvider, error) {
	exporter, err := stdouttrace.New(
		stdouttrace.WithWriter(w),
		stdouttrace.WithPrettyPrint(),
	)
	if err != nil {
		return nil, err
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("cloudcostguard-backend"),
		)),
	)
	return tp, nil
}

// InitTracer initializes the global tracer provider.
func InitTracer() func() {
	tp, err := NewTracerProvider(log.Writer())
	if err != nil {
		log.Fatal(err)
	}

	otel.SetTracerProvider(tp)

	return func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}
}
