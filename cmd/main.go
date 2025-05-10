package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	"yanm/internal/config"
	"yanm/internal/logger"
	"yanm/internal/monitor"
	"yanm/internal/network"
	"yanm/internal/storage"
)

func main() {
	configFile := flag.String("config", "config.yml", "Path to the configuration file")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration", err)
	}

	logger, err := logger.New(cfg.Logging)
	if err != nil {
		log.Fatal("Failed to initialize logger", err)
	}

	logger.Info("Yet Another Network Monitor (YANM) starting up...", "configFile", *configFile)
	logger.Info("loaded configuration", "settings", cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		sig := <-sigChan
		logger.Info("Received signal, initiating shutdown...", "signal", sig.String())
		cancel()
		os.Exit(1)
	}()

	var dataStorage storage.MetricsStorage
	switch cfg.Metrics.Engine {
	case "prometheus":
		dataStorage, err = storage.NewPrometheusStorage(
			logger,
			cfg.Metrics.Prometheus.PushGatewayURL,
			cfg.Metrics.Prometheus.JobName,
		)
	case "influxdb":
		dataStorage, err = storage.NewInfluxDBStorage(
			logger,
			cfg.Metrics.InfluxDB.URL,
			cfg.Metrics.InfluxDB.Token,
			cfg.Metrics.InfluxDB.Org,
			cfg.Metrics.InfluxDB.Bucket,
		)
	case "no-op":
		fallthrough // Fallthrough to default for no-op
	default:
		if cfg.Metrics.Engine != "no-op" && cfg.Metrics.Engine != "" {
			logger.Warn("Unsupported metrics engine specified in config, falling back to NoOpStorage.", "specifiedEngine", cfg.Metrics.Engine)
		}
		dataStorage = storage.NewNoOpStorage(logger)
	}
	if err != nil {
		log.Fatal("Failed to initialize metrics storage", "error", err, "engine", cfg.Metrics.Engine)
	}
	defer dataStorage.Close(ctx)

	speedTestClient := network.NewSpeedTestClient(logger)
	monitorSvc := monitor.NewNetwork(logger, dataStorage, speedTestClient, cfg)
	monitorSvc.StartMonitor(ctx)
}
