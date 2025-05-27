package storage

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusStorage manages sending metrics to Prometheus
type PrometheusStorage struct {
	handler       http.Handler
	downloadSpeed *prometheus.HistogramVec
	uploadSpeed   *prometheus.HistogramVec
	pingLatency   *prometheus.HistogramVec

	logger *slog.Logger
}

// Verify PrometheusStorage implements MetricsStorage interface
var _ MetricsStorage = (*PrometheusStorage)(nil)

// NewPrometheusStorage creates a new Prometheus storage client
func NewPrometheusStorage(logger *slog.Logger) (*PrometheusStorage, error) {
	// Create metrics using promauto
	downloadSpeed := promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:      "network_download_speed_mbps",
		Help:      "Network download speed in Mbps",
		Subsystem: "speedtest",
		Buckets:   prometheus.LinearBuckets(0, 25, 20), // up to 500mbps
	}, []string{"server"})

	uploadSpeed := promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:      "network_upload_speed_mbps",
		Help:      "Network upload speed in Mbps",
		Subsystem: "speedtest",
		Buckets:   prometheus.LinearBuckets(0, 25, 20), // up to 500mbps
	}, []string{"server"})

	pingLatency := promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:      "network_latency_ms",
		Help:      "Network ping latency in milliseconds",
		Subsystem: "ping",
		Buckets:   prometheus.ExponentialBucketsRange(1, 4096, 32), // up to 4 seconds, 32 buckets
	}, []string{"server"})

	return &PrometheusStorage{
		handler:       promhttp.Handler(),
		downloadSpeed: downloadSpeed,
		uploadSpeed:   uploadSpeed,
		pingLatency:   pingLatency,
		logger:        logger,
	}, nil
}

// StoreNetworkPerformance sends network performance metrics to Prometheus
func (p *PrometheusStorage) StoreNetworkPerformance(
	_ context.Context,
	_ time.Time,
	downloadSpeedMbps, uploadSpeedMbps float64,
	pingMs int64,
	serverName string,
) error {
	// Set metric values
	p.downloadSpeed.WithLabelValues(serverName).Observe(downloadSpeedMbps)
	p.uploadSpeed.WithLabelValues(serverName).Observe(uploadSpeedMbps)
	p.pingLatency.WithLabelValues(serverName).Observe(float64(pingMs))
	return nil
}

func (p *PrometheusStorage) StorePingResult(
	_ context.Context,
	_ time.Time,
	pingMs int64,
	serverName string,
) error {
	// Set metric values with server label
	p.pingLatency.WithLabelValues(serverName).Observe(float64(pingMs))
	return nil
}
func (p *PrometheusStorage) MetricsHTTPHandler() http.Handler {
	return p.handler
}

// Close terminates the Prometheus storage connection
func (p *PrometheusStorage) Close(_ context.Context) {}
