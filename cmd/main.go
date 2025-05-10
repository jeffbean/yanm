package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"yanm/internal/config"
	"yanm/internal/monitor"
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
	log.Printf("Configuration: %+v", config)
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

	monitor := monitor.NewNetwork(store, network.NewSpeedTestClient(), config)
	monitor.StartMonitor(ctx)
}
