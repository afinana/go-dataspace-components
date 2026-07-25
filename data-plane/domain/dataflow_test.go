package domain

import (
	"context"
	"io"
	"strings"
	"testing"

	cp "github.com/afinana/go-dataspace-components/control-plane/domain"
)

type mockController struct{}

func (m *mockController) CanHandle(request *DataFlowRequest) bool {
	return request.SourceDataAddress.Type == "HttpData"
}

func (m *mockController) Initiate(ctx context.Context, request *DataFlowRequest) (DataFlowResponse, error) {
	if !m.CanHandle(request) {
		return DataFlowResponse{Success: false, ErrorDetail: "unsupported format"}, nil
	}
	return DataFlowResponse{Success: true, DataPlaneID: "dp-01"}, nil
}

type mockSource struct{}

func (s *mockSource) OpenPartStream(ctx context.Context) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("sample stream data")), nil
}

func TestDataFlow_DomainModels(t *testing.T) {
	req := DataFlowRequest{
		ID:                  "flow-01",
		ContractAgreementID: "agreement-01",
		SourceDataAddress: cp.DataAddress{
			Type: "HttpData",
			Properties: map[string]string{
				"endpoint": "http://127.0.0.1:8080/data",
			},
		},
		DestinationDataAddress: cp.DataAddress{
			Type: "HttpProxy",
		},
		Properties: map[string]string{
			"auth_token": "token-xyz",
		},
	}

	ctrl := &mockController{}
	if !ctrl.CanHandle(&req) {
		t.Errorf("expected mockController to handle request")
	}

	resp, err := ctrl.Initiate(context.Background(), &req)
	if err != nil || !resp.Success {
		t.Errorf("expected successful initiation, got err: %v, resp: %+v", err, resp)
	}

	src := &mockSource{}
	rc, err := src.OpenPartStream(context.Background())
	if err != nil {
		t.Fatalf("unexpected error opening stream: %v", err)
	}
	defer rc.Close()
	data, _ := io.ReadAll(rc)
	if string(data) != "sample stream data" {
		t.Errorf("expected 'sample stream data', got %s", string(data))
	}
}
