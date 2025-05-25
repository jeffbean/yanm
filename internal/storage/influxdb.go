package storage

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

// InfluxDBStorage manages sending metrics to InfluxDB
type InfluxDBStorage struct {
	client   influxdb2.Client
	writeAPI api.WriteAPIBlocking
	org      string
	bucket   string

	logger *slog.Logger
}

// Verify InfluxDBStorage implements MetricsStorage interface
var _ MetricsStorage = (*InfluxDBStorage)(nil)

// NewInfluxDBStorage creates a new InfluxDB storage client
func NewInfluxDBStorage(logger *slog.Logger, url, token, org, bucket string) (*InfluxDBStorage, error) {
	if url == "" || token == "" || org == "" || bucket == "" {
		return nil, fmt.Errorf("InfluxDB configuration cannot have empty values")
	}

	// Create InfluxDB client
	client := influxdb2.NewClient(url, token)

	// Create write API
	writeAPI := client.WriteAPIBlocking(org, bucket)

	return &InfluxDBStorage{
		client:   client,
		writeAPI: writeAPI,
		org:      org,
		bucket:   bucket,
		logger:   logger,
	}, nil
}

// StoreNetworkPerformance sends network performance metrics to InfluxDB
func (i *InfluxDBStorage) StoreNetworkPerformance(
	ctx context.Context,
	timestamp time.Time,
	downloadSpeed, uploadSpeed float64,
	ping int64,
	serverName string,
) error {
	// Create point
	point := influxdb2.NewPoint("network_performance",
		map[string]string{
			"server": serverName,
		},
		map[string]interface{}{
			"download_speed": downloadSpeed,
			"upload_speed":   uploadSpeed,
			"ping":           ping,
		},
		timestamp)

	// Write point
	if err := i.writeAPI.WritePoint(ctx, point); err != nil {
		i.logger.ErrorContext(ctx, "Failed to write metrics to InfluxDB", "error", err)
		return fmt.Errorf("failed to write metrics: %v", err)
	}

	i.logger.InfoContext(ctx, "Successfully sent network performance metrics to InfluxDB", "server", serverName)
	return nil
}

// StorePingResult sends ping metrics to InfluxDB
func (i *InfluxDBStorage) StorePingResult(
	ctx context.Context,
	timestamp time.Time,
	latencyMs int64,
	serverName string,
) error {
	// Create point
	point := influxdb2.NewPoint("ping",
		map[string]string{
			"server": serverName,
		},
		map[string]interface{}{
			"latency_ms": latencyMs,
		},
		timestamp)

	// Write point
	if err := i.writeAPI.WritePoint(ctx, point); err != nil {
		i.logger.ErrorContext(ctx, "Failed to write metrics to InfluxDB", "error", err)
		return fmt.Errorf("failed to write metrics: %v", err)
	}

	i.logger.InfoContext(ctx, "Successfully sent ping metrics to InfluxDB", "server", serverName)
	return nil
}

// Close terminates the InfluxDB storage connection
func (i *InfluxDBStorage) Close(_ context.Context) {
	// Close the InfluxDB client
	i.client.Close()
}

func (i *InfluxDBStorage) MetricsHTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	})
}
