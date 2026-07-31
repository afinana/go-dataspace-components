package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// Event represents a generic dataspace event payload.
type Event struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // e.g., "negotiation.started", "transfer.completed", "catalog.updated"
	Source    string    `json:"source"`
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data"`
}

// Client wraps NATS JetStream connectivity for event publishing and consuming.
type Client struct {
	nc *nats.Conn
	js jetstream.JetStream
}

// NewClient establishes connection to NATS JetStream server and ensures default stream exists.
func NewClient(url string) (*Client, error) {
	if url == "" {
		url = nats.DefaultURL
	}

	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS server at %s: %w", url, err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("failed to initialize JetStream context: %w", err)
	}

	// Create or update default stream for DATASPACE_EVENTS
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      "DATASPACE_EVENTS",
		Subjects:  []string{"dataspace.>"},
		Retention: jetstream.WorkQueuePolicy,
		Storage:   jetstream.FileStorage,
	})
	if err != nil {
		// Fallback to limits policy if workqueue fails or exists
		_, _ = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
			Name:     "DATASPACE_EVENTS",
			Subjects: []string{"dataspace.>"},
			Storage:  jetstream.MemoryStorage,
		})
	}

	return &Client{
		nc: nc,
		js: js,
	}, nil
}

// Publish Event sends a structured event to JetStream under subject dataspace.<eventType>.
func (c *Client) Publish(ctx context.Context, subject string, event Event) error {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	fullSubject := "dataspace." + subject
	_, err = c.js.Publish(ctx, fullSubject, data)
	if err != nil {
		return fmt.Errorf("failed to publish event to JetStream subject %s: %w", fullSubject, err)
	}

	return nil
}

// Close closes the NATS connection cleanly.
func (c *Client) Close() {
	if c.nc != nil {
		c.nc.Close()
	}
}
