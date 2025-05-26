package config

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"yanm/internal/logger"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestConfigPage_ServeHTTP(t *testing.T) {
	// Create a sample configuration
	sampleConfig := &Configuration{
		Network: struct {
			PingTest struct {
				IntervalSeconds  int     `yaml:"interval_seconds"`
				ThresholdSeconds float64 `yaml:"threshold_seconds"`
			} `yaml:"ping_test"`
			SpeedTest struct {
				IntervalMinutes int `yaml:"interval_minutes"`
				Servers         struct {
					MaxPingTimeout   string `yaml:"max_ping_timeout"`
					MaxServersToTest int    `yaml:"max_servers_to_test"`
				} `yaml:"servers"`
			} `yaml:"speedtest"`
		}{
			PingTest: struct {
				IntervalSeconds  int     `yaml:"interval_seconds"`
				ThresholdSeconds float64 `yaml:"threshold_seconds"`
			}{
				IntervalSeconds:  60,
				ThresholdSeconds: 5.0, // Example value
			},
			SpeedTest: struct {
				IntervalMinutes int `yaml:"interval_minutes"`
				Servers         struct {
					MaxPingTimeout   string `yaml:"max_ping_timeout"`
					MaxServersToTest int    `yaml:"max_servers_to_test"`
				} `yaml:"servers"`
			}{
				IntervalMinutes: 60,
				Servers: struct {
					MaxPingTimeout   string `yaml:"max_ping_timeout"`
					MaxServersToTest int    `yaml:"max_servers_to_test"`
				}{
					MaxPingTimeout:   "1s", // Example value
					MaxServersToTest: 5,    // Example value
				},
			},
		},
		Metrics: struct {
			Engine     string `yaml:"engine"`
			Prometheus struct {
			} `yaml:"prometheus"`
			InfluxDB struct {
				URL    string `yaml:"url"`
				Token  string `yaml:"token"`
				Org    string `yaml:"org"`
				Bucket string `yaml:"bucket"`
			} `yaml:"influxdb"`
		}{
			Engine: "prometheus",
		},
		Logging: logger.Config{
			Level:  "info",
			Format: "text",
		},
		DebugServer: struct {
			Disabled      bool   `yaml:"disabled"`
			ListenAddress string `yaml:"listen_address"`
		}{
			Disabled:      false,
			ListenAddress: ":8081",
		},
	}

	expectedYAMLBytes, err := yaml.Marshal(sampleConfig)
	if err != nil {
		t.Fatalf("Failed to marshal sample config to YAML: %v", err)
	}
	expectedYAMLString := string(expectedYAMLBytes)

	req := httptest.NewRequest("GET", "/debug/config", nil)
	rr := httptest.NewRecorder()

	handler := NewConfigDebugPageProvider(sampleConfig)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code")

	body := rr.Body.String()
	expectedHeading := "<h2>Application Configuration</h2>"
	assert.Contains(t, body, expectedHeading, "handler response body does not contain expected heading")

	expectedPre := "<pre>" + expectedYAMLString + "</pre>"
	assert.Contains(t, body, expectedPre, "handler response body does not contain expected preformatted YAML")

	assert.Contains(t, body, expectedYAMLString, "handler response body does not contain the exact YAML string")
}
