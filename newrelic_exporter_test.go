package main

import (
	"flag"
	"fmt"
	"github.com/mrf/newrelic_exporter/config"
	"github.com/mrf/newrelic_exporter/exporter"
	"github.com/mrf/newrelic_exporter/newrelic"
	"github.com/prometheus/client_golang/prometheus"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)
var testApiKey string = "205071e37e95bdaa327c62ccd3201da9289ccd17"
var testApiAppId int = 9045822
var testTimeout time.Duration = 5 * time.Second
var testPostData string = "names[]=Datastore%2Fstatement%2FJDBC%2Fmessages%2Finsert&raw=true&summarize=true&period=0&from=0001-01-01T00:00:00Z&to=0001-01-01T00:00:00Z"

func TestAppListGet(t *testing.T) {

	ts, err := testServer()
	if err != nil {
		t.Fatal(err)
	}

	defer ts.Close()

	var configFile string
	flag.StringVar(&configFile, "config", "_testing/newrelic_exporter_test_config.yml", "Config file path. Defaults to 'newrelic_exporter.yml'")
	flag.Parse()
	cfg, err := config.GetConfig(configFile)
	cfg.NRApiServer = ts.URL
	api := newrelic.NewAPI(cfg)

	app,err := api.GetApplications()
	if err != nil {
		t.Fatal(err)
	}

	if len(app) != 1 {
		t.Fatal("Expected 1 application, got", len(app))
	}

	a := app[0]

	switch {

	case a.ID != testApiAppId:
		t.Fatal("Wrong ID")

	case a.Health != "green":
		t.Fatal("Wrong health status")

	case a.Name != "Test/Client/Name":
		t.Fatal("Wrong name")

	case a.AppSummary["throughput"] != 54.7:
		t.Fatal("Wrong throughput")

	case a.AppSummary["host_count"] != 3:
		t.Fatal("Wrong host count")

	case a.UsrSummary["response_time"] != 4.61:
		t.Fatal("Wrong response time")

	}

}

func TestMetricNamesGet(t *testing.T) {

	ts, err := testServer()
	if err != nil {
		t.Fatal(err)
	}

	defer ts.Close()

	var configFile string
	flag.StringVar(&configFile, "config", "_testing/newrelic_exporter_test_config.yml", "Config file path. Defaults to 'newrelic_exporter.yml'")
	flag.Parse()
	cfg, err := config.GetConfig(configFile)
	cfg.NRApiServer = ts.URL
	api := newrelic.NewAPI(cfg)

	names,err := api.GetMetricNames(testApiAppId)
	if err != nil {
		t.Fatal(err)
	}

	if len(names) != 2 {
		t.Fatal("Expected 2 name sets, got", len(names))
	}

	if len(names[0].ValueNames) != 10 {
		t.Fatal("Expected 10 metric names")
	}

	if names[0].Name != "Datastore/statement/JDBC/messages/insert" {
		t.Fatal("Wrong application name")
	}
	if names[1].Name != "Datastore/statement/JDBC/messages/update" {
		t.Fatal("Wrong application name")
	}

}

func TestMetricValuesGet(t *testing.T) {

	ts, err := testServer()
	if err != nil {
		t.Fatal(err)
	}

	defer ts.Close()

	var configFile string
	flag.StringVar(&configFile, "config", "_testing/newrelic_exporter_test_config.yml", "Config file path. Defaults to 'newrelic_exporter.yml'")
	flag.Parse()
	cfg, err := config.GetConfig(configFile)
	cfg.NRApiServer = ts.URL
	api := newrelic.NewAPI(cfg)

	names,err := api.GetMetricNames(testApiAppId)
	if err != nil {
		t.Fatal(err)
	}

	data,err := api.GetMetricData(testApiAppId,names, time.Now(),time.Now())
	if err != nil {
		t.Fatal(err)
	}

	if len(data) != 1 {
		t.Fatal("Expected 1 metric sets")
	}

	if len(data[0].Timeslices) != 1 {
		t.Fatal("Expected 1 timeslice")
	}

	appData := data[0].Timeslices[0]

	if len(appData.Values) != 10 {
		t.Fatal("Expected 10 data points")
	}

	if appData.Values["call_count"].(float64) != 2 {
		t.Fatal("Wrong call_count value")
	}

	if appData.Values["calls_per_minute"].(float64) != 2.03 {
		t.Fatal("Wrong calls_per_minute value")
	}

}

func TestScrapeAPI(t *testing.T) {

	ts, err := testServer()
	if err != nil {
		t.Fatal(err)
	}

	defer ts.Close()

	var configFile string
	flag.StringVar(&configFile, "config", "_testing/newrelic_exporter_test_config.yml", "Config file path. Defaults to 'newrelic_exporter.yml'")
	flag.Parse()
	cfg, err := config.GetConfig(configFile)
	cfg.NRApiServer = ts.URL
	api := newrelic.NewAPI(cfg)
	exporter := exporter.NewExporter(api,cfg)

	var recieved []prometheus.Metric

	metrics := make(chan prometheus.Metric)

	go exporter.Collect(metrics)

	for m := range metrics {
		recieved = append(recieved, m)
	}

	if len(recieved) != 21 {
		t.Fatal("Expected 21 metrics")
	}

}

func testServer() (ts *httptest.Server, err error) {

	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Header.Get("X-Api-Key") != testApiKey {
			w.WriteHeader(403)
		}

		var body []byte
		var sourceFile string

		firstLink := fmt.Sprintf(
			"<%s%s?page=%d>; rel=%s, <%s%s?page=%d>; rel=%s",
			ts.URL, r.URL.Path, 2, `"next"`,
			ts.URL, r.URL.Path, 2, `"last"`)

		secondLink := fmt.Sprintf(
			"<%s%s?page=%d>; rel=%s, <%s%s?page=%d>; rel=%s",
			ts.URL, r.URL.Path, 1, `"first"`,
			ts.URL, r.URL.Path, 1, `"prev"`)

		switch r.URL.Path {

		case "/v2/applications.json":
			sourceFile = "_testing/application_list.json"

		case "/v2/applications/9045822/metrics.json":
			if r.URL.Query().Get("page") == "2" {
				sourceFile = ("_testing/metric_names_2.json")
				w.Header().Set("Link", secondLink)
			} else {
				sourceFile = ("_testing/metric_names.json")
				w.Header().Set("Link", firstLink)
			}

		case "/v2/applications/9045822/metrics/data.json":
			sourceFile = ("_testing/metric_data.json")

		default:
			w.WriteHeader(404)
			return

		}

		body, err = ioutil.ReadFile(sourceFile)
		if err != nil {
			return
		}

		w.WriteHeader(200)
		w.Write(body)

	}))

	return
}
