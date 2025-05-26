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


## Deployment Options

### Using Ansible
We provide an Ansible playbook for automated deployment:

```bash
# Install Ansible
sudo apt-get install ansible

# Run the playbook
ansible-playbook --connection=local 127.0.0.1 deploy.yml
```

The playbook will:
- Clone the repository
- Build the application
- Set up systemd service
- Configure logging
- Start the service

### Manual Installation
1. Clone the repository
2. Build the application
3. Set up systemd service (example provided in README)
4. Configure logging
5. Start the service

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
