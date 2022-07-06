package exporter

import (
	"fmt"
	"sync"
	"time"

	"github.com/klinux/newrelic_exporter/config"
	"github.com/klinux/newrelic_exporter/newrelic"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/log"
)

// Namespace for metrics
const NameSpace = "newrelic"

type Metric struct {
	App   string
	Name  string
	Value float64
	Label string
}

type Exporter struct {
	mu                                       sync.Mutex
	duration, error                          prometheus.Gauge
	totalScrapes                             prometheus.Counter
	metrics                                  map[string]prometheus.GaugeVec
	api                                      *newrelic.API
	cfg                                      config.Config
	apps                                     []newrelic.Application
	names                                    sync.Map
	values                                   []string
	appListLastScrape, metricNamesLastScrape time.Time
}

func NewExporter(api *newrelic.API, cfg config.Config) *Exporter {
	return &Exporter{
		duration: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: NameSpace,
			Name:      "exporter_last_scrape_duration_seconds",
			Help:      "The last scrape duration.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: NameSpace,
			Name:      "exporter_scrapes_total",
			Help:      "Total scraped metrics",
		}),
		error: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: NameSpace,
			Name:      "exporter_last_scrape_error",
			Help:      "The last scrape error status.",
		}),
		metrics: map[string]prometheus.GaugeVec{},
		api:     api,
		cfg:     cfg,
		apps:    make([]newrelic.Application, 0),
		names:   sync.Map{},
		values:  make([]string, 0),
	}
}

func (e *Exporter) scrape(from time.Time, to time.Time, ch chan<- Metric) {
	e.error.Set(0)
	e.totalScrapes.Inc()

	startTime := time.Now()
	log.Infof("Starting new scrape at %v for period from %v to %v.", startTime, from.Format(time.Stamp), to.Format(time.Stamp))

	if time.Since(e.appListLastScrape) >= e.cfg.NRAppListCacheTime {
		var err error
		e.apps, err = e.api.GetApplications()
		if err != nil {
			log.Error(err)
			e.error.Set(1)
		} else {
			// Only successful tries should touch cache times
			e.appListLastScrape = time.Now()
			log.Debugf("Application list updated at %v", e.appListLastScrape)
		}
	} else {
		log.Debug("Applications list taken from cache")
	}

	for _, app := range e.apps {
		for name, value := range app.AppSummary {
			ch <- Metric{
				App:   app.Name,
				Name:  name,
				Value: value,
				Label: "application_summary",
			}
		}

		for name, value := range app.UsrSummary {
			ch <- Metric{
				App:   app.Name,
				Name:  name,
				Value: value,
				Label: "end_user_summary",
			}
		}
	}

	if e.cfg.NRUseOnlySummary == false {
		var wg sync.WaitGroup
		for _, app := range e.apps {
			wg.Add(1)

			go func(app newrelic.Application) {
				defer wg.Done()

				var err error
				var names []newrelic.MetricName

				if time.Since(e.metricNamesLastScrape) >= e.cfg.NRAppListCacheTime {
					names, err = e.api.GetMetricNames(app.ID)
					e.names.Store(app.ID, names)
					log.Infof("Scraped %v metric names for app %v", len(names), app.ID)
					if err != nil {
						log.Error(err)
						e.error.Set(1)
					} else {
						// Only successful tries should touch cache times
						e.metricNamesLastScrape = time.Now()
						log.Debugf("Metric names list updated at %v", e.appListLastScrape)
					}
				} else {
					log.Debug("Metrics names list taken from cache")
				}

				// Getting metric data
				var data []newrelic.MetricData

				name, _ := e.names.Load(app.ID)
				data, err = e.api.GetMetricData(app.ID, name.([]newrelic.MetricName), from, to)
				log.Infof("Scraped %v metric datas for app %v", len(data), app.ID)
				if err != nil {
					log.Error(err)
					e.error.Set(1)
				}

				// Sending metrics
				for _, set := range data {
					if len(set.Timeslices) == 0 {
						continue
					}

					// As we set summarise=true there will only be one timeseries.
					for name, value := range set.Timeslices[0].Values {
						if v, ok := value.(float64); ok {
							ch <- Metric{
								App:   app.Name,
								Name:  name,
								Value: v,
								Label: set.Name,
							}
						}
					}
				}
			}(app)
		}
		wg.Wait()
	}

	close(ch)

	e.duration.Set(float64(time.Now().UnixNano()-startTime.UnixNano()) / 1000000000)
	log.Infof("Scrape finished in %v", time.Since(startTime))
}

func (e *Exporter) receive(ch <-chan Metric) {
	for metric := range ch {
		id := fmt.Sprintf("%s_%s", NameSpace, metric.Name)

		if m, ok := e.metrics[id]; ok {
			m.WithLabelValues(metric.App, metric.Label).Set(metric.Value)
		} else {
			g := prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: NameSpace,
					Name:      metric.Name,
				},
				[]string{"app", "component"})

			e.metrics[id] = *g
		}
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, m := range e.metrics {
		m.Describe(ch)
	}

	ch <- e.duration.Desc()
	ch <- e.totalScrapes.Desc()
	ch <- e.error.Desc()
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mu.Lock()
	defer e.mu.Unlock()

	var from, to time.Time
	from = time.Now().Add(-1 * time.Minute).Truncate(time.Minute)
	to = from.Add(time.Minute)

	metricChan := make(chan Metric)

	go e.scrape(from, to, metricChan)

	e.receive(metricChan)

	ch <- e.duration
	ch <- e.totalScrapes
	ch <- e.error

	for _, m := range e.metrics {
		m.Collect(ch)
	}
}
