# New Relic Exporter

[![GitHub Container Registry](https://img.shields.io/badge/ghcr.io-mrf%2Fnewrelic--exporter-blue)](https://github.com/mrf/newrelic_exporter/pkgs/container/newrelic-exporter)
[![GitHub issues](https://img.shields.io/github/issues/mrf/newrelic_exporter)](https://github.com/mrf/newrelic_exporter/issues)
[![GitHub pull requests](https://img.shields.io/github/issues-pr/mrf/newrelic_exporter)](https://github.com/mrf/newrelic_exporter/pulls)
[![License](https://img.shields.io/github/license/mrf/newrelic_exporter)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/mrf/newrelic_exporter)](https://goreportcard.com/report/github.com/mrf/newrelic_exporter)
[![Coverage](https://raw.githubusercontent.com/mrf/newrelic_exporter/badges/coverage-badge.svg)](https://github.com/mrf/newrelic_exporter/actions)

Prometheus exporter for New Relic data. This exporter allows you to collect metrics from New Relic's API and expose them in Prometheus format.

## Project Status

**✅ Actively Maintained** - This project is actively maintained and accepting contributions.

We welcome:
- 🐛 Bug reports and fixes
- ✨ Feature requests and implementations
- 📖 Documentation improvements
- 🧪 Test coverage improvements

## Features

- ✅ Export New Relic APM metrics to Prometheus
- ✅ Support for multiple applications
- ✅ Configurable metric filtering to reduce API usage
- ✅ Built-in caching to optimize API calls
- ✅ Docker support with multi-architecture images (amd64, arm64)
- ✅ Kubernetes deployment via Helm chart
- ✅ Comprehensive configuration with validation
- ✅ Automated CI/CD with GitHub Actions and CircleCI

## Requirements

- A New Relic account with API access
- New Relic REST API key
- Prometheus server (for collecting metrics)

## Quick Start

### Using Docker (Recommended)

```bash
# Pull the latest image
docker pull ghcr.io/mrf/newrelic-exporter:latest

# Create a configuration file
cat > newrelic_exporter.yml <<EOF
api.key: YOUR_API_KEY_HERE
api.service: applications
api.timeout: 15s
api.include-metric-filters:
  - "HttpDispatcher"
  - "Database"
  - "Apdex"
EOF

# Run the exporter
docker run -d \
  -p 9126:9126 \
  -v $(pwd)/newrelic_exporter.yml:/app/newrelic_exporter.yml \
  ghcr.io/mrf/newrelic-exporter:latest
```

### Using Kubernetes/Helm

```bash
# Add the Helm chart
helm install newrelic-exporter ./helm/newrelic-exporter \
  --set newrelic.apiKey=YOUR_API_KEY \
  --set newrelic.service=applications \
  --set newrelic.includeMetricFilters[0]="HttpDispatcher"
```

See the [Helm chart README](helm/newrelic-exporter/README.md) for detailed configuration options.

## Building and Running

### Running in a container

```bash
cp newrelic_exporter.yml.example newrelic_exporter.yml
# Edit newrelic_exporter.yml with your API key and settings
docker run -v $(pwd)/newrelic_exporter.yml:/app/newrelic_exporter.yml ghcr.io/mrf/newrelic-exporter
```

### From source

```bash
git clone https://github.com/mrf/newrelic_exporter.git
cd newrelic_exporter
make
cp newrelic_exporter.yml.example newrelic_exporter.yml
# Edit newrelic_exporter.yml with your API key and settings
./newrelic_exporter
```

## Configuration

### Command Line Flags

| Name | Description | Default |
|------|-------------|---------|
| `--config` | Config file path | `newrelic_exporter.yml` |

### Configuration File

The exporter uses a YAML configuration file. See [`newrelic_exporter.yml.example`](newrelic_exporter.yml.example) for a comprehensive example with detailed documentation.

#### Required Settings

| Setting | Description |
|---------|-------------|
| `api.key` | Your New Relic API key |
| `api.service` | Service type (e.g., `applications`, `mobile`) |
| `api.timeout` | API request timeout (e.g., `15s`) - must include time unit |
| `api.include-metric-filters` | List of metric filters (at least one required) |

#### Optional Settings

| Setting | Description | Default |
|---------|-------------|---------|
| `api.server` | API server URL | `https://api.newrelic.com` |
| `api.period` | Data request period in seconds | `60` |
| `api.apps-list-cache-time` | Application list cache duration | `0` (no cache) |
| `api.metric-names-cache-time` | Metric names cache duration | `0` (no cache) |
| `api.include-apps` | Filter specific applications | `[]` (all apps) |
| `api.include-values` | Filter specific metric values | `[]` (all values) |
| `web.listen-address` | HTTP server listen address | `:9126` |
| `web.telemetry-path` | Metrics endpoint path | `/metrics` |
| `debug.proxy-address` | Proxy for debugging | `""` |

### Example Configurations

#### Basic APM Monitoring

```yaml
api.key: YOUR_API_KEY
api.service: applications
api.timeout: 15s
api.include-metric-filters:
  - "HttpDispatcher"
  - "Apdex"
```

#### Multiple Applications with Caching

```yaml
api.key: YOUR_API_KEY
api.service: applications
api.timeout: 30s
api.apps-list-cache-time: 1h
api.metric-names-cache-time: 1h
api.include-apps:
  - name: "Production Web"
  - name: "Production API"
api.include-metric-filters:
  - "HttpDispatcher"
  - "Database"
  - "External"
```

See [`newrelic_exporter.yml.example`](newrelic_exporter.yml.example) for more examples.

## Prometheus Integration

Add the following to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'newrelic'
    static_configs:
      - targets: ['localhost:9126']
    scrape_interval: 60s  # Match api.period in exporter config
```

## Development

### Running Tests

```bash
go test -v ./...
```

### Building

```bash
make
# or
go build -o newrelic_exporter .
```

### Docker Build

```bash
docker build -t ghcr.io/mrf/newrelic-exporter:latest .
```

## CI/CD

This project includes automated CI/CD pipelines:

- **GitHub Actions**: Builds, tests, and publishes Docker images on every push and release
- **Helm Chart Testing**: Automated Helm chart validation and testing with Kind
- **CircleCI**: Alternative CI/CD pipeline with advanced features

See [`.github/workflows/`](.github/workflows/) and [`.circleci/`](.circleci/) for configuration details.

### Helm Chart Testing

The Helm chart is automatically tested using Kind (Kubernetes in Docker) to ensure it works correctly. Tests include:

- Chart linting and validation
- Installation in a real Kubernetes cluster
- Resource validation (deployments, services, pods, etc.)
- Connectivity and metrics endpoint testing
- Upgrade scenario testing

For local testing and detailed information, see the [Helm Testing Guide](HELM_TESTING.md).

**Quick local test:**
```bash
./scripts/test-helm-chart.sh
```

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### How to Contribute

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests:
   - Go tests: `go test ./...`
   - Helm chart tests: `./scripts/test-helm-chart.sh`
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## Troubleshooting

### Common Issues

**Error: "Config validation failed: api.timeout is not set"**
- Solution: Ensure timeout includes time units (e.g., `"15s"` not `"15"`)

**Error: "Config validation failed: api.key is required"**
- Solution: Set your New Relic API key in the config file

**No metrics appearing**
- Check that your API key is valid
- Verify application names/IDs are correct
- Ensure metric filters match available metrics in New Relic

**Slow scrapes or timeouts**
- Reduce the number of metric filters
- Enable and increase cache times
- Filter to specific applications only
- Increase the timeout value

See the [configuration example](newrelic_exporter.yml.example) for a comprehensive troubleshooting guide.

## Support

- **Issues**: [GitHub Issues](https://github.com/mrf/newrelic_exporter/issues)
- **Pull Requests**: [GitHub Pull Requests](https://github.com/mrf/newrelic_exporter/pulls)
- **Discussions**: [GitHub Discussions](https://github.com/mrf/newrelic_exporter/discussions)

## License

This project is licensed under the terms specified in the [LICENSE](LICENSE) file.

## Acknowledgments

- Built with [Prometheus Go client library](https://github.com/prometheus/client_golang)
- New Relic API documentation: https://docs.newrelic.com/docs/apis/rest-api-v2/

## Related Projects

- [Prometheus](https://prometheus.io/) - Monitoring system and time series database
- [New Relic](https://newrelic.com/) - Application performance monitoring platform
- [Grafana](https://grafana.com/) - Visualization platform (works great with this exporter)
