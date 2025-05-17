package monitor

import (
	"context"
	"log"
	"log/slog"
	"sync"
	"time"
	"yanm/internal/network"
	"yanm/internal/storage"

	"github.com/benbjohnson/clock"
	"golang.org/x/time/rate"
)

type Network struct {
	storage storage.MetricsStorage
	client  network.SpeedTester
	logger  *slog.Logger

	pingLimiter          *rate.Limiter
	networkLimiter       *rate.Limiter
	networkTicker        *time.Ticker
	pingTriggerThreshold time.Duration

	stop                chan struct{}
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

	m := &Network{
		storage: storage,
		client:  client,
		logger:  logger,

		pingLimiter:          rate.NewLimiter(rate.Every(opt.pingInterval), 1),
		networkLimiter:       rate.NewLimiter(rate.Every(opt.networkInterval), 3),
		networkTicker:        time.NewTicker(opt.networkInterval),
		pingTriggerThreshold: opt.pingTriggerThreshold,

		stop:                make(chan struct{}, 1),
		triggerNetworkCheck: make(chan struct{}, 1),

		clock: clock.New(),
	}

	return m
}

// StartMonitor starts the monitor asynchronously.
//
// monitoring will stop when the parentContext is done.
func (m *Network) StartMonitor(parentContext context.Context) {
	go m.run(parentContext)
}

// StopMonitor stops the monitor.
func (m *Network) StopMonitor() {
	close(m.stop)
}

func (m *Network) run(parentContext context.Context) {
	ctx, cancel := context.WithCancel(parentContext)
	m.registerCancel(cancel)

	log.Println("Starting monitoring loop...")

	var wg sync.WaitGroup

	wg.Add(2)

	// Goroutine for Ping Checks
	go func() {
		defer wg.Done()

		for {
			if err := m.pingLimiter.Wait(ctx); err != nil {
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
				m.triggerNetwork(ctx)
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
				m.logger.InfoContext(ctx, "TRIGGER: Performing network check due to high ping latency...")
				if !m.networkLimiter.Allow() { // Respect the limiter even for triggered checks
					m.logger.InfoContext(ctx, "Network check rate limit active, triggered check skipped.", "tokens", m.networkLimiter.Tokens())
					return
				}
				m.performNetworkCheck(ctx)
			case <-m.networkTicker.C:
				m.logger.InfoContext(ctx, "SCHEDULED: Performing network check...")
				if !m.networkLimiter.Allow() {
					m.logger.InfoContext(ctx, "Network check rate limit active, scheduled check skipped.", "tokens", m.networkLimiter.Tokens())
					return
				}
				m.performNetworkCheck(ctx)
			}
		}
	}()

	m.logger.InfoContext(ctx, "Monitoring goroutines started.")
	wg.Wait() // Wait for all goroutines to finish
	m.logger.InfoContext(ctx, "Monitor shut down gracefully.")
}

func (m *Network) registerCancel(cancel context.CancelFunc) {
	go func() {
		// signal to goroutines here to stop.
		<-m.stop
		m.logger.InfoContext(context.Background(), "stop signal received, canceling context.")
		cancel()
	}()
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

func (m *Network) triggerNetwork(ctx context.Context) {
	select {
	case m.triggerNetworkCheck <- struct{}{}:
	default:
		m.logger.InfoContext(ctx, "Network check trigger channel is full. Skipping immediate check.")
	}
}
