package network

import (
	"context"
	"net/http"
	"time"
)

// Performance represents the results of a network speed test
//
//go:generate mockgen -source interface.go -destination networkmock/storage_mock.go -package networkmock
type Performance struct {
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
	PerformSpeedTest(ctx context.Context) (*Performance, error)

	// PerformPingTest runs a network ping and returns the latency in milliseconds.
	PerformPingTest(ctx context.Context) (*PingResult, error)

	// Debug returns a DebugRoute to optioanlly expose functions.
	Debug() http.Handler
}
