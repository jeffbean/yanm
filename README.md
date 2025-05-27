# Home Network Internet Monitor

A comprehensive Golang-based system for monitoring home network internet health and performance, designed to run on a Raspberry Pi.
![image](https://github.com/user-attachments/assets/860b88b7-a31d-413a-af85-91f60786ce57)

## Features
- 🚀 Periodic Internet Speed Testing
- 📊 Network Performance Tracking
- 📈 Historical Data Storage
- 🌐 Grafana Dashboard Integration

## Monitoring Metrics
- Download Speed (Mbps)
- Upload Speed (Mbps)
- Ping Latency (ms)
- Server Information

## Requirements

### Hardware
- Raspberry Pi (3B+ or 4 recommended)
- Stable Internet Connection

### Software
- Golang 1.24+
- Prometheus Collector
- Grafana Cloud
  
## Monitoring Metrics
- Download Speed (Mbps)
- Upload Speed (Mbps)
- Ping Latency (ms)
- Server Information


## Quick Start

```bash
# Build the application
go build -o yanm cmd/main.go

# Run with default configuration
./yanm -config config.yml

# Or specify a custom config file
./yanm -config /path/to/config.yml
```

## Configuration

The application uses a YAML configuration file. By default, it looks for `config.yml` in the current directory, but you can specify a different location using the `-config` flag.

You can view the current configuration by accessing the debug server at `http://localhost:8090/config/` (when debug server is enabled).

## Contributing
Contributions are welcome! Please read our contributing guidelines before submitting a pull request.

## License

MIT License

## Support

For support, please open an issue on GitHub or contact the maintainers.
