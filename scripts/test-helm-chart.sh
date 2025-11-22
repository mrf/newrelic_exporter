#!/bin/bash

# Helm Chart Testing Script for New Relic Exporter
# This script creates a Kind cluster, installs the Helm chart, validates it, and cleans up

set -e  # Exit on error
set -o pipefail  # Catch errors in pipes

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CLUSTER_NAME="${CLUSTER_NAME:-newrelic-exporter-test}"
NAMESPACE="${NAMESPACE:-default}"
RELEASE_NAME="${RELEASE_NAME:-newrelic-exporter-test}"
KIND_CONFIG="${KIND_CONFIG:-.kind/kind-config.yaml}"
HELM_CHART="${HELM_CHART:-./helm/newrelic-exporter}"
HELM_VALUES="${HELM_VALUES:-./helm/newrelic-exporter/values-test.yaml}"
TIMEOUT="${TIMEOUT:-300}"  # 5 minutes timeout

# Cleanup flag
CLEANUP="${CLEANUP:-true}"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

cleanup() {
    if [ "$CLEANUP" = "true" ]; then
        log_info "Cleaning up resources..."

        # Delete Helm release
        if helm list -n "$NAMESPACE" | grep -q "$RELEASE_NAME"; then
            log_info "Uninstalling Helm release: $RELEASE_NAME"
            helm uninstall "$RELEASE_NAME" -n "$NAMESPACE" --wait || true
        fi

        # Delete Kind cluster
        if kind get clusters | grep -q "^${CLUSTER_NAME}$"; then
            log_info "Deleting Kind cluster: $CLUSTER_NAME"
            kind delete cluster --name "$CLUSTER_NAME"
        fi

        log_success "Cleanup completed"
    else
        log_warning "Skipping cleanup (CLEANUP=false)"
        log_info "To manually cleanup, run: kind delete cluster --name $CLUSTER_NAME"
    fi
}

# Trap cleanup on exit
trap cleanup EXIT

check_prerequisites() {
    log_info "Checking prerequisites..."

    local missing_tools=()

    if ! command -v kind &> /dev/null; then
        missing_tools+=("kind")
    fi

    if ! command -v kubectl &> /dev/null; then
        missing_tools+=("kubectl")
    fi

    if ! command -v helm &> /dev/null; then
        missing_tools+=("helm")
    fi

    if ! command -v docker &> /dev/null; then
        missing_tools+=("docker")
    fi

    if [ ${#missing_tools[@]} -ne 0 ]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        log_info "Please install the missing tools:"
        for tool in "${missing_tools[@]}"; do
            case $tool in
                kind)
                    echo "  - kind: https://kind.sigs.k8s.io/docs/user/quick-start/#installation"
                    ;;
                kubectl)
                    echo "  - kubectl: https://kubernetes.io/docs/tasks/tools/"
                    ;;
                helm)
                    echo "  - helm: https://helm.sh/docs/intro/install/"
                    ;;
                docker)
                    echo "  - docker: https://docs.docker.com/get-docker/"
                    ;;
            esac
        done
        exit 1
    fi

    log_success "All prerequisites met"
}

create_kind_cluster() {
    log_info "Creating Kind cluster: $CLUSTER_NAME"

    # Check if cluster already exists
    if kind get clusters | grep -q "^${CLUSTER_NAME}$"; then
        log_warning "Cluster $CLUSTER_NAME already exists, deleting it first..."
        kind delete cluster --name "$CLUSTER_NAME"
    fi

    # Create cluster
    if [ -f "$KIND_CONFIG" ]; then
        log_info "Using Kind config from: $KIND_CONFIG"
        kind create cluster --name "$CLUSTER_NAME" --config "$KIND_CONFIG" --wait 60s
    else
        log_warning "Kind config not found at $KIND_CONFIG, using default configuration"
        kind create cluster --name "$CLUSTER_NAME" --wait 60s
    fi

    # Wait for cluster to be ready
    log_info "Waiting for cluster to be ready..."
    kubectl wait --for=condition=Ready nodes --all --timeout=60s

    log_success "Kind cluster created successfully"
}

validate_helm_chart() {
    log_info "Validating Helm chart syntax..."

    # Lint the chart
    log_info "Running helm lint..."
    if ! helm lint "$HELM_CHART" --values "$HELM_VALUES"; then
        log_error "Helm chart linting failed"
        return 1
    fi

    # Template the chart to check for rendering errors
    log_info "Running helm template..."
    if ! helm template "$RELEASE_NAME" "$HELM_CHART" --values "$HELM_VALUES" --namespace "$NAMESPACE" > /dev/null; then
        log_error "Helm chart templating failed"
        return 1
    fi

    log_success "Helm chart validation passed"
}

install_helm_chart() {
    log_info "Installing Helm chart..."

    helm install "$RELEASE_NAME" "$HELM_CHART" \
        --values "$HELM_VALUES" \
        --namespace "$NAMESPACE" \
        --create-namespace \
        --wait \
        --timeout "${TIMEOUT}s" \
        --debug

    log_success "Helm chart installed successfully"
}

wait_for_deployment() {
    log_info "Waiting for deployment to be ready..."

    local deployment_name
    deployment_name=$(kubectl get deployment -n "$NAMESPACE" -l "app.kubernetes.io/name=newrelic-exporter" -o jsonpath='{.items[0].metadata.name}')

    if [ -z "$deployment_name" ]; then
        log_error "Could not find deployment"
        return 1
    fi

    log_info "Found deployment: $deployment_name"

    kubectl wait --for=condition=Available deployment/"$deployment_name" \
        -n "$NAMESPACE" \
        --timeout="${TIMEOUT}s"

    log_success "Deployment is ready"
}

validate_resources() {
    log_info "Validating Kubernetes resources..."

    # Check deployment
    log_info "Checking deployment..."
    local deployment_count
    deployment_count=$(kubectl get deployment -n "$NAMESPACE" -l "app.kubernetes.io/name=newrelic-exporter" --no-headers | wc -l)
    if [ "$deployment_count" -eq 0 ]; then
        log_error "No deployment found"
        return 1
    fi
    log_success "Deployment found: $deployment_count"

    # Check pods
    log_info "Checking pods..."
    local pod_count
    pod_count=$(kubectl get pods -n "$NAMESPACE" -l "app.kubernetes.io/name=newrelic-exporter" --field-selector=status.phase=Running --no-headers | wc -l)
    if [ "$pod_count" -eq 0 ]; then
        log_error "No running pods found"
        kubectl get pods -n "$NAMESPACE" -l "app.kubernetes.io/name=newrelic-exporter"
        kubectl describe pods -n "$NAMESPACE" -l "app.kubernetes.io/name=newrelic-exporter"
        return 1
    fi
    log_success "Running pods: $pod_count"

    # Check service
    log_info "Checking service..."
    local service_count
    service_count=$(kubectl get service -n "$NAMESPACE" -l "app.kubernetes.io/name=newrelic-exporter" --no-headers | wc -l)
    if [ "$service_count" -eq 0 ]; then
        log_error "No service found"
        return 1
    fi
    log_success "Service found: $service_count"

    # Check service account
    log_info "Checking service account..."
    local sa_count
    sa_count=$(kubectl get serviceaccount -n "$NAMESPACE" -l "app.kubernetes.io/name=newrelic-exporter" --no-headers | wc -l)
    if [ "$sa_count" -eq 0 ]; then
        log_error "No service account found"
        return 1
    fi
    log_success "Service account found: $sa_count"

    # Check configmap
    log_info "Checking configmap..."
    local cm_count
    cm_count=$(kubectl get configmap -n "$NAMESPACE" -l "app.kubernetes.io/name=newrelic-exporter" --no-headers | wc -l)
    if [ "$cm_count" -eq 0 ]; then
        log_error "No configmap found"
        return 1
    fi
    log_success "ConfigMap found: $cm_count"

    # Check secret
    log_info "Checking secret..."
    local secret_count
    secret_count=$(kubectl get secret -n "$NAMESPACE" -l "app.kubernetes.io/name=newrelic-exporter" --no-headers | wc -l)
    if [ "$secret_count" -eq 0 ]; then
        log_error "No secret found"
        return 1
    fi
    log_success "Secret found: $secret_count"

    log_success "All Kubernetes resources validated"
}

test_connectivity() {
    log_info "Testing pod connectivity..."

    # Get pod name
    local pod_name
    pod_name=$(kubectl get pods -n "$NAMESPACE" -l "app.kubernetes.io/name=newrelic-exporter" -o jsonpath='{.items[0].metadata.name}')

    if [ -z "$pod_name" ]; then
        log_error "Could not find pod"
        return 1
    fi

    log_info "Testing pod: $pod_name"

    # Check if metrics endpoint is accessible
    log_info "Testing metrics endpoint..."
    if kubectl exec -n "$NAMESPACE" "$pod_name" -- wget -q -O- http://localhost:9126/metrics > /dev/null; then
        log_success "Metrics endpoint is accessible"
    else
        log_error "Metrics endpoint is not accessible"
        log_info "Pod logs:"
        kubectl logs -n "$NAMESPACE" "$pod_name" --tail=50
        return 1
    fi

    # Check for Prometheus metrics format
    log_info "Validating Prometheus metrics format..."
    local metrics
    metrics=$(kubectl exec -n "$NAMESPACE" "$pod_name" -- wget -q -O- http://localhost:9126/metrics)

    if echo "$metrics" | grep -q "^# HELP"; then
        log_success "Valid Prometheus metrics format detected"
    else
        log_warning "Metrics format may not be standard Prometheus format"
    fi

    # Display some sample metrics
    log_info "Sample metrics:"
    echo "$metrics" | grep -E "^(# HELP|# TYPE|newrelic_)" | head -10

    log_success "Connectivity tests passed"
}

show_cluster_info() {
    log_info "Cluster information:"
    echo ""
    echo "=== Nodes ==="
    kubectl get nodes
    echo ""
    echo "=== Deployments ==="
    kubectl get deployments -n "$NAMESPACE" -l "app.kubernetes.io/name=newrelic-exporter"
    echo ""
    echo "=== Pods ==="
    kubectl get pods -n "$NAMESPACE" -l "app.kubernetes.io/name=newrelic-exporter"
    echo ""
    echo "=== Services ==="
    kubectl get services -n "$NAMESPACE" -l "app.kubernetes.io/name=newrelic-exporter"
    echo ""
}

# Main execution
main() {
    log_info "Starting Helm chart testing for New Relic Exporter"
    log_info "================================================"
    echo ""

    check_prerequisites
    create_kind_cluster
    validate_helm_chart
    install_helm_chart
    wait_for_deployment
    validate_resources
    test_connectivity
    show_cluster_info

    echo ""
    log_success "================================================"
    log_success "All tests passed successfully!"
    log_success "================================================"
    echo ""

    if [ "$CLEANUP" = "true" ]; then
        log_info "Cleanup will run automatically on exit"
    else
        log_info "Cluster is still running. Access it with:"
        log_info "  kubectl config use-context kind-$CLUSTER_NAME"
        log_info "  kubectl get pods -n $NAMESPACE"
        log_info ""
        log_info "To delete the cluster manually:"
        log_info "  kind delete cluster --name $CLUSTER_NAME"
    fi
}

# Run main function
main "$@"
