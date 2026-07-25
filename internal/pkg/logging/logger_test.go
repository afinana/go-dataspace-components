package logging_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/afinana/go-dataspace-components/internal/pkg/logging"
)

func TestInitLogger_Levels(t *testing.T) {
	levels := []string{"DEBUG", "INFO", "WARN", "ERROR", "UNKNOWN"}

	for _, lvl := range levels {
		logger := logging.InitLogger(lvl)
		if logger == nil {
			t.Errorf("expected non-nil logger for level '%s'", lvl)
		}
	}
}

func TestWithContext(t *testing.T) {
	logger := slog.Default()
	ctx := context.Background()

	ctxLogger := logging.WithContext(ctx, logger)
	if ctxLogger == nil {
		t.Error("expected non-nil logger from WithContext")
	}
}
