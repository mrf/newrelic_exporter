package config

import (
	"fmt"
	"github.com/prometheus/log"
	"gopkg.in/yaml.v2"
	"io/ioutil"
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

func GetConfig(path string) (Config, error) {
	config := Config{}
	configSource, err := ioutil.ReadFile(path)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(configSource, &config)
	if err != nil {
		return config, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	err = config.Validate()
	if err != nil {
		return config, fmt.Errorf("config validation failed: %w", err)
	}

	log.Debugf("Config loaded: %v", config)

	return config, nil
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

	// Check duration fields - if they're 0, it might indicate parsing failure
	// YAML duration parsing expects proper units like "15s", "1h", "30m"
	if c.NRTimeout == 0 {
		return fmt.Errorf("api.timeout is not set or invalid - please specify a duration with units (e.g., '15s', '30s')")
	}

	// Warn about potentially incorrect duration values (less than 1 second suggests missing units)
	if c.NRTimeout < time.Second {
		log.Warnf("api.timeout is set to %v which is less than 1 second - did you forget to add time units? (e.g., '15s' instead of '15')", c.NRTimeout)
	}

	// Set defaults for optional fields
	if c.NRApiServer == "" {
		c.NRApiServer = "https://api.newrelic.com"
	}

	if c.MetricPath == "" {
		c.MetricPath = "/metrics"
	}

	if c.ListenAddress == "" {
		c.ListenAddress = ":9126"
	}

	if c.NRPeriod == 0 {
		c.NRPeriod = 60
	}

	return nil
}
