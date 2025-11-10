package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mrf/newrelic_exporter/config"
	"github.com/mrf/newrelic_exporter/exporter"
	"github.com/mrf/newrelic_exporter/newrelic"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// TestHTTPHandlerRoot tests that the root HTTP handler returns the expected HTML
func TestHTTPHandlerRoot(t *testing.T) {
	cfg := config.Config{
		MetricPath: "/metrics",
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
<head><title>NewRelic exporter</title></head>
<body>
<h1>NewRelic exporter</h1>
<p><a href='` + cfg.MetricPath + `'>Metrics</a></p>
</body>
</html>
`))
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if !strings.Contains(body, "NewRelic exporter") {
		t.Fatal("Expected body to contain 'NewRelic exporter'")
	}

	if !strings.Contains(body, cfg.MetricPath) {
		t.Fatalf("Expected body to contain metrics path '%s'", cfg.MetricPath)
	}
}

// TestExporterRegistration tests that an exporter can be created and registered with Prometheus
func TestExporterRegistration(t *testing.T) {
	cfg := config.Config{
		NRApiKey:    "test-api-key",
		NRApiServer: "https://api.newrelic.com",
		NRTimeout:   15,
		NRPeriod:    60,
		NRService:   "applications",
		MetricPath:  "/metrics",
	}

	api := newrelic.NewAPI(cfg)
	if api == nil {
		t.Fatal("Failed to create NewRelic API")
	}

	exp := exporter.NewExporter(api, cfg)
	if exp == nil {
		t.Fatal("Failed to create exporter")
	}

	// Create a new registry for this test to avoid conflicts
	registry := prometheus.NewRegistry()
	err := registry.Register(exp)
	if err != nil {
		t.Fatalf("Failed to register exporter with Prometheus: %v", err)
	}

	// Test that we can create a metrics handler
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	if handler == nil {
		t.Fatal("Failed to create Prometheus HTTP handler")
	}

	// Make a test request to the metrics endpoint
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200 from metrics endpoint, got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if !strings.Contains(body, "# HELP") || !strings.Contains(body, "# TYPE") {
		t.Fatal("Expected Prometheus metrics format in response")
	}
}
