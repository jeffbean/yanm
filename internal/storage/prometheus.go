package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/push"
)

// PrometheusStorage manages sending metrics to Prometheus
type PrometheusStorage struct {
	pusher        *push.Pusher
	downloadSpeed prometheus.Gauge
	uploadSpeed   prometheus.Gauge
	pingLatency   prometheus.Gauge
}

// Verify PrometheusStorage implements MetricsStorage interface
var _ MetricsStorage = (*PrometheusStorage)(nil)

// NewPrometheusStorage creates a new Prometheus storage client
func NewPrometheusStorage(pushGatewayURL, jobName string) (*PrometheusStorage, error) {
	if pushGatewayURL == "" {
		return nil, fmt.Errorf("push gateway URL cannot be empty")
	}

	// Create metrics using promauto
	downloadSpeed := promauto.NewGauge(prometheus.GaugeOpts{
		Name: "network_download_speed_mbps",
		Help: "Network download speed in Mbps",
	})

	uploadSpeed := promauto.NewGauge(prometheus.GaugeOpts{
		Name: "network_upload_speed_mbps",
		Help: "Network upload speed in Mbps",
	})

	pingLatency := promauto.NewGauge(prometheus.GaugeOpts{
		Name: "network_ping_latency_ms",
		Help: "Network ping latency in milliseconds",
	})

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
	}, nil
}

// StoreNetworkPerformance sends network performance metrics to Prometheus
func (p *PrometheusStorage) StoreNetworkPerformance(
	ctx context.Context,
	timestamp time.Time,
	downloadSpeedMbps, uploadSpeedMbps, pingMs float64,
	serverName string,
) error {
	// Set metric values
	p.downloadSpeed.Set(downloadSpeedMbps)
	p.uploadSpeed.Set(uploadSpeedMbps)
	p.pingLatency.Set(pingMs)

	// Push metrics to Prometheus Push Gateway
	if err := p.pusher.AddContext(ctx); err != nil {
		log.Printf("Failed to push metrics: %v", err)
		return fmt.Errorf("failed to push metrics: %v", err)
	}

	log.Printf("Successfully sent network performance metrics to Prometheus (Server: %s)", serverName)
	return nil
}

// Close terminates the Prometheus storage connection
func (p *PrometheusStorage) Close(ctx context.Context) {
	// Optional: Final push before closing
	if err := p.pusher.AddContext(ctx); err != nil {
		log.Printf("Error during final metrics push: %v", err)
	}
}
