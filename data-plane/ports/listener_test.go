package ports_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	cp "github.com/afinana/go-dataspace-components/control-plane/domain"
	dp "github.com/afinana/go-dataspace-components/data-plane/domain"
	"github.com/afinana/go-dataspace-components/data-plane/ports"
)

type mockController struct {
	canHandle bool
	resp      dp.DataFlowResponse
	err       error
}

func (m *mockController) CanHandle(req *dp.DataFlowRequest) bool {
	return m.canHandle
}

func (m *mockController) Initiate(ctx context.Context, req *dp.DataFlowRequest) (dp.DataFlowResponse, error) {
	return m.resp, m.err
}

func TestSignalingListener_HandleStart(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctrl := &mockController{
		canHandle: true,
		resp: dp.DataFlowResponse{
			Success:     true,
			DataPlaneID: "dp-01",
		},
	}

	listener := ports.NewSignalingListener(logger, []dp.DataFlowController{ctrl})
	mux := http.NewServeMux()
	listener.RegisterRoutes(mux)

	payload := dp.DataFlowRequest{
		ID: "flow-123",
		SourceDataAddress: cp.DataAddress{
			Type: "HttpData",
		},
		DestinationDataAddress: cp.DataAddress{
			Type: "HttpProxy",
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/v1/dataflows/start", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status OK 200, got %d", rec.Code)
	}

	var resp dp.DataFlowResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if !resp.Success || resp.DataPlaneID != "dp-01" {
		t.Errorf("unexpected response payload: %+v", resp)
	}
}

func TestSignalingListener_HandleStart_NoController(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctrl := &mockController{
		canHandle: false,
	}

	listener := ports.NewSignalingListener(logger, []dp.DataFlowController{ctrl})
	mux := http.NewServeMux()
	listener.RegisterRoutes(mux)

	payload := dp.DataFlowRequest{
		ID: "flow-unhandled",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/v1/dataflows/start", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status BadRequest 400 when no controller handles request, got %d", rec.Code)
	}
}

func TestSignalingListener_HandleTerminate(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	listener := ports.NewSignalingListener(logger, nil)
	mux := http.NewServeMux()
	listener.RegisterRoutes(mux)

	termPayload := map[string]string{
		"reason": "Contract expired",
	}
	body, _ := json.Marshal(termPayload)

	req := httptest.NewRequest(http.MethodPost, "/v1/dataflows/flow-123/terminate", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status OK 200, got %d", rec.Code)
	}
}
