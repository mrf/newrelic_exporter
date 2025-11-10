package main

import (
	"flag"
	"net/http"

	"github.com/mrf/newrelic_exporter/config"
	"github.com/mrf/newrelic_exporter/exporter"
	"github.com/mrf/newrelic_exporter/logger"
	"github.com/mrf/newrelic_exporter/newrelic"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	var configFile string
	var logLevel string
	var logJSON bool

	flag.StringVar(&configFile, "config", "newrelic_exporter.yml", "Config file path")
	flag.StringVar(&logLevel, "log.level", "info", "Log level (debug, info, warn, error)")
	flag.BoolVar(&logJSON, "log.json", false, "Output logs in JSON format")
	flag.Parse()

	// Setup logging
	logger.Setup(logger.Config{
		Level:      logger.LogLevel(logLevel),
		JSONFormat: logJSON,
	})

	cfg, err := config.GetConfig(configFile)
	if err != nil {
		logger.Fatalf("Error loading config file '%s': %v", configFile, err)
	}

	logger.Infof("Configuration loaded successfully from %s", configFile)

	api := newrelic.NewAPI(cfg)

	exp := exporter.NewExporter(api, cfg)

	prometheus.MustRegister(exp)

	http.Handle(cfg.MetricPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
<head><title>NewRelic exporter</title></head>
<body>
<h1>NewRelic exporter</h1>
<p><a href='` + cfg.MetricPath + `'>Metrics</a></p>
</body>
</html>
`))
	})

	logger.Infof("Listening on %s", cfg.ListenAddress)
	err = http.ListenAndServe(cfg.ListenAddress, nil)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Info("HTTP server stopped")
}
