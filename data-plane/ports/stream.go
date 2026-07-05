package ports

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	cp "github.com/afinana/go-dataspace-components/control-plane/domain"
	dp "github.com/afinana/go-dataspace-components/data-plane/domain"
)

// FileStreamController implements dp.DataFlowController to stream large assets (e.g., local files, S3 blocks)
// safely to HTTP destinations or response streams without loading them into memory.
type FileStreamController struct {
	logger *slog.Logger
}

// NewFileStreamController initializes the streaming controller.
func NewFileStreamController(logger *slog.Logger) *FileStreamController {
	return &FileStreamController{logger: logger}
}

// CanHandle determines if this controller handles file or object storage streaming.
func (c *FileStreamController) CanHandle(req *dp.DataFlowRequest) bool {
	srcType := req.SourceDataAddress.Type
	destType := req.DestinationDataAddress.Type
	return (srcType == "LocalFile" || srcType == "AmazonS3" || srcType == "GoogleCloudStorage") &&
		(destType == "HttpData" || destType == "LocalFile")
}

// Initiate executes the streaming data flow (Push model).
// It opens the source reader (local file or external storage API) and writes it directly to the destination endpoint.
func (c *FileStreamController) Initiate(ctx context.Context, req *dp.DataFlowRequest) (dp.DataFlowResponse, error) {
	if !c.CanHandle(req) {
		return dp.DataFlowResponse{Success: false, ErrorDetail: "Unsupported data address types"}, nil
	}

	go func() {
		transferCtx := context.Background()
		c.logger.Info("Starting file stream data transfer", "transferId", req.ID)
		
		err := c.executeTransfer(transferCtx, req)
		if err != nil {
			c.logger.Error("File stream transfer failed", "transferId", req.ID, "err", err)
			return
		}
		c.logger.Info("File stream transfer completed successfully", "transferId", req.ID)
	}()

	return dp.DataFlowResponse{
		Success:     true,
		DataPlaneID: "file-streamer-01",
	}, nil
}

// executeTransfer orchestrates opening the source reader and copying to the destination writer.
func (c *FileStreamController) executeTransfer(ctx context.Context, req *dp.DataFlowRequest) error {
	// 1. Resolve the DataSource
	dataSource, err := c.resolveDataSource(&req.SourceDataAddress)
	if err != nil {
		return fmt.Errorf("failed to resolve data source: %w", err)
	}

	// 2. Open part stream
	srcStream, err := dataSource.OpenPartStream(ctx)
	if err != nil {
		return fmt.Errorf("failed to open part stream: %w", err)
	}
	defer srcStream.Close()

	// 3. Resolve destination writer (DataSink)
	destWriter, destCloser, err := c.resolveDestinationSink(ctx, &req.DestinationDataAddress)
	if err != nil {
		return fmt.Errorf("failed to resolve destination sink: %w", err)
	}
	defer destCloser()

	// 4. Stream data using a fixed-size 32KB circular buffer to enforce constant memory footprint
	buffer := make([]byte, 32*1024) // 32KB buffer
	written, err := io.CopyBuffer(destWriter, srcStream, buffer)
	if err != nil {
		return fmt.Errorf("streaming copy failed: %w", err)
	}

	c.logger.Info("Data streaming complete", "bytesWritten", written)
	return nil
}

// Data Source implementations

type localFileDataSource struct {
	path string
}

func (s *localFileDataSource) OpenPartStream(ctx context.Context) (io.ReadCloser, error) {
	cleanPath := filepath.Clean(s.path)
	return os.Open(cleanPath)
}

type cloudStorageDataSource struct {
	logger   *slog.Logger
	provider string
	bucket   string
	key      string
}

func (s *cloudStorageDataSource) OpenPartStream(ctx context.Context) (io.ReadCloser, error) {
	s.logger.Info("Opening stream from object storage bucket", "provider", s.provider, "bucket", s.bucket, "key", s.key)
	// Simulating an active reader of a large object block (10MB dummy stream)
	simulatedReader := io.LimitReader(newDummyDataGenerator(), 10*1024*1024)
	return io.NopCloser(simulatedReader), nil
}

func (c *FileStreamController) resolveDataSource(addr *cp.DataAddress) (dp.DataSource, error) {
	switch addr.Type {
	case "LocalFile":
		path := addr.GetProperty("path")
		if path == "" {
			return nil, fmt.Errorf("missing path property for LocalFile source")
		}
		return &localFileDataSource{path: path}, nil

	case "AmazonS3", "GoogleCloudStorage":
		bucket := addr.GetProperty("bucket")
		key := addr.GetProperty("key")
		return &cloudStorageDataSource{
			logger:   c.logger,
			provider: addr.Type,
			bucket:   bucket,
			key:      key,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported source data address type: %s", addr.Type)
	}
}

// resolveDestinationSink resolves the output sink where data will be pushed.
func (c *FileStreamController) resolveDestinationSink(ctx context.Context, addr *cp.DataAddress) (io.Writer, func() error, error) {
	switch addr.Type {
	case "LocalFile":
		path := addr.GetProperty("path")
		if path == "" {
			return nil, nil, fmt.Errorf("missing path property for LocalFile destination")
		}
		cleanPath := filepath.Clean(path)
		file, err := os.Create(cleanPath)
		if err != nil {
			return nil, nil, err
		}
		return file, file.Close, nil

	case "HttpData":
		endpoint := addr.GetProperty("endpoint")
		if endpoint == "" {
			return nil, nil, fmt.Errorf("missing endpoint for HttpData destination")
		}

		// Set up an HTTP client request that streams its body via pipe.
		pr, pw := io.Pipe()

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, pr)
		if err != nil {
			pr.Close()
			pw.Close()
			return nil, nil, err
		}

		req.ContentLength = -1 // Triggers HTTP/1.1 chunked transfer encoding
		req.Header.Set("Content-Type", "application/octet-stream")

		if authSecret := addr.GetProperty("authSecret"); authSecret != "" {
			req.Header.Set("Authorization", "Bearer "+authSecret)
		}

		errChan := make(chan error, 1)
		go func() {
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				errChan <- err
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
				errChan <- fmt.Errorf("http upload failed with status %d: %s", resp.StatusCode, string(bodyBytes))
				return
			}
			errChan <- nil
		}()

		writerCloser := func() error {
			pw.Close()
			return <-errChan
		}

		return pw, writerCloser, nil

	default:
		return nil, nil, fmt.Errorf("unsupported destination data address type: %s", addr.Type)
	}
}

// dummyDataGenerator simulates a large continuous data source block
type dummyDataGenerator struct{}

func newDummyDataGenerator() io.Reader {
	return &dummyDataGenerator{}
}

func (d *dummyDataGenerator) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = 'A' // Fill with mock bytes
	}
	return len(p), nil
}
