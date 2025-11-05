# CloudCostGuard Runbook

Comprehensive operational guide for deploying, monitoring, troubleshooting, and maintaining CloudCostGuard in production environments.

## Table of Contents

1. [Deployment](#deployment)
2. [Configuration](#configuration)
3. [Monitoring & Observability](#monitoring--observability)
4. [Troubleshooting](#troubleshooting)
5. [Incident Response](#incident-response)
6. [Disaster Recovery](#disaster-recovery)
7. [Performance Tuning](#performance-tuning)

---

## Deployment

### Pre-Deployment Checklist

- [ ] Database is provisioned and accessible
- [ ] API keys are generated for client authentication
- [ ] Kubernetes cluster is available and healthy
- [ ] Docker registry credentials are configured
- [ ] SSL certificates are valid

### Build and Push Docker Image

```bash
# Build the image
docker build -t cloudcostguard/backend:v1.0.0 -f backend/Dockerfile .

# Tag for registry
docker tag cloudcostguard/backend:v1.0.0 registry.example.com/cloudcostguard/backend:v1.0.0
docker tag cloudcostguard/backend:v1.0.0 registry.example.com/cloudcostguard/backend:latest

# Push to registry
docker push registry.example.com/cloudcostguard/backend:v1.0.0
docker push registry.example.com/cloudcostguard/backend:latest
```

### Database Setup

```bash
# Create Kubernetes secret for database credentials
kubectl create secret generic db-credentials \
  --from-literal=user=postgres \
  --from-literal=password=$(openssl rand -base64 32) \
  --from-literal=host=postgres.default.svc.cluster.local \
  --from-literal=port=5432

# Run migrations
docker run -it --rm \
  -e DATABASE_URL="postgres://user:password@host:5432/cloudcostguard" \
  registry.example.com/cloudcostguard/backend:v1.0.0 \
  migrate
```

### Deploy to Kubernetes

```bash
# Update image references in deployment manifest
sed -i 's|cloudcostguard/backend:latest|registry.example.com/cloudcostguard/backend:v1.0.0|g' k8s/deployment.yaml

# Create namespace
kubectl create namespace cloudcostguard

# Apply manifests
kubectl apply -f k8s/ -n cloudcostguard

# Verify deployment
kubectl rollout status deployment/cloudcostguard-backend -n cloudcostguard
kubectl get pods -n cloudcostguard

# Verify health
kubectl port-forward -n cloudcostguard svc/cloudcostguard-backend 8080:8080 &
curl http://localhost:8080/health/ready
```

---

## Configuration

### Environment Variables

Required environment variables for production:

```bash
# Database
DATABASE_URL=postgres://user:password@host:5432/cloudcostguard

# API Configuration
CCG_API_KEYS=key1,key2,key3  # Comma-separated API keys
CCG_RATE_LIMIT_PER_SECOND=100
CCG_RATE_LIMIT_BURST=50

# Server
CCG_SERVER_PORT=8080
CCG_SERVER_READ_TIMEOUT=30s
CCG_SERVER_WRITE_TIMEOUT=30s
CCG_SERVER_SHUTDOWN_TIMEOUT=30s

# Observability
OTEL_ENABLED=true
OTEL_EXPORTER_TYPE=jaeger
OTEL_JAEGER_ENDPOINT=http://jaeger-collector:4318/v1/traces

# Cache
CCG_CACHE_REFRESH_INTERVAL=24h
```

### Kubernetes ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cloudcostguard-config
  namespace: cloudcostguard
data:
  CCG_SERVER_PORT: "8080"
  OTEL_ENABLED: "true"
  OTEL_EXPORTER_TYPE: "jaeger"
  OTEL_JAEGER_ENDPOINT: "http://jaeger-collector:4318/v1/traces"
```

---

## Monitoring & Observability

### Metrics Endpoints

- **Prometheus Metrics:** `http://localhost:8080/metrics`
- **Health Check (Live):** `http://localhost:8080/health/live`
- **Readiness Check:** `http://localhost:8080/health/ready`

### Key Metrics to Monitor

**HTTP Performance:**
- `http_requests_total` - Total requests by status and path
- `http_request_duration_seconds` - Request latency histogram
- `http_errors_total` - API errors by type and endpoint

**Business Logic:**
- `estimations_total` - Total estimations processed
- `cost_estimated_usd_total` - Total estimated costs
- `recommendations_total` - Recommendations generated
- `potential_savings_usd_total` - Potential savings identified

**Infrastructure:**
- `db_query_duration_seconds` - Database query latency
- `pricing_cache_hit_rate` - Cache hit percentage
- `active_estimations` - Active estimation operations

### Distributed Tracing with Jaeger

View traces at `http://jaeger-ui:16686`

Example query to find slow requests:
```
service.name:cloudcostguard-backend AND duration>1s
```

### Logging

Structured JSON logging includes:
```json
{
  "timestamp": "2024-01-01T00:00:00Z",
  "level": "info",
  "request_id": "uuid",
  "method": "POST",
  "path": "/estimate",
  "status_code": 200,
  "duration_ms": 250,
  "user_agent": "cli/1.0"
}
```

Access logs:
```bash
kubectl logs -f deployment/cloudcostguard-backend -n cloudcostguard
kubectl logs -f deployment/cloudcostguard-backend -n cloudcostguard --tail=100 | grep ERROR
```

---

## Troubleshooting

### High Error Rate (>5%)

**Investigation Steps:**

1. **Check error types:**
   ```bash
   kubectl logs -f deploy/cloudcostguard-backend -n cloudcostguard | grep ERROR | head -20
   ```

2. **Query error metrics:**
   ```promql
   rate(http_errors_total[5m])
   ```

3. **Common causes and solutions:**

| Error Type | Cause | Solution |
|-----------|-------|----------|
| 401 Unauthorized | Invalid API key | Verify CCG_API_KEYS config |
| 503 Service Unavailable | Pricing cache not initialized | Wait 5-10 min for cache population |
| 500 Internal Server Error | Database connection failed | Check database credentials and network |
| 429 Too Many Requests | Rate limit exceeded | Reduce client request rate or increase limits |

### High Latency (>1000ms)

**Investigation Steps:**

1. **Check database performance:**
   ```bash
   kubectl exec -it <postgres-pod> -n cloudcostguard -- \
    psql -U postgres -d cloudcostguard -c "\dt"
   ```

2. **Analyze slow queries:**
   ```promql
   histogram_quantile(0.95, db_query_duration_seconds)
   ```

3. **Check pricing cache:**
   ```bash
   curl http://localhost:8080/health/ready -i
   ```

### Pod Not Starting

**Check pod status:**
```bash
kubectl describe pod <pod-name> -n cloudcostguard
kubectl logs <pod-name> -n cloudcostguard --previous
```

**Common issues:**
- Image pull failed: Check registry credentials
- Resource limits exceeded: Increase memory/CPU in deployment
- Database connection failed: Verify DATABASE_URL

### Disk Space Issues

```bash
# Check disk usage
kubectl exec -it <pod> -n cloudcostguard -- df -h

# Clean up old logs
kubectl logs --tail=0 -f deploy/cloudcostguard-backend -n cloudcostguard > /dev/null

# Remove old pods
kubectl delete pod <pod-name> -n cloudcostguard
```

---

## Incident Response

### Page Duty Alert Workflow

1. **Acknowledge Alert** (within 5 minutes)
   - Document alert time and affected service
   - Assign team member to investigate

2. **Assess Severity**
   - Critical: Service unavailable to all users
   - High: Service degraded or errors >10%
   - Medium: Performance issues or partial outages
   - Low: Warnings or minor issues

3. **Investigate Root Cause**
   - Check recent deployments: `kubectl rollout history deploy/cloudcostguard-backend`
   - Review application logs and metrics
   - Check external dependencies (database, pricing API)

4. **Execute Remediation**

**For Database Issues:**
```bash
kubectl port-forward -n cloudcostguard svc/postgres 5432:5432 &
psql -h localhost -U postgres -d cloudcostguard -c "SELECT version();"
```

**For High Error Rate:**
```bash
# Scale down pods to clear state
kubectl scale deployment cloudcostguard-backend --replicas=1 -n cloudcostguard
sleep 30

# Scale back up
kubectl scale deployment cloudcostguard-backend --replicas=3 -n cloudcostguard
```

**For Memory Issues:**
```bash
# Check memory usage
kubectl top pod -n cloudcostguard

# Restart pod
kubectl rollout restart deployment/cloudcostguard-backend -n cloudcostguard
```

5. **Post-Incident Review**
   - Document timeline and impact
   - Identify prevention measures
   - Create follow-up tickets

---

## Disaster Recovery

### Backup Database

```bash
# Create backup
kubectl exec -it <postgres-pod> -n cloudcostguard -- \
  pg_dump -U postgres cloudcostguard > backup-$(date +%Y%m%d).sql

# Store in external location
gsutil cp backup-*.sql gs://cloudcostguard-backups/
```

### Restore from Backup

```bash
# Connect to database pod
kubectl exec -it <postgres-pod> -n cloudcostguard -- psql -U postgres

# Restore dump
psql -U postgres cloudcostguard < backup-YYYYMMDD.sql
```

### Rollback Deployment

```bash
# View rollout history
kubectl rollout history deployment/cloudcostguard-backend -n cloudcostguard

# Rollback to previous version
kubectl rollout undo deployment/cloudcostguard-backend -n cloudcostguard

# Verify rollback
kubectl rollout status deployment/cloudcostguard-backend -n cloudcostguard
```

---

## Performance Tuning

### Optimize Database Queries

```bash
# Enable query logging
kubectl exec -it <postgres-pod> -n cloudcostguard -- \
  psql -U postgres -d cloudcostguard -c \
  "ALTER SYSTEM SET log_min_duration_statement = 1000;"

# Check slow queries
kubectl logs <postgres-pod> -n cloudcostguard | grep duration
```

### Increase Cache Refresh Interval

For high-traffic environments, increase cache refresh:
```bash
kubectl set env deployment/cloudcostguard-backend \
  CCG_CACHE_REFRESH_INTERVAL=48h -n cloudcostguard
```

### Scale Horizontally

```bash
# Increase replicas
kubectl scale deployment cloudcostguard-backend --replicas=5 -n cloudcostguard

# Configure horizontal pod autoscaler
kubectl autoscale deployment cloudcostguard-backend \
  --min=3 --max=10 --cpu-percent=80 -n cloudcostguard
```

### Resource Requests & Limits

Recommended for production:

```yaml
resources:
  requests:
    memory: "512Mi"
    cpu: "500m"
  limits:
    memory: "1Gi"
    cpu: "1000m"
```

---

## Maintenance

### Regular Tasks

**Daily:**
- Monitor error rate dashboard
- Check system health metrics
- Review application logs for warnings

**Weekly:**
- Review performance metrics
- Check database size and growth
- Validate backup integrity

**Monthly:**
- Performance optimization review
- Security audit of API keys
- Capacity planning analysis
