package storage

import (
	"context"
	"fmt"
	"log"
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
}

// Verify InfluxDBStorage implements MetricsStorage interface
var _ MetricsStorage = (*InfluxDBStorage)(nil)

// NewInfluxDBStorage creates a new InfluxDB storage client
func NewInfluxDBStorage(url, token, org, bucket string) (*InfluxDBStorage, error) {
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
		log.Printf("Failed to write metrics to InfluxDB: %v", err)
		return fmt.Errorf("failed to write metrics: %v", err)
	}

	log.Printf("Successfully sent network performance metrics to InfluxDB (Server: %s)", serverName)
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
		log.Printf("Failed to write metrics to InfluxDB: %v", err)
		return fmt.Errorf("failed to write metrics: %v", err)
	}

	log.Printf("Successfully sent ping metrics to InfluxDB (Server: %s)", serverName)
	return nil
}

// Close terminates the InfluxDB storage connection
func (i *InfluxDBStorage) Close(ctx context.Context) {
	// Close the InfluxDB client
	i.client.Close()
}
