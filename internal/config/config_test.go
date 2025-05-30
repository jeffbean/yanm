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

func TestLoadFile_ErrorConditions(t *testing.T) {
	testCases := []struct {
		name          string
		pathArgument  string // Path to pass to LoadFile
		configContent string // Content for the temporary config file, if one needs to be created (e.g., for invalid YAML)
		errorMessage  string
	}{
		{
			name:         "Empty path argument, default config file does not exist in CWD",
			pathArgument: "no-op",
			errorMessage: "failed to read config file",
		},
		{
			name:         "Path to non-existent file",
			pathArgument: "./non_existent_config.yml",
			errorMessage: "failed to read config file",
		},
		{
			name:          "Invalid YAML content",
			pathArgument:  "USE_TEMP_FILE",
			configContent: "metrics: { engine: prometheus, prometheus: { push_gateway_url: http://localhost:9091 }", // Malformed YAML
			errorMessage:  "failed to parse config data: yaml: line 1: did not find expected ',' or '}'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			pathToLoad := tc.pathArgument

			// If we need to use a temp file with specific content
			if tc.pathArgument == "USE_TEMP_FILE" {
				tempDir := t.TempDir()
				pathToLoad = filepath.Join(tempDir, "config.yml")
				err := os.WriteFile(pathToLoad, []byte(tc.configContent), 0644)
				require.NoError(t, err)
			}

			_, err := LoadFile(pathToLoad)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.errorMessage)
		})
	}
}

func TestLoadFile_ContentAndValidation(t *testing.T) {
	testCases := []struct {
		name          string
		pathArgument  string
		configContent string
		wantConfig    *Configuration
		errorMessage  string // For validation errors
	}{
		{
			name:          "Default Configuration (empty actual file)",
			pathArgument:  "USE_TEMP_FILE",
			configContent: "", // Empty content, leads to zero-value Configuration struct
			wantConfig: &Configuration{
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
						IntervalSeconds:  2,
						ThresholdSeconds: 5.0,
					},
					SpeedTest: struct {
						IntervalMinutes int `yaml:"interval_minutes"`
						Servers         struct {
							MaxPingTimeout   string `yaml:"max_ping_timeout"`
							MaxServersToTest int    `yaml:"max_servers_to_test"`
						} `yaml:"servers"`
					}{
						IntervalMinutes: 720,
					},
				},
				Logging: logger.Config{
					Level:  "info",
					Format: "json",
				},
				DebugServer: struct {
					Disabled      bool   `yaml:"disabled"`
					ListenAddress string `yaml:"listen_address"`
				}{
					Disabled:      false,
					ListenAddress: "127.0.0.1:8090",
				},
			},
		},
		{
			name:         "Prometheus Configuration (valid)",
			pathArgument: "USE_TEMP_FILE",
			configContent: `
metrics:
  engine: prometheus
`,
			wantConfig: &Configuration{
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
						IntervalSeconds:  2,
						ThresholdSeconds: 5.0,
					},
					SpeedTest: struct {
						IntervalMinutes int `yaml:"interval_minutes"`
						Servers         struct {
							MaxPingTimeout   string `yaml:"max_ping_timeout"`
							MaxServersToTest int    `yaml:"max_servers_to_test"`
						} `yaml:"servers"`
					}{
						IntervalMinutes: 720,
					},
				},
				Logging: logger.Config{
					Level:  "info",
					Format: "json",
				},
				DebugServer: struct {
					Disabled      bool   `yaml:"disabled"`
					ListenAddress string `yaml:"listen_address"`
				}{
					Disabled:      false,
					ListenAddress: "127.0.0.1:8090",
				},
			},
		},
		{
			name:         "Invalid Metrics Engine (validation)",
			pathArgument: "USE_TEMP_FILE",
			configContent: `metrics:
  engine: invalid_engine
`,
			wantConfig:   nil,
			errorMessage: "metrics.engine must be 'prometheus' or 'no-op'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualPathToLoad := tc.pathArgument
			if tc.pathArgument == "USE_TEMP_FILE" {
				actualPathToLoad = createTempConfigFile(t, tc.configContent)
			}

			cfg, err := LoadFile(actualPathToLoad)

			if tc.errorMessage != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorMessage)
				require.Nil(t, cfg, "cfg should be nil on validation error")
			} else {
				require.NoError(t, err)
				require.NotNil(t, cfg, "cfg should not be nil on successful load")

				// Compare the entire configuration structure
				assert.Equal(t, tc.wantConfig, cfg, "Configuration should match expected value")
			}
		})
	}
}
