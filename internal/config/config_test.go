package config

import (
	"os"
	"path/filepath"
	"testing"
	"yanm/internal/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTempConfigFile(t *testing.T, content string) string {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yml")

	err := os.WriteFile(configPath, []byte(content), 0644)
	require.NoError(t, err)

	return configPath
}

func TestLoad(t *testing.T) {
	testCases := []struct {
		name           string
		configContent  string
		expectedConfig Configuration
		errorMessage   string
	}{
		{
			name: "Default Configuration",
			expectedConfig: Configuration{
				Metrics: struct {
					Engine     string `yaml:"engine"`
					Prometheus struct {
						PushGatewayURL string `yaml:"push_gateway_url"`
						JobName        string `yaml:"job_name"`
						InstanceName   string `yaml:"instance_name"`
					} `yaml:"prometheus"`
					InfluxDB struct {
						URL    string `yaml:"url"`
						Token  string `yaml:"token"`
						Org    string `yaml:"org"`
						Bucket string `yaml:"bucket"`
					} `yaml:"influxdb"`
				}{
					Engine: "no-op",
					Prometheus: struct {
						PushGatewayURL string `yaml:"push_gateway_url"`
						JobName        string `yaml:"job_name"`
						InstanceName   string `yaml:"instance_name"`
					}{},
					InfluxDB: struct {
						URL    string `yaml:"url"`
						Token  string `yaml:"token"`
						Org    string `yaml:"org"`
						Bucket string `yaml:"bucket"`
					}{},
				},
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
						IntervalSeconds:  15,  // Default value
						ThresholdSeconds: 5.0, // Default value
					},
					SpeedTest: struct {
						IntervalMinutes int `yaml:"interval_minutes"`
						Servers         struct {
							MaxPingTimeout   string `yaml:"max_ping_timeout"`
							MaxServersToTest int    `yaml:"max_servers_to_test"`
						} `yaml:"servers"`
					}{
						IntervalMinutes: 1,
					},
				},
				Logging: logger.Config{
					Level:      "info",
					Format:     "json",
					OutputFile: "/var/log/yanm.log",
				},
			},
		},
		{
			name: "Prometheus Configuration",
			configContent: `
metrics:
  engine: prometheus
  prometheus:
    push_gateway_url: http://localhost:9091
    job_name: test_job
    instance_name: test_instance
`,
			expectedConfig: Configuration{
				Metrics: struct {
					Engine     string `yaml:"engine"`
					Prometheus struct {
						PushGatewayURL string `yaml:"push_gateway_url"`
						JobName        string `yaml:"job_name"`
						InstanceName   string `yaml:"instance_name"`
					} `yaml:"prometheus"`
					InfluxDB struct {
						URL    string `yaml:"url"`
						Token  string `yaml:"token"`
						Org    string `yaml:"org"`
						Bucket string `yaml:"bucket"`
					} `yaml:"influxdb"`
				}{
					Engine: "prometheus",
					Prometheus: struct {
						PushGatewayURL string `yaml:"push_gateway_url"`
						JobName        string `yaml:"job_name"`
						InstanceName   string `yaml:"instance_name"`
					}{
						PushGatewayURL: "http://localhost:9091",
						JobName:        "test_job",
						InstanceName:   "test_instance",
					},
				},
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
						IntervalSeconds:  15,  // Default value
						ThresholdSeconds: 5.0, // Default value
					},
					SpeedTest: struct {
						IntervalMinutes int `yaml:"interval_minutes"`
						Servers         struct {
							MaxPingTimeout   string `yaml:"max_ping_timeout"`
							MaxServersToTest int    `yaml:"max_servers_to_test"`
						} `yaml:"servers"`
					}{
						IntervalMinutes: 1,
					},
				},
				Logging: logger.Config{
					Level:      "info",
					Format:     "json",
					OutputFile: "/var/log/yanm.log",
				},
			},
		},
		{
			name: "Invalid Metrics Engine",
			configContent: `
metrics:
  engine: invalid_engine
`,
			errorMessage: "metrics.engine must be 'prometheus' or 'no-op'",
		},
		{
			name: "Prometheus Without URL",
			configContent: `
metrics:
  engine: prometheus
`,
			errorMessage: "metrics.prometheus.push_gateway_url is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary config file
			configPath := createTempConfigFile(t, tc.configContent)

			// Set the CONFIG_PATH environment variable
			originalConfigPath := os.Getenv("CONFIG_PATH")
			t.Setenv("CONFIG_PATH", configPath)
			defer os.Setenv("CONFIG_PATH", originalConfigPath)

			// Load configuration
			cfg, err := Load()

			if tc.errorMessage != "" {
				// Check for error
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMessage)
				return
			}

			// Check for no error
			require.NoError(t, err)
			require.NotNil(t, cfg)

			assert.Equal(t, tc.expectedConfig.Metrics.Engine, cfg.Metrics.Engine, "Metrics Engine")
			assert.Equal(t, tc.expectedConfig.Metrics.Prometheus.PushGatewayURL, cfg.Metrics.Prometheus.PushGatewayURL, "Prometheus Push Gateway URL")
			assert.Equal(t, tc.expectedConfig.Metrics.Prometheus.JobName, cfg.Metrics.Prometheus.JobName, "Prometheus Job Name")
			assert.Equal(t, tc.expectedConfig.Metrics.Prometheus.InstanceName, cfg.Metrics.Prometheus.InstanceName, "Prometheus Instance Name")
			assert.Equal(t, tc.expectedConfig.Network.PingTest.IntervalSeconds, cfg.Network.PingTest.IntervalSeconds, "Ping Test Interval")
			assert.Equal(t, tc.expectedConfig.Network.PingTest.ThresholdSeconds, cfg.Network.PingTest.ThresholdSeconds, "Ping Test Threshold")
			assert.Equal(t, tc.expectedConfig.Network.SpeedTest.IntervalMinutes, cfg.Network.SpeedTest.IntervalMinutes, "Speed Test Interval")
			assert.Equal(t, tc.expectedConfig.Logging.Level, cfg.Logging.Level, "Logging Level")
			assert.Equal(t, tc.expectedConfig.Logging.Format, cfg.Logging.Format, "Logging Format")
			assert.Equal(t, tc.expectedConfig.Logging.OutputFile, cfg.Logging.OutputFile, "Logging Output File")
		})
	}
}
