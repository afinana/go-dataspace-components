package events_test

import (
	"context"
	"testing"
	"time"

	"github.com/afinana/go-dataspace-components/internal/pkg/events"
)

func TestEventStruct(t *testing.T) {
	evt := events.Event{
		ID:        "evt-1",
		Type:      "negotiation.started",
		Source:    "control-plane",
		Timestamp: time.Now(),
		Data: map[string]string{
			"negotiationId": "neg-123",
		},
	}

	if evt.ID != "evt-1" {
		t.Errorf("expected event ID evt-1, got %s", evt.ID)
	}

	if evt.Type != "negotiation.started" {
		t.Errorf("expected type negotiation.started, got %s", evt.Type)
	}
}

func TestNatsClient_InvalidURL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	client, err := events.NewClient("nats://127.0.0.1:54321")
	if err == nil {
		client.Close()
		t.Errorf("expected error connecting to non-existent nats server, got nil")
	}

	_ = ctx
}
