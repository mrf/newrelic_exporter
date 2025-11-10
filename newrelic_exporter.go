package main

import (
	"flag"
	"net/http"

	"github.com/mrf/newrelic_exporter/config"
	"github.com/mrf/newrelic_exporter/exporter"
	"github.com/mrf/newrelic_exporter/newrelic"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/log"
)

func main() {
	var configFile string
	var printDefaults bool
	var generateConfig string

	flag.StringVar(&configFile, "config", "newrelic_exporter.yml", "Config file path")
	flag.BoolVar(&printDefaults, "print-defaults", false, "Print default configuration values and exit")
	flag.StringVar(&generateConfig, "generate-config", "", "Generate a default configuration file at the specified path and exit")
	flag.Parse()

	// Handle utility flags
	if printDefaults {
		config.PrintDefaults()
		return
	}

	if generateConfig != "" {
		err := config.WriteDefaultConfig(generateConfig)
		if err != nil {
			log.Fatalf("Failed to generate config: %v", err)
		}
		log.Infof("Default configuration written to: %s", generateConfig)
		log.Info("Please edit the file to set your API key and other required fields")
		return
	}

	// Load configuration with defaults
	cfg, err := config.GetConfig(configFile)
	if err != nil {
		log.Fatalf("Error loading config file '%s': %v", configFile, err)
	}

	log.Infof("Configuration loaded successfully from %s", configFile)

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

	log.Printf("Listening on %s.", cfg.ListenAddress)
	err = http.ListenAndServe(cfg.ListenAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("HTTP server stopped.")
}
