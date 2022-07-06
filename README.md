# New Relic Exporter

Prometheus exporter for New Relic data.
Requires a New Relic account.
## Building and running

### Running in a container

    cp newrelic_exporter.yml.example newrelic_exporter.yml
	docker run -p 9126:9126 -v $PWD/newrelic_exporter.yml:/app/newrelic_exporter.yml klinux/newrelic-exporter

### From source

	git clone https://github.com/klinux/newrelic_exporter.git --branch release
	cd newrelic_exporter
    make
    cp newrelic_exporter.yml.example newrelic_exporter.yml
    ./newrelic_exporter

## Flags

Name               | Description
-------------------|------------
config             | Config file path. Defaults to `newrelic_exporter.yml` in current directory.

## Available Configuration Values

Name                        | Description
----------------------------|------------
api.key                     | API key
api.server                  | API location.  Defaults to https://api.newrelic.com
api.period                  | Period of data to request, in seconds.  Defaults to 60.
api.timeout                 | Period of time to wait for an API response in seconds (default 5s)
api.apps-list-cache-time    | Length of time to cache list of available applications
api.metric-names-cache-time | Length of time to cache names of metrics (not values)
api.service                 | Define section of API to limit requests to (applications, mobile, etc)
api.include-apps            | List of applications to query (optional). []
api.use-only-summary        | Use only summary metrics, default false
api.include-metric-filters  | List of metric groups to filter by to reduce number of API calls (optional). ["WebTransactionTotalTime", "Errors/allOther"]
api.include-values          | List of values to filter by to reduce number of API calls (optional)
web.listen-address          | Address to listen on for web interface and telemetry.  Port defaults to 9126.
web.telemetry-path          | Path under which to expose metrics.
debug.proxy-address         | Proxy settings for debugging
