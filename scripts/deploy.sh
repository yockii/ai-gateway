#!/usr/bin/env bash
set -euo pipefail

# AI Gateway Deployment Orchestrator
# Main entry point for all deployment operations

# ============================================
# Configuration
# ============================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DEPLOYMENTS_DIR="$PROJECT_ROOT/deployments"
LOG_DIR="$DEPLOYMENTS_DIR/logs"
LOG_FILE="$DEPLOYMENTS_DIR/deployments.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

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

log_step() {
    echo -e "${CYAN}[STEP]${NC} $*"
}

print_usage() {
    cat << EOF
Usage: $(basename "$0") <environment> <action> [options]

AI Gateway Deployment Orchestrator - Manage deployments across environments.

ENVIRONMENTS:
    dev         Development environment
    staging     Staging environment
    prod        Production environment

ACTIONS:
    deploy      Full deployment (blue-green for prod)
    update      Update existing deployment (no color switch)
    rollback    Rollback to previous version (prod only)
    status      Show deployment status
    logs        Show logs from all services
    restart     Restart services
    stop        Stop services
    health      Run health checks

OPTIONS:
    --tag TAG           Docker image tag to deploy (default: latest)
    --dry-run           Show what would be done without executing
    --no-health-check   Skip health check after deployment
    --no-migrations     Skip database migrations
    --scale N           Scale to N replicas
    --service NAME      Operate on specific service only

ENVIRONMENT VARIABLES:
    DOCKER_REGISTRY     Docker registry (default: ai-gateway)
    IMAGE_TAG           Default image tag (default: latest)
    VERBOSE             Enable verbose output (true/false)

EXAMPLES:
    $(basename "$0") dev deploy                    # Deploy to dev
    $(basename "$0") staging deploy --tag v1.0.0   # Deploy v1.0.0 to staging
    $(basename "$0") prod deploy --tag v1.2.3      # Deploy to production
    $(basename "$0") prod rollback                 # Rollback production
    $(basename "$0") prod status                   # Show production status
    $(basename "$0") staging logs --service ai-gateway  # Show service logs
    $(basename "$0") dev stop                      # Stop dev environment

EOF
    exit 1
}

# ============================================
# Validation Functions
# ============================================

validate_environment() {
    local env="$1"
    case "$env" in
        dev|development)
            echo "dev"
            ;;
        staging|stage)
            echo "staging"
            ;;
        prod|production)
            echo "prod"
            ;;
        *)
            log_error "Invalid environment: $env"
            log_info "Valid environments: dev, staging, prod"
            exit 1
            ;;
    esac
}

validate_action() {
    local action="$1"
    case "$action" in
        deploy|update|rollback|status|logs|restart|stop|health)
            echo "$action"
            ;;
        *)
            log_error "Invalid action: $action"
            log_info "Valid actions: deploy, update, rollback, status, logs, restart, stop, health"
            exit 1
            ;;
    esac
}

# ============================================
# Prerequisites Check
# ============================================

check_prerequisites() {
    log_info "Checking prerequisites..."

    # Check Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        exit 1
    fi

    # Check Docker Compose
    if ! docker compose version &> /dev/null && ! docker-compose version &> /dev/null; then
        log_error "Docker Compose is not installed"
        exit 1
    fi

    # Check Docker daemon
    if ! docker info &> /dev/null; then
        log_error "Docker daemon is not running"
        exit 1
    fi

    # Check .env file
    local env_file="$DEPLOYMENTS_DIR/docker/.env"
    if [[ ! -f "$env_file" ]]; then
        if [[ -f "$env_file.example" ]]; then
            log_warning ".env file not found, copying from .env.example"
            cp "$env_file.example" "$env_file"
        else
            log_error "No .env file found. Please create $env_file"
            exit 1
        fi
    fi

    # Create log directory
    mkdir -p "$LOG_DIR"

    log_success "Prerequisites check passed"
}

# ============================================
# Docker Compose Commands
# ============================================

get_compose_files() {
    local env="$1"

    case "$env" in
        dev)
            echo "-f $DEPLOYMENTS_DIR/docker/docker-compose.base.yml -f $DEPLOYMENTS_DIR/docker/docker-compose.dev.yml"
            ;;
        staging)
            echo "-f $DEPLOYMENTS_DIR/docker/docker-compose.base.yml -f $DEPLOYMENTS_DIR/docker/docker-compose.staging.yml"
            ;;
        prod)
            echo "-f $DEPLOYMENTS_DIR/docker/docker-compose.base.yml -f $DEPLOYMENTS_DIR/docker/docker-compose.prod.yml"
            ;;
    esac
}

compose_cmd() {
    local env="$1"
    shift
    local compose_files
    compose_files=$(get_compose_files "$env")

    cd "$DEPLOYMENTS_DIR/docker"
    docker compose $compose_files "$@"
}

# ============================================
# Action Handlers
# ============================================

action_deploy() {
    local env="$1"
    local tag="${2:-latest}"

    log_step "Deploying to $env environment (tag: $tag)..."

    if [[ "$env" == "prod" ]]; then
        # Production uses blue-green deployment
        log_info "Using blue-green deployment for production..."

        if [[ "${DRY_RUN:-false}" == "true" ]]; then
            log_info "[DRY RUN] Would execute blue-green deployment"
            log_info "  Environment: $env"
            log_info "  Image tag: $tag"
            return 0
        fi

        # Call blue-green deployment script
        if bash "$SCRIPT_DIR/blue-green-deploy.sh" "$env" "$tag"; then
            log_success "Production deployment completed"
        else
            log_error "Production deployment failed"
            return 1
        fi
    else
        # Dev and staging use standard deployment
        log_info "Pulling latest images..."
        compose_cmd "$env" pull || log_warning "Some images could not be pulled"

        log_info "Starting services..."
        compose_cmd "$env" up -d || {
            log_error "Failed to start services"
            return 1
        }

        # Run migrations unless skipped
        if [[ "${NO_MIGRATIONS:-false}" != "true" ]]; then
            log_info "Running database migrations..."
            # Add migration command here if needed
            log_info "Migrations completed (or not configured)"
        fi

        # Health check unless skipped
        if [[ "${NO_HEALTH_CHECK:-false}" != "true" ]]; then
            log_info "Running health checks..."
            run_health_check "$env"
        fi

        log_success "Deployment to $env completed"
    fi
}

action_update() {
    local env="$1"
    local tag="${2:-latest}"

    log_step "Updating $env environment (tag: $tag)..."

    if [[ "${DRY_RUN:-false}" == "true" ]]; then
        log_info "[DRY RUN] Would update services"
        return 0
    fi

    # Pull new images
    compose_cmd "$env" pull

    # Update services without recreating (unless image changed)
    compose_cmd "$env" up -d --no-deps --build ai-gateway

    log_success "Update completed"
}

action_rollback() {
    local env="$1"

    if [[ "$env" != "prod" ]]; then
        log_warning "Rollback is only supported for production environment"
        log_info "For $env, use: $(basename "$0") $env deploy --tag <previous-tag>"
        return 0
    fi

    log_step "Rolling back production deployment..."

    if [[ "${DRY_RUN:-false}" == "true" ]]; then
        log_info "[DRY RUN] Would execute rollback"
        return 0
    fi

    # Call rollback script
    if bash "$SCRIPT_DIR/rollback.sh"; then
        log_success "Rollback completed"
    else
        log_error "Rollback failed"
        return 1
    fi
}

action_status() {
    local env="$1"

    log_step "Deployment status for $env environment..."

    echo ""
    echo "========================================="
    echo "  Deployment Status: $env"
    echo "========================================="
    echo ""

    compose_cmd "$env" ps

    echo ""
    echo "========================================="
    echo "  Service Health"
    echo "========================================="
    echo ""

    # Check health endpoints
    local base_url
    case "$env" in
        dev)
            base_url="http://localhost:8080"
            ;;
        staging)
            base_url="http://localhost:8080"
            ;;
        prod)
            # Check blue and green endpoints
            echo "Blue deployment:"
            if curl -f -s http://localhost:8081/health > /dev/null 2>&1; then
                echo -e "  ${GREEN}HEALTHY${NC}"
            else
                echo -e "  ${RED}UNHEALTHY${NC}"
            fi

            echo "Green deployment:"
            if curl -f -s http://localhost:8082/health > /dev/null 2>&1; then
                echo -e "  ${GREEN}HEALTHY${NC}"
            else
                echo -e "  ${RED}UNHEALTHY${NC}"
            fi

            echo "Active upstream:"
            if [[ -f "$DEPLOYMENTS_DIR/.active-upstream" ]]; then
                cat "$DEPLOYMENTS_DIR/.active-upstream"
            else
                echo "  blue (default)"
            fi

            return 0
            ;;
    esac

    if curl -f -s "$base_url/health" > /dev/null 2>&1; then
        echo -e "API Gateway: ${GREEN}HEALTHY${NC}"
    else
        echo -e "API Gateway: ${RED}UNHEALTHY${NC}"
    fi
}

action_logs() {
    local env="$1"
    local service="${2:-}"
    local tail_lines="${TAIL_LINES:-100}"

    log_step "Showing logs for $env environment..."

    if [[ -n "$service" ]]; then
        compose_cmd "$env" logs --tail="$tail_lines" -f "$service"
    else
        compose_cmd "$env" logs --tail="$tail_lines" -f
    fi
}

action_restart() {
    local env="$1"
    local service="${2:-}"

    log_step "Restarting services in $env environment..."

    if [[ "${DRY_RUN:-false}" == "true" ]]; then
        log_info "[DRY RUN] Would restart services"
        return 0
    fi

    if [[ -n "$service" ]]; then
        compose_cmd "$env" restart "$service"
    else
        compose_cmd "$env" restart
    fi

    log_success "Restart completed"
}

action_stop() {
    local env="$1"

    log_step "Stopping services in $env environment..."

    if [[ "${DRY_RUN:-false}" == "true" ]]; then
        log_info "[DRY RUN] Would stop services"
        return 0
    fi

    compose_cmd "$env" down

    log_success "Services stopped"
}

run_health_check() {
    local env="$1"
    local base_url

    case "$env" in
        dev|staging)
            base_url="http://localhost:8080"
            ;;
        *)
            log_warning "Health check not configured for environment: $env"
            return 0
            ;;
    esac

    local health_url="$base_url/health"

    log_info "Checking health at $health_url..."

    local attempts=0
    local max_attempts=30

    while [[ $attempts -lt $max_attempts ]]; do
        if curl -f -s "$health_url" > /dev/null 2>&1; then
            log_success "Health check passed"
            return 0
        fi

        ((attempts++))
        echo -ne "\rChecking... ($attempts/$max_attempts) "
        sleep 2
    done

    echo ""
    log_error "Health check failed after $max_attempts attempts"
    return 1
}

action_health() {
    local env="$1"

    log_step "Running health checks for $env environment..."

    run_health_check "$env"

    # Run smoke tests if available
    local smoke_test="$SCRIPT_DIR/smoke-test.sh"
    if [[ -f "$smoke_test" ]]; then
        log_info "Running smoke tests..."
        bash "$smoke_test" || log_warning "Smoke tests failed"
    fi
}

# ============================================
# Main Execution
# ============================================

main() {
    # Parse arguments
    if [[ $# -lt 2 ]]; then
        print_usage
    fi

    local env
    local action
    local tag=""
    local service=""
    local dry_run=false

    env=$(validate_environment "$1")
    action=$(validate_action "$2")

    shift 2

    # Parse options
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --tag)
                tag="$2"
                shift 2
                ;;
            --dry-run)
                dry_run=true
                shift
                ;;
            --no-health-check)
                export NO_HEALTH_CHECK=true
                shift
                ;;
            --no-migrations)
                export NO_MIGRATIONS=true
                shift
                ;;
            --scale)
                export SCALE="$2"
                shift 2
                ;;
            --service)
                service="$2"
                shift 2
                ;;
            -h|--help)
                print_usage
                ;;
            *)
                log_error "Unknown option: $1"
                print_usage
                ;;
        esac
    done

    # Set dry run flag
    if [[ "$dry_run" == "true" ]]; then
        export DRY_RUN=true
    fi

    # Set image tag if provided
    if [[ -n "$tag" ]]; then
        export IMAGE_TAG="$tag"
    fi

    # Print deployment info
    log_info "========================================="
    log_info "AI Gateway Deployment"
    log_info "========================================="
    log_info "Environment: $env"
    log_info "Action: $action"
    [[ -n "$tag" ]] && log_info "Image tag: $tag"
    log_info "Started at: $(date -u +"%Y-%m-%dT%H:%M:%SZ")"
    log_info "========================================="
    echo ""

    # Check prerequisites
    check_prerequisites

    # Execute action
    case "$action" in
        deploy)
            action_deploy "$env" "$tag"
            ;;
        update)
            action_update "$env" "$tag"
            ;;
        rollback)
            action_rollback "$env"
            ;;
        status)
            action_status "$env"
            ;;
        logs)
            action_logs "$env" "$service"
            ;;
        restart)
            action_restart "$env" "$service"
            ;;
        stop)
            action_stop "$env"
            ;;
        health)
            action_health "$env"
            ;;
    esac

    local exit_code=$?

    echo ""
    log_info "========================================="
    if [[ $exit_code -eq 0 ]]; then
        log_success "Operation completed successfully!"
    else
        log_error "Operation failed with exit code $exit_code"
    fi
    log_info "Completed at: $(date -u +"%Y-%m-%dT%H:%M:%SZ")"
    log_info "========================================="

    return $exit_code
}

# Run main function
main "$@"
