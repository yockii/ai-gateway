# AI Gateway Deployment Guide

This guide covers deploying the AI Gateway application across different environments using Docker Compose, nginx reverse proxy, and blue-green deployment strategy.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start (Development)](#quick-start-development)
- [Environment Configuration](#environment-configuration)
- [Staging Deployment](#staging-deployment)
- [Production Deployment](#production-deployment)
- [Blue-Green Deployment Process](#blue-green-deployment-process)
- [Rollback Procedures](#rollback-procedures)
- [Monitoring and Logging](#monitoring-and-logging)
- [Security Considerations](#security-considerations)
- [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Required Software

- **Docker**: Version 20.10 or later
- **Docker Compose**: Version 2.0 or later
- **Bash**: For deployment scripts
- **curl**: For health checks

### System Requirements

| Environment | CPU | Memory | Disk |
|-------------|-----|--------|------|
| Development | 2 cores | 4 GB | 10 GB |
| Staging | 4 cores | 8 GB | 20 GB |
| Production | 8 cores | 16 GB | 50 GB |

### Network Requirements

- **Ports**: 80 (HTTP), 443 (HTTPS), 8080-8082 (Application), 9090-9092 (Metrics)
- **Firewall**: Allow inbound on ports 80 and 443
- **Outbound**: Allow access to Docker registry, AI provider APIs

### SSL Certificates

For production deployment, you need SSL certificates. Options:

1. **Let's Encrypt** (recommended for public domains)
2. **AWS Certificate Manager** (for AWS deployments)
3. **Self-signed** (for testing only)
4. **Commercial certificates**

---

## Quick Start (Development)

### 1. Clone and Configure

```bash
# Clone the repository
git clone https://github.com/yockii/ai_gateway.git
cd ai_gateway

# Copy environment configuration
cp deployments/docker/.env.example deployments/docker/.env

# Edit .env as needed (development defaults work)
nano deployments/docker/.env
```

### 2. Start Development Environment

```bash
# Using the deployment orchestrator
bash scripts/deploy.sh dev deploy

# Or directly with Docker Compose
cd deployments/docker
docker compose -f docker-compose.base.yml -f docker-compose.dev.yml up -d
```

### 3. Verify Deployment

```bash
# Check service status
bash scripts/deploy.sh dev status

# Test health endpoint
curl http://localhost:8080/health

# Run smoke tests
bash scripts/smoke-test.sh http://localhost:8080
```

### 4. View Logs

```bash
# View all logs
bash scripts/deploy.sh dev logs

# Follow logs
docker compose -f deployments/docker/docker-compose.dev.yml \
    -f deployments/docker/docker-compose.base.yml logs -f

# View specific service logs
bash scripts/deploy.sh dev logs --service ai-gateway
```

### 5. Stop Services

```bash
bash scripts/deploy.sh dev stop
```

---

## Environment Configuration

### Environment File Structure

Create `.env` files per environment in `deployments/docker/`:

```bash
# Development
cp .env.example .env

# Staging
cp .env.example .env.staging

# Production
cp .env.example .env.prod
```

### Key Configuration Variables

```bash
# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Database
DB_HOST=postgres
DB_PORT=5432
DB_NAME=ai_gateway
DB_USER=postgres
DB_PASSWORD=changeme

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=changeme

# Docker Registry
DOCKER_REGISTRY=ai-gateway
IMAGE_TAG=latest

# Logging
LOG_LEVEL=info          # debug, info, warn, error
LOG_FORMAT=text         # text, json

# Blue-Green Deployment
DEPLOYMENT_COLOR=blue   # blue or green

# SSL Certificates
SSL_CERT_PATH=./certs/fullchain.pem
SSL_KEY_PATH=./certs/privkey.pem
```

### Environment-Specific Settings

| Variable | Development | Staging | Production |
|----------|-------------|---------|------------|
| LOG_LEVEL | debug | info | warn |
| LOG_FORMAT | text | text | json |
| Replicas | 1 | 2 | 3 |
| Resource Limits | None | 512M | 1G |

---

## Staging Deployment

### 1. Prepare Staging Environment

```bash
# Configure environment
cp deployments/docker/.env.example deployments/docker/.env.staging
# Edit staging-specific values

# Build and tag image
docker build -t ai-gateway:staging-latest .
```

### 2. Deploy to Staging

```bash
# Using deployment orchestrator
bash scripts/deploy.sh staging deploy --tag staging-latest

# Or directly
cd deployments/docker
docker compose -f docker-compose.base.yml \
    -f docker-compose.staging.yml up -d
```

### 3. Verify Staging Deployment

```bash
# Check status
bash scripts/deploy.sh staging status

# Run smoke tests
bash scripts/smoke-test.sh http://staging.example.com

# Run integration tests
bash scripts/test-integration.sh
```

### 4. Access Staging Services

- **API Gateway**: http://staging.example.com
- **Health Check**: http://staging.example.com/health
- **Metrics**: http://staging.example.com/metrics (internal only)
- **Grafana**: http://staging.example.com:3000

---

## Production Deployment

### Production Deployment Checklist

Before deploying to production, verify:

- [ ] SSL certificates configured and valid
- [ ] Firewall rules configured (80, 443 open)
- [ ] Database backups enabled
- [ ] Monitoring and alerting configured
- [ ] Log rotation configured
- [ ] Resource limits set in docker-compose.prod.yml
- [ ] Health checks verified
- [ ] Rollback procedure tested
- [ ] Environment variables reviewed
- [ ] CI/CD pipeline configured (if applicable)

### 1. Prepare Production Environment

```bash
# Configure production environment
cp deployments/docker/.env.example deployments/docker/.env.prod

# Edit production values - IMPORTANT:
# - Change all passwords
# - Set strong DB password
# - Configure proper Redis password
# - Set production log level
nano deployments/docker/.env.prod
```

### 2. Build and Push Production Image

```bash
# Build production image
docker build -t your-registry.com/ai-gateway:v1.0.0 .

# Push to registry
docker push your-registry.com/ai-gateway:v1.0.0
```

### 3. Deploy Using Blue-Green Strategy

```bash
# Deploy to production (blue-green)
bash scripts/deploy.sh prod deploy --tag v1.0.0
```

This will:
1. Deploy new version to inactive color (green if blue is active)
2. Wait for health checks
3. Run smoke tests
4. Switch traffic to new deployment
5. Scale down old deployment

### 4. Verify Production Deployment

```bash
# Check deployment status
bash scripts/deploy.sh prod status

# Check health endpoints
curl https://api.example.com/health

# View logs
bash scripts/deploy.sh prod logs --service ai-gateway
```

---

## Blue-Green Deployment Process

### Architecture Overview

```
                    ┌─────────────────┐
                    │     nginx       │
                    │  (Reverse Proxy)│
                    └────────┬────────┘
                             │
              ┌──────────────┴──────────────┐
              │                             │
         ┌────▼─────┐                 ┌────▼─────┐
         │   BLUE   │                 │   GREEN  │
         │ :8081    │                 │  :8082   │
         │ (Active) │                 │(Standby) │
         └──────────┘                 └──────────┘
```

### Deployment Steps

The blue-green deployment process (`scripts/blue-green-deploy.sh`) performs:

1. **Determine Active Color**: Read from `.deployment-state` file
2. **Calculate Target Color**: Opposite of current active
3. **Deploy New Version**: Start containers with new image
4. **Health Check**: Wait for `/health` endpoint to respond
5. **Smoke Tests**: Run basic functionality tests
6. **Switch Traffic**: Update nginx upstream configuration
7. **Scale Down Old**: Reduce old deployment replicas to 0

### Manual Blue-Green Deployment

```bash
# Deploy specific version
bash scripts/blue-green-deploy.sh production v1.0.0

# The script will:
# - Deploy v1.0.0 to inactive color
# - Verify health at :8081 or :8082
# - Run smoke tests
# - Switch nginx upstream
# - Update deployment state
```

### Traffic Switching

Traffic is controlled by nginx upstream configuration in `nginx/nginx.conf`:

```nginx
upstream ai_gateway_active {
    least_conn;
    server ai-gateway-blue:8080;  # or ai-gateway-green
    keepalive 32;
}
```

The `switch-traffic.sh` script updates this configuration:

```bash
# Switch traffic manually (not recommended - use blue-green-deploy.sh)
bash scripts/switch-traffic.sh green
```

---

## Rollback Procedures

### Automatic Rollback on Failure

The blue-green deployment process includes automatic safety checks:

1. Health check fails → Deployment stops, traffic not switched
2. Smoke tests fail → Deployment stops, traffic not switched
3. Traffic switch fails → Manual intervention required

### Manual Rollback

If a deployment is successful but issues are discovered later:

```bash
# Rollback to previous version
bash scripts/deploy.sh prod rollback

# This will:
# - Identify previous deployment color
# - Switch traffic back
# - Scale down current deployment
```

### Rollback Script Details

The `rollback.sh` script:

1. Reads `.deployment-state` to find previous color
2. Calls `switch-traffic.sh` to revert
3. Verifies health of reverted deployment
4. Optionally scales down failed deployment
5. Updates deployment state

### Emergency Rollback

If immediate rollback is needed:

```bash
# Quick traffic switch
bash scripts/switch-traffic.sh blue  # or green

# Verify health
curl http://localhost:8081/health  # blue
curl http://localhost:8082/health  # green

# Check which is active
cat deployments/.deployment-state
```

---

## Monitoring and Logging

### Health Checks

```bash
# Application health endpoint
curl http://localhost:8080/health

# Expected response: {"status":"healthy"}

# Blue deployment health
curl http://localhost:8081/health

# Green deployment health
curl http://localhost:8082/health
```

### Viewing Logs

```bash
# All services in environment
bash scripts/deploy.sh prod logs

# Specific service
bash scripts/deploy.sh prod logs --service ai-gateway

# Last N lines
TAIL_LINES=500 bash scripts/deploy.sh prod logs

# Docker Compose directly
docker compose -f deployments/docker/docker-compose.prod.yml \
    -f deployments/docker/docker-compose.base.yml logs -f ai-gateway
```

### Metrics

The application exposes Prometheus metrics on port 9090:

```bash
# Blue metrics
curl http://localhost:9091/metrics

# Green metrics
curl http://localhost:9092/metrics
```

### Monitoring Stack

Start monitoring services:

```bash
cd deployments/monitoring
docker compose up -d
```

Access points:
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)
- **Logs**: Available in Grafana with Loki datasource

### Log Locations

| Component | Log Location |
|-----------|--------------|
| Application | Container stdout/stderr |
| nginx | /var/log/nginx/ (volume) |
| PostgreSQL | Container stdout/stderr |
| Redis | Container stdout/stderr |
| Deployment logs | deployments/deployments.log |

---

## Security Considerations

### Container Security

All production containers run with security hardening:

- **Non-root user**: Containers run as `appuser`
- **Read-only filesystem**: Except /tmp (tmpfs)
- **Dropped capabilities**: All except NET_BIND_SERVICE, CHOWN, SETGID, SETUID
- **No new privileges**: Security option enabled
- **Resource limits**: CPU and memory constrained

### Network Security

- **nginx TLS termination**: TLSv1.2 and TLSv1.3 only
- **Security headers**: HSTS, X-Frame-Options, CSP
- **Rate limiting**: Configured in nginx
- **Firewall**: Restrict access to metrics endpoint

### Secrets Management

**Never commit secrets to git.** Use `.env` files (in `.gitignore`) for:

- Database passwords
- Redis passwords
- API keys
- SSL certificate paths
- Registry credentials

For production, consider:
- HashiCorp Vault
- AWS Secrets Manager
- Azure Key Vault
- Docker Swarm secrets
- Kubernetes secrets

### SSL/TLS Configuration

Production nginx uses Mozilla Intermediate configuration:

```nginx
ssl_protocols TLSv1.2 TLSv1.3;
ssl_ciphers 'ECDHE-ECDSA-AES128-GCM-SHA256:...';
ssl_prefer_server_ciphers off;
```

Renew certificates with Let's Encrypt:

```bash
# Certbot container auto-renews every 12 hours
# Manual renewal:
docker compose -f deployments/docker/docker-compose.prod.yml \
    -f deployments/docker/docker-compose.base.yml exec certbot renew
```

---

## Troubleshooting

### Common Issues

#### Container Won't Start

```bash
# Check logs
docker logs ai-gateway-app

# Common causes:
# 1. Port already in use
# 2. Database not ready
# 3. Invalid environment variables

# Check port usage
netstat -tuln | grep -E '8080|8081|8082|5432|6379'

# Check database connection
docker exec ai-gateway-postgres pg_isready -U postgres
```

#### Health Check Failing

```bash
# Test health endpoint directly
docker exec ai-gateway-app wget -O- http://localhost:8080/health

# Check if service is listening
docker exec ai-gateway-app netstat -tuln | grep 8080

# View application logs for errors
docker logs ai-gateway-app --tail 100
```

#### nginx 502 Bad Gateway

```bash
# Check upstream is healthy
curl http://localhost:8081/health  # blue
curl http://localhost:8082/health  # green

# Check nginx configuration
docker exec ai-gateway-nginx nginx -t

# Reload nginx if config changed
docker exec ai-gateway-nginx nginx -s reload

# View nginx error logs
docker logs ai-gateway-nginx
```

#### Database Connection Issues

```bash
# Check database is running
docker ps | grep postgres

# Test connection from application container
docker exec ai-gateway-app pg_isready -h postgres -U postgres

# View database logs
docker logs ai-gateway-postgres

# Connect to database
docker exec -it ai-gateway-postgres psql -U postgres -d ai_gateway
```

#### High Memory Usage

```bash
# Check container resource usage
docker stats

# Adjust limits in docker-compose.prod.yml
# deploy.resources.limits.memory: 1G

# Restart with new limits
bash scripts/deploy.sh prod restart
```

### Debug Mode

Enable debug logging:

```bash
# In .env file
LOG_LEVEL=debug

# Restart services
bash scripts/deploy.sh prod restart
```

### Clean Restart

If all else fails:

```bash
# Stop all services
bash scripts/deploy.sh prod stop

# Remove volumes (WARNING: deletes data)
docker compose -f deployments/docker/docker-compose.prod.yml \
    -f deployments/docker/docker-compose.base.yml down -v

# Start fresh
bash scripts/deploy.sh prod deploy
```

### Getting Help

1. Check logs: `bash scripts/deploy.sh prod logs`
2. Review this documentation
3. Check GitHub issues: https://github.com/yockii/ai_gateway/issues
4. Contact operations team

---

## Deployment Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         Production Environment                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                     nginx (Reverse Proxy)                │    │
│  │              TLS Termination | Load Balancing            │    │
│  │              Security Headers | Rate Limiting            │    │
│  └────────────────────┬────────────────────────────────────┘    │
│                       │                                           │
│         ┌─────────────┴─────────────┐                            │
│         │                           │                            │
│  ┌──────▼──────┐             ┌──────▼──────┐                    │
│  │    BLUE     │             │    GREEN    │                    │
│  │ Deployment  │             │ Deployment  │                    │
│  │             │             │             │                    │
│  │ ┌────────┐  │             │ ┌────────┐  │                    │
│  │ │ App 1  │  │             │ │ App 1  │  │                    │
│  │ ├────────┤  │             │ ├────────┤  │                    │
│  │ │ App 2  │  │             │ │ App 2  │  │                    │
│  │ ├────────┤  │             │ ├────────┤  │                    │
│  │ │ App 3  │  │             │ │ App 3  │  │                    │
│  │ └────────┘  │             │ └────────┘  │                    │
│  │  :8081      │             │  :8082      │                    │
│  └─────────────┘             └─────────────┘                    │
│                                                                   │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                      Data Layer                          │    │
│  │  ┌────────────┐           ┌────────────┐                │    │
│  │  │ PostgreSQL │           │   Redis    │                │    │
│  │  │  :5432     │           │   :6379    │                │    │
│  │  └────────────┘           └────────────┘                │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                   │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                    Monitoring Stack                      │    │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐              │    │
│  │  │Prometheus│  │  Grafana │  │   Loki   │              │    │
│  │  │  :9090   │  │  :3000   │  │  :3100   │              │    │
│  │  └──────────┘  └──────────┘  └──────────┘              │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```

---

## Appendix: Deployment Scripts Reference

### deploy.sh

Main deployment orchestrator.

```bash
bash scripts/deploy.sh <environment> <action> [options]

Environments: dev, staging, prod
Actions: deploy, update, rollback, status, logs, restart, stop, health
Options: --tag TAG, --dry-run, --no-health-check, --service NAME
```

### blue-green-deploy.sh

Blue-green deployment automation.

```bash
bash scripts/blue-green-deploy.sh <environment> <image_tag>

Performs: Deploy → Health Check → Smoke Test → Traffic Switch → Scale Down
```

### switch-traffic.sh

Switch traffic between blue and green.

```bash
bash scripts/switch-traffic.sh <color>

Updates nginx upstream configuration and reloads nginx
```

### rollback.sh

Rollback to previous deployment.

```bash
bash scripts/rollback.sh [environment]

Switches traffic back to previous deployment color
```

### smoke-test.sh

Basic functionality tests.

```bash
bash scripts/smoke-test.sh <base_url>

Tests: health endpoint, models endpoint, authentication
```

---

**Document Version**: 1.0
**Last Updated**: 2025-01-09
**Maintained By**: Operations Team
