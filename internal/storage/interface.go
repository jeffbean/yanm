package storage

import (
	"context"
	"time"
)

// MetricsStorage defines the interface for storing network performance metrics
type MetricsStorage interface {
	// StoreNetworkPerformance stores the network performance metrics
	StoreNetworkPerformance(
		ctx context.Context,
		timestamp time.Time,
		downloadSpeedMbps, uploadSpeedMbps float64,
		pingMs int64,
		serverName string,
	) error

	// StorePingResult stores the ping result.
	StorePingResult(
		ctx context.Context,
		timestamp time.Time,
		pingMs int64,
		serverName string,
	) error

	// Close terminates the storage connection and performs any final operations
	Close(ctx context.Context)
}
