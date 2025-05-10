package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"yanm/internal/config"
	"yanm/internal/network"
	"yanm/internal/storage"
)

func main() {
	log.Println("Home Network Internet Monitor")
	log.Println("Initializing monitoring system...")

	// Load configuration
	config, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Start monitoring loop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)
		<-signals
		log.Println("Received interrupt, stopping...")
		cancel()
	}()

	var (
		store storage.MetricsStorage
	)

	switch config.Metrics.Engine {
	case "prometheus":
		store, err = storage.NewPrometheusStorage(
			config.Metrics.Prometheus.PushGatewayURL,
			config.Metrics.Prometheus.JobName,
		)
		if err != nil {
			log.Fatalf("Failed to create Prometheus storage: %v", err)
		}
		defer store.Close(ctx)
		log.Println("Metrics storage enabled")
	case "no-op":
		store = storage.NewNoOpStorage()
		log.Println("Metrics storage disabled")
	default:
		log.Fatalf("Invalid metrics engine: %s", config.Metrics.Engine)
	}

	// Use interval from configuration, default to 5 minutes if not set
	interval := time.Duration(config.Network.Speedtest.IntervalMinutes) * time.Minute

	monitorNetwork(ctx, interval, store)
}

func monitorNetwork(ctx context.Context, interval time.Duration, storage storage.MetricsStorage) {
	performNetworkCheck(ctx, storage)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Starting network monitoring with %s interval", interval)

	for {
		select {
		case <-ticker.C:
			performNetworkCheck(ctx, storage)
		case <-ctx.Done():
			log.Println("Stopping monitoring system...")
			return
		}
	}
}

func performNetworkCheck(ctx context.Context, storage storage.MetricsStorage) {
	log.Println("Performing network health check...")

	// Perform speed test
	speedTester := network.NewSpeedTestClient()
	performance, err := speedTester.PerformSpeedTest(ctx)
	if err != nil {
		log.Printf("Speed test failed: %v", err)
		return
	}

	// Store performance data
	err = storage.StoreNetworkPerformance(
		ctx,
		performance.Timestamp,
		performance.DownloadSpeedMbps,
		performance.UploadSpeedMbps,
		performance.PingMs,
		performance.TargetName,
	)
	if err != nil {
		log.Printf("Failed to store performance data: %v", err)
	}
}
