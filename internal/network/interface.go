package network

import (
	"context"
	"net/http"
	"time"
)

type Geo struct {
	Lat string
	Lon string
}

// PerformanceResult represents the results of a network speed test
//
//go:generate mockgen -source interface.go -destination networkmock/storage_mock.go -package networkmock
type PerformanceResult struct {
	TargetName        string
	Timestamp         time.Time
	DownloadSpeedMbps float64
	UploadSpeedMbps   float64
	PingLatency       time.Duration
	Geo               Geo
}

// PingResult represents the result of a network ping
type PingResult struct {
	TargetName string
	Timestamp  time.Time
	Latency    time.Duration
	Geo        Geo
}

// SpeedTester defines the interface for performing network speed tests
type SpeedTester interface {
	// PerformSpeedTest runs a network speed test and returns performance metrics
	PerformSpeedTest(ctx context.Context) (*PerformanceResult, error)

	// PerformPingTest runs a network ping and returns the latency in milliseconds.
	PerformPingTest(ctx context.Context) (*PingResult, error)

	// Debug returns a DebugRoute to optioanlly expose functions.
	Debug() http.Handler
}
