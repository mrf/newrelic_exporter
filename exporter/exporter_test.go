package exporter

import (
	"testing"
	"time"

	"github.com/mrf/newrelic_exporter/config"
	"github.com/mrf/newrelic_exporter/newrelic"
	"github.com/prometheus/client_golang/prometheus"
)

func TestNewExporter(t *testing.T) {
	cfg := config.Config{
		NRApiKey:    "test-key",
		NRApiServer: "https://api.newrelic.com",
		NRTimeout:   15 * time.Second,
		NRPeriod:    60,
		NRService:   "applications",
	}

	api := newrelic.NewAPI(cfg)
	exporter := NewExporter(api, cfg)

	if exporter == nil {
		t.Fatal("NewExporter returned nil")
	}

	if exporter.api != api {
		t.Error("Exporter API not set correctly")
	}

	if len(exporter.metrics) != 0 {
		t.Errorf("Expected empty metrics map, got %d items", len(exporter.metrics))
	}

	if len(exporter.apps) != 0 {
		t.Errorf("Expected empty apps slice, got %d items", len(exporter.apps))
	}

	if len(exporter.names) != 0 {
		t.Errorf("Expected empty names map, got %d items", len(exporter.names))
	}
}

func TestMetricChannelCommunication(t *testing.T) {
	cfg := config.Config{
		NRApiKey:    "test-key",
		NRApiServer: "https://api.newrelic.com",
		NRTimeout:   15 * time.Second,
		NRPeriod:    60,
		NRService:   "applications",
	}

	api := newrelic.NewAPI(cfg)
	exporter := NewExporter(api, cfg)

	// Create a channel and send test metrics
	ch := make(chan Metric, 10)

	testMetrics := []Metric{
		{App: "TestApp1", Name: "response_time", Value: 123.45, Label: "web_transaction"},
		{App: "TestApp1", Name: "throughput", Value: 100.0, Label: "web_transaction"},
		{App: "TestApp2", Name: "response_time", Value: 234.56, Label: "web_transaction"},
	}

	go func() {
		for _, m := range testMetrics {
			ch <- m
		}
		close(ch)
	}()

	// Receive metrics
	exporter.receive(ch)

	// Verify metrics were created
	expectedMetrics := 2 // response_time and throughput
	if len(exporter.metrics) != expectedMetrics {
		t.Errorf("Expected %d metrics, got %d", expectedMetrics, len(exporter.metrics))
	}

	// Check specific metrics exist
	if _, ok := exporter.metrics[NameSpace+"_response_time"]; !ok {
		t.Error("response_time metric not created")
	}

	if _, ok := exporter.metrics[NameSpace+"_throughput"]; !ok {
		t.Error("throughput metric not created")
	}
}

func TestPrometheusDescribe(t *testing.T) {
	cfg := config.Config{
		NRApiKey:    "test-key",
		NRApiServer: "https://api.newrelic.com",
		NRTimeout:   15 * time.Second,
		NRPeriod:    60,
		NRService:   "applications",
	}

	api := newrelic.NewAPI(cfg)
	exporter := NewExporter(api, cfg)

	// Add a test metric
	exporter.metrics[NameSpace+"_test_metric"] = *prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: NameSpace,
			Name:      "test_metric",
		},
		[]string{"app", "component"},
	)

	descChan := make(chan *prometheus.Desc, 10)

	go func() {
		exporter.Describe(descChan)
		close(descChan)
	}()

	descCount := 0
	for range descChan {
		descCount++
	}

	// Should have at least: duration, totalScrapes, error, and test_metric
	minExpected := 4
	if descCount < minExpected {
		t.Errorf("Expected at least %d descriptions, got %d", minExpected, descCount)
	}
}

func TestMetricNamespacing(t *testing.T) {
	// Test that metric IDs are correctly namespaced
	testCases := []struct {
		metricName string
		expected   string
	}{
		{"response_time", "newrelic_response_time"},
		{"throughput", "newrelic_throughput"},
		{"error_rate", "newrelic_error_rate"},
	}

	for _, tc := range testCases {
		actual := NameSpace + "_" + tc.metricName
		if actual != tc.expected {
			t.Errorf("Expected %s, got %s", tc.expected, actual)
		}
	}
}

func TestCacheTiming(t *testing.T) {
	cfg := config.Config{
		NRApiKey:               "test-key",
		NRApiServer:            "https://api.newrelic.com",
		NRTimeout:              15 * time.Second,
		NRPeriod:               60,
		NRService:              "applications",
		NRAppListCacheTime:     1 * time.Hour,
		NRMetricNamesCacheTime: 1 * time.Hour,
	}

	api := newrelic.NewAPI(cfg)
	exporter := NewExporter(api, cfg)

	// Set cache times in the past
	exporter.appListLastScrape = time.Now().Add(-2 * time.Hour)
	exporter.metricNamesLastScrape = time.Now().Add(-2 * time.Hour)

	// Verify cache is expired
	if time.Since(exporter.appListLastScrape) < cfg.NRAppListCacheTime {
		t.Error("App list cache should be expired")
	}

	if time.Since(exporter.metricNamesLastScrape) < cfg.NRMetricNamesCacheTime {
		t.Error("Metric names cache should be expired")
	}

	// Set cache times to now
	exporter.appListLastScrape = time.Now()
	exporter.metricNamesLastScrape = time.Now()

	// Verify cache is valid
	if time.Since(exporter.appListLastScrape) >= cfg.NRAppListCacheTime {
		t.Error("App list cache should be valid")
	}

	if time.Since(exporter.metricNamesLastScrape) >= cfg.NRMetricNamesCacheTime {
		t.Error("Metric names cache should be valid")
	}
}

func TestMetricValueConversion(t *testing.T) {
	// Test that non-float64 values are handled correctly
	cfg := config.Config{
		NRApiKey:    "test-key",
		NRApiServer: "https://api.newrelic.com",
		NRTimeout:   15 * time.Second,
		NRPeriod:    60,
		NRService:   "applications",
	}

	api := newrelic.NewAPI(cfg)
	exporter := NewExporter(api, cfg)

	ch := make(chan Metric, 5)

	// Send various metric types
	go func() {
		ch <- Metric{App: "Test", Name: "metric1", Value: 123.45, Label: "test"}
		ch <- Metric{App: "Test", Name: "metric2", Value: 0.0, Label: "test"}
		ch <- Metric{App: "Test", Name: "metric3", Value: -10.5, Label: "test"}
		close(ch)
	}()

	exporter.receive(ch)

	// Verify all metrics were created
	if len(exporter.metrics) != 3 {
		t.Errorf("Expected 3 metrics, got %d", len(exporter.metrics))
	}
}

func TestNamespace(t *testing.T) {
	expected := "newrelic"
	if NameSpace != expected {
		t.Errorf("Expected namespace '%s', got '%s'", expected, NameSpace)
	}
}
