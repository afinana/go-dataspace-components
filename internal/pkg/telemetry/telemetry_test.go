package telemetry_test

import (
	"context"
	"testing"
	"time"

	"github.com/afinana/go-dataspace-components/internal/pkg/telemetry"
)

func TestInitTelemetry(t *testing.T) {
	tel, shutdown, err := telemetry.InitTelemetry("test-service")
	if err != nil {
		t.Fatalf("unexpected error initializing telemetry: %v", err)
	}
	if tel == nil {
		t.Fatal("expected non-nil Telemetry struct")
	}
	if tel.Tracer == nil {
		t.Error("expected non-nil Tracer")
	}
	if tel.Meter == nil {
		t.Error("expected non-nil Meter")
	}

	ctx := context.Background()
	if err := shutdown(ctx); err != nil {
		t.Errorf("unexpected error shutting down telemetry: %v", err)
	}
}

func TestStartSpan(t *testing.T) {
	tel, _, err := telemetry.InitTelemetry("test-service")
	if err != nil {
		t.Fatalf("unexpected error initializing telemetry: %v", err)
	}

	ctx := context.Background()
	spanCtx, span := telemetry.StartSpan(ctx, tel.Tracer, "test-span")
	if spanCtx == nil {
		t.Error("expected non-nil span context")
	}
	if span == nil {
		t.Error("expected non-nil span")
	}
	span.End()
}

func TestRecordDuration(t *testing.T) {
	tel, _, err := telemetry.InitTelemetry("test-service")
	if err != nil {
		t.Fatalf("unexpected error initializing telemetry: %v", err)
	}

	ctx := context.Background()
	telemetry.RecordDuration(ctx, tel.Meter, "test_operation", 150*time.Millisecond)
}
