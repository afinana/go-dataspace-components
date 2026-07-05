package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Telemetry provides handles for tracing and metrics.
type Telemetry struct {
	Tracer trace.Tracer
	Meter  metric.Meter
}

// InitTelemetry sets up trace and metric providers.
// In a production environment, this would initialize an exporter (e.g. OTLP) to Collector/Jaeger/Prometheus.
// Here we set up standard OTel configurations and return a shutdown function.
func InitTelemetry(serviceName string) (*Telemetry, func(context.Context) error, error) {
	// Set global propagator to W3C Trace Context.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Get default tracer and meter (these default to no-op until registered,
	// but we can register standard OTel globals)
	tracer := otel.Tracer(serviceName)
	meter := otel.Meter(serviceName)

	shutdown := func(ctx context.Context) error {
		// Flush traces and metrics.
		return nil
	}

	return &Telemetry{
		Tracer: tracer,
		Meter:  meter,
	}, shutdown, nil
}

// StartSpan helper function to make starting traces cleaner.
func StartSpan(ctx context.Context, tracer trace.Tracer, name string) (context.Context, trace.Span) {
	return tracer.Start(ctx, name)
}

// RecordDuration records a duration metric for a operation.
func RecordDuration(ctx context.Context, meter metric.Meter, name string, duration time.Duration, attrs ...any) {
	// Utility for recording durations or latency of operations.
	histogram, err := meter.Float64Histogram(
		fmt.Sprintf("%s_duration_seconds", name),
		metric.WithDescription("Duration of operations in seconds"),
		metric.WithUnit("s"),
	)
	if err == nil {
		histogram.Record(ctx, duration.Seconds())
	}
}
