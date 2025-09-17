package tracing

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

// InitTracing initializes the tracing system
func InitTracing(serviceName string, jaegerURL string) error {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerURL)))
	if err != nil {
		return err
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)

	otel.SetTracerProvider(tp)
	tracer = otel.Tracer(serviceName)

	log.Printf("Tracing initialized with service name: %s", serviceName)
	return nil
}

// StartSpan starts a new span with the given name and context
func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	if tracer == nil {
		// Return a no-op span if tracer is not initialized
		return ctx, trace.SpanFromContext(ctx)
	}
	return tracer.Start(ctx, name)
}

// GetTracer returns the tracer instance
func GetTracer() trace.Tracer {
	return tracer
}
