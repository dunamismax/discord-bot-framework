#!/bin/bash

# Discord Bot Framework Deployment Script
# Usage: ./scripts/deploy.sh [environment] [action]
# Example: ./scripts/deploy.sh production deploy

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
REGISTRY="ghcr.io"
IMAGE_NAME="sawyer/discord-bot-framework"

# Default values
ENVIRONMENT="${1:-development}"
ACTION="${2:-deploy}"
VERSION="${VERSION:-latest}"
KUBECTL_TIMEOUT="${KUBECTL_TIMEOUT:-600s}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Validation functions
validate_environment() {
    local env="$1"
    case "$env" in
        development|staging|production)
            return 0
            ;;
        *)
            log_error "Invalid environment: $env"
            log_error "Valid environments: development, staging, production"
            return 1
            ;;
    esac
}

validate_kubectl() {
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed or not in PATH"
        return 1
    fi

    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster"
        return 1
    fi
}

validate_kustomize() {
    if ! command -v kustomize &> /dev/null; then
        log_warn "kustomize not found, using kubectl kustomize"
        return 1
    fi
    return 0
}

# Pre-deployment checks
pre_deployment_checks() {
    log_info "Running pre-deployment checks..."
    
    validate_environment "$ENVIRONMENT"
    validate_kubectl
    
    # Check if overlay exists
    local overlay_path="$PROJECT_ROOT/k8s/overlays/$ENVIRONMENT"
    if [[ ! -d "$overlay_path" ]]; then
        log_error "Overlay directory not found: $overlay_path"
        exit 1
    fi

    # Validate Kubernetes manifests
    log_info "Validating Kubernetes manifests..."
    if validate_kustomize; then
        kustomize build "$overlay_path" | kubectl apply --dry-run=client -f -
    else
        kubectl kustomize "$overlay_path" | kubectl apply --dry-run=client -f -
    fi

    log_success "Pre-deployment checks passed"
}

# Deploy function
deploy() {
    log_info "Starting deployment to $ENVIRONMENT environment"
    
    pre_deployment_checks
    
    local overlay_path="$PROJECT_ROOT/k8s/overlays/$ENVIRONMENT"
    local namespace="discord-bots"
    
    # Set namespace based on environment
    case "$ENVIRONMENT" in
        production)
            namespace="discord-bots-prod"
            ;;
        staging)
            namespace="discord-bots-staging"
            ;;
        development)
            namespace="discord-bots-dev"
            ;;
    esac

    log_info "Deploying to namespace: $namespace"

    # Create namespace if it doesn't exist
    kubectl create namespace "$namespace" --dry-run=client -o yaml | kubectl apply -f -

    # Apply the configuration
    if validate_kustomize; then
        kustomize build "$overlay_path" | kubectl apply -f -
    else
        kubectl apply -k "$overlay_path"
    fi

    # Wait for deployment to be ready
    log_info "Waiting for deployment to be ready..."
    kubectl rollout status deployment/discord-bot-framework \
        -n "$namespace" \
        --timeout="$KUBECTL_TIMEOUT"

    # Verify deployment
    verify_deployment "$namespace"
    
    log_success "Deployment completed successfully"
}

# Verify deployment
verify_deployment() {
    local namespace="$1"
    
    log_info "Verifying deployment in namespace: $namespace"
    
    # Check pod status
    local ready_pods
    ready_pods=$(kubectl get pods -n "$namespace" -l app=discord-bot-framework -o jsonpath='{.items[*].status.conditions[?(@.type=="Ready")].status}' | tr ' ' '\n' | grep -c "True" || true)
    
    local total_pods
    total_pods=$(kubectl get pods -n "$namespace" -l app=discord-bot-framework --no-headers | wc -l)
    
    log_info "Ready pods: $ready_pods/$total_pods"
    
    if [[ "$ready_pods" -eq 0 ]]; then
        log_error "No pods are ready!"
        kubectl get pods -n "$namespace" -l app=discord-bot-framework
        kubectl describe pods -n "$namespace" -l app=discord-bot-framework
        return 1
    fi

    # Check service endpoints
    local service_endpoints
    service_endpoints=$(kubectl get endpoints discord-bot-framework -n "$namespace" -o jsonpath='{.subsets[*].addresses[*].ip}' | wc -w || true)
    
    log_info "Service endpoints: $service_endpoints"
    
    if [[ "$service_endpoints" -eq 0 ]]; then
        log_warn "No service endpoints found"
    fi

    # Test health endpoint if possible
    test_health_endpoint "$namespace"
    
    log_success "Deployment verification completed"
}

# Test health endpoint
test_health_endpoint() {
    local namespace="$1"
    
    log_info "Testing health endpoint..."
    
    # Port forward to test health endpoint
    local pod_name
    pod_name=$(kubectl get pods -n "$namespace" -l app=discord-bot-framework -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
    
    if [[ -z "$pod_name" ]]; then
        log_warn "No pods found for health check"
        return 1
    fi

    log_info "Testing health on pod: $pod_name"
    
    # Test with timeout
    if timeout 10s kubectl exec -n "$namespace" "$pod_name" -- curl -f http://localhost:8081/health > /dev/null 2>&1; then
        log_success "Health check passed"
    else
        log_warn "Health check failed or timed out"
        return 1
    fi
}

# Rollback function
rollback() {
    local namespace="discord-bots"
    
    case "$ENVIRONMENT" in
        production)
            namespace="discord-bots-prod"
            ;;
        staging)
            namespace="discord-bots-staging"
            ;;
        development)
            namespace="discord-bots-dev"
            ;;
    esac

    log_info "Rolling back deployment in $ENVIRONMENT environment"
    
    kubectl rollout undo deployment/discord-bot-framework -n "$namespace"
    kubectl rollout status deployment/discord-bot-framework -n "$namespace" --timeout="$KUBECTL_TIMEOUT"
    
    log_success "Rollback completed"
}

# Scale function
scale() {
    local replicas="${3:-1}"
    local namespace="discord-bots"
    
    case "$ENVIRONMENT" in
        production)
            namespace="discord-bots-prod"
            ;;
        staging)
            namespace="discord-bots-staging"
            ;;
        development)
            namespace="discord-bots-dev"
            ;;
    esac

    log_info "Scaling deployment to $replicas replicas in $ENVIRONMENT environment"
    
    kubectl scale deployment discord-bot-framework --replicas="$replicas" -n "$namespace"
    kubectl rollout status deployment/discord-bot-framework -n "$namespace" --timeout="$KUBECTL_TIMEOUT"
    
    log_success "Scaling completed"
}

# Status function
status() {
    local namespace="discord-bots"
    
    case "$ENVIRONMENT" in
        production)
            namespace="discord-bots-prod"
            ;;
        staging)
            namespace="discord-bots-staging"
            ;;
        development)
            namespace="discord-bots-dev"
            ;;
    esac

    log_info "Checking status in $ENVIRONMENT environment (namespace: $namespace)"
    
    echo "=== Deployment Status ==="
    kubectl get deployment discord-bot-framework -n "$namespace" -o wide
    
    echo -e "\n=== Pod Status ==="
    kubectl get pods -n "$namespace" -l app=discord-bot-framework
    
    echo -e "\n=== Service Status ==="
    kubectl get service discord-bot-framework -n "$namespace"
    
    echo -e "\n=== Recent Events ==="
    kubectl get events -n "$namespace" --sort-by=.metadata.creationTimestamp | tail -10
}

# Logs function
logs() {
    local namespace="discord-bots"
    
    case "$ENVIRONMENT" in
        production)
            namespace="discord-bots-prod"
            ;;
        staging)
            namespace="discord-bots-staging"
            ;;
        development)
            namespace="discord-bots-dev"
            ;;
    esac

    log_info "Fetching logs from $ENVIRONMENT environment"
    
    kubectl logs -n "$namespace" -l app=discord-bot-framework --tail=100 -f
}

# Delete function
delete() {
    local namespace="discord-bots"
    
    case "$ENVIRONMENT" in
        production)
            namespace="discord-bots-prod"
            ;;
        staging)
            namespace="discord-bots-staging"
            ;;
        development)
            namespace="discord-bots-dev"
            ;;
    esac

    log_warn "Deleting deployment from $ENVIRONMENT environment"
    read -p "Are you sure you want to delete the deployment? (y/N): " -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Deletion cancelled"
        exit 0
    fi
    
    local overlay_path="$PROJECT_ROOT/k8s/overlays/$ENVIRONMENT"
    if validate_kustomize; then
        kustomize build "$overlay_path" | kubectl delete -f -
    else
        kubectl delete -k "$overlay_path"
    fi
    
    log_success "Deployment deleted"
}

# Main script logic
main() {
    case "$ACTION" in
        deploy)
            deploy
            ;;
        rollback)
            rollback
            ;;
        scale)
            scale
            ;;
        status)
            status
            ;;
        logs)
            logs
            ;;
        delete)
            delete
            ;;
        *)
            log_error "Invalid action: $ACTION"
            echo "Valid actions: deploy, rollback, scale, status, logs, delete"
            exit 1
            ;;
    esac
}

# Show usage if no arguments provided
if [[ $# -eq 0 ]]; then
    echo "Discord Bot Framework Deployment Script"
    echo ""
    echo "Usage: $0 [environment] [action] [options]"
    echo ""
    echo "Environments:"
    echo "  development  Deploy to development environment"
    echo "  staging      Deploy to staging environment"
    echo "  production   Deploy to production environment"
    echo ""
    echo "Actions:"
    echo "  deploy       Deploy the application (default)"
    echo "  rollback     Rollback to previous version"
    echo "  scale        Scale the deployment (requires replica count)"
    echo "  status       Show deployment status"
    echo "  logs         Show application logs"
    echo "  delete       Delete the deployment"
    echo ""
    echo "Examples:"
    echo "  $0 production deploy"
    echo "  $0 staging rollback"
    echo "  $0 development scale 2"
    echo "  $0 production status"
    echo ""
    echo "Environment variables:"
    echo "  VERSION            Image version to deploy (default: latest)"
    echo "  KUBECTL_TIMEOUT    Kubectl timeout (default: 600s)"
    exit 1
fi

# Run main function
main "$@"