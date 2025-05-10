package monitor

import (
	"context"
	"log"
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

	pingInterval         time.Duration
	networkInterval      time.Duration
	pingTriggerThreshold time.Duration

	clock clock.Clock
}

func NewNetwork(
	storage storage.MetricsStorage,
	client network.SpeedTester,
	config *config.Configuration,
) *Network {
	return &Network{
		storage: storage,
		client:  client,

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
				log.Printf("Ping limiter error: %v", err)
				return
			}

			log.Println("Performing ping check...")
			pingResult, err := m.performPingCheck(ctx)
			if err != nil {
				// TODO: trigger network check for some ping error conditions.
				log.Printf("Ping failed: %v", err)
				continue
			}

			if pingResult != nil && pingResult.Latency > m.pingTriggerThreshold {
				log.Printf("Ping latency is high: %v. Triggering network check.", pingResult.Latency)
				select {
				case triggerNetworkCheck <- struct{}{}: // Non-blocking send
				default:
					log.Println("Network check trigger channel is full. Skipping immediate check.")
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
				log.Println("Network check goroutine stopping...")
				return
			case <-triggerNetworkCheck:
				log.Println("TRIGGER: Performing network check due to high ping latency...")
				if networkLimiter.Allow() { // Respect the limiter even for triggered checks
					m.performNetworkCheck(ctx)
				} else {
					log.Println("Network check rate limit active, triggered check skipped. tokens: ", networkLimiter.Tokens())
				}
			case <-time.Tick(m.networkInterval):
				log.Println("SCHEDULED: Performing network check...")
				if networkLimiter.Allow() {
					m.performNetworkCheck(ctx)
				} else {
					log.Println("Network check rate limit active, scheduled check skipped. tokens: ", networkLimiter.Tokens())
				}
			}
		}
	}()

	log.Println("Monitoring goroutines started.")
	<-ctx.Done() // Wait for context cancellation (e.g., SIGINT)
	log.Println("Shutting down monitor...")
	wg.Wait() // Wait for all goroutines to finish
	log.Println("Monitor shut down gracefully.")
}

func (m *Network) performPingCheck(ctx context.Context) (*network.PingResult, error) {
	pingResult, err := m.client.PerformPingTest(ctx)

	if err != nil {
		log.Printf("Ping failed: %v", err)
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
		log.Printf("Failed to store ping result: %v", err)
	}

	return pingResult, nil
}

func (m *Network) performNetworkCheck(ctx context.Context) {
	speedResult, err := m.client.PerformSpeedTest(ctx)
	if err != nil {
		log.Printf("Speed test failed: %v", err)
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
		log.Printf("Failed to store speed result: %v", err)
	}
}
