# Home Network Internet Monitor

A comprehensive Golang-based system for monitoring home network internet health and performance, designed to run on a Raspberry Pi.

## Features
- 🚀 Periodic Internet Speed Testing
- 📊 Network Performance Tracking
- 📈 Historical Data Storage
- 🌐 Grafana Dashboard Integration

## Requirements

### Hardware
- Raspberry Pi (3B+ or 4 recommended)
- Stable Internet Connection
- Optional: External Storage for Long-Term Data

### Software
- Golang 1.24+
- Prometheus Client SDK
- Prometheus Push Gateway
- Grafana Cloud

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

### Using GitHub Actions Build (Recommended)

1. Download the latest release tarball from the [Releases](https://github.com/jeffbean/yanm/releases) page
2. Extract the tarball:
```bash
tar -xzf yanm-binary.tar.gz
```
3. Copy the files to your Raspberry Pi:
```bash
scp yanm config.yml pi@raspberrypi:/opt/yanm/
scp yanm.service pi@raspberrypi:/etc/systemd/system/
```
4. Enable and start the service:
```bash
sudo systemctl daemon-reload
sudo systemctl enable yanm
sudo systemctl start yanm
```
5. Check service status:
```bash
sudo systemctl status yanm
```
6. View logs:
```bash
journalctl -u yanm -f
```

### Using Docker (Development)

```bash
# Build the Docker image
docker build -t yanm-monitor .

# Run the container with required environment variables
docker run -d \
    -p 8090:8090 \
    yanm-monitor

# View container logs
docker logs yanm-monitor
```

### Building from Source (Development)

```bash
# Build the application
go build -o network-monitor cmd/main.go

# Run the monitor
./network-monitor
```

### Grafana Cloud Setup
1. Create a Grafana Cloud account
2. Get your Prometheus Push Gateway URL
3. Configure your dashboard to ingest metrics

## Monitoring Metrics
- Download Speed (Mbps)
- Upload Speed (Mbps)
- Ping Latency (ms)
- Server Information

## Contributing
Contributions are welcome! Please read our contributing guidelines before submitting a pull request.

## License

MIT License

## Support

For support, please open an issue on GitHub or contact the maintainers.
