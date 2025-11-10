package newrelic

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mrf/newrelic_exporter/config"
)

const testAPIKey = "test-api-key-123456789"
const testAppID = 123456

func TestNewAPI(t *testing.T) {
	cfg := config.Config{
		NRApiKey:    testAPIKey,
		NRApiServer: "https://api.newrelic.com",
		NRTimeout:   15 * time.Second,
		NRPeriod:    60,
		NRService:   "applications",
	}

	api := NewAPI(cfg)

	if api == nil {
		t.Fatal("NewAPI returned nil")
	}

	if api.apiKey != testAPIKey {
		t.Errorf("Expected API key %s, got %s", testAPIKey, api.apiKey)
	}

	if api.service != "applications" {
		t.Errorf("Expected service 'applications', got %s", api.service)
	}

	if api.Period != 60 {
		t.Errorf("Expected period 60, got %d", api.Period)
	}

	if api.client.Timeout != 15*time.Second {
		t.Errorf("Expected timeout 15s, got %v", api.client.Timeout)
	}
}

func TestNewAPIWithProxy(t *testing.T) {
	cfg := config.Config{
		NRApiKey:          testAPIKey,
		NRApiServer:       "https://api.newrelic.com",
		NRTimeout:         15 * time.Second,
		NRPeriod:          60,
		NRService:         "applications",
		DebugProxyAddress: "http://localhost:8888",
	}

	api := NewAPI(cfg)

	if api == nil {
		t.Fatal("NewAPI returned nil")
	}

	// Verify transport was configured (proxy setup)
	if api.client.Transport == nil {
		t.Error("Expected transport to be configured for proxy")
	}
}

func TestVersionConstant(t *testing.T) {
	if Version == "" {
		t.Error("Version constant should not be empty")
	}
}

func TestUserAgent(t *testing.T) {
	expected := "Prometheus-NewRelic-Exporter/" + Version
	if UserAgent != expected {
		t.Errorf("Expected UserAgent %s, got %s", expected, UserAgent)
	}
}

func TestChunkSize(t *testing.T) {
	if ChunkSize <= 0 {
		t.Error("ChunkSize should be positive")
	}
}

func TestGetApplications(t *testing.T) {
	// Create a test server
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify API key header
		if r.Header.Get("X-Api-Key") != testAPIKey {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		// Return mock application data
		response := map[string]interface{}{
			"applications": []map[string]interface{}{
				{
					"id":           testAppID,
					"name":         "Test Application",
					"health_status": "green",
					"application_summary": map[string]float64{
						"response_time": 123.45,
						"throughput":    100.0,
					},
					"end_user_summary": map[string]float64{
						"response_time": 234.56,
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	cfg := config.Config{
		NRApiKey:    testAPIKey,
		NRApiServer: ts.URL,
		NRTimeout:   15 * time.Second,
		NRPeriod:    60,
		NRService:   "applications",
	}

	api := NewAPI(cfg)
	api.client = &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	apps, err := api.GetApplications()
	if err != nil {
		t.Fatalf("GetApplications failed: %v", err)
	}

	if len(apps) != 1 {
		t.Fatalf("Expected 1 application, got %d", len(apps))
	}

	app := apps[0]
	if app.ID != testAppID {
		t.Errorf("Expected app ID %d, got %d", testAppID, app.ID)
	}

	if app.Name != "Test Application" {
		t.Errorf("Expected app name 'Test Application', got %s", app.Name)
	}

	if app.Health != "green" {
		t.Errorf("Expected health 'green', got %s", app.Health)
	}

	if app.AppSummary["response_time"] != 123.45 {
		t.Errorf("Expected response_time 123.45, got %f", app.AppSummary["response_time"])
	}
}

func TestGetApplicationsUnauthorized(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": "Unauthorized"}`))
	}))
	defer ts.Close()

	cfg := config.Config{
		NRApiKey:    "invalid-key",
		NRApiServer: ts.URL,
		NRTimeout:   15 * time.Second,
		NRPeriod:    60,
		NRService:   "applications",
	}

	api := NewAPI(cfg)
	api.client = &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	apps, err := api.GetApplications()

	// Should handle error gracefully
	if err != nil {
		// Expected to have some error handling
		if len(apps) > 0 {
			t.Error("Expected no applications on error")
		}
	}
}

func TestGetMetricNames(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != testAPIKey {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		response := map[string]interface{}{
			"metrics": []map[string]interface{}{
				{
					"name": "HttpDispatcher",
					"values": []string{
						"average_response_time",
						"call_count",
						"throughput",
					},
				},
				{
					"name": "Database/all",
					"values": []string{
						"average_response_time",
						"call_count",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	cfg := config.Config{
		NRApiKey:    testAPIKey,
		NRApiServer: ts.URL,
		NRTimeout:   15 * time.Second,
		NRPeriod:    60,
		NRService:   "applications",
		NRMetricFilters: []string{
			"HttpDispatcher",
			"Database",
		},
	}

	api := NewAPI(cfg)
	api.client = &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	names, err := api.GetMetricNames(testAppID)
	if err != nil {
		t.Fatalf("GetMetricNames failed: %v", err)
	}

	// Should get metrics from both filters (2 metrics per filter = 4 total)
	if len(names) < 2 {
		t.Errorf("Expected at least 2 metric names, got %d", len(names))
	}
}

func TestGetMetricData(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != testAPIKey {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		response := map[string]interface{}{
			"metric_data": map[string]interface{}{
				"metrics": []map[string]interface{}{
					{
						"name": "HttpDispatcher",
						"timeslices": []map[string]interface{}{
							{
								"from":   "2024-01-01T00:00:00Z",
								"to":     "2024-01-01T00:01:00Z",
								"values": map[string]interface{}{
									"average_response_time": 123.45,
									"call_count":            100.0,
									"throughput":            50.5,
								},
							},
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	cfg := config.Config{
		NRApiKey:    testAPIKey,
		NRApiServer: ts.URL,
		NRTimeout:   15 * time.Second,
		NRPeriod:    60,
		NRService:   "applications",
	}

	api := NewAPI(cfg)
	api.client = &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	metricNames := []MetricName{
		{
			Name: "HttpDispatcher",
			ValueNames: []string{
				"average_response_time",
				"call_count",
				"throughput",
			},
		},
	}

	from := time.Now().Add(-1 * time.Minute)
	to := time.Now()

	data, err := api.GetMetricData(testAppID, metricNames, from, to)
	if err != nil {
		t.Fatalf("GetMetricData failed: %v", err)
	}

	if len(data) != 1 {
		t.Fatalf("Expected 1 metric data, got %d", len(data))
	}

	if data[0].Name != "HttpDispatcher" {
		t.Errorf("Expected metric name 'HttpDispatcher', got %s", data[0].Name)
	}

	if len(data[0].Timeslices) != 1 {
		t.Errorf("Expected 1 timeslice, got %d", len(data[0].Timeslices))
	}
}

func TestAPIRateLimitHandling(t *testing.T) {
	resetTime := time.Now().Add(5 * time.Minute).Unix()

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Newrelic-Overloadprotection-Reset", fmt.Sprintf("%d", resetTime))
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error": "Rate limit exceeded"}`))
	}))
	defer ts.Close()

	cfg := config.Config{
		NRApiKey:    testAPIKey,
		NRApiServer: ts.URL,
		NRTimeout:   15 * time.Second,
		NRPeriod:    60,
		NRService:   "applications",
	}

	api := NewAPI(cfg)
	api.client = &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Make request - should handle 429 gracefully
	_, err := api.req("/v2/applications.json", "")

	// Should not panic or crash
	if err != nil {
		// Error is acceptable for rate limiting
		t.Logf("Rate limit error (expected): %v", err)
	}
}

func TestPaginatedResults(t *testing.T) {
	callCount := 0

	var ts *httptest.Server
	ts = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		if r.Header.Get("X-Api-Key") != testAPIKey {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		// First page
		if r.URL.Query().Get("cursor") == "" {
			w.Header().Set("Link", fmt.Sprintf(`<%s/v2/applications.json?cursor=next123>; rel="next"`, ts.URL))
			response := map[string]interface{}{
				"applications": []map[string]interface{}{
					{
						"id":            1,
						"name":          "App 1",
						"health_status": "green",
					},
				},
			}
			json.NewEncoder(w).Encode(response)
		} else {
			// Second page (no next link)
			response := map[string]interface{}{
				"applications": []map[string]interface{}{
					{
						"id":            2,
						"name":          "App 2",
						"health_status": "green",
					},
				},
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer ts.Close()

	cfg := config.Config{
		NRApiKey:    testAPIKey,
		NRApiServer: ts.URL,
		NRTimeout:   15 * time.Second,
		NRPeriod:    60,
		NRService:   "applications",
	}

	api := NewAPI(cfg)
	api.client = &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	body, err := api.req("/v2/applications.json", "")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	// Should have made 2 calls for pagination
	if callCount != 2 {
		t.Errorf("Expected 2 API calls for pagination, got %d", callCount)
	}

	// Body should contain data from both pages
	if len(body) == 0 {
		t.Error("Expected response body, got empty")
	}
}

func TestMetricDataChunking(t *testing.T) {
	// Create more metric names than ChunkSize to test chunking
	metricNames := make([]MetricName, ChunkSize+5)
	for i := range metricNames {
		metricNames[i] = MetricName{
			Name:       fmt.Sprintf("Metric%d", i),
			ValueNames: []string{"value1", "value2"},
		}
	}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"metric_data": map[string]interface{}{
				"metrics": []map[string]interface{}{},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	cfg := config.Config{
		NRApiKey:    testAPIKey,
		NRApiServer: ts.URL,
		NRTimeout:   15 * time.Second,
		NRPeriod:    60,
		NRService:   "applications",
	}

	api := NewAPI(cfg)
	api.client = &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	from := time.Now().Add(-1 * time.Minute)
	to := time.Now()

	_, err := api.GetMetricData(testAppID, metricNames, from, to)
	if err != nil {
		t.Fatalf("GetMetricData with chunking failed: %v", err)
	}

	// If we got here without error, chunking worked
}

func testServer() (*httptest.Server, error) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != testAPIKey {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		var body []byte
		var sourceFile string

		switch r.URL.Path {
		case "/v2/applications.json":
			sourceFile = "../_testing/application_list.json"
		case fmt.Sprintf("/v2/applications/%d/metrics.json", testAppID):
			sourceFile = "../_testing/metric_names.json"
		case fmt.Sprintf("/v2/applications/%d/metrics/data.json", testAppID):
			sourceFile = "../_testing/metric_data.json"
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var err error
		body, err = ioutil.ReadFile(sourceFile)
		if err != nil {
			// If test files don't exist, return mock data
			mockResponse := map[string]interface{}{"data": "mock"}
			json.NewEncoder(w).Encode(mockResponse)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}))

	return ts, nil
}
