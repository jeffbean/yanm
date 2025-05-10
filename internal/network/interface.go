package network

import (
	"context"
	"time"
)

// NetworkPerformance represents the results of a network speed test
type NetworkPerformance struct {
	TargetName        string
	Timestamp         time.Time
	DownloadSpeedMbps float64
	UploadSpeedMbps   float64
	PingLatency       time.Duration
}

// PingResult represents the result of a network ping
type PingResult struct {
	TargetName string
	Timestamp  time.Time
	Latency    time.Duration
}

// SpeedTester defines the interface for performing network speed tests
type SpeedTester interface {
	// PerformSpeedTest runs a network speed test and returns performance metrics
	PerformSpeedTest(ctx context.Context) (*NetworkPerformance, error)

	// PerformPingTest runs a network ping and returns the latency in milliseconds.
	PerformPingTest(ctx context.Context) (*PingResult, error)
}
