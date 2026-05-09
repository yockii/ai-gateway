# AI Gateway Operations Manual

This manual provides operational procedures, runbooks, and guidelines for day-to-day management of the AI Gateway platform.

## Table of Contents

- [Daily Operations](#daily-operations)
- [Backup Procedures](#backup-procedures)
- [Scaling Procedures](#scaling-procedures)
- [Monitoring Interpretation](#monitoring-interpretation)
- [Common Operational Tasks](#common-operational-tasks)
- [Incident Response](#incident-response)
- [Maintenance Procedures](#maintenance-procedures)
- [Performance Tuning](#performance-tuning)
- [Security Operations](#security-operations)
- [Capacity Planning](#capacity-planning)

---

## Daily Operations

### Daily Operations Checklist

Complete these tasks daily to ensure system health:

```bash
# 1. Check service health
bash scripts/deploy.sh prod status

# 2. Review error logs
bash scripts/deploy.sh prod logs --service ai-gateway | grep ERROR

# 3. Verify backup completion
ls -lh /backups/postgres/$(date +%Y%m%d)*

# 4. Check alert status
curl http://localhost:9090/api/v1/alerts | jq .

# 5. Review metrics dashboards
# Access Grafana: http://localhost:3000
```

### Daily Health Check Script

Create a cron job for automated daily checks:

```bash
#!/bin/bash
# scripts/daily-health-check.sh

LOG_FILE="/var/log/ai-gateway/daily-check.log"
date > "$LOG_FILE"

echo "=== Service Health ===" >> "$LOG_FILE"
bash scripts/deploy.sh prod status >> "$LOG_FILE" 2>&1

echo "=== Disk Space ===" >> "$LOG_FILE"
df -h >> "$LOG_FILE"

echo "=== Container Resources ===" >> "$LOG_FILE"
docker stats --no-stream >> "$LOG_FILE"

echo "=== Recent Errors ===" >> "$LOG_FILE"
docker logs ai-gateway-app --since 24h 2>&1 | grep -i error >> "$LOG_FILE"

echo "=== Backup Status ===" >> "$LOG_FILE"
ls -lh /backups/postgres/ | tail -5 >> "$LOG_FILE"
```

### Log Review Procedure

1. **Application Logs**

```bash
# View last hour of logs
docker logs ai-gateway-app --since 1h

# Check for errors
docker logs ai-gateway-app --since 24h | grep -i error

# Check for specific patterns
docker logs ai-gateway-app --since 24h | grep -E "timeout|connection|failed"
```

2. **nginx Logs**

```bash
# View nginx access logs
docker exec ai-gateway-nginx tail -f /var/log/nginx/access.log

# Check for 5xx errors
docker exec ai-gateway-nginx grep " 5[0-9][0-9] " /var/log/nginx/access.log | tail -50

# View error logs
docker exec ai-gateway-nginx tail -f /var/log/nginx/error.log
```

3. **Database Logs**

```bash
# PostgreSQL logs
docker logs ai-gateway-postgres --tail 100

# Check for connection issues
docker logs ai-gateway-postgres | grep -i "connection"
```

### Daily Metrics Review

Key metrics to review in Grafana:

1. **Request Rate**: Requests per second
2. **Error Rate**: Percentage of failed requests
3. **Latency**: P50, P95, P99 response times
4. **Database Connections**: Active vs. maximum
5. **Cache Hit Rate**: Redis effectiveness
6. **Resource Usage**: CPU, memory, disk I/O

---

## Backup Procedures

### Database Backup

#### Automated Daily Backups

Configure in `.env`:

```bash
BACKUP_ENABLED=true
BACKUP_SCHEDULE="0 2 * * *"  # Daily at 2 AM
BACKUP_RETENTION_DAYS=7
BACKUP_PATH=/backups
```

#### Manual Database Backup

```bash
# Backup PostgreSQL database
docker exec ai-gateway-postgres pg_dump -U postgres ai_gateway \
    | gzip > /backups/postgres/ai_gateway_$(date +%Y%m%d_%H%M%S).sql.gz

# Backup with custom format (faster restore)
docker exec ai-gateway-postgres pg_dump -U postgres -F c -f \
    /tmp/backup.dump ai_gateway

# Copy from container
docker cp ai-gateway-postgres:/tmp/backup.dump \
    /backups/postgres/backup_$(date +%Y%m%d).dump
```

#### Restore Database

```bash
# Restore from SQL dump
gunzip < /backups/postgres/ai_gateway_20250109.sql.gz \
    | docker exec -i ai-gateway-postgres psql -U postgres ai_gateway

# Restore from custom format
docker cp /backups/postgres/backup_20250109.dump ai-gateway-postgres:/tmp/restore.dump
docker exec ai-gateway-postgres pg_restore -U postgres -d ai_gateway -j 4 /tmp/restore.dump

# Restore to new database (for testing)
docker exec ai-gateway-postgres createdb -U postgres ai_gateway_restore
docker exec -i ai-gateway-postgres psql -U postgres ai_gateway_restore < backup.sql
```

### Configuration Backup

```bash
# Backup all configuration files
tar -czf /backups/config_$(date +%Y%m%d).tar.gz \
    deployments/docker/.env \
    deployments/docker/docker-compose.*.yml \
    nginx/nginx.conf \
    nginx/conf.d/
```

### Certificate Backup

```bash
# Backup SSL certificates
tar -czf /backups/certs_$(date +%Y%m%d).tar.gz \
    deployments/docker/certs/
```

### Backup Retention Policy

| Backup Type | Retention | Location |
|-------------|-----------|----------|
| Database Daily | 7 days | /backups/postgres/ |
| Database Weekly | 4 weeks | /backups/postgres/weekly/ |
| Database Monthly | 12 months | /backups/postgres/monthly/ |
| Configuration | 3 months | /backups/config/ |
| Certificates | Until expiry | /backups/certs/ |

### Backup Verification

```bash
# Test backup integrity
#!/bin/bash
# scripts/verify-backup.sh

BACKUP_FILE="/backups/postgres/ai_gateway_$(date +%Y%m%d).sql.gz"

if [ ! -f "$BACKUP_FILE" ]; then
    echo "ERROR: Backup file not found"
    exit 1
fi

# Check file size
SIZE=$(stat -f%z "$BACKUP_FILE" 2>/dev/null || stat -c%s "$BACKUP_FILE")
if [ "$SIZE" -lt 1000 ]; then
    echo "ERROR: Backup file too small ($SIZE bytes)"
    exit 1
fi

# Test decompression
if ! gunzip -t "$BACKUP_FILE" 2>/dev/null; then
    echo "ERROR: Backup file is corrupted"
    exit 1
fi

echo "OK: Backup verified"
```

---

## Scaling Procedures

### Horizontal Scaling

#### Scale Application Containers

```bash
# Scale to 5 replicas (production)
cd deployments/docker
docker compose -f docker-compose.base.yml -f docker-compose.prod.yml \
    up -d --scale ai-gateway-blue=5

# Verify scaling
docker compose -f docker-compose.base.yml -f docker-compose.prod.yml \
    ps ai-gateway-blue
```

#### Scaling Considerations

- **Database Connections**: Each replica needs a connection pool
- **Load Balancer**: nginx distributes traffic evenly
- **Session Storage**: Use Redis for shared session state
- **Resource Limits**: Ensure sufficient host capacity

### Vertical Scaling

#### Increase Container Resources

Edit `docker-compose.prod.yml`:

```yaml
deploy:
  resources:
    limits:
      cpus: '4'      # Increase from 2
      memory: 2G     # Increase from 1G
    reservations:
      cpus: '2'
      memory: 1G
```

Apply changes:

```bash
bash scripts/deploy.sh prod restart
```

### Database Scaling

#### Connection Pooling

Adjust database pool in application configuration:

```bash
# In .env
DB_MAX_CONNECTIONS=50  # Increase for more replicas
DB_MIN_CONNECTIONS=10
```

#### Read Replicas (Future)

For read-heavy workloads:

1. Set up PostgreSQL read replicas
2. Configure application to route reads to replicas
3. Use PgBouncer for connection pooling

---

## Monitoring Interpretation

### Key Metrics

#### Application Metrics

| Metric | Description | Healthy Range | Alert Threshold |
|--------|-------------|---------------|-----------------|
| `http_requests_total` | Total HTTP requests | - | - |
| `http_request_duration_seconds` | Request latency | P95 < 500ms | P95 > 1s |
| `http_requests_errors_total` | Error count | < 1% | > 5% |
| `db_connections_active` | Active DB connections | < 80% max | > 90% max |
| `cache_hit_ratio` | Redis cache hit rate | > 80% | < 60% |

#### System Metrics

| Metric | Description | Healthy Range | Alert Threshold |
|--------|-------------|---------------|-----------------|
| `container_cpu_usage` | CPU utilization | < 70% | > 90% |
| `container_memory_usage` | Memory utilization | < 80% | > 95% |
| `container_network_io` | Network I/O | - | - |
| `disk_usage` | Disk utilization | < 80% | > 90% |

### Prometheus Queries

#### Average Request Latency

```promql
rate(http_request_duration_seconds_sum[5m]) /
rate(http_request_duration_seconds_count[5m])
```

#### Error Rate

```promql
rate(http_requests_errors_total[5m]) /
rate(http_requests_total[5m]) * 100
```

#### Database Connection Pool Usage

```promql
db_connections_active / db_connections_max * 100
```

#### Request Rate by Endpoint

```promql
sum by (endpoint) (rate(http_requests_total{status!~"5.."}[5m]))
```

### Grafana Dashboards

Access pre-configured dashboards:

- **API Gateway Overview**: Overall system health
- **System Metrics**: CPU, memory, network
- **Database Performance**: Queries, connections, locks
- **Application Logs**: Error patterns and frequency

---

## Common Operational Tasks

### Restart Services

```bash
# Restart all services
bash scripts/deploy.sh prod restart

# Restart specific service
bash scripts/deploy.sh prod restart --service ai-gateway

# Graceful restart (zero downtime)
bash scripts/blue-green-deploy.sh production current-version
```

### Update Configuration

```bash
# 1. Edit configuration
nano deployments/docker/.env

# 2. Restart services
bash scripts/deploy.sh prod restart

# 3. Verify changes
docker logs ai-gateway-app --tail 50
```

### View Real-time Metrics

```bash
# Application metrics
curl http://localhost:9090/metrics | grep http_requests

# Health status
watch -n 5 'curl -s http://localhost:8080/health | jq .'

# Container stats
watch -n 2 'docker stats --no-stream'
```

### Database Maintenance

#### Reindex Database

```bash
docker exec -it ai-gateway-postgres psql -U postgres -d ai_gateway \
    -c "REINDEX DATABASE ai_gateway;"
```

#### Vacuum Analyze

```bash
docker exec -it ai-gateway-postgres psql -U postgres -d ai_gateway \
    -c "VACUUM ANALYZE;"
```

#### Check Table Sizes

```bash
docker exec -it ai-gateway-postgres psql -U postgres -d ai_gateway \
    -c "SELECT relname, pg_size_pretty(pg_total_relation_size(relid)) \
         FROM pg_catalog.pg_statio_user_tables \
         ORDER BY pg_total_relation_size(relid) DESC;"
```

### Log Rotation

Configure log rotation for containers:

```bash
# /etc/logrotate.d/docker-containers
/var/lib/docker/containers/*/*.log {
    rotate 7
    daily
    compress
    missingok
    delaycompress
    copytruncate
}
```

---

## Incident Response

### Severity Levels

| Severity | Description | Response Time | Examples |
|----------|-------------|---------------|----------|
| **P0 - Critical** | Complete system outage | 15 minutes | All services down, data loss |
| **P1 - High** | Major functionality broken | 1 hour | API unavailable, degraded performance |
| **P2 - Medium** | Partial functionality affected | 4 hours | Single feature broken, high error rate |
| **P3 - Low** | Minor issues | 1 day | UI issues, non-critical bugs |

### Incident Response Runbook

#### 1. Detection and Triage

```bash
# Check system status
bash scripts/deploy.sh prod status

# Check active alerts
curl http://localhost:9090/api/v1/alerts | jq '.data.alerts[] | select(.state=="firing")'

# Review recent errors
docker logs ai-gateway-app --since 1h | grep -i error | tail -50
```

#### 2. Assessment

- Determine severity level
- Identify affected components
- Estimate user impact
- Check recent deployments

```bash
# Check recent deployments
cat deployments/.deployment-state

# Check recent changes
git log --since="2 hours ago" --oneline
```

#### 3. Containment

For P0/P1 incidents:

```bash
# If deployment-related, rollback immediately
bash scripts/deploy.sh prod rollback

# If resource exhaustion, scale up
cd deployments/docker
docker compose -f docker-compose.prod.yml up -d --scale ai-gateway-blue=5

# If database issue, check connections
docker exec ai-gateway-postgres psql -U postgres -c \
    "SELECT count(*) FROM pg_stat_activity;"
```

#### 4. Resolution

- Apply fix or rollback
- Verify service restoration
- Monitor for recurrence

#### 5. Communication

Notify stakeholders:

```bash
# Example notification template
SUBJECT="[P0] AI Gateway Service Outage - RESOLVED"

MESSAGE="
The AI Gateway experienced a complete service outage from 14:30-15:15 UTC.

Root Cause: Database connection pool exhaustion

Resolution: Increased connection pool size and restarted services

Impact: All API requests failed during the outage window

Prevention: Connection pool monitoring alert added

Updates will be posted at: status.example.com
"
```

#### 6. Post-Incident Review

Create incident report:

```markdown
# Incident Report: [ID]

## Summary
[Brief description]

## Timeline
- 14:30: Detection
- 14:35: Triage started
- 14:45: Rollback initiated
- 15:15: Service restored

## Root Cause
[Analysis]

## Impact
- [Users affected]
- [Duration]
- [Requests lost]

## Resolution
[Actions taken]

## Prevention
[Future measures]
```

### Emergency Contacts

| Role | Name | Contact | Hours |
|------|------|---------|-------|
| On-Call Engineer | [Name] | [Phone/Slack] | 24/7 |
| Engineering Lead | [Name] | [Phone/Slack] | Business hours |
| DevOps Engineer | [Name] | [Phone/Slack] | Business hours |

---

## Maintenance Procedures

### Scheduled Maintenance

#### Maintenance Checklist

1. **Pre-Maintenance**
   - [ ] Notify users 24 hours in advance
   - [ ] Create pre-maintenance backup
   - [ ] Verify rollback procedure
   - [ ] Set up maintenance page

2. **During Maintenance**
   - [ ] Stop accepting new requests
   - - [ ] Drain existing connections
   - [ ] Apply changes
   - [ ] Verify functionality

3. **Post-Maintenance**
   - [ ] Run smoke tests
   - [ ] Monitor metrics
   - [ ] Verify backups
   - [ ] Send completion notice

#### Maintenance Mode

Enable maintenance mode:

```bash
# Update nginx to return maintenance page
cp nginx/maintenance.conf nginx/conf.d/maintenance.conf
docker exec ai-gateway-nginx nginx -s reload
```

### Database Migration

```bash
# 1. Backup database
bash scripts/backup-database.sh

# 2. Test migration on staging
bash scripts/deploy.sh staging deploy --tag migration-test

# 3. Run migration
docker exec -it ai-gateway-postgres psql -U postgres -d ai_gateway < migration.sql

# 4. Verify
bash scripts/smoke-test.sh http://localhost:8080

# 5. If failed, restore from backup
```

### Application Update

```bash
# 1. Build new image
docker build -t ai-gateway:v1.1.0 .

# 2. Deploy using blue-green
bash scripts/blue-green-deploy.sh production v1.1.0

# 3. Monitor
bash scripts/deploy.sh prod logs --service ai-gateway

# 4. If issues, rollback
bash scripts/deploy.sh prod rollback
```

### Dependency Updates

```bash
# 1. Update Go modules
cd cmd/api_gateway
go get -u ./...
go mod tidy

# 2. Build and test
docker build -t ai-gateway:test .

# 3. Deploy to staging first
bash scripts/deploy.sh staging deploy --tag test

# 4. After verification, deploy to production
bash scripts/blue-green-deploy.sh production v1.2.0
```

---

## Performance Tuning

### Application Tuning

#### Database Connection Pool

```bash
# In .env
DB_MAX_CONNECTIONS=50
DB_MAX_IDLE_CONNECTIONS=10
DB_CONNECTION_MAX_LIFETIME=1h
```

#### Cache Configuration

```bash
# Redis settings
REDIS_MAX_MEMORY=256mb
REDIS_EVICTION_POLICY=allkeys-lru

# Application cache
CACHE_TTL_DEFAULT=300
CACHE_TTL_SHORT=60
CACHE_TTL_LONG=3600
```

#### Rate Limiting

```nginx
# In nginx.conf
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=100r/s;
limit_req_zone $binary_remote_addr zone=auth_limit:10m rate=20r/s;
```

### System Tuning

#### Kernel Parameters

```bash
# /etc/sysctl.conf
net.core.somaxconn = 65535
net.ipv4.tcp_max_syn_backlog = 8192
net.ipv4.ip_local_port_range = 1024 65535
vm.swappiness = 10
```

#### Docker Daemon

```json
// /etc/docker/daemon.json
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  },
  "storage-driver": "overlay2",
  "max-concurrent-downloads": 10
}
```

### Performance Testing

```bash
# Load test with Apache Bench
ab -n 10000 -c 100 http://localhost:8080/v1/models

# Load test with wrk
wrk -t12 -c400 -d30s http://localhost:8080/v1/chat/completions

# Benchmark script
bash scripts/benchmark.sh
```

---

## Security Operations

### Security Daily Checklist

- [ ] Review authentication logs for suspicious activity
- [ ] Check for failed login attempts
- [ ] Verify SSL certificate validity
- [ ] Review access logs for unusual patterns
- [ ] Check security advisories for dependencies

### Security Monitoring

```bash
# Check for failed authentication
docker logs ai-gateway-app --since 24h | grep -i "failed auth"

# Check rate limit violations
docker exec ai-gateway-nginx grep "limiting requests" /var/log/nginx/error.log

# Check for SQL injection patterns
docker logs ai-gateway-app | grep -iE "union.*select|drop.*table|'--'"
```

### Certificate Management

```bash
# Check certificate expiry
openssl x509 -in deployments/docker/certs/fullchain.pem -noout -dates

# Renew with Let's Encrypt
docker compose -f deployments/docker/docker-compose.prod.yml \
    -f deployments/docker/docker-compose.base.yml exec certbot renew

# Force renewal
docker compose -f deployments/docker/docker-compose.prod.yml \
    -f deployments/docker/docker-compose.base.yml exec certbot renew --force-renewal
```

### Security Updates

```bash
# Update base images
docker pull postgres:16-alpine
docker pull redis:7-alpine
docker pull nginx:alpine

# Rebuild application
docker build --no-cache -t ai-gateway:v1.2.1 .

# Deploy
bash scripts/blue-green-deploy.sh production v1.2.1
```

---

## Capacity Planning

### Capacity Metrics

Track these metrics for capacity planning:

| Metric | Current | Warning | Critical |
|--------|---------|---------|----------|
| CPU Utilization | % | 70% | 90% |
| Memory Utilization | % | 80% | 95% |
| Disk Usage | % | 80% | 90% |
| Request Rate | req/s | - | - |
| Database Size | GB | - | - |

### Growth Planning

```bash
# Calculate growth rate
# Current requests/day: 100,000
# Expected growth: 10%/month
# 6-month projection: 100,000 * 1.1^6 = 177,000 requests/day

# Plan capacity accordingly
```

### Scaling Triggers

Define triggers for scaling:

- **CPU > 80% for 5 minutes** → Scale up
- **Memory > 85%** → Add memory or containers
- **Disk > 80%** → Clean up or expand storage
- **Request latency P95 > 1s** → Scale up
- **Queue depth > 1000** → Scale up

---

## Appendix: Quick Reference

### Essential Commands

```bash
# Service Status
bash scripts/deploy.sh prod status

# View Logs
bash scripts/deploy.sh prod logs

# Restart Services
bash scripts/deploy.sh prod restart

# Deploy New Version
bash scripts/deploy.sh prod deploy --tag v1.0.0

# Rollback
bash scripts/deploy.sh prod rollback

# Health Check
curl http://localhost:8080/health

# Database Backup
docker exec ai-gateway-postgres pg_dump -U postgres ai_gateway | gzip > backup.sql.gz

# Container Stats
docker stats
```

### File Locations

| Component | Location |
|-----------|----------|
| Configuration | deployments/docker/.env |
| Docker Compose | deployments/docker/docker-compose.*.yml |
| nginx Config | nginx/nginx.conf |
| SSL Certificates | deployments/docker/certs/ |
| Backups | /backups/ |
| Logs | deployments/deployments.log |

### Service Ports

| Service | Port | Notes |
|---------|------|-------|
| HTTP | 80 | Redirects to HTTPS |
| HTTPS | 443 | nginx |
| API Gateway (dev) | 8080 | Direct access |
| API Gateway (blue) | 8081 | Blue deployment |
| API Gateway (green) | 8082 | Green deployment |
| Metrics (blue) | 9091 | Prometheus |
| Metrics (green) | 9092 | Prometheus |
| PostgreSQL | 5432 | Internal |
| Redis | 6379 | Internal |
| Prometheus | 9090 | Monitoring |
| Grafana | 3000 | Dashboards |

---

**Document Version**: 1.0
**Last Updated**: 2025-01-09
**Maintained By**: Operations Team
**Next Review**: 2025-02-09
