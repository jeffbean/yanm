package storage

import (
	"context"
	"log"
	"time"
)

// NoOpStorage is a no-operation implementation of MetricsStorage
// It does nothing when storing metrics, effectively turning off metrics storage
type NoOpStorage struct{}

// NewNoOpStorage creates a new NoOpStorage instance
func NewNoOpStorage() *NoOpStorage {
	return &NoOpStorage{}
}

// StoreNetworkPerformance does nothing and always returns nil
func (n *NoOpStorage) StoreNetworkPerformance(
	ctx context.Context,
	timestamp time.Time,
	downloadSpeedMbps, uploadSpeedMbps, pingMs float64,
	serverName string,
) error {
	log.Printf("NoOpStorage: logging metrics: Download: %f Mbps, Upload: %f Mbps, Ping: %f ms, Server: %s",
		downloadSpeedMbps, uploadSpeedMbps, pingMs, serverName,
	)
	return nil
}

// Close does nothing
func (n *NoOpStorage) Close(ctx context.Context) {
	// No-op
}

// Verify that NoOpStorage implements MetricsStorage interface
var _ MetricsStorage = (*NoOpStorage)(nil)
