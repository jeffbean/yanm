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
		downloadSpeedMbps, uploadSpeedMbps, pingMs float64,
		serverName string,
	) error

	// Close terminates the storage connection and performs any final operations
	Close(ctx context.Context)
}
