#!/usr/bin/env bash
set -euo pipefail

# AI Gateway Blue-Green Deployment Script
# Automates zero-downtime deployments using blue-green strategy

# ============================================
# Configuration
# ============================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DEPLOYMENTS_DIR="$PROJECT_ROOT/deployments"
LOG_FILE="$DEPLOYMENTS_DIR/deployments.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Deployment settings
MAX_HEALTH_CHECKS=30
HEALTH_CHECK_INTERVAL=2
BASE_URL="${BASE_URL:-http://localhost:8080}"

# ============================================
# Functions
# ============================================

log() {
    local level="$1"
    shift
    local message="$*"
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    echo "[$timestamp] [$level] $message" | tee -a "$LOG_FILE"
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $*" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*" | tee -a "$LOG_FILE"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $*" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*" | tee -a "$LOG_FILE"
}

print_usage() {
    cat << EOF
Usage: $(basename "$0") <environment> <image_tag>

Deploy AI Gateway using blue-green strategy.

ARGUMENTS:
    environment    Target environment (staging|production)
    image_tag      Docker image tag to deploy

ENVIRONMENT VARIABLES:
    BASE_URL        Base URL for health checks (default: http://localhost:8080)
    MAX_HEALTH_CHECKS  Maximum health check attempts (default: 30)
    HEALTH_CHECK_INTERVAL  Seconds between health checks (default: 2)

EXAMPLES:
    $(basename "$0") staging v1.0.0
    $(basename "$0") production v1.2.3

EOF
    exit 1
}

validate_arguments() {
    if [[ "$#" -ne 2 ]]; then
        log_error "Invalid arguments"
        print_usage
    fi

    ENVIRONMENT="$1"
    IMAGE_TAG="$2"

    if [[ ! "$ENVIRONMENT" =~ ^(staging|production)$ ]]; then
        log_error "Invalid environment: $ENVIRONMENT"
        log_info "Valid environments: staging, production"
        exit 1
    fi

    if [[ -z "$IMAGE_TAG" ]]; then
        log_error "Image tag cannot be empty"
        exit 1
    fi

    log_info "Deployment configuration:"
    log_info "  Environment: $ENVIRONMENT"
    log_info "  Image Tag: $IMAGE_TAG"
    log_info "  Base URL: $BASE_URL"
}

check_prerequisites() {
    log_info "Checking prerequisites..."

    # Check Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed or not in PATH"
        exit 1
    fi

    # Check Docker Compose
    if ! docker compose version &> /dev/null && ! docker-compose version &> /dev/null; then
        log_error "Docker Compose is not installed"
        exit 1
    fi

    # Check if Docker daemon is running
    if ! docker info &> /dev/null; then
        log_error "Docker daemon is not running"
        exit 1
    fi

    # Check if .env file exists
    if [[ ! -f "$DEPLOYMENTS_DIR/docker/.env" ]]; then
        log_warning ".env file not found, using .env.example"
        if [[ ! -f "$DEPLOYMENTS_DIR/docker/.env.example" ]]; then
            log_error "Neither .env nor .env.example found"
            exit 1
        fi
    fi

    log_success "Prerequisites check passed"
}

get_current_color() {
    # Read current deployment color from state file or environment
    local state_file="$DEPLOYMENTS_DIR/.deployment-state"

    if [[ -f "$state_file" ]]; then
        source "$state_file"
        echo "${DEPLOYMENT_COLOR:-blue}"
    else
        echo "blue"
    fi
}

get_new_color() {
    local current="$1"
    if [[ "$current" == "blue" ]]; then
        echo "green"
    else
        echo "blue"
    fi
}

deploy_new_version() {
    local new_color="$1"
    local image_tag="$2"

    log_info "Deploying new version to $new_color environment..."

    cd "$DEPLOYMENTS_DIR/docker"

    # Set image tag for the new color
    export "IMAGE_TAG_${new_color^^}=$image_tag"

    # Start the new deployment
    log_info "Starting $new_color containers..."
    if [[ "$ENVIRONMENT" == "production" ]]; then
        docker compose -f docker-compose.base.yml -f docker-compose.prod.yml \
            up -d "ai-gateway-$new_color" \
            || { log_error "Failed to start $new_color deployment"; return 1; }
    else
        docker compose -f docker-compose.base.yml -f docker-compose.staging.yml \
            up -d "ai-gateway-$new_color" \
            || { log_error "Failed to start $new_color deployment"; return 1; }
    fi

    log_success "Started $new_color deployment"
}

wait_for_health_check() {
    local new_color="$1"
    local port="$2"
    local health_url="http://localhost:$port/health"

    log_info "Waiting for $new_color deployment to be healthy..."
    log_info "Health check URL: $health_url"

    local attempts=0
    local healthy=false

    while [[ $attempts -lt $MAX_HEALTH_CHECKS ]]; do
        if curl -f -s "$health_url" > /dev/null 2>&1; then
            healthy=true
            log_success "Health check passed for $new_color"
            break
        fi

        ((attempts++))
        local progress=$((attempts * 100 / MAX_HEALTH_CHECKS))
        echo -ne "\rProgress: [$progress%] ($attempts/$MAX_HEALTH_CHECKS) "
        sleep "$HEALTH_CHECK_INTERVAL"
    done

    echo ""

    if [[ "$healthy" != "true" ]]; then
        log_error "Health check failed after $MAX_HEALTH_CHECKS attempts"
        log_error "Deployment did not become healthy in time"
        return 1
    fi

    return 0
}

run_smoke_tests() {
    local new_color="$1"
    local port="$2"

    log_info "Running smoke tests against $new_color..."

    local smoke_test_script="$SCRIPT_DIR/smoke-test.sh"

    if [[ ! -f "$smoke_test_script" ]]; then
        log_warning "Smoke test script not found, skipping"
        return 0
    fi

    # Make script executable
    chmod +x "$smoke_test_script"

    # Run smoke tests
    if bash "$smoke_test_script" "http://localhost:$port"; then
        log_success "Smoke tests passed for $new_color"
        return 0
    else
        log_error "Smoke tests failed for $new_color"
        return 1
    fi
}

switch_traffic() {
    local new_color="$1"

    log_info "Switching traffic to $new_color..."

    # Call the switch-traffic script
    if bash "$SCRIPT_DIR/switch-traffic.sh" "$new_color"; then
        log_success "Traffic switched to $new_color"
        return 0
    else
        log_error "Failed to switch traffic to $new_color"
        return 1
    fi
}

scale_down_old() {
    local old_color="$1"

    log_info "Scaling down $old_color deployment..."

    cd "$DEPLOYMENTS_DIR/docker"

    # Scale down old deployment to 0 replicas
    if [[ "$ENVIRONMENT" == "production" ]]; then
        docker compose -f docker-compose.base.yml -f docker-compose.prod.yml \
            up -d --scale "ai-gateway-$old_color=0" \
            || { log_warning "Failed to scale down $old_color deployment"; return 1; }
    else
        docker compose -f docker-compose.base.yml -f docker-compose.staging.yml \
            up -d --scale "ai-gateway-$old_color=0" \
            || { log_warning "Failed to scale down $old_color deployment"; return 1; }
    fi

    log_success "Scaled down $old_color deployment"
}

update_deployment_state() {
    local new_color="$1"
    local image_tag="$2"
    local state_file="$DEPLOYMENTS_DIR/.deployment-state"

    cat > "$state_file" << EOF
# Deployment State File
# Generated by blue-green-deploy.sh

DEPLOYMENT_COLOR="$new_color"
IMAGE_TAG="$image_tag"
DEPLOYMENT_TIME="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
ENVIRONMENT="$ENVIRONMENT"
EOF

    log_success "Updated deployment state: $new_color (image: $image_tag)"
}

# ============================================
# Main Execution
# ============================================

main() {
    log_info "========================================="
    log_info "AI Gateway Blue-Green Deployment"
    log_info "========================================="
    log_info "Started at: $(date -u +"%Y-%m-%dT%H:%M:%SZ")"

    # Validate arguments
    validate_arguments "$@"

    # Check prerequisites
    check_prerequisites

    # Get current and new deployment colors
    local current_color
    local new_color
    current_color=$(get_current_color)
    new_color=$(get_new_color "$current_color")

    log_info "Current deployment: $current_color"
    log_info "New deployment: $new_color"

    # Deploy new version
    if ! deploy_new_version "$new_color" "$IMAGE_TAG"; then
        log_error "Deployment failed at deploy stage"
        exit 1
    fi

    # Determine port based on color
    local new_port
    if [[ "$new_color" == "blue" ]]; then
        new_port=8081
    else
        new_port=8082
    fi

    # Wait for health check
    if ! wait_for_health_check "$new_color" "$new_port"; then
        log_error "Deployment failed at health check stage"
        log_error "Manual intervention required - new deployment may not be healthy"
        log_info "To manually investigate: docker logs ai-gateway-app-$new_color"
        exit 1
    fi

    # Run smoke tests
    if ! run_smoke_tests "$new_color" "$new_port"; then
        log_error "Deployment failed at smoke test stage"
        log_error "New deployment failed smoke tests - traffic NOT switched"
        log_info "To investigate: bash $SCRIPT_DIR/smoke-test.sh http://localhost:$new_port"
        exit 1
    fi

    # Switch traffic
    if ! switch_traffic "$new_color"; then
        log_error "Traffic switch failed - manual intervention required"
        exit 1
    fi

    # Scale down old deployment
    scale_down_old "$current_color"

    # Update deployment state
    update_deployment_state "$new_color" "$IMAGE_TAG"

    log_success "========================================="
    log_success "Deployment completed successfully!"
    log_success "Active deployment: $new_color"
    log_success "Image tag: $IMAGE_TAG"
    log_success "Completed at: $(date -u +"%Y-%m-%dT%H:%M:%SZ")"
    log_success "========================================="

    exit 0
}

# Run main function
main "$@"
