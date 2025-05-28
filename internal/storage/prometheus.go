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

// want whole numbers, but not linerar.
var _pingBuckets = []float64{
	2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
	20, 25, 30, 35, 40,
	50, 75, 100,
	200, 300, 500, 750, 1000,
	2500, 5000, 10000,
}

// NewPrometheusStorage creates a new Prometheus storage client
func NewPrometheusStorage(logger *slog.Logger) (*PrometheusStorage, error) {
	// Create metrics using promauto
	downloadSpeed := promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:      "network_download_speed_mbps",
		Help:      "Network download speed in Mbps",
		Subsystem: "speedtest",
		Buckets:   prometheus.LinearBuckets(0, 25, 20), // up to 500mbps
	}, []string{"server", "latitude", "longitude"})

	uploadSpeed := promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:      "network_upload_speed_mbps",
		Help:      "Network upload speed in Mbps",
		Subsystem: "speedtest",
		Buckets:   prometheus.LinearBuckets(0, 25, 20), // up to 500mbps
	}, []string{"server", "latitude", "longitude"})

	pingLatency := promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:      "network_latency_ms",
		Help:      "Network ping latency in milliseconds",
		Subsystem: "ping",
		Buckets:   _pingBuckets,
	}, []string{"server", "latitude", "longitude"})

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
	latitude, longitude string,
) error {
	// Set metric values
	p.downloadSpeed.WithLabelValues(serverName, latitude, longitude).Observe(downloadSpeedMbps)
	p.uploadSpeed.WithLabelValues(serverName, latitude, longitude).Observe(uploadSpeedMbps)
	p.pingLatency.WithLabelValues(serverName, latitude, longitude).Observe(float64(pingMs))
	return nil
}

func (p *PrometheusStorage) StorePingResult(
	_ context.Context,
	_ time.Time,
	pingMs int64,
	serverName string,
	latitude, longitude string,
) error {
	// Set metric values with server label
	p.pingLatency.WithLabelValues(serverName, latitude, longitude).Observe(float64(pingMs))
	return nil
}
func (p *PrometheusStorage) MetricsHTTPHandler() http.Handler {
	return p.handler
}

// Close terminates the Prometheus storage connection
func (p *PrometheusStorage) Close(_ context.Context) {}
