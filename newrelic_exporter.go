package main

import (
	"flag"
	"github.com/invokermain/newrelic_exporter/config"
	"github.com/invokermain/newrelic_exporter/exporter"
	"github.com/invokermain/newrelic_exporter/newrelic"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"net/http"
)

func main() {
	var configFile string

	flag.StringVar(&configFile, "config", "newrelic_exporter.yml", "Config file path. Defaults to 'newrelic_exporter.yml'")
	flag.Parse()

	cfg, err := config.GetConfig(configFile)

	api := newrelic.NewAPI(cfg)

	exp := exporter.NewExporter(api, cfg)

	prometheus.MustRegister(exp)

	http.Handle(cfg.MetricPath, prometheus.Handler())
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

	log.Infof("Listening on %s.", cfg.ListenAddress)
	err = http.ListenAndServe(cfg.ListenAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Infor("HTTP server stopped.")
}
