package config

import (
	"fmt"
	"github.com/mrf/newrelic_exporter/logger"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"time"
)

type Config struct {
	// NewRelic related settings
	NRApiKey               string        `yaml:"api.key"`
	NRApiServer            string        `yaml:"api.server"`
	NRPeriod               int           `yaml:"api.period"`
	NRTimeout              time.Duration `yaml:"api.timeout"`
	NRAppListCacheTime     time.Duration `yaml:"api.apps-list-cache-time"`
	NRMetricNamesCacheTime time.Duration `yaml:"api.metric-names-cache-time"`
	NRService              string        `yaml:"api.service"`
	NRApps                 []Application `yaml:"api.include-apps"`
	NRMetricFilters        []string      `yaml:"api.include-metric-filters"`
	NRValueFilters         []string      `yaml:"api.include-values"`

	// Prometheus Exporter related settings
	MetricPath    string `yaml:"web.telemetry-path"`
	ListenAddress string `yaml:"web.listen-address"`

	// Debugging settings
	DebugProxyAddress string `yaml:"debug.proxy-address"`
}

type Application struct {
	Id   int    `yaml:"id"`
	Name string `yaml:"name"`
}

// Default configuration values
const (
	DefaultAPIServer            = "https://api.newrelic.com"
	DefaultPeriod               = 60
	DefaultTimeout              = 15 * time.Second
	DefaultAppListCacheTime     = 1 * time.Hour
	DefaultMetricNamesCacheTime = 1 * time.Hour
	DefaultListenAddress        = ":9126"
	DefaultMetricPath           = "/metrics"
)

// GetConfig loads configuration from a file and applies defaults
func GetConfig(path string) (Config, error) {
	// Start with defaults
	config := NewDefaultConfig()

	configSource, err := ioutil.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal will override defaults with user-provided values
	err = yaml.Unmarshal(configSource, &config)
	if err != nil {
		return config, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	// Apply defaults for any values still unset
	config.ApplyDefaults()

	// Validate the final configuration
	err = config.Validate()
	if err != nil {
		return config, fmt.Errorf("config validation failed: %w", err)
	}

	logger.Debugf("Config loaded successfully with defaults applied: %v", config)

	return config, nil
}

// NewDefaultConfig creates a new Config with all default values set
func NewDefaultConfig() Config {
	return Config{
		NRApiServer:            DefaultAPIServer,
		NRPeriod:               DefaultPeriod,
		NRTimeout:              DefaultTimeout,
		NRAppListCacheTime:     DefaultAppListCacheTime,
		NRMetricNamesCacheTime: DefaultMetricNamesCacheTime,
		ListenAddress:          DefaultListenAddress,
		MetricPath:             DefaultMetricPath,
		NRApps:                 make([]Application, 0),
		NRMetricFilters:        make([]string, 0),
		NRValueFilters:         make([]string, 0),
	}
}

// ApplyDefaults sets default values for any unset configuration fields
func (c *Config) ApplyDefaults() {
	if c.NRApiServer == "" {
		c.NRApiServer = DefaultAPIServer
		logger.Debugf("Applied default API server: %s", DefaultAPIServer)
	}

	if c.NRPeriod == 0 {
		c.NRPeriod = DefaultPeriod
		logger.Debugf("Applied default period: %d seconds", DefaultPeriod)
	}

	if c.NRTimeout == 0 {
		c.NRTimeout = DefaultTimeout
		logger.Debugf("Applied default timeout: %v", DefaultTimeout)
	}

	if c.NRAppListCacheTime == 0 {
		c.NRAppListCacheTime = DefaultAppListCacheTime
		logger.Debugf("Applied default app list cache time: %v", DefaultAppListCacheTime)
	}

	if c.NRMetricNamesCacheTime == 0 {
		c.NRMetricNamesCacheTime = DefaultMetricNamesCacheTime
		logger.Debugf("Applied default metric names cache time: %v", DefaultMetricNamesCacheTime)
	}

	if c.ListenAddress == "" {
		c.ListenAddress = DefaultListenAddress
		logger.Debugf("Applied default listen address: %s", DefaultListenAddress)
	}

	if c.MetricPath == "" {
		c.MetricPath = DefaultMetricPath
		logger.Debugf("Applied default metric path: %s", DefaultMetricPath)
	}

	// Initialize empty slices if nil
	if c.NRApps == nil {
		c.NRApps = make([]Application, 0)
	}

	if c.NRMetricFilters == nil {
		c.NRMetricFilters = make([]string, 0)
	}

	if c.NRValueFilters == nil {
		c.NRValueFilters = make([]string, 0)
	}
}

// Validate checks if the configuration is valid and provides helpful error messages
func (c *Config) Validate() error {
	// Check required fields
	if c.NRApiKey == "" {
		return fmt.Errorf("api.key is required but not set")
	}

	if c.NRService == "" {
		return fmt.Errorf("api.service is required but not set (e.g., 'applications', 'mobile')")
	}

	if len(c.NRMetricFilters) == 0 {
		return fmt.Errorf("api.include-metric-filters is required but empty - at least one metric filter must be specified")
	}

	// Warn about potentially incorrect duration values
	if c.NRTimeout < time.Second {
		logger.Warnf("api.timeout is set to %v which is less than 1 second - this may be too short for API requests", c.NRTimeout)
	}

	return nil
}

// WriteDefaultConfig writes a configuration file with all default values and comments
func WriteDefaultConfig(path string) error {
	defaultConfig := `# New Relic Exporter Configuration with Defaults
# All values shown are the defaults that will be used if not specified

# REQUIRED: Your New Relic API key
api.key: YOUR_API_KEY_HERE

# API server (default: https://api.newrelic.com)
api.server: https://api.newrelic.com

# Data request period in seconds (default: 60)
api.period: 60

# API request timeout (default: 15s)
# IMPORTANT: Must include time unit (s, m, h)
api.timeout: 15s

# REQUIRED: New Relic service type
# Common values: applications, mobile, servers, plugins
api.service: applications

# Application list cache duration (default: 1h)
# Set to 0 to disable caching
api.apps-list-cache-time: 1h

# Metric names cache duration (default: 1h)
# Set to 0 to disable caching
api.metric-names-cache-time: 1h

# Filter specific applications (optional, default: all applications)
# api.include-apps:
#   - name: "My Application"
#   - id: 123456

# REQUIRED: Metric filters (at least one required)
# Common filters: HttpDispatcher, Database, External, Apdex, Memory, CPU
api.include-metric-filters:
  - HttpDispatcher
  - Database

# Value filters (optional, default: all values)
# api.include-values:
#   - average_response_time
#   - throughput

# Web server listen address (default: :9126)
web.listen-address: ":9126"

# Metrics endpoint path (default: /metrics)
web.telemetry-path: "/metrics"

# Debug proxy address (optional, default: none)
# debug.proxy-address: "http://localhost:8888"
`

	err := ioutil.WriteFile(path, []byte(defaultConfig), 0644)
	if err != nil {
		return fmt.Errorf("failed to write default config: %w", err)
	}

	logger.Infof("Default configuration written to: %s", path)
	return nil
}

// PrintDefaults prints all default values to stdout
func PrintDefaults() {
	fmt.Println("Default Configuration Values:")
	fmt.Printf("  api.server: %s\n", DefaultAPIServer)
	fmt.Printf("  api.period: %d seconds\n", DefaultPeriod)
	fmt.Printf("  api.timeout: %v\n", DefaultTimeout)
	fmt.Printf("  api.apps-list-cache-time: %v\n", DefaultAppListCacheTime)
	fmt.Printf("  api.metric-names-cache-time: %v\n", DefaultMetricNamesCacheTime)
	fmt.Printf("  web.listen-address: %s\n", DefaultListenAddress)
	fmt.Printf("  web.telemetry-path: %s\n", DefaultMetricPath)
	fmt.Println("\nRequired fields (no defaults):")
	fmt.Println("  api.key - Your New Relic API key")
	fmt.Println("  api.service - Service type (e.g., applications, mobile)")
	fmt.Println("  api.include-metric-filters - At least one metric filter")
}

// GetConfigOrDefault attempts to load config from file, or returns defaults if file doesn't exist
func GetConfigOrDefault(path string) (Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		logger.Warnf("Config file not found at %s, using defaults", path)
		config := NewDefaultConfig()
		config.ApplyDefaults()
		return config, nil
	}

	return GetConfig(path)
}
