package domain

import (
	"context"
	"io"

	cp "github.com/afinana/go-dataspace-components/control-plane/domain"
)

// DataFlowRequest defines the command payload dispatched from the Control Plane
// to the Data Plane to trigger a data transfer process.
type DataFlowRequest struct {
	ID                     string         `json:"id"`
	ContractAgreementID    string         `json:"contractAgreementId"`
	SourceDataAddress      cp.DataAddress `json:"sourceDataAddress"`
	DestinationDataAddress cp.DataAddress `json:"destinationDataAddress"`
	Properties             map[string]string `json:"properties,omitempty"`
}

// DataFlowResponse contains the result details of initiating a flow request.
type DataFlowResponse struct {
	Success      bool   `json:"success"`
	DataPlaneID  string `json:"dataPlaneId"`
	ErrorDetail  string `json:"errorDetail,omitempty"`
}

// DataFlowController represents the core Port in the Data Plane.
// Every backend adapter (e.g., HTTP proxy, S3 streamer) implements this interface
// to handle specific Source/Destination address types.
type DataFlowController interface {
	// CanHandle inspects the request addresses to verify if this controller can orchestrate the flow.
	CanHandle(request *DataFlowRequest) bool

	// Initiate triggers the actual data flow in a non-blocking or blocking manner.
	Initiate(ctx context.Context, request *DataFlowRequest) (DataFlowResponse, error)
}

// DataSource represents the source boundary of a data flow.
type DataSource interface {
	// OpenPartStream opens a partition/part stream from the asset storage block.
	OpenPartStream(ctx context.Context) (io.ReadCloser, error)
}

// DataSink represents the destination boundary of a data flow.
type DataSink interface {
	io.Writer
}
