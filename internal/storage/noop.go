package storage

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// NoOpStorage is a no-operation implementation of MetricsStorage
// It does nothing when storing metrics, effectively turning off metrics storage
type NoOpStorage struct {
	logger *slog.Logger
}

// Verify that NoOpStorage implements MetricsStorage interface
var _ MetricsStorage = (*NoOpStorage)(nil)

// NewNoOpStorage creates a new NoOpStorage instance
func NewNoOpStorage(logger *slog.Logger) *NoOpStorage {
	return &NoOpStorage{
		logger: logger,
	}
}

// StoreNetworkPerformance does nothing and always returns nil
func (n *NoOpStorage) StoreNetworkPerformance(
	ctx context.Context,
	timestamp time.Time,
	downloadSpeedMbps, uploadSpeedMbps float64,
	ping int64,
	serverName string,
) error {
	n.logger.InfoContext(ctx, "NoOpStorage: logging metrics",
		"downloadSpeedMbps", downloadSpeedMbps,
		"uploadSpeedMbps", uploadSpeedMbps,
		"ping", ping,
		"serverName", serverName)
	return nil
}

// StorePingResult does nothing and always returns nil
func (n *NoOpStorage) StorePingResult(
	ctx context.Context,
	timestamp time.Time,
	pingMs int64,
	serverName string,
) error {
	n.logger.InfoContext(ctx, "NoOpStorage: logging ping result",
		"pingMs", pingMs,
		"serverName", serverName)
	return nil
}

// MetricsHandler returns an http.Handler to optioanlly expose functions.
func (n *NoOpStorage) MetricsHandler() http.Handler {
	return nil
}

// Close does nothing
func (n *NoOpStorage) Close(ctx context.Context) {
	// No-op
}
