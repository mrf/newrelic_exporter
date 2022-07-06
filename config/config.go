package config

import (
	"io/ioutil"
	"time"

	"github.com/prometheus/log"
	"gopkg.in/yaml.v2"
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
	NRUseOnlySummary       bool          `yaml:"api.use-only-summary"`
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
		return config, err
	}

	log.Debugf("Config loaded: %v", config)

	return config, err
}
