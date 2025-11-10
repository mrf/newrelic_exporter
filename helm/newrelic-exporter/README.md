# New Relic Exporter Helm Chart

This Helm chart deploys the New Relic Prometheus Exporter to a Kubernetes cluster.

## Prerequisites

- Kubernetes 1.16+
- Helm 3.0+
- A New Relic account with API access

## Installing the Chart

To install the chart with the release name `my-newrelic-exporter`:

```bash
helm install my-newrelic-exporter ./helm/newrelic-exporter \
  --set newrelic.apiKey=YOUR_NEWRELIC_API_KEY \
  --set newrelic.service=applications \
  --set newrelic.includeMetricFilters[0]="HttpDispatcher"
```

## Uninstalling the Chart

To uninstall/delete the `my-newrelic-exporter` deployment:

```bash
helm uninstall my-newrelic-exporter
```

## Configuration

The following table lists the configurable parameters of the New Relic Exporter chart and their default values.

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `1` |
| `image.repository` | Image repository | `ghcr.io/mrf/newrelic-exporter` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `image.tag` | Image tag | `""` (uses appVersion) |
| `serviceAccount.create` | Create service account | `true` |
| `serviceAccount.name` | Service account name | `""` |
| `service.type` | Kubernetes service type | `ClusterIP` |
| `service.port` | Service port | `9126` |
| `newrelic.apiKey` | New Relic API key (required) | `""` |
| `newrelic.existingSecret` | Use existing secret for API key | `""` |
| `newrelic.server` | New Relic API server URL | `https://api.newrelic.com` |
| `newrelic.period` | Data request period in seconds | `60` |
| `newrelic.timeout` | API timeout | `15s` |
| `newrelic.service` | New Relic service type | `""` |
| `newrelic.includeApps` | List of applications to query | `[]` |
| `newrelic.includeMetricFilters` | Metric filters (required) | `[]` |
| `newrelic.includeValues` | Value filters | `[]` |
| `web.listenAddress` | Listen address | `:9126` |
| `web.telemetryPath` | Metrics endpoint path | `/metrics` |
| `resources` | CPU/Memory resource requests/limits | `{}` |

## Example Configuration

### Basic Installation

```bash
helm install newrelic-exporter ./helm/newrelic-exporter \
  --set newrelic.apiKey=YOUR_API_KEY \
  --set newrelic.service=applications \
  --set newrelic.includeMetricFilters[0]="HttpDispatcher"
```

### Using an Existing Secret

Create a secret with your API key:

```bash
kubectl create secret generic newrelic-api-key --from-literal=api-key=YOUR_API_KEY
```

Install the chart using the existing secret:

```bash
helm install newrelic-exporter ./helm/newrelic-exporter \
  --set newrelic.existingSecret=newrelic-api-key \
  --set newrelic.service=applications \
  --set newrelic.includeMetricFilters[0]="HttpDispatcher"
```

### Advanced Configuration with values.yaml

Create a `custom-values.yaml` file:

```yaml
replicaCount: 2

newrelic:
  apiKey: YOUR_API_KEY
  service: applications
  period: 60
  includeApps:
    - "My Application"
  includeMetricFilters:
    - "HttpDispatcher"
    - "Database"
  includeValues:
    - "average_response_time"
    - "throughput"

resources:
  limits:
    cpu: 200m
    memory: 256Mi
  requests:
    cpu: 100m
    memory: 128Mi
```

Install with custom values:

```bash
helm install newrelic-exporter ./helm/newrelic-exporter -f custom-values.yaml
```

## Prometheus Integration

The exporter automatically adds Prometheus scrape annotations to the pod. If you're using Prometheus Operator, you can create a ServiceMonitor:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: newrelic-exporter
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: newrelic-exporter
  endpoints:
  - port: metrics
    interval: 30s
```

## Troubleshooting

### Check if the exporter is running

```bash
kubectl get pods -l app.kubernetes.io/name=newrelic-exporter
```

### View exporter logs

```bash
kubectl logs -l app.kubernetes.io/name=newrelic-exporter
```

### Test metrics endpoint

```bash
kubectl port-forward svc/newrelic-exporter 9126:9126
curl http://localhost:9126/metrics
```

## Support

For issues and questions, please visit: https://github.com/mrf/newrelic_exporter/issues
