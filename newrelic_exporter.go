package main

import (
	"github.com/invokermain/newrelic_exporter/config"
	"github.com/invokermain/newrelic_exporter/exporter"
	"github.com/invokermain/newrelic_exporter/newrelic"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"net/http"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	var (
		configFile = kingpin.Flag("config", "Config file path. Defaults to 'newrelic_exporter.yml'").Default("newrelic_exporter.yml").String()
	)

	// Parse Flags
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("newrelic_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	cfg, err := config.GetConfig(*configFile)

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
	log.Info("HTTP server stopped.")
}
