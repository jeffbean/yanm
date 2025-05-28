package storage

import (
	"context"
	"net/http"
	"time"
)

// MetricsStorage defines the interface for storing network performance metrics
//
//go:generate mockgen -source interface.go -destination storagemock/storage_mock.go -package storagemock
type MetricsStorage interface {
	// StoreNetworkPerformance stores the network performance metrics
	StoreNetworkPerformance(
		ctx context.Context,
		timestamp time.Time,
		downloadSpeedMbps, uploadSpeedMbps float64,
		pingMs int64,
		serverName string,
		lat, lon string,
	) error

	// StorePingResult stores the ping result.
	StorePingResult(
		ctx context.Context,
		timestamp time.Time,
		pingMs int64,
		serverName string,
		lat, lon string,
	) error

	// Close terminates the storage connection and performs any final operations
	Close(ctx context.Context)

	// MetricsHTTPHandler returns the HTTP handler for metrics
	MetricsHTTPHandler() http.Handler
}
