# Audit Logs System - Quick Reference

## 🚀 Quick Start

### Access Endpoints
```bash
# List audit logs (requires authentication)
curl http://localhost:8080/api/v1/audit/logs \
  -H "Authorization: Bearer YOUR_TOKEN"

# Get statistics
curl http://localhost:8080/api/v1/audit/stats \
  -H "Authorization: Bearer YOUR_TOKEN"

# Export logs
curl "http://localhost:8080/api/v1/audit/export?format=json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -o audit_logs.json

# View metrics (no auth required)
curl http://localhost:8080/metrics | Select-String "audit_"
```

### Import Grafana Dashboards
1. Open http://localhost:3000
2. Login (admin/admin)
3. Go to **Dashboards** → **Import**
4. Upload files from `monitoring/grafana/dashboards/`
5. Select Prometheus datasource
6. Click **Import**

## 📊 API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/audit/logs` | GET | List audit logs with filters |
| `/api/v1/audit/logs/:id` | GET | Get single audit log |
| `/api/v1/audit/stats` | GET | Get aggregate statistics |
| `/api/v1/audit/timeline` | GET | Get time-series data |
| `/api/v1/audit/users/:id/logs` | GET | Get user activity logs |
| `/api/v1/audit/resources/:type` | GET | Get resource activity |
| `/api/v1/audit/traces/:traceId` | GET | Get logs by trace ID |
| `/api/v1/audit/export` | GET | Export logs (JSON/CSV) |

## 🔍 Common Filters

```bash
# Filter by user
?user_id=1

# Filter by action
?action=DELETE

# Filter by resource
?resource_type=users

# Filter by date range
?start_date=2024-01-01&end_date=2024-12-31

# Pagination
?limit=20&offset=0

# Combined filters
?user_id=1&action=UPDATE&limit=50
```

## 📈 Prometheus Metrics

### Key Metrics
```promql
# Total actions
audit_actions_total

# Response time percentiles
histogram_quantile(0.95, rate(audit_action_duration_seconds_bucket[5m]))

# Error rate
sum(rate(audit_errors_total[5m])) / sum(rate(audit_actions_total[5m]))

# Top users
topk(10, sum(increase(audit_user_actions_total[24h])) by (user_email))

# Suspicious activities
sum(increase(audit_suspicious_activity_total[1h])) by (activity_type)
```

## 🎨 Grafana Dashboards

### Dashboard 1: Audit Overview
- **UID:** `audit-overview`
- **Purpose:** General operational metrics
- **Panels:** 10 (stats, graphs, tables)

### Dashboard 2: Security Monitoring
- **UID:** `audit-security`
- **Purpose:** Security threat detection
- **Panels:** 11 (failed logins, suspicious activities)

### Dashboard 3: Performance Analytics
- **UID:** `audit-performance`
- **Purpose:** Performance optimization
- **Panels:** 13 (response times, throughput, errors)

## 🔧 Troubleshooting

### No Data in Dashboards
```bash
# 1. Check API is running
curl http://localhost:8080/health

# 2. Check metrics endpoint
curl http://localhost:8080/metrics | Select-String "audit_"

# 3. Check Prometheus is scraping
curl http://localhost:9090/api/v1/targets

# 4. Generate some traffic
curl http://localhost:8080/api/v1/audit/logs \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Database Queries
```sql
-- Count total logs
SELECT COUNT(*) FROM audit_logs;

-- Get recent logs
SELECT * FROM audit_logs 
ORDER BY created_at DESC 
LIMIT 10;

-- Get logs by user
SELECT * FROM audit_logs 
WHERE user_id = 1 
ORDER BY created_at DESC;

-- Get DELETE operations
SELECT * FROM audit_logs 
WHERE action = 'DELETE' 
ORDER BY created_at DESC;

-- Get failed requests
SELECT * FROM audit_logs 
WHERE status_code >= 400 
ORDER BY created_at DESC;
```

## 📚 Documentation

| File | Purpose |
|------|---------|
| `AUDIT_ARCHITECTURE.md` | System design & architecture |
| `AUDIT_COMPLETE.md` | Project completion summary |
| `AUDIT_TESTING_GUIDE.md` | Comprehensive testing procedures |
| `monitoring/grafana/dashboards/README.md` | Dashboard installation guide |

## 🎯 Quick Checks

### Health Check
```bash
# API
curl http://localhost:8080/health

# Database
docker exec dashtrack-db-1 psql -U postgres -d dashtrack -c "SELECT 1;"

# Jaeger
curl http://localhost:16686

# Grafana
curl http://localhost:3000/api/health
```

### View Logs
```bash
# API logs
docker logs dashtrack-api-1 --tail 50

# Filter for audit-related logs
docker logs dashtrack-api-1 | Select-String "audit"

# Follow logs in real-time
docker logs dashtrack-api-1 -f
```

## 🚨 Alerts (Recommended)

Configure in Grafana:

1. **Failed logins > 10 in 1 hour**
   ```promql
   sum(increase(audit_authentication_total{success="false"}[1h])) > 10
   ```

2. **Suspicious activities > 5 in 1 hour**
   ```promql
   sum(increase(audit_suspicious_activity_total[1h])) > 5
   ```

3. **P95 response time > 500ms**
   ```promql
   histogram_quantile(0.95, rate(audit_action_duration_seconds_bucket[5m])) > 0.5
   ```

4. **Error rate > 5%**
   ```promql
   sum(rate(audit_errors_total[5m])) / sum(rate(audit_actions_total[5m])) > 0.05
   ```

## 📞 Support

- **Architecture:** See `AUDIT_ARCHITECTURE.md`
- **Testing:** See `AUDIT_TESTING_GUIDE.md`
- **Dashboards:** See `monitoring/grafana/dashboards/README.md`
- **Status:** See `AUDIT_STATUS.md`

---

**Version:** 1.0  
**Status:** Production-Ready  
**Last Updated:** 2024
