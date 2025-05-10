package config

import (
	"os"
	"path/filepath"
	"testing"

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
			name: "Default No-Op Configuration",
			configContent: `
metrics:
  engine: no-op
network:
  speedtest:
    interval_minutes: 10
`,
			expectedConfig: Configuration{
				Metrics: struct {
					Engine     string `yaml:"engine"`
					Prometheus struct {
						PushGatewayURL string `yaml:"push_gateway_url"`
						JobName        string `yaml:"job_name"`
						InstanceName   string `yaml:"instance_name"`
					} `yaml:"prometheus"`
				}{
					Engine: "no-op",
				},
				Network: struct {
					Speedtest struct {
						IntervalMinutes int `yaml:"interval_minutes"`
						Servers         struct {
							MaxPingTimeout   string `yaml:"max_ping_timeout"`
							MaxServersToTest int    `yaml:"max_servers_to_test"`
						} `yaml:"servers"`
					} `yaml:"speedtest"`
				}{
					Speedtest: struct {
						IntervalMinutes int `yaml:"interval_minutes"`
						Servers         struct {
							MaxPingTimeout   string `yaml:"max_ping_timeout"`
							MaxServersToTest int    `yaml:"max_servers_to_test"`
						} `yaml:"servers"`
					}{
						IntervalMinutes: 10,
					},
				},
				Logging: struct {
					Level      string `yaml:"level"`
					OutputFile string `yaml:"output_file"`
				}{
					Level:      "info",
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
					Speedtest struct {
						IntervalMinutes int `yaml:"interval_minutes"`
						Servers         struct {
							MaxPingTimeout   string `yaml:"max_ping_timeout"`
							MaxServersToTest int    `yaml:"max_servers_to_test"`
						} `yaml:"servers"`
					} `yaml:"speedtest"`
				}{
					Speedtest: struct {
						IntervalMinutes int `yaml:"interval_minutes"`
						Servers         struct {
							MaxPingTimeout   string `yaml:"max_ping_timeout"`
							MaxServersToTest int    `yaml:"max_servers_to_test"`
						} `yaml:"servers"`
					}{
						IntervalMinutes: 5,
					},
				},
				Logging: struct {
					Level      string `yaml:"level"`
					OutputFile string `yaml:"output_file"`
				}{
					Level:      "info",
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

			// Compare configuration
			assert.Equal(t, tc.expectedConfig.Metrics.Engine, cfg.Metrics.Engine, "Metrics Engine")

			// Prometheus configuration
			if tc.expectedConfig.Metrics.Engine == "prometheus" {
				assert.Equal(t, tc.expectedConfig.Metrics.Prometheus.PushGatewayURL,
					cfg.Metrics.Prometheus.PushGatewayURL, "Prometheus Push Gateway URL")
				assert.Equal(t, tc.expectedConfig.Metrics.Prometheus.JobName,
					cfg.Metrics.Prometheus.JobName, "Prometheus Job Name")
				assert.Equal(t, tc.expectedConfig.Metrics.Prometheus.InstanceName,
					cfg.Metrics.Prometheus.InstanceName, "Prometheus Instance Name")
			}

			// Network configuration
			assert.Equal(t, tc.expectedConfig.Network.Speedtest.IntervalMinutes,
				cfg.Network.Speedtest.IntervalMinutes, "Network Speedtest Interval")

			// Logging configuration
			assert.Equal(t, tc.expectedConfig.Logging.Level,
				cfg.Logging.Level, "Logging Level")
			assert.Equal(t, tc.expectedConfig.Logging.OutputFile,
				cfg.Logging.OutputFile, "Logging Output File")

		})
	}
}
