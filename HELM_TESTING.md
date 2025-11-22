# Helm Chart Testing Guide

This document describes how to test the New Relic Exporter Helm chart using Kind (Kubernetes in Docker).

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Testing Locally](#testing-locally)
- [CI/CD Testing](#cicd-testing)
- [Test Configuration](#test-configuration)
- [Troubleshooting](#troubleshooting)

## Overview

The Helm chart testing uses Kind to create an isolated Kubernetes cluster where the chart is installed and validated. This ensures:

- ✅ Chart syntax is valid
- ✅ All Kubernetes resources are created correctly
- ✅ Pods start and run successfully
- ✅ The exporter is accessible via the service
- ✅ Metrics endpoint responds correctly
- ✅ Chart upgrades work as expected

## Prerequisites

### Required Tools

1. **Docker** - Container runtime
   ```bash
   # Check if Docker is installed
   docker --version

   # Install Docker: https://docs.docker.com/get-docker/
   ```

2. **Kind** - Kubernetes in Docker
   ```bash
   # Check if Kind is installed
   kind version

   # Install Kind
   # Linux/MacOS
   curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
   chmod +x ./kind
   sudo mv ./kind /usr/local/bin/kind

   # Or use package manager
   # MacOS
   brew install kind

   # Windows (using Chocolatey)
   choco install kind
   ```

3. **kubectl** - Kubernetes CLI
   ```bash
   # Check if kubectl is installed
   kubectl version --client

   # Install kubectl: https://kubernetes.io/docs/tasks/tools/
   ```

4. **Helm** - Kubernetes package manager
   ```bash
   # Check if Helm is installed
   helm version

   # Install Helm
   curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

   # Or use package manager
   # MacOS
   brew install helm

   # Windows (using Chocolatey)
   choco install kubernetes-helm
   ```

## Quick Start

### Automated Testing (Recommended)

The easiest way to test the Helm chart is using the automated test script:

```bash
# From the project root directory
./scripts/test-helm-chart.sh
```

This script will:
1. Check prerequisites
2. Create a Kind cluster
3. Validate the Helm chart syntax
4. Install the chart
5. Wait for all resources to be ready
6. Validate all Kubernetes resources
7. Test connectivity to the exporter
8. Show cluster information
9. Clean up (delete cluster)

### Keep Cluster Running

To keep the cluster running after tests (for debugging):

```bash
CLEANUP=false ./scripts/test-helm-chart.sh
```

Then you can interact with the cluster:

```bash
# Switch to the Kind cluster context
kubectl config use-context kind-newrelic-exporter-test

# View pods
kubectl get pods

# View logs
kubectl logs -l app.kubernetes.io/name=newrelic-exporter

# Test metrics endpoint
POD_NAME=$(kubectl get pods -l "app.kubernetes.io/name=newrelic-exporter" -o jsonpath='{.items[0].metadata.name}')
kubectl exec $POD_NAME -- wget -q -O- http://localhost:9126/metrics

# Clean up when done
kind delete cluster --name newrelic-exporter-test
```

## Testing Locally

### Manual Step-by-Step Testing

If you prefer to run tests manually:

#### 1. Create Kind Cluster

```bash
# Create cluster with custom configuration
kind create cluster --name newrelic-exporter-test --config .kind/kind-config.yaml

# Or create with default configuration
kind create cluster --name newrelic-exporter-test

# Verify cluster is ready
kubectl cluster-info
kubectl get nodes
```

#### 2. Validate Helm Chart

```bash
# Lint the chart
helm lint ./helm/newrelic-exporter

# Lint with test values
helm lint ./helm/newrelic-exporter --values ./helm/newrelic-exporter/values-test.yaml

# Template the chart (dry-run)
helm template newrelic-exporter ./helm/newrelic-exporter \
  --values ./helm/newrelic-exporter/values-test.yaml \
  --namespace default
```

#### 3. Install Helm Chart

```bash
# Install the chart
helm install newrelic-exporter-test ./helm/newrelic-exporter \
  --values ./helm/newrelic-exporter/values-test.yaml \
  --namespace default \
  --create-namespace \
  --wait \
  --timeout 300s

# Check installation status
helm list -n default
helm status newrelic-exporter-test -n default
```

#### 4. Verify Resources

```bash
# Check all resources
kubectl get all -n default -l "app.kubernetes.io/name=newrelic-exporter"

# Check specific resources
kubectl get deployment -n default
kubectl get pods -n default
kubectl get service -n default
kubectl get configmap -n default
kubectl get secret -n default

# Describe deployment
kubectl describe deployment -n default -l "app.kubernetes.io/name=newrelic-exporter"
```

#### 5. Test Pod Functionality

```bash
# Get pod name
POD_NAME=$(kubectl get pods -n default -l "app.kubernetes.io/name=newrelic-exporter" -o jsonpath='{.items[0].metadata.name}')

# View pod logs
kubectl logs -n default $POD_NAME

# Test metrics endpoint
kubectl exec -n default $POD_NAME -- wget -q -O- http://localhost:9126/metrics

# Port-forward to access locally (in separate terminal)
kubectl port-forward -n default $POD_NAME 9126:9126

# Then access from your local machine
curl http://localhost:9126/metrics
```

#### 6. Test Chart Upgrade

```bash
# Modify values and upgrade
helm upgrade newrelic-exporter-test ./helm/newrelic-exporter \
  --values ./helm/newrelic-exporter/values-test.yaml \
  --set replicaCount=2 \
  --namespace default \
  --wait

# Verify upgrade
kubectl get pods -n default -l "app.kubernetes.io/name=newrelic-exporter"
helm history newrelic-exporter-test -n default
```

#### 7. Cleanup

```bash
# Uninstall Helm release
helm uninstall newrelic-exporter-test -n default

# Delete Kind cluster
kind delete cluster --name newrelic-exporter-test
```

## CI/CD Testing

### GitHub Actions Workflow

The project includes a comprehensive GitHub Actions workflow (`.github/workflows/helm-test.yml`) that automatically tests the Helm chart on every push and pull request.

The workflow includes:

1. **Lint Chart** - Validates chart syntax and templates
2. **Test with Kind** - Full integration test in a Kind cluster
3. **Multiple Configurations** - Tests different value combinations
4. **Upgrade Testing** - Validates chart upgrade scenarios

### Triggering CI Tests

Tests run automatically on:
- Push to `main` or `master` branches (when Helm files change)
- Pull requests to `main` or `master` branches (when Helm files change)

Manual trigger:
```bash
# Push changes to trigger workflow
git add .
git commit -m "Test Helm chart changes"
git push origin your-branch
```

### Viewing CI Results

1. Go to your repository on GitHub
2. Click on "Actions" tab
3. Select the "Helm Chart Testing" workflow
4. View the results of each job

## Test Configuration

### Kind Cluster Configuration

The Kind cluster configuration is in `.kind/kind-config.yaml`:

```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: newrelic-exporter-test
nodes:
  - role: control-plane
    extraPortMappings:
      - containerPort: 30926
        hostPort: 9126
        protocol: TCP
```

This configuration:
- Creates a single-node cluster
- Maps the exporter NodePort (30926) to host port 9126

### Test Values

Test values are in `helm/newrelic-exporter/values-test.yaml`:

```yaml
replicaCount: 1

image:
  repository: ghcr.io/mrf/newrelic-exporter
  pullPolicy: IfNotPresent
  tag: "latest"

service:
  type: NodePort
  port: 9126
  nodePort: 30926

newrelic:
  apiKey: "test-api-key-for-validation-only"
  service: "applications"
  timeout: "15s"
  includeMetricFilters:
    - "HttpDispatcher"
    - "Apdex"

resources:
  limits:
    cpu: 200m
    memory: 256Mi
  requests:
    cpu: 100m
    memory: 128Mi
```

### Customizing Test Configuration

To test with different values:

```bash
# Create custom values file
cat > my-test-values.yaml <<EOF
replicaCount: 2
newrelic:
  apiKey: "my-test-key"
  service: "applications"
  timeout: "30s"
  includeMetricFilters:
    - "HttpDispatcher"
    - "Database"
    - "External"
EOF

# Use custom values with test script
HELM_VALUES=my-test-values.yaml ./scripts/test-helm-chart.sh

# Or manually
helm install test ./helm/newrelic-exporter --values my-test-values.yaml
```

## Troubleshooting

### Common Issues

#### 1. Kind Cluster Creation Fails

```bash
# Error: failed to create cluster

# Solution: Check Docker is running
docker ps

# Try deleting existing cluster
kind delete cluster --name newrelic-exporter-test

# Recreate cluster
kind create cluster --name newrelic-exporter-test
```

#### 2. Pods Not Starting

```bash
# Check pod status
kubectl get pods -n default

# Describe pod to see events
kubectl describe pod -n default <pod-name>

# Check pod logs
kubectl logs -n default <pod-name>

# Common causes:
# - Image pull errors (check image name/tag)
# - Resource constraints (check node resources)
# - Configuration errors (check configmap/secret)
```

#### 3. Image Pull Errors

```bash
# Error: ImagePullBackOff

# Load local image into Kind cluster
docker pull ghcr.io/mrf/newrelic-exporter:latest
kind load docker-image ghcr.io/mrf/newrelic-exporter:latest --name newrelic-exporter-test

# Or build and load local image
docker build -t ghcr.io/mrf/newrelic-exporter:latest .
kind load docker-image ghcr.io/mrf/newrelic-exporter:latest --name newrelic-exporter-test
```

#### 4. Metrics Endpoint Not Responding

```bash
# Check if pod is running
kubectl get pods -n default

# Check pod logs for errors
kubectl logs -n default <pod-name>

# Test from within pod
kubectl exec -n default <pod-name> -- wget -q -O- http://localhost:9126/metrics

# Check service
kubectl get service -n default
kubectl describe service -n default <service-name>
```

#### 5. Test Script Fails

```bash
# Make script executable
chmod +x scripts/test-helm-chart.sh

# Run with debug output
bash -x scripts/test-helm-chart.sh

# Check prerequisites
./scripts/test-helm-chart.sh
# Script will show which tools are missing
```

### Debugging Tips

#### View All Resources

```bash
kubectl get all -n default -l "app.kubernetes.io/name=newrelic-exporter"
```

#### Check Events

```bash
kubectl get events -n default --sort-by='.lastTimestamp'
```

#### Interactive Pod Shell

```bash
POD_NAME=$(kubectl get pods -n default -l "app.kubernetes.io/name=newrelic-exporter" -o jsonpath='{.items[0].metadata.name}')
kubectl exec -it -n default $POD_NAME -- /bin/sh
```

#### Port Forward for Local Access

```bash
kubectl port-forward -n default service/newrelic-exporter-test 9126:9126

# Then access from browser or curl
curl http://localhost:9126/metrics
```

#### Helm Debug Output

```bash
helm install test ./helm/newrelic-exporter \
  --values ./helm/newrelic-exporter/values-test.yaml \
  --debug \
  --dry-run
```

### Getting Help

If you encounter issues:

1. Check this troubleshooting section
2. Review pod logs: `kubectl logs -n default <pod-name>`
3. Review pod events: `kubectl describe pod -n default <pod-name>`
4. Check GitHub Issues: https://github.com/mrf/newrelic_exporter/issues
5. Open a new issue with:
   - Steps to reproduce
   - Error messages
   - Relevant logs
   - Environment details (OS, Docker/Kind/Helm versions)

## Best Practices

### Before Committing Changes

1. **Lint the chart**
   ```bash
   helm lint ./helm/newrelic-exporter
   ```

2. **Test locally with Kind**
   ```bash
   ./scripts/test-helm-chart.sh
   ```

3. **Verify templates render correctly**
   ```bash
   helm template test ./helm/newrelic-exporter --values ./helm/newrelic-exporter/values-test.yaml
   ```

### Continuous Validation

- Run tests before pushing changes
- Review CI test results before merging PRs
- Test with multiple value configurations
- Verify upgrade scenarios work

### Documentation

- Update this guide when adding new tests
- Document any new test values or configurations
- Keep examples up to date with chart changes

## Additional Resources

- [Kind Documentation](https://kind.sigs.k8s.io/)
- [Helm Documentation](https://helm.sh/docs/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Helm Chart Testing](https://github.com/helm/chart-testing)
- [Project README](README.md)
- [Testing Guide](TESTING.md)

## Questions?

If you have questions about Helm chart testing:
- Check this guide
- Review the automated test script: `scripts/test-helm-chart.sh`
- Review the GitHub Actions workflow: `.github/workflows/helm-test.yml`
- Ask in GitHub Issues or Discussions
