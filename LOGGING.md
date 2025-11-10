# Logging Framework

This document describes the logging framework used in the New Relic Exporter.

## Overview

The exporter uses a modern, structured logging framework based on [logrus](https://github.com/sirupsen/logrus), providing:

- **Structured logging** with fields
- **Multiple log levels** (debug, info, warn, error, fatal)
- **JSON and text output formats**
- **Configurable via command-line flags**
- **Thread-safe logging**
- **Performance optimized**

## Migration from prometheus/log

The project has been migrated from the deprecated `github.com/prometheus/log` to a custom logging package built on logrus. This provides:

- Better structured logging capabilities
- More flexible configuration
- Active maintenance and community support
- Better performance
- Standard logging interface

## Usage

### Command-Line Flags

Configure logging when starting the exporter:

```bash
# Default (info level, text format)
./newrelic_exporter

# Debug level logging
./newrelic_exporter --log.level=debug

# JSON format (for log aggregation systems)
./newrelic_exporter --log.json

# Combined
./newrelic_exporter --log.level=debug --log.json
```

### Log Levels

Available log levels (in order of verbosity):

| Level | Description | Use Case |
|-------|-------------|----------|
| `debug` | Very detailed information | Development, troubleshooting |
| `info` | General informational messages | Normal operation (default) |
| `warn` | Warning messages | Potential issues |
| `error` | Error messages | Errors that don't stop execution |
| `fatal` | Fatal errors | Critical errors, exits program |

### Log Formats

#### Text Format (Default)

Human-readable format for console output:

```
2025-01-10 15:04:05 INFO Configuration loaded successfully from newrelic_exporter.yml
2025-01-10 15:04:05 INFO Listening on :9126
2025-01-10 15:04:10 INFO Starting new scrape at Jan 10 15:04:10 for period from Jan 10 15:03:00 to Jan 10 15:04:00
```

#### JSON Format

Machine-readable format for log aggregation:

```json
{"level":"info","msg":"Configuration loaded successfully from newrelic_exporter.yml","time":"2025-01-10T15:04:05Z"}
{"level":"info","msg":"Listening on :9126","time":"2025-01-10T15:04:05Z"}
{"level":"info","msg":"Starting new scrape","time":"2025-01-10T15:04:10Z"}
```

## Code Examples

### Basic Logging

```go
import "github.com/mrf/newrelic_exporter/logger"

// Simple messages
logger.Debug("Detailed debug information")
logger.Info("Informational message")
logger.Warn("Warning message")
logger.Error("Error occurred")
logger.Fatal("Fatal error, will exit")

// Formatted messages
logger.Debugf("Processing %d items", count)
logger.Infof("Server started on port %d", port)
logger.Errorf("Failed to connect: %v", err)
```

### Structured Logging

Add context fields to log messages:

```go
import (
	"github.com/mrf/newrelic_exporter/logger"
	"github.com/sirupsen/logrus"
)

// Single field
logger.WithField("app_id", 12345).Info("Processing application")

// Multiple fields
logger.WithFields(logrus.Fields{
	"app_id":   12345,
	"app_name": "MyApp",
	"duration": "2.5s",
}).Info("Scrape completed")
```

Output in JSON format:
```json
{"app_id":12345,"level":"info","msg":"Processing application","time":"2025-01-10T15:04:05Z"}
{"app_id":12345,"app_name":"MyApp","duration":"2.5s","level":"info","msg":"Scrape completed","time":"2025-01-10T15:04:05Z"}
```

### Programmatic Configuration

```go
import "github.com/mrf/newrelic_exporter/logger"

// Configure logger
logger.Setup(logger.Config{
	Level:      logger.DebugLevel,
	JSONFormat: true,
})

// Get the underlying logrus instance
log := logger.GetLogger()
```

## Best Practices

### 1. Choose Appropriate Log Levels

```go
// Debug - detailed information for diagnosing problems
logger.Debug("Cache hit for application list")

// Info - general informational messages
logger.Info("Scrape completed successfully")

// Warn - something unexpected but not critical
logger.Warn("API rate limit approaching")

// Error - error occurred but processing continues
logger.Error("Failed to fetch metrics for one application")

// Fatal - critical error, cannot continue
logger.Fatal("Cannot connect to New Relic API")
```

### 2. Use Structured Logging for Context

```go
// Bad - hard to parse
logger.Infof("Scraped %d metrics for app %d in %v", count, appID, duration)

// Good - structured and parseable
logger.WithFields(logrus.Fields{
	"metric_count": count,
	"app_id":       appID,
	"duration":     duration,
}).Info("Metrics scraped")
```

### 3. Don't Log Sensitive Information

```go
// Bad - logs API key
logger.Infof("Using API key: %s", apiKey)

// Good - masks sensitive data
logger.Info("API key configured")
```

### 4. Use Appropriate Verbosity

```go
// Debug level - very detailed
logger.Debugf("Processing metric: %s with value: %f", name, value)

// Info level - key events
logger.Info("Scrape started")

// Too verbose for info level
// logger.Info("Processing metric 1 of 1000")
```

## Integration with Log Aggregation

### Elasticsearch/Logstash

Run with JSON format:
```bash
./newrelic_exporter --log.json | logstash -f logstash.conf
```

### Fluentd

```xml
<source>
  @type tail
  path /var/log/newrelic_exporter.log
  pos_file /var/log/newrelic_exporter.log.pos
  tag newrelic.exporter
  format json
</source>
```

### Splunk

Configure forwarding of JSON logs:
```bash
./newrelic_exporter --log.json >> /var/log/newrelic_exporter.log
```

### CloudWatch Logs

Use the CloudWatch agent to collect JSON-formatted logs.

### Loki (Grafana)

```yaml
scrape_configs:
  - job_name: newrelic-exporter
    static_configs:
      - targets:
          - localhost
        labels:
          job: newrelic-exporter
          __path__: /var/log/newrelic_exporter.log
```

## Performance Considerations

### Log Level Impact

- **debug**: High overhead, logs everything
- **info**: Normal overhead, recommended for production
- **warn/error**: Low overhead, only logs issues

### JSON vs Text Format

- **JSON**: Slightly higher overhead, better for parsing
- **Text**: Lower overhead, better for humans

### Best Practices

1. **Use info level in production** unless troubleshooting
2. **Enable debug level temporarily** for issue diagnosis
3. **Use JSON format** when sending to log aggregation systems
4. **Use text format** for local development and debugging

## Troubleshooting

### Not Seeing Debug Logs

Ensure debug level is enabled:
```bash
./newrelic_exporter --log.level=debug
```

### Logs Not Formatting Correctly

Check that JSON flag is set if expecting JSON:
```bash
./newrelic_exporter --log.json
```

### Too Much Output

Reduce log level:
```bash
./newrelic_exporter --log.level=warn
```

### No Logs Appearing

Check that the logger is initialized:
```go
logger.Setup(logger.Config{Level: logger.InfoLevel})
```

## Testing

The logging package includes comprehensive tests:

```bash
# Run logger tests
go test ./logger/

# Run with verbose output
go test -v ./logger/

# Run specific test
go test -v -run TestJSONFormat ./logger/
```

## Migration Guide

### For Contributors

If you're adding new logging to the code:

**Old way (deprecated):**
```go
import "github.com/prometheus/log"

log.Info("message")
log.Debugf("formatted %s", value)
```

**New way:**
```go
import "github.com/mrf/newrelic_exporter/logger"

logger.Info("message")
logger.Debugf("formatted %s", value)
```

### For Users

No changes required! The logging is configured via command-line flags.

## Additional Resources

- [Logrus Documentation](https://github.com/sirupsen/logrus)
- [Structured Logging Best Practices](https://www.honeycomb.io/blog/structured-logging-and-your-team/)
- [The Twelve-Factor App: Logs](https://12factor.net/logs)

## Future Enhancements

Potential future improvements:

- [ ] Log rotation support
- [ ] Remote syslog support
- [ ] Configurable timestamp formats
- [ ] Custom log hooks
- [ ] Context-aware logging
- [ ] Sampling for high-volume logs

## Questions?

For questions or issues with logging:
- Check this documentation
- Review the logger package code
- Ask in GitHub Issues or Discussions
