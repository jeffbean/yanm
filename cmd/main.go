package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"yanm/internal/network"
	"yanm/internal/storage"
)

func main() {
	fmt.Println("Home Network Internet Monitor")
	log.Println("Initializing monitoring system...")

	// Prometheus Push Gateway configuration
	pushGatewayURL := os.Getenv("PROMETHEUS_PUSH_GATEWAY_URL")
	jobName := os.Getenv("PROMETHEUS_JOB_NAME")

	if pushGatewayURL == "" {
		log.Fatal("Missing Prometheus Push Gateway configuration. Set PROMETHEUS_PUSH_GATEWAY_URL")
	}

	// Use default job name if not set
	if jobName == "" {
		jobName = "home_network_monitor"
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

	// Initialize Prometheus storage
	storage, err := storage.NewPrometheusStorage(pushGatewayURL, jobName)
	if err != nil {
		log.Fatalf("Failed to create Prometheus storage: %v", err)
	}
	defer storage.Close(ctx)

	monitorNetwork(ctx, storage)
}

func monitorNetwork(ctx context.Context, storage storage.MetricsStorage) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping monitoring system...")
			return
		case <-ticker.C:
			performNetworkCheck(ctx, storage)
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
