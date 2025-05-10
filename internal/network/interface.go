package network

import (
	"context"
	"time"
)

// NetworkPerformance represents the results of a network speed test
type NetworkPerformance struct {
	TargetName string
	Timestamp  time.Time
	// Download speed in Mbps
	DownloadSpeedMbps float64
	// Upload speed in Mbps
	UploadSpeedMbps float64
	// Ping in milliseconds
	PingMs float64
}

// SpeedTester defines the interface for performing network speed tests
type SpeedTester interface {
	// PerformSpeedTest runs a network speed test and returns performance metrics
	PerformSpeedTest(ctx context.Context) (*NetworkPerformance, error)
}
