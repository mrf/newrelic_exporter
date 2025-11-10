package config

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestDefaultConstants(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected interface{}
	}{
		{"DefaultAPIServer", DefaultAPIServer, "https://api.newrelic.com"},
		{"DefaultPeriod", DefaultPeriod, 60},
		{"DefaultTimeout", DefaultTimeout, 15 * time.Second},
		{"DefaultAppListCacheTime", DefaultAppListCacheTime, 1 * time.Hour},
		{"DefaultMetricNamesCacheTime", DefaultMetricNamesCacheTime, 1 * time.Hour},
		{"DefaultListenAddress", DefaultListenAddress, ":9126"},
		{"DefaultMetricPath", DefaultMetricPath, "/metrics"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.value, tt.expected)
			}
		})
	}
}

func TestNewDefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig()

	if cfg.NRApiServer != DefaultAPIServer {
		t.Errorf("Expected API server %s, got %s", DefaultAPIServer, cfg.NRApiServer)
	}

	if cfg.NRPeriod != DefaultPeriod {
		t.Errorf("Expected period %d, got %d", DefaultPeriod, cfg.NRPeriod)
	}

	if cfg.NRTimeout != DefaultTimeout {
		t.Errorf("Expected timeout %v, got %v", DefaultTimeout, cfg.NRTimeout)
	}

	if cfg.NRAppListCacheTime != DefaultAppListCacheTime {
		t.Errorf("Expected app list cache time %v, got %v", DefaultAppListCacheTime, cfg.NRAppListCacheTime)
	}

	if cfg.NRMetricNamesCacheTime != DefaultMetricNamesCacheTime {
		t.Errorf("Expected metric names cache time %v, got %v", DefaultMetricNamesCacheTime, cfg.NRMetricNamesCacheTime)
	}

	if cfg.ListenAddress != DefaultListenAddress {
		t.Errorf("Expected listen address %s, got %s", DefaultListenAddress, cfg.ListenAddress)
	}

	if cfg.MetricPath != DefaultMetricPath {
		t.Errorf("Expected metric path %s, got %s", DefaultMetricPath, cfg.MetricPath)
	}

	// Check slices are initialized
	if cfg.NRApps == nil {
		t.Error("NRApps should be initialized, got nil")
	}

	if cfg.NRMetricFilters == nil {
		t.Error("NRMetricFilters should be initialized, got nil")
	}

	if cfg.NRValueFilters == nil {
		t.Error("NRValueFilters should be initialized, got nil")
	}
}

func TestApplyDefaults(t *testing.T) {
	// Test with empty config
	cfg := Config{}
	cfg.ApplyDefaults()

	if cfg.NRApiServer != DefaultAPIServer {
		t.Errorf("Expected default API server, got %s", cfg.NRApiServer)
	}

	if cfg.NRPeriod != DefaultPeriod {
		t.Errorf("Expected default period, got %d", cfg.NRPeriod)
	}

	if cfg.ListenAddress != DefaultListenAddress {
		t.Errorf("Expected default listen address, got %s", cfg.ListenAddress)
	}
}

func TestApplyDefaultsPreservesUserValues(t *testing.T) {
	// Test that user-provided values are not overridden
	cfg := Config{
		NRApiServer:   "https://custom.api.com",
		NRPeriod:      120,
		ListenAddress: ":8080",
	}

	cfg.ApplyDefaults()

	if cfg.NRApiServer != "https://custom.api.com" {
		t.Error("ApplyDefaults should not override user-provided API server")
	}

	if cfg.NRPeriod != 120 {
		t.Error("ApplyDefaults should not override user-provided period")
	}

	if cfg.ListenAddress != ":8080" {
		t.Error("ApplyDefaults should not override user-provided listen address")
	}
}

func TestGetConfigWithDefaults(t *testing.T) {
	// Create a minimal config file
	configContent := `
api.key: test-api-key
api.service: applications
api.timeout: 15s
api.include-metric-filters:
  - HttpDispatcher
`

	tmpfile, err := ioutil.TempFile("", "config-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(configContent)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	cfg, err := GetConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}

	// Check required fields were set
	if cfg.NRApiKey != "test-api-key" {
		t.Error("API key not set from config")
	}

	if cfg.NRService != "applications" {
		t.Error("Service not set from config")
	}

	// Check defaults were applied
	if cfg.NRApiServer != DefaultAPIServer {
		t.Errorf("Expected default API server, got %s", cfg.NRApiServer)
	}

	if cfg.NRPeriod != DefaultPeriod {
		t.Errorf("Expected default period, got %d", cfg.NRPeriod)
	}

	if cfg.ListenAddress != DefaultListenAddress {
		t.Errorf("Expected default listen address, got %s", cfg.ListenAddress)
	}

	if cfg.MetricPath != DefaultMetricPath {
		t.Errorf("Expected default metric path, got %s", cfg.MetricPath)
	}
}

func TestGetConfigOverridesDefaults(t *testing.T) {
	// Create a config with custom values
	configContent := `
api.key: test-api-key
api.server: https://custom.api.com
api.period: 120
api.timeout: 30s
api.service: applications
api.include-metric-filters:
  - HttpDispatcher
web.listen-address: ":8080"
web.telemetry-path: "/custom-metrics"
`

	tmpfile, err := ioutil.TempFile("", "config-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(configContent)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	cfg, err := GetConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}

	// Check custom values were used instead of defaults
	if cfg.NRApiServer != "https://custom.api.com" {
		t.Errorf("Expected custom API server, got %s", cfg.NRApiServer)
	}

	if cfg.NRPeriod != 120 {
		t.Errorf("Expected custom period 120, got %d", cfg.NRPeriod)
	}

	if cfg.ListenAddress != ":8080" {
		t.Errorf("Expected custom listen address, got %s", cfg.ListenAddress)
	}

	if cfg.MetricPath != "/custom-metrics" {
		t.Errorf("Expected custom metric path, got %s", cfg.MetricPath)
	}
}

func TestWriteDefaultConfig(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "default-config-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	err = WriteDefaultConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("WriteDefaultConfig failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(tmpfile.Name()); os.IsNotExist(err) {
		t.Error("Default config file was not created")
	}

	// Read and verify content
	content, err := ioutil.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	contentStr := string(content)

	// Check that file contains expected elements
	expectedStrings := []string{
		"api.key:",
		"api.server:",
		"api.timeout:",
		"api.service:",
		"api.include-metric-filters:",
		"web.listen-address:",
		"web.telemetry-path:",
	}

	for _, expected := range expectedStrings {
		if !contains(contentStr, expected) {
			t.Errorf("Default config should contain '%s'", expected)
		}
	}
}

func TestGetConfigOrDefaultWithExistingFile(t *testing.T) {
	configContent := `
api.key: test-key
api.service: applications
api.timeout: 15s
api.include-metric-filters:
  - HttpDispatcher
`

	tmpfile, err := ioutil.TempFile("", "config-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(configContent)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	cfg, err := GetConfigOrDefault(tmpfile.Name())
	if err != nil {
		t.Fatalf("GetConfigOrDefault failed: %v", err)
	}

	if cfg.NRApiKey != "test-key" {
		t.Error("Config should be loaded from file")
	}
}

func TestGetConfigOrDefaultWithMissingFile(t *testing.T) {
	// Use a path that doesn't exist
	cfg, err := GetConfigOrDefault("/tmp/nonexistent-config-file.yml")

	// Should not error, should return defaults
	if err != nil {
		t.Errorf("GetConfigOrDefault should not error on missing file, got: %v", err)
	}

	// Should have defaults
	if cfg.NRApiServer != DefaultAPIServer {
		t.Error("Should have default API server")
	}

	if cfg.ListenAddress != DefaultListenAddress {
		t.Error("Should have default listen address")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
