package storage

import (
	"context"
	"log"
	"time"
)

// NoOpStorage is a no-operation implementation of MetricsStorage
// It does nothing when storing metrics, effectively turning off metrics storage
type NoOpStorage struct{}

// Verify that NoOpStorage implements MetricsStorage interface
var _ MetricsStorage = (*NoOpStorage)(nil)

// NewNoOpStorage creates a new NoOpStorage instance
func NewNoOpStorage() *NoOpStorage {
	return &NoOpStorage{}
}

// StoreNetworkPerformance does nothing and always returns nil
func (n *NoOpStorage) StoreNetworkPerformance(
	ctx context.Context,
	timestamp time.Time,
	downloadSpeedMbps, uploadSpeedMbps float64,
	ping int64,
	serverName string,
) error {
	log.Printf("NoOpStorage: logging metrics: Download: %f Mbps, Upload: %f Mbps, Ping: %d ms, Server: %s",
		downloadSpeedMbps, uploadSpeedMbps, ping, serverName,
	)
	return nil
}

// StorePingResult does nothing and always returns nil
func (n *NoOpStorage) StorePingResult(
	ctx context.Context,
	timestamp time.Time,
	pingMs int64,
	serverName string,
) error {
	log.Printf("NoOpStorage: logging ping result: Ping: %d ms, Server: %s", pingMs, serverName)
	return nil
}

// Close does nothing
func (n *NoOpStorage) Close(ctx context.Context) {
	// No-op
}
