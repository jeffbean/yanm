package monitor

import (
	"context"
	"log"
	"log/slog"
	"sync"
	"time"
	"yanm/internal/config"
	"yanm/internal/network"
	"yanm/internal/storage"

	"github.com/benbjohnson/clock"
	"golang.org/x/time/rate"
)

type Network struct {
	storage storage.MetricsStorage
	client  network.SpeedTester
	logger  *slog.Logger

	pingInterval         time.Duration
	networkInterval      time.Duration
	pingTriggerThreshold time.Duration

	clock clock.Clock
}

func NewNetwork(
	logger *slog.Logger,
	storage storage.MetricsStorage,
	client network.SpeedTester,
	config *config.Configuration,
) *Network {
	return &Network{
		storage: storage,
		client:  client,
		logger:  logger,

		pingInterval:         time.Duration(config.Network.PingTest.IntervalSeconds) * time.Second,
		networkInterval:      time.Duration(config.Network.SpeedTest.IntervalMinutes) * time.Minute,
		pingTriggerThreshold: time.Duration(config.Network.PingTest.ThresholdSeconds) * time.Second,
		clock:                clock.New(),
	}
}

func (m *Network) StartMonitor(ctx context.Context) {
	log.Println("Starting monitoring loop...")

	networkLimiter := rate.NewLimiter(rate.Every(m.networkInterval), 3) // Allow bursting for network checks
	pingLimiter := rate.NewLimiter(rate.Every(m.pingInterval), 1)

	var (
		triggerNetworkCheck = make(chan struct{}, 1)
		wg                  sync.WaitGroup
	)

	wg.Add(2)

	// Goroutine for Ping Checks
	go func() {
		defer wg.Done()

		for {
			if err := pingLimiter.Wait(ctx); err != nil {
				m.logger.ErrorContext(ctx, "Ping limiter error", "error", err)
				return
			}

			m.logger.InfoContext(ctx, "Performing ping check...")
			pingResult, err := m.performPingCheck(ctx)
			if err != nil {
				// TODO: trigger network check for some ping error conditions.
				m.logger.ErrorContext(ctx, "Ping failed", "error", err)
				continue
			}

			if pingResult != nil && pingResult.Latency > m.pingTriggerThreshold {
				m.logger.InfoContext(ctx, "Ping latency is high", "latency", pingResult.Latency)
				select {
				case triggerNetworkCheck <- struct{}{}: // Non-blocking send
				default:
					m.logger.InfoContext(ctx, "Network check trigger channel is full. Skipping immediate check.")
				}
			}
		}
	}()

	// Goroutine for Network Checks
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				m.logger.InfoContext(ctx, "Network check goroutine stopping...")
				return
			case <-triggerNetworkCheck:
				m.logger.InfoContext(ctx, "TRIGGER: Performing network check due to high ping latency...")
				if networkLimiter.Allow() { // Respect the limiter even for triggered checks
					m.performNetworkCheck(ctx)
				} else {
					m.logger.InfoContext(ctx, "Network check rate limit active, triggered check skipped.", "tokens", networkLimiter.Tokens())
				}
			case <-time.Tick(m.networkInterval):
				m.logger.InfoContext(ctx, "SCHEDULED: Performing network check...")
				if networkLimiter.Allow() {
					m.performNetworkCheck(ctx)
				} else {
					m.logger.InfoContext(ctx, "Network check rate limit active, scheduled check skipped.", "tokens", networkLimiter.Tokens())
				}
			}
		}
	}()

	m.logger.InfoContext(ctx, "Monitoring goroutines started.")
	<-ctx.Done() // Wait for context cancellation (e.g., SIGINT)
	m.logger.InfoContext(ctx, "Shutting down monitor...")
	wg.Wait() // Wait for all goroutines to finish
	m.logger.InfoContext(ctx, "Monitor shut down gracefully.")
}

func (m *Network) performPingCheck(ctx context.Context) (*network.PingResult, error) {
	pingResult, err := m.client.PerformPingTest(ctx)

	if err != nil {
		m.logger.ErrorContext(ctx, "Ping failed", "error", err)
		return nil, err
	}

	// Store ping result
	err = m.storage.StorePingResult(
		ctx,
		m.clock.Now(),
		pingResult.Latency.Milliseconds(),
		pingResult.TargetName,
	)
	if err != nil {
		m.logger.ErrorContext(ctx, "Failed to store ping result", "error", err)
	}

	return pingResult, nil
}

func (m *Network) performNetworkCheck(ctx context.Context) {
	speedResult, err := m.client.PerformSpeedTest(ctx)
	if err != nil {
		m.logger.ErrorContext(ctx, "Speed test failed", "error", err)
		return
	}

	// Store speed result
	err = m.storage.StoreNetworkPerformance(
		ctx,
		m.clock.Now(),
		speedResult.DownloadSpeedMbps,
		speedResult.UploadSpeedMbps,
		speedResult.PingLatency.Milliseconds(),
		speedResult.TargetName,
	)
	if err != nil {
		m.logger.ErrorContext(ctx, "Failed to store speed result", "error", err)
	}
}
