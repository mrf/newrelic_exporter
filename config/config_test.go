package config

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestConfigValidation_MissingAPIKey(t *testing.T) {
	configContent := `
api.server: https://api.newrelic.com
api.timeout: 15s
api.service: applications
api.include-metric-filters:
  - HttpDispatcher
`
	tmpfile := createTempConfig(t, configContent)
	defer os.Remove(tmpfile)

	_, err := GetConfig(tmpfile)
	if err == nil {
		t.Fatal("Expected error for missing API key, got nil")
	}
	if !strings.Contains(err.Error(), "api.key") {
		t.Errorf("Expected error message to mention 'api.key', got: %v", err)
	}
}

func TestConfigValidation_MissingService(t *testing.T) {
	configContent := `
api.key: test-key-123
api.timeout: 15s
api.include-metric-filters:
  - HttpDispatcher
`
	tmpfile := createTempConfig(t, configContent)
	defer os.Remove(tmpfile)

	_, err := GetConfig(tmpfile)
	if err == nil {
		t.Fatal("Expected error for missing service, got nil")
	}
	if !strings.Contains(err.Error(), "api.service") {
		t.Errorf("Expected error message to mention 'api.service', got: %v", err)
	}
}

func TestConfigValidation_MissingMetricFilters(t *testing.T) {
	configContent := `
api.key: test-key-123
api.timeout: 15s
api.service: applications
`
	tmpfile := createTempConfig(t, configContent)
	defer os.Remove(tmpfile)

	_, err := GetConfig(tmpfile)
	if err == nil {
		t.Fatal("Expected error for missing metric filters, got nil")
	}
	if !strings.Contains(err.Error(), "api.include-metric-filters") {
		t.Errorf("Expected error message to mention 'api.include-metric-filters', got: %v", err)
	}
}

func TestConfigValidation_MissingTimeoutUnit(t *testing.T) {
	configContent := `
api.key: test-key-123
api.service: applications
api.timeout: 15
api.include-metric-filters:
  - HttpDispatcher
`
	tmpfile := createTempConfig(t, configContent)
	defer os.Remove(tmpfile)

	_, err := GetConfig(tmpfile)
	// YAML will parse "15" as 15 nanoseconds, which is less than 1 second
	// Our validation should catch this
	if err == nil {
		t.Fatal("Expected error or warning for timeout without units, got nil")
	}
}

func TestConfigValidation_MissingTimeout(t *testing.T) {
	configContent := `
api.key: test-key-123
api.service: applications
api.include-metric-filters:
  - HttpDispatcher
`
	tmpfile := createTempConfig(t, configContent)
	defer os.Remove(tmpfile)

	_, err := GetConfig(tmpfile)
	if err == nil {
		t.Fatal("Expected error for missing timeout, got nil")
	}
	if !strings.Contains(err.Error(), "api.timeout") {
		t.Errorf("Expected error message to mention 'api.timeout', got: %v", err)
	}
}

func TestConfigValidation_ValidConfig(t *testing.T) {
	configContent := `
api.key: test-key-123
api.server: https://api.newrelic.com
api.timeout: 15s
api.period: 60
api.service: applications
api.apps-list-cache-time: 1h
api.metric-names-cache-time: 1h
api.include-metric-filters:
  - HttpDispatcher
  - Database
web.listen-address: ":9126"
web.telemetry-path: "/metrics"
`
	tmpfile := createTempConfig(t, configContent)
	defer os.Remove(tmpfile)

	cfg, err := GetConfig(tmpfile)
	if err != nil {
		t.Fatalf("Expected no error for valid config, got: %v", err)
	}

	// Verify values are parsed correctly
	if cfg.NRApiKey != "test-key-123" {
		t.Errorf("Expected API key 'test-key-123', got: %s", cfg.NRApiKey)
	}

	if cfg.NRTimeout.Seconds() != 15 {
		t.Errorf("Expected timeout 15s, got: %v", cfg.NRTimeout)
	}

	if len(cfg.NRMetricFilters) != 2 {
		t.Errorf("Expected 2 metric filters, got: %d", len(cfg.NRMetricFilters))
	}
}

func TestConfigValidation_Defaults(t *testing.T) {
	configContent := `
api.key: test-key-123
api.timeout: 15s
api.service: applications
api.include-metric-filters:
  - HttpDispatcher
`
	tmpfile := createTempConfig(t, configContent)
	defer os.Remove(tmpfile)

	cfg, err := GetConfig(tmpfile)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check defaults are set
	if cfg.NRApiServer != "https://api.newrelic.com" {
		t.Errorf("Expected default API server, got: %s", cfg.NRApiServer)
	}

	if cfg.MetricPath != "/metrics" {
		t.Errorf("Expected default metric path '/metrics', got: %s", cfg.MetricPath)
	}

	if cfg.ListenAddress != ":9126" {
		t.Errorf("Expected default listen address ':9126', got: %s", cfg.ListenAddress)
	}

	if cfg.NRPeriod != 60 {
		t.Errorf("Expected default period 60, got: %d", cfg.NRPeriod)
	}
}

func createTempConfig(t *testing.T, content string) string {
	tmpfile, err := ioutil.TempFile("", "config-*.yml")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}

	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	return tmpfile.Name()
}
