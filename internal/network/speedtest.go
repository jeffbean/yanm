package network

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/showwin/speedtest-go/speedtest"
	"go.uber.org/multierr"
)

const _pingTimeout = time.Second * 10

// SpeedTestClient implements the SpeedTester interface
type SpeedTestClient struct {
	st *speedtest.Speedtest

	logger *slog.Logger

	mu                 sync.RWMutex
	lastNetworkResults []*PerformanceResult
	lastPingResults    []*PingResult

	// testing fields
	clock clock.Clock
}

// Verify SpeedTestClient implements SpeedTester interface
var _ SpeedTester = (*SpeedTestClient)(nil)

const maxHistory = 10

// NewSpeedTestClient creates a new speed test client
func NewSpeedTestClient(logger *slog.Logger) *SpeedTestClient {
	return &SpeedTestClient{
		st:     speedtest.New(),
		clock:  clock.New(),
		logger: logger,
	}
}

// PerformSpeedTest conducts a network speed test
func (s *SpeedTestClient) PerformSpeedTest(ctx context.Context) (*PerformanceResult, error) {
	serverList, err := s.st.FetchServerListContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch server list: %v", err)
	}

	targets, err := serverList.Available().FindServer([]int{})
	if err != nil {
		return nil, fmt.Errorf("no suitable speedtest servers found: %v", err)
	}

	if len(targets) < 1 {
		return nil, fmt.Errorf("no target ")
	}

	target := targets[0]
	s.logger.DebugContext(ctx, "Selected server", "serverName", target.Name)

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.performTests(ctx, target); err != nil {
		return nil, fmt.Errorf("failed to perform tests: %v", err)
	}

	performance := &PerformanceResult{
		TargetName:        target.Name,
		Timestamp:         s.clock.Now(),
		DownloadSpeedMbps: float64(target.DLSpeed.Mbps()),
		UploadSpeedMbps:   float64(target.ULSpeed.Mbps()),
		PingLatency:       target.Latency,
		Geo:               Geo{Lat: target.Lat, Lon: target.Lon},
	}

	s.lastNetworkResults = append([]*PerformanceResult{performance}, s.lastNetworkResults...)
	if len(s.lastNetworkResults) > maxHistory {
		s.lastNetworkResults = s.lastNetworkResults[:maxHistory]
	}

	return performance, nil
}

func (s *SpeedTestClient) PerformPingTest(ctx context.Context) (*PingResult, error) {
	result := &PingResult{}

	serverList, err := s.st.FetchServerListContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch server list: %v", err)
	}

	targets, err := serverList.Available().FindServer([]int{})
	if err != nil {
		return nil, fmt.Errorf("no suitable speedtest servers found: %v", err)
	}

	if len(targets) < 1 {
		return nil, fmt.Errorf("no target ")
	}

	target := targets[0]
	s.logger.DebugContext(ctx, "Selected server", "serverName", target.Name)

	s.mu.Lock()
	defer s.mu.Unlock()

	_callback := func(latency time.Duration) {
		result.Latency = latency
		result.Timestamp = s.clock.Now()
		result.TargetName = target.Name
		result.Geo = Geo{Lat: target.Lat, Lon: target.Lon}
	}

	s.lastPingResults = append([]*PingResult{result}, s.lastPingResults...)
	if len(s.lastPingResults) > maxHistory {
		s.lastPingResults = s.lastPingResults[:maxHistory]
	}

	pingCtx, cancel := context.WithTimeout(ctx, _pingTimeout)
	defer cancel()
	if err := target.PingTestContext(pingCtx, _callback); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *SpeedTestClient) performTests(ctx context.Context, target *speedtest.Server) error {
	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs error // protected with sync.Mutex
	)

	wg.Add(1)
	go func() {
		defer wg.Done()

		s.logger.InfoContext(ctx, "Testing download speed on server", "serverName", target.Name)
		if err := target.DownloadTestContext(ctx); err != nil {
			mu.Lock()
			defer mu.Unlock()
			errs = multierr.Append(errs, fmt.Errorf("download test failed: %v", err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		s.logger.InfoContext(ctx, "Testing upload speed on server", "serverName", target.Name)
		if err := target.UploadTestContext(ctx); err != nil {
			mu.Lock()
			defer mu.Unlock()
			errs = multierr.Append(errs, fmt.Errorf("upload test failed: %v", err))
		}
	}()

	wg.Wait()
	return errs
}
