# NewRelic Exporter Helm Installation
This Helm Chart will deploy the NewRelic Exorter.

Value                       | Description                                      | Default
----------------------------|--------------------------------------------------|-------------------
imagePullSecrets            | Image Pull Secret                                | ""
containerName               | The name of the container                        | newrelic-exporter
image.registry              | The image registry: default docker hub           | klinux
image.image                 | The image name                                   | newrelic-exporter
image.tag                   | The image tag                                    | 1.0.0
image.pullPolicy            | The image policy                                 | IfNotPresent
service.port                | The service and deployment port                  | 9126
nr-api-key                  | NewRelic API Key                                 | ""
nr-server                   | API location.                                    | https://api.newrelic.com
nr-period                   | Period of data to request, in seconds            | Defaults to 60.
nr-timeout                  | Period of time to wait for an API response in seconds | 5s
nr-apps-list-cache-time     | Length of time to cache list of available applications | 1m
nr-metric-names-cache-time  | Length of time to cache names of metrics (not values) | 0
nr-service                  | Define section of API to limit requests to (applications, mobile, etc) | applications
nr-include-apps             | List of applications to query (optional) | ""
nr-include-metric-filters   | List of metric groups to filter by to reduce number of API calls (optional) | ""
nr-include-values           | List of values to filter by to reduce number of API calls (optional) | ""
nr-telemetry-path           | Path under which to expose metrics               | /metrics
nr-debug                    | If you want to start debug server                | false
nr-debug.proxy-address      | Proxy settings for debugging                     | https://localhost:8888
extraEnvs                   | Extra environments                               | []
labels                      | Deployment labels                                | {}
commonLabels                | Deployment commons labels                        | {}
podLabels                   | Pod labels                                       | {}
annotations                 | Deployment annotations                           | {}
resources                   | Deployment resources (cpu and memory)            | limits: 100m, 128Mi, resources: 50m, 64Mi
tolerations                 | Deployment tolerations                           | []
nodeSelector                | Deployment nodeselector                          | {}
