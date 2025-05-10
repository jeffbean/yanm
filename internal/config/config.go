package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Configuration represents the application's configuration structure
type Configuration struct {
	Network struct {
		Speedtest struct {
			IntervalMinutes int `yaml:"interval_minutes"`
			Servers         struct {
				MaxPingTimeout   string `yaml:"max_ping_timeout"`
				MaxServersToTest int    `yaml:"max_servers_to_test"`
			} `yaml:"servers"`
		} `yaml:"speedtest"`
	} `yaml:"network"`

	Metrics struct {
		Engine string `yaml:"engine"`

		Prometheus struct {
			PushGatewayURL string `yaml:"push_gateway_url"`
			JobName        string `yaml:"job_name"`
			InstanceName   string `yaml:"instance_name"`
		} `yaml:"prometheus"`
	} `yaml:"metrics"`

	Logging struct {
		Level      string `yaml:"level"`
		OutputFile string `yaml:"output_file"`
	} `yaml:"logging"`
}

// Load reads the configuration from the config file
func Load() (*Configuration, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yml"
	}

	absConfigPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config path: %v", err)
	}

	configData, err := os.ReadFile(absConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var configuration Configuration
	err = yaml.Unmarshal(configData, &configuration)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	if err := configuration.validate(); err != nil {
		return nil, err
	}

	return &configuration, nil
}

func (c *Configuration) validate() error {
	// Validate metrics configuration
	if err := c.validateMetrics(); err != nil {
		return err
	}

	// Set default logging configuration
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}

	if c.Logging.OutputFile == "" {
		c.Logging.OutputFile = "/var/log/yanm.log"
	}

	// Set default network speedtest configuration
	if c.Network.Speedtest.IntervalMinutes <= 0 {
		c.Network.Speedtest.IntervalMinutes = 5
	}

	return nil
}

func (c *Configuration) validateMetrics() error {
	// Set default metrics engine
	if c.Metrics.Engine == "" {
		c.Metrics.Engine = "no-op"
	}

	// Validate metrics engine
	if c.Metrics.Engine != "prometheus" && c.Metrics.Engine != "no-op" {
		return fmt.Errorf("metrics.engine must be 'prometheus' or 'no-op'")
	}

	// Validate and set defaults for Prometheus configuration
	if c.Metrics.Engine == "prometheus" {
		if c.Metrics.Prometheus.PushGatewayURL == "" {
			return fmt.Errorf("metrics.prometheus.push_gateway_url is required")
		}

		if c.Metrics.Prometheus.JobName == "" {
			c.Metrics.Prometheus.JobName = "home_network_monitor"
		}
	}

	return nil
}
