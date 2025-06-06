package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log/slog"

	"yanm/internal/config"
	"yanm/internal/debughttp"
	"yanm/internal/debughttp/debughandler"
	"yanm/internal/logger"
	"yanm/internal/monitor"
	"yanm/internal/network"
	"yanm/internal/storage"
)

var (
	configFile string
)

func main() {
	flag.StringVar(&configFile, "config", "config.yml", "Path to the configuration file")
	flag.Parse()

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.LoadFile(configFile)
	if err != nil {
		return err
	}

	logger, err := logger.New(cfg.Logging)
	if err != nil {
		return err
	}

	logger.Info("Yet Another Network Monitor (YANM) starting up...", "configFile", configFile)
	logger.Info("Loaded configuration", "settings", cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var dataStorage storage.MetricsStorage
	switch cfg.Metrics.Engine {
	case "prometheus":
		dataStorage, err = storage.NewPrometheusStorage(logger)
	case "no-op":
		fallthrough
	default:
		dataStorage = storage.NewNoOpStorage(logger)
	}
	if err != nil {
		return err
	}
	defer dataStorage.Close(ctx)

	speedTestClient := network.NewSpeedTestClient(logger)

	// Create handler for the config debug page
	configDebugHandler := config.NewConfigDebugPageProvider(cfg)
	monitorSvc := monitor.NewNetwork(logger, dataStorage, speedTestClient,
		monitor.WithNetworkInterval(time.Duration(cfg.Network.SpeedTest.IntervalMinutes)*time.Minute),
		monitor.WithPingInterval(time.Duration(cfg.Network.PingTest.IntervalSeconds)*time.Second),
		monitor.WithPingTriggerThreshold(time.Duration(cfg.Network.PingTest.ThresholdSeconds)*time.Second),
	)

	routes := []debughttp.DebugRoute{
		{
			Path:        "/debug/speedtest",
			Name:        "Speed Test Results",
			Description: "Displays recent speed test and ping results.",
			Handler:     debughandler.NewHTMLProducingHandler(speedTestClient.Debug()),
		},
		{
			Path:        "/debug/config",
			Name:        "Configuration",
			Description: "Displays the current application configuration.",
			Handler:     debughandler.NewHTMLProducingHandler(configDebugHandler),
		},
		{
			Path:        "/debug/monitor",
			Name:        "Monitor",
			Description: "Controls the monitor service.",
			Handler: debughandler.NewHTMLProducingHandler(
				monitor.NewMonitorDebugPageProvider(monitorSvc)),
		},
	}

	debugSrv, err := setupDebugServer(cfg.DebugServer.ListenAddress, logger, routes)
	if err != nil {
		return err
	}

	if err := debugSrv.RegisterPage(debughttp.DebugRoute{
		Path:        "/metrics",
		Name:        "Metrics",
		Description: "Displays metrics data.",
		Handler:     dataStorage.MetricsHTTPHandler(),
	}); err != nil {
		return err
	}

	if !cfg.DebugServer.Disabled { // setupDebugServer can return nil if disabled
		logger.Info("Starting debug server", "address", cfg.DebugServer.ListenAddress)
		debugSrv.Start(ctx)
		defer func() {
			if err := debugSrv.Stop(ctx); err != nil {
				logger.Error("Failed to stop debug server", "error", err)
			}
		}()
	} else {
		logger.Info("Debug server is not enabled, skipping start.")
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.Info("Received signal, initiating shutdown...", "signal", sig.String())
		cancel()
	}()

	// blocks until ctx is done.
	monitorSvc.Monitor(ctx)
	return nil
}

// setupDebugServer initializes the debug HTTP server and registers all known debug pages.
func setupDebugServer(
	listenAddress string,
	logger *slog.Logger,
	routes []debughttp.DebugRoute,
) (*debughttp.Server, error) {
	debugServerConfig := debughttp.Config{ListenAddress: listenAddress}
	debugSrv, err := debughttp.NewServer(debugServerConfig, logger)
	if err != nil {
		return nil, err
	}

	for _, route := range routes {
		if err := debugSrv.RegisterPage(route); err != nil {
			return nil, err
		}
	}

	return debugSrv, nil
}
