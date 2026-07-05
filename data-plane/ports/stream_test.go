package ports

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	cp "github.com/afinana/go-dataspace-components/control-plane/domain"
	dp "github.com/afinana/go-dataspace-components/data-plane/domain"
)

func TestFileStreamController_Transfer(t *testing.T) {
	// Create logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	controller := NewFileStreamController(logger)

	// Create temp directory for testing local file streaming
	tmpDir := t.TempDir()
	sourcePath := filepath.Join(tmpDir, "source.txt")
	destPath := filepath.Join(tmpDir, "dest.txt")

	// Write mock test data
	testContent := "Sovereign Data Exchange via Custom Go Dataspace Connector Scaffold"
	err := os.WriteFile(sourcePath, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	// Prepare data flow request configuration
	req := &dp.DataFlowRequest{
		ID: "test-flow-01",
		SourceDataAddress: cp.DataAddress{
			Type: "LocalFile",
			Properties: map[string]string{
				"path": sourcePath,
			},
		},
		DestinationDataAddress: cp.DataAddress{
			Type: "LocalFile",
			Properties: map[string]string{
				"path": destPath,
			},
		},
	}

	// Execute transfer
	ctx := context.Background()
	err = controller.executeTransfer(ctx, req)
	if err != nil {
		t.Fatalf("executeTransfer returned error: %v", err)
	}

	// Verify target contents match exactly
	copiedContent, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("failed to read copied destination file: %v", err)
	}

	if string(copiedContent) != testContent {
		t.Errorf("copied content mismatch: got %q, expected %q", string(copiedContent), testContent)
	}
}
