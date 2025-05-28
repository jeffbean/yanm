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
	_ time.Time,
	downloadSpeedMbps, uploadSpeedMbps float64,
	ping int64,
	serverName string,
	lat, lon string,
) error {
	n.logger.InfoContext(ctx, "NoOpStorage: logging metrics",
		"downloadSpeedMbps", downloadSpeedMbps,
		"uploadSpeedMbps", uploadSpeedMbps,
		"ping", ping,
		"serverName", serverName,
		"lat", lat,
		"lon", lon)
	return nil
}

// StorePingResult does nothing and always returns nil
func (n *NoOpStorage) StorePingResult(
	ctx context.Context,
	_ time.Time,
	pingMs int64,
	serverName string,
	lat, lon string,
) error {
	n.logger.InfoContext(ctx, "NoOpStorage: logging ping result",
		"pingMs", pingMs,
		"serverName", serverName,
		"lat", lat,
		"lon", lon)
	return nil
}

// Close does nothing
func (n *NoOpStorage) Close(_ context.Context) {
	// No-op
}

func (n *NoOpStorage) MetricsHTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	})
}
