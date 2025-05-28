package monitor

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
	"yanm/internal/network"
	"yanm/internal/storage"

	"github.com/benbjohnson/clock"
	"golang.org/x/time/rate"
)

const (
	_burstPing    = 1
	_burstNetwork = 3
)

// trackingLimiter is a rate limiter that tracks the limit and the current rate.
type trackingLimiter struct {
	*rate.Limiter

	originalLimit rate.Limit
}

func (l *trackingLimiter) Status() string {
	// returning the current limit is more useful than the stored limit.
	switch l.Limit() {
	case 0:
		return "Paused"
	case rate.Inf:
		return "Unlimited"
	default:
		return fmt.Sprintf("%v / %v", l.Limit(), l.Burst())
	}
}

type Network struct {
	storage storage.MetricsStorage
	client  network.SpeedTester
	logger  *slog.Logger

	pingLimiter          trackingLimiter
	networkLimiter       trackingLimiter
	networkTicker        *time.Ticker
	pingTriggerThreshold time.Duration

	triggerNetworkCheck chan struct{}

	clock clock.Clock
}

func NewNetwork(
	logger *slog.Logger,
	storage storage.MetricsStorage,
	client network.SpeedTester,
	opts ...Option,
) *Network {
	opt := &options{
		pingInterval:         time.Second * 15,
		networkInterval:      time.Minute,
		pingTriggerThreshold: time.Second * 10,
	}

	for _, o := range opts {
		o.apply(opt)
	}

	pingLimit := rate.Every(opt.pingInterval)
	networkLimit := rate.Every(opt.networkInterval)
	m := &Network{
		storage: storage,
		client:  client,
		logger:  logger,

		pingLimiter: trackingLimiter{
			Limiter:       rate.NewLimiter(pingLimit, _burstPing),
			originalLimit: pingLimit,
		},
		networkLimiter: trackingLimiter{
			Limiter:       rate.NewLimiter(networkLimit, _burstNetwork),
			originalLimit: networkLimit,
		},
		networkTicker:        time.NewTicker(opt.networkInterval),
		pingTriggerThreshold: opt.pingTriggerThreshold,

		triggerNetworkCheck: make(chan struct{}, 1),

		clock: clock.New(),
	}

	return m
}

// Monitor starts the monitor asynchronously.
//
// monitoring will stop when the parentContext is done.
func (m *Network) Monitor(ctx context.Context) {
	m.run(ctx)
}

// PausePing pauses the ping checks.
func (m *Network) PausePing() {
	m.pingLimiter.SetLimit(rate.Limit(0))
	m.pingLimiter.SetBurst(0)
}

// ResumePing resumes the ping checks.
func (m *Network) ResumePing() {
	m.pingLimiter.SetLimit(m.pingLimiter.originalLimit)
	m.pingLimiter.SetBurst(_burstPing)
}

// PauseNetwork pauses the network checks.
func (m *Network) PauseNetwork() {
	m.networkLimiter.SetLimit(rate.Limit(0))
	m.networkLimiter.SetBurst(0)
}

// ResumeNetwork resumes the network checks.
func (m *Network) ResumeNetwork() {
	m.networkLimiter.SetLimit(m.networkLimiter.originalLimit)
	m.networkLimiter.SetBurst(_burstNetwork)
}

func (m *Network) run(ctx context.Context) {
	m.logger.InfoContext(ctx, "Starting monitoring loop...")

	var wg sync.WaitGroup

	wg.Add(2)

	// Goroutine for Ping Checks
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				m.logger.InfoContext(ctx, "Ping check goroutine stopping...")
				return
			case <-time.Tick(time.Second):
				if !m.pingLimiter.Allow() {
					continue
				}

				m.logger.DebugContext(ctx, "Performing ping check...")
				pingResult, err := m.performPingCheck(ctx)
				if err != nil {
					// TODO: trigger network check for some ping error conditions.
					m.logger.ErrorContext(ctx, "Ping failed", "error", err)
					continue
				}

				if pingResult != nil && pingResult.Latency > m.pingTriggerThreshold {
					m.logger.InfoContext(ctx, "Ping latency is high", "latency", pingResult.Latency)
					m.triggerNetwork(ctx)
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
			case <-m.triggerNetworkCheck:
				m.logger.DebugContext(ctx, "TRIGGER: Performing network check due to high ping latency...")
				if !m.networkLimiter.Allow() { // Respect the limiter even for triggered checks
					m.logger.InfoContext(ctx, "Network check rate limit active, triggered check skipped.", "tokens", m.networkLimiter.Tokens())
					continue
				}
				m.performNetworkCheck(ctx)
			case <-m.networkTicker.C:
				m.logger.DebugContext(ctx, "SCHEDULED: Performing network check...")
				if !m.networkLimiter.Allow() {
					m.logger.InfoContext(ctx, "Network check rate limit active, scheduled check skipped.", "tokens", m.networkLimiter.Tokens())
					continue
				}
				m.performNetworkCheck(ctx)
			}
		}
	}()

	m.logger.InfoContext(ctx, "Monitoring goroutines started.")
	<-ctx.Done()
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
		pingResult.Geo.Lat,
		pingResult.Geo.Lon,
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
		speedResult.Geo.Lat,
		speedResult.Geo.Lon,
	)
	if err != nil {
		m.logger.ErrorContext(ctx, "Failed to store speed result", "error", err)
	}
}

func (m *Network) triggerNetwork(ctx context.Context) {
	select {
	case m.triggerNetworkCheck <- struct{}{}:
	default:
		m.logger.InfoContext(ctx, "Network check trigger channel is full. Skipping immediate check.")
	}
}
