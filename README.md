# Home Network Internet Monitor

## Overview
A comprehensive Golang-based system for monitoring home network internet health and performance, designed to run on a Raspberry Pi.

## Features
- 🚀 Periodic Internet Speed Testing
- 📊 Network Performance Tracking
- 📈 Historical Data Storage
- 🔔 Configurable Alerts
- 🌐 Grafana Dashboard Integration

## Hardware Requirements
- Raspberry Pi (3B+ or 4 recommended)
- Stable Internet Connection
- Optional: External Storage for Long-Term Data

## Software Dependencies
- Golang 1.16+
- Prometheus Client SDK
- Prometheus Push Gateway
- Grafana Cloud
- Optional: Slack/Email for Alerts

## Project Structure
```
yanm/
├── cmd/           # Application entry points
├── internal/      # Private application packages
│   ├── network/   # Network testing logic
│   └── storage/   # Data storage implementations
├── pkg/           # Shared packages
├── config.yml     # Configuration management
└── README.md      # Project documentation
```

## Quick Start
1. Clone the repository
2. Configure `config.yml`
3. Set environment variables
4. Build and run

```bash
# Set Prometheus Push Gateway configuration
export PROMETHEUS_PUSH_GATEWAY_URL=https://your-grafana-cloud-push-gateway.com/metrics
export PROMETHEUS_JOB_NAME=home_network_monitor

# Build the application
go build -o network-monitor cmd/main.go

# Run the monitor
./network-monitor
```

## Configuration
Configure your monitoring by setting these environment variables:
- `PROMETHEUS_PUSH_GATEWAY_URL`: URL of your Grafana Cloud Prometheus Push Gateway
- `PROMETHEUS_JOB_NAME`: Optional job name for metrics (default: home_network_monitor)

### Grafana Cloud Setup
1. Create a Grafana Cloud account
2. Get your Prometheus Push Gateway URL
3. Configure your dashboard to ingest metrics

## Monitoring Metrics
- Download Speed (Mbps)
- Upload Speed (Mbps)
- Ping Latency (ms)
- Server Information

## Grafana Dashboard
A pre-configured Grafana dashboard is available to visualize your network performance over time.

## Contributing
Contributions are welcome! Please read our contributing guidelines before submitting a pull request.

## License
[Specify your license]
