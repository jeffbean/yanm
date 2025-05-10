package network

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/showwin/speedtest-go/speedtest"
	"go.uber.org/multierr"
)

// SpeedTestClient implements the SpeedTester interface
type SpeedTestClient struct {
	st *speedtest.Speedtest

	// testing fields
	clock clock.Clock
}

// Verify SpeedTestClient implements SpeedTester interface
var _ SpeedTester = (*SpeedTestClient)(nil)

// NewSpeedTestClient creates a new speed test client
func NewSpeedTestClient() *SpeedTestClient {
	return &SpeedTestClient{
		st:    speedtest.New(),
		clock: clock.New(),
	}
}

// PerformSpeedTest conducts a network speed test
func (s *SpeedTestClient) PerformSpeedTest(ctx context.Context) (*NetworkPerformance, error) {
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
	log.Printf("Selected server: %s", target.Name)

	if err := s.performTests(ctx, target); err != nil {
		return nil, fmt.Errorf("failed to perform tests: %v", err)
	}

	performance := &NetworkPerformance{
		TargetName:        target.Name,
		Timestamp:         s.clock.Now(),
		DownloadSpeedMbps: float64(target.DLSpeed.Mbps()),
		UploadSpeedMbps:   float64(target.ULSpeed.Mbps()),
		PingLatency:       target.Latency,
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
	log.Printf("Selected server: %s", target.Name)

	_callback := func(latency time.Duration) {
		result.Latency = latency
		result.Timestamp = s.clock.Now()
		result.TargetName = target.Name
	}

	if err := target.PingTestContext(ctx, _callback); err != nil {
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

		log.Printf("Testing download speed on server: %s", target.Name)
		if err := target.DownloadTestContext(ctx); err != nil {
			mu.Lock()
			defer mu.Unlock()
			errs = multierr.Append(errs, fmt.Errorf("download test failed: %v", err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		log.Printf("Testing upload speed on server: %s", target.Name)
		if err := target.UploadTestContext(ctx); err != nil {
			mu.Lock()
			defer mu.Unlock()
			errs = multierr.Append(errs, fmt.Errorf("upload test failed: %v", err))
		}
	}()

	wg.Wait()
	return errs
}
