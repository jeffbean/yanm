package storage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/push"
)

// PrometheusStorage manages sending metrics to Prometheus
type PrometheusStorage struct {
	pusher        *push.Pusher
	downloadSpeed *prometheus.GaugeVec
	uploadSpeed   *prometheus.GaugeVec
	pingLatency   *prometheus.GaugeVec

	logger *slog.Logger
}

// Verify PrometheusStorage implements MetricsStorage interface
var _ MetricsStorage = (*PrometheusStorage)(nil)

// NewPrometheusStorage creates a new Prometheus storage client
func NewPrometheusStorage(logger *slog.Logger, pushGatewayURL, jobName string) (*PrometheusStorage, error) {
	if pushGatewayURL == "" {
		return nil, fmt.Errorf("push gateway URL cannot be empty")
	}

	// Create metrics using promauto
	downloadSpeed := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "network_download_speed_mbps",
		Help: "Network download speed in Mbps",
	}, []string{"server"})

	uploadSpeed := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "network_upload_speed_mbps",
		Help: "Network upload speed in Mbps",
	}, []string{"server"})

	pingLatency := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "network_ping_latency_ms",
		Help: "Network ping latency in milliseconds",
	}, []string{"server"})

	// Create Prometheus pusher
	pusher := push.New(pushGatewayURL, jobName).
		Collector(downloadSpeed).
		Collector(uploadSpeed).
		Collector(pingLatency)

	return &PrometheusStorage{
		pusher:        pusher,
		downloadSpeed: downloadSpeed,
		uploadSpeed:   uploadSpeed,
		pingLatency:   pingLatency,
		logger:        logger,
	}, nil
}

// StoreNetworkPerformance sends network performance metrics to Prometheus
func (p *PrometheusStorage) StoreNetworkPerformance(
	ctx context.Context,
	timestamp time.Time,
	downloadSpeedMbps, uploadSpeedMbps float64,
	pingMs int64,
	serverName string,
) error {
	// Set metric values
	p.downloadSpeed.WithLabelValues(serverName).Set(downloadSpeedMbps)
	p.uploadSpeed.WithLabelValues(serverName).Set(uploadSpeedMbps)
	p.pingLatency.WithLabelValues(serverName).Set(float64(pingMs))

	// Push metrics to Prometheus Push Gateway
	if err := p.pusher.AddContext(ctx); err != nil {
		p.logger.ErrorContext(ctx, "Failed to push metrics", "error", err)
		return fmt.Errorf("failed to push metrics: %v", err)
	}

	p.logger.InfoContext(ctx, "Successfully sent network performance metrics to Prometheus", "server", serverName)
	return nil
}

func (p *PrometheusStorage) StorePingResult(
	ctx context.Context,
	timestamp time.Time,
	pingMs int64,
	serverName string,
) error {
	// Set metric values with server label
	p.pingLatency.WithLabelValues(serverName).Set(float64(pingMs))

	// Push metrics to Prometheus Push Gateway
	if err := p.pusher.AddContext(ctx); err != nil {
		p.logger.ErrorContext(ctx, "Failed to push metrics", "error", err)
		return fmt.Errorf("failed to push metrics: %v", err)
	}

	p.logger.InfoContext(ctx, "Successfully sent ping result to Prometheus", "server", serverName)
	return nil
}

// Close terminates the Prometheus storage connection
func (p *PrometheusStorage) Close(ctx context.Context) {
	// Optional: Final push before closing
	if err := p.pusher.AddContext(ctx); err != nil {
		p.logger.ErrorContext(ctx, "Error during final metrics push", "error", err)
	}
}
