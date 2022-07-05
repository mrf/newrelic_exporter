package newrelic

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/antonholmquist/jason"
	"github.com/klinux/newrelic_exporter/config"
	"github.com/prometheus/log"
	"github.com/tomnomnom/linkheader"
)

// Module version
const Version = "0.4b"

// User-Agent string
const UserAgent = "Prometheus-NewRelic-Exporter/" + Version

// Chunk size of metric requests
const ChunkSize = 10

var cfg config.Config

type API struct {
	server          url.URL
	apiKey          string
	service         string
	Period          int
	unreportingApps bool
	client          *http.Client
}

type Application struct {
	ID         int
	Name       string
	Health     string             `json:"health_status"`
	AppSummary map[string]float64 `json:"application_summary"`
	UsrSummary map[string]float64 `json:"end_user_summary"`
}

type MetricName struct {
	Name       string   `json:"name"`
	ValueNames []string `json:"values"`
}

type MetricData struct {
	Name       string
	Timeslices []struct {
		Values map[string]interface{}
	}
}

func NewAPI(c config.Config) *API {
	cfg = c

	serverURL, err := url.Parse(cfg.NRApiServer)
	if err != nil {
		log.Fatal("Could not parse API URL: ", err)
	}
	if cfg.NRApiKey == "" {
		log.Fatal("Cannot continue without an API key.")
	}
	if cfg.NRService == "" {
		log.Fatal("Cannot continue without NewRelic service selected")
	}

	client := &http.Client{Timeout: cfg.NRTimeout}

	if len(cfg.DebugProxyAddress) > 0 {
		proxyUrl, _ := url.Parse(cfg.DebugProxyAddress)
		transport := &http.Transport{}
		transport.Proxy = http.ProxyURL(proxyUrl)
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		client.Transport = transport
	}

	return &API{
		server:  *serverURL,
		apiKey:  cfg.NRApiKey,
		service: cfg.NRService,
		client:  client,
		Period:  cfg.NRPeriod,
	}
}

func (api *API) GetApplications() ([]Application, error) {
	log.Infof("Requesting application list from %s.", api.server.String())

	body, err := api.req(fmt.Sprintf("/v2/%s.json", api.service), "")
	if err != nil {
		log.Error("Error getting application list: ", err)
		return nil, err
	}

	v, err := jason.NewObjectFromBytes(body)

	var applications []Application

	if err == nil {
		appsArray, err := v.GetObjectArray("applications")

		applications = make([]Application, len(appsArray))
		if err == nil {
			for i, a := range appsArray {
				application := new(Application)
				aBytes, _ := a.Marshal()
				json.Unmarshal(aBytes, application)
				applications[i] = *application
			}
		}

		log.Debugf("Found %v applications: %v", len(applications), applications)
	}

	return applications, err
}

func (api *API) GetMetricNames(appID int) ([]MetricName, error) {
	log.Infof("Requesting metrics names for application id %d with %v filters", appID, len(cfg.NRMetricFilters))
	path := fmt.Sprintf("/v2/%s/%s/metrics.json", api.service, strconv.Itoa(appID))

	channel := make(chan MetricName)
	metricNames := make([]MetricName, 0)

	// We will only make filtered requests for metric names. Otherwise there are too many of them (tens of thousands)
	go func(ch chan MetricName) {
		var filter string
		var wg sync.WaitGroup

		for _, filter = range cfg.NRMetricFilters {
			log.Debugf("Scraping filter %v for app %v", filter, appID)

			wg.Add(1)

			go func(filter string) error {
				defer wg.Done()

				params := url.Values{}
				params.Add("name", filter)

				body, err := api.req(path, params.Encode())
				if err != nil {
					log.Error("Error getting metric names:", err)
					return err
				}

				v, err := jason.NewObjectFromBytes(body)
				if err != nil {
					log.Error("Error parsing metric names from JSON:", err)
					return err
				}

				metricsArray, err := v.GetObjectArray("metrics")
				if err != nil {
					log.Error("Error parsing metric names from JSON:", err)
					return err
				}

				for _, mn := range metricsArray {
					metric := new(MetricName)

					mnBytes, err := mn.Marshal()
					if err != nil {
						log.Error("Error marshalling metric to JSON object:", err)
						return err
					}

					err = json.Unmarshal(mnBytes, metric)
					if err != nil {
						log.Error("Error unmarshalling metric from JSON object:", err)
						return err
					}

					ch <- *metric
				}

				log.Debugf("Found %v possible metric names for app %v and filter %v", len(metricsArray), appID, filter)

				return nil
			}(filter)
		}

		wg.Wait() // wait for all goroutines to finish
		close(ch)
	}(channel)

	// receiving
	for mn := range channel {
		metricNames = append(metricNames, mn)
	}

	return metricNames, nil
}

func (api *API) GetMetricData(appId int, names []MetricName, from time.Time, to time.Time) ([]MetricData, error) {
	path := fmt.Sprintf("/v2/%s/%s/metrics/data.json", api.service, strconv.Itoa(appId))

	var valueNamesList []string

	// If Values Filter is set in config we will use it. Otherwise - gather all possible value names from metric names
	if len(cfg.NRValueFilters) == 0 {
		valueNamesSet := make(map[string]struct{})

		for _, name := range names {
			for _, v := range name.ValueNames {
				valueNamesSet[v] = struct{}{}
			}
		}

		for k := range valueNamesSet {
			valueNamesList = append(valueNamesList, k)
		}
	} else {
		valueNamesList = append(valueNamesList, cfg.NRValueFilters...)
	}

	// Because the Go client does not yet support 100-continue
	// ( see issue #3665 ),
	// we have to process this in chunks, to ensure the response
	// fits within api single request.

	channel := make(chan MetricData)
	metricDatas := make([]MetricData, 0)

	go func(ch chan MetricData) {
		var wg sync.WaitGroup

		for i := 0; i < len(names); i += ChunkSize {
			var thisList []MetricName

			if i+ChunkSize > len(names) {
				thisList = names[i:]
			} else {
				thisList = names[i : i+ChunkSize]
			}

			wg.Add(1)

			go func(names []MetricName) error {
				defer wg.Done()

				params := url.Values{}

				for _, thisName := range names {
					params.Add("names[]", thisName.Name)
				}

				for _, valueFilter := range valueNamesList {
					params.Add("values[]", valueFilter)
				}

				params.Add("raw", "true")
				params.Add("summarize", "true")
				params.Add("period", strconv.Itoa(api.Period))
				params.Add("from", from.Format(time.RFC3339))
				params.Add("to", to.Format(time.RFC3339))

				body, err := api.req(path, params.Encode())
				if err != nil {
					log.Error("Error requesting metrics: ", err)
					return err
				}

				v, err := jason.NewObjectFromBytes(body)
				if err != nil {
					log.Error("Error parsing metric names from JSON:", err)
					return err
				}

				metricsData, err := v.GetObject("metric_data")
				metricsArray, err := metricsData.GetObjectArray("metrics")
				if err != nil {
					log.Error("Error parsing metric names from JSON:", err)
					return err
				}

				for _, md := range metricsArray {
					metric := new(MetricData)

					mdBytes, err := md.Marshal()
					if err != nil {
						log.Error("Error marshalling metric to JSON object:", err)
						return err
					}

					err = json.Unmarshal(mdBytes, metric)
					if err != nil {
						log.Error("Error unmarshalling metric from JSON object:", err)
						return err
					}
					ch <- *metric
				}

				return nil
			}(thisList)
		}

		wg.Wait() // wait for all goroutines to finish
		close(ch)
	}(channel)

	// receiving
	datasCollected := 0
	for mn := range channel {
		metricDatas = append(metricDatas, mn)

		datasCollected++
	}

	return metricDatas, nil
}

func (api *API) req(path string, params string) ([]byte, error) {
	u, err := url.Parse(api.server.String() + path)
	if err != nil {
		return nil, err
	}
	u.RawQuery = params

	//log.Debug("Making API call: ", u.String())

	req := &http.Request{
		Method: "GET",
		URL:    u,
		Header: http.Header{
			"User-Agent": {UserAgent},
			"X-Api-Key":  {api.apiKey},
		},
	}

	var data []byte

	return api.httpget(req, data)
}

func (api *API) httpget(req *http.Request, in []byte) (out []byte, err error) {
	resp, err := api.client.Do(req)
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if resp.StatusCode == 429 {
		log.Info("API Limit Exceeded, New Relic Returning 429 see: https://docs.newrelic.com/docs/apis/rest-api-v2/requirements/api-overload-protection-handling-429-errors")
		overload_header := resp.Header["Newrelic-Overloadprotection-Reset"][0]
		overload_time, _ := strconv.ParseInt(overload_header, 10, 64)
		log.Info("Overload protection resets at: ", time.Unix(overload_time, 0))
		return
	}

	resp.Body.Close()
	out = append(in, body...)

	// Read the link header to see if we need to read more pages.
	links := linkheader.Parse(resp.Header.Get("Link"))

	relLast := links.FilterByRel("last")
	if len(relLast) > 0 {
		u, err := url.Parse(relLast[0].URL)
		if err != nil {
			log.Errorf("Error parsing 'last' relation link. %v", err)
		}
		log.Debugf("Found %v pages for %s", u.Query().Get("page"), req.URL)
	}

	relNext := links.FilterByRel("next")
	if len(relNext) > 0 {
		u := new(url.URL)

		u, err = url.Parse(relNext[0].URL)
		if err != nil {
			return
		}
		cursor := u.Query().Get("cursor")

		if cursor != "" {
			query := req.URL.Query()

			query.Set("cursor", cursor)
			query.Del("page")

			req.URL.RawQuery = query.Encode()

			return api.httpget(req, out)
		}
	}

	return
}
