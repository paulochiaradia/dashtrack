# Grafana Dashboards - Audit Logs System

## Overview

This directory contains 3 comprehensive Grafana dashboards for monitoring the audit logs system.

## Dashboards

### 1. Audit Overview (`audit_overview.json`)
**UID:** `audit-overview`  
**Time Range:** Last 24 hours  
**Refresh:** 30 seconds

Main operational dashboard with general metrics:

- **Stats Panels:**
  - Total Actions (24h)
  - Total Errors (24h)
  - Active Users
  - P95 Response Time

- **Time Series:**
  - Actions Rate by Type
  - Response Time Percentiles (P50, P95, P99)
  - HTTP Status Codes Rate

- **Charts:**
  - Actions by Resource (donut chart)
  - Top 10 Most Active Users (bar chart)

- **Tables:**
  - Most Accessed Resources

**Variables:**
- `action`: Filter by action type
- `resource`: Filter by resource type

---

### 2. Security Monitoring (`audit_security.json`)
**UID:** `audit-security`  
**Time Range:** Last 24 hours  
**Refresh:** 30 seconds

Security-focused dashboard for threat detection:

- **Security Stats:**
  - Failed Logins (1h)
  - Suspicious Activities (1h)
  - DELETE Operations (24h)
  - Permission Denied (24h)

- **Security Trends:**
  - Failed Authentication Attempts (by action)
  - Suspicious Activity by Type
  - Error Rate by Type

- **Threat Analysis:**
  - Top Failed Login IPs (last hour)
  - Users with Most DELETE Operations (24h)
  - Recent Suspicious Activities
  - Error Distribution by Status Code

**Use Cases:**
- Monitor failed login attempts
- Detect brute force attacks
- Track destructive operations
- Identify unauthorized access attempts
- Security incident investigation

---

### 3. Performance Analytics (`audit_performance.json`)
**UID:** `audit-performance`  
**Time Range:** Last 6 hours  
**Refresh:** 30 seconds

Performance monitoring and optimization dashboard:

- **Response Time Metrics:**
  - P50, P95, P99 Response Time (stat panels)
  - Request Throughput
  - Response Time by Action Type
  - Throughput by Action Type

- **System Performance:**
  - Middleware Processing Time
  - Database Write Performance
  - Async Processing Queue Size

- **Performance Analysis:**
  - Slowest Resources (table)
  - Error Rate (gauge)
  - Slow Requests >1s (last hour)
  - Request Distribution by HTTP Method

**Use Cases:**
- Identify slow endpoints
- Monitor system throughput
- Track database write performance
- Detect performance degradation
- Optimize resource-heavy operations

---

## Installation

### Import via Grafana UI

1. Open Grafana web interface (default: http://localhost:3000)
2. Login with admin credentials
3. Navigate to **Dashboards** → **Import**
4. Click **Upload JSON file** or paste JSON content
5. Select Prometheus datasource
6. Click **Import**

### Import via API

```powershell
# Dashboard 1: Audit Overview
$body = Get-Content "monitoring/grafana/dashboards/audit_overview.json" -Raw
Invoke-RestMethod -Uri "http://localhost:3000/api/dashboards/db" -Method POST -Headers @{"Content-Type"="application/json"} -Body $body -Credential (Get-Credential)

# Dashboard 2: Security Monitoring
$body = Get-Content "monitoring/grafana/dashboards/audit_security.json" -Raw
Invoke-RestMethod -Uri "http://localhost:3000/api/dashboards/db" -Method POST -Headers @{"Content-Type"="application/json"} -Body $body -Credential (Get-Credential)

# Dashboard 3: Performance Analytics
$body = Get-Content "monitoring/grafana/dashboards/audit_performance.json" -Raw
Invoke-RestMethod -Uri "http://localhost:3000/api/dashboards/db" -Method POST -Headers @{"Content-Type"="application/json"} -Body $body -Credential (Get-Credential)
```

### Import via Provisioning

Add to `docker-compose.monitoring.yml` or Grafana configuration:

```yaml
volumes:
  - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards
```

---

## Configuration

### Prometheus Datasource

All dashboards require a Prometheus datasource named **"Prometheus"**.

Configure in Grafana:
1. **Configuration** → **Data Sources** → **Add data source**
2. Select **Prometheus**
3. URL: `http://prometheus:9090`
4. Click **Save & Test**

### Time Ranges

Default time ranges can be modified:
- **Audit Overview:** 24 hours
- **Security Monitoring:** 24 hours
- **Performance Analytics:** 6 hours

### Auto-Refresh

All dashboards auto-refresh every **30 seconds**. Adjust via dashboard settings if needed.

---

## Metrics Reference

### Core Metrics Used

| Metric | Type | Description |
|--------|------|-------------|
| `audit_actions_total` | Counter | Total audit actions by labels |
| `audit_action_duration_seconds` | Histogram | Request duration distribution |
| `audit_errors_total` | Counter | Total errors by type |
| `audit_authentication_total` | Counter | Authentication attempts |
| `audit_suspicious_activity_total` | Counter | Suspicious activities detected |
| `audit_middleware_processing_duration_seconds` | Histogram | Middleware processing time |
| `audit_database_writes_total` | Counter | Successful database writes |
| `audit_database_write_errors_total` | Counter | Failed database writes |
| `audit_queue_size` | Gauge | Async processing queue size |
| `audit_slow_requests_total` | Counter | Requests >1s |

### Labels

- `action`: CREATE, READ, UPDATE, DELETE
- `resource_type`: users, companies, sensors, etc.
- `method`: GET, POST, PUT, PATCH, DELETE
- `status_code`: 200, 201, 400, 401, 403, 404, 500
- `user_email`: User email address
- `company_id`: Company identifier
- `ip_address`: Client IP address
- `error_type`: authentication, validation, forbidden, etc.
- `activity_type`: Type of suspicious activity

---

## Usage Examples

### Monitor Failed Logins

Use **Security Monitoring** dashboard:
1. Check "Failed Logins (1h)" stat panel
2. Review "Failed Authentication Attempts" graph
3. Inspect "Top Failed Login IPs" table
4. Alert threshold: >10 in 1 hour

### Identify Performance Issues

Use **Performance Analytics** dashboard:
1. Check P95/P99 response times
2. Review "Slowest Resources" table
3. Monitor "Error Rate" gauge
4. Check "Slow Requests" for bottlenecks

### Track User Activity

Use **Audit Overview** dashboard:
1. View "Top 10 Most Active Users"
2. Filter by user using variables
3. Check "Actions Rate by Type"
4. Export data via /api/audit/export

### Security Incident Investigation

Use **Security Monitoring** dashboard:
1. Check "Suspicious Activities" stat
2. Review "Recent Suspicious Activities" table
3. Filter by time range (incident window)
4. Correlate with Jaeger traces via trace_id

---

## Alerting

### Recommended Alerts

Configure alerts in Grafana for:

**Security Alerts:**
- Failed logins > 10 in 1 hour
- Suspicious activities > 5 in 1 hour
- DELETE operations > 20 in 1 day
- Permission denied > 15 in 1 hour

**Performance Alerts:**
- P95 response time > 500ms
- Error rate > 5%
- Database write errors > 10 in 5 minutes
- Slow requests > 50 in 1 hour

**System Alerts:**
- Queue size > 1000
- Request throughput < 1 req/s (system idle)
- Middleware processing > 100ms

---

## Customization

### Adding Panels

1. Edit dashboard in Grafana UI
2. Click **Add panel**
3. Select **Time series** or other visualization
4. Add Prometheus query:
   ```promql
   sum(rate(audit_actions_total[5m])) by (label)
   ```
5. Configure thresholds and styling
6. Save dashboard

### Modifying Queries

All queries use PromQL. Common patterns:

```promql
# Rate of increase
rate(audit_actions_total[5m])

# Sum by label
sum(increase(audit_actions_total[1h])) by (action)

# Histogram quantile (percentiles)
histogram_quantile(0.95, rate(audit_action_duration_seconds_bucket[5m]))

# Top K resources
topk(10, sum(increase(audit_actions_total[24h])) by (resource_type))

# Error rate
sum(rate(audit_errors_total[5m])) / sum(rate(audit_actions_total[5m]))
```

### Exporting Dashboards

```powershell
# Export dashboard JSON
Invoke-RestMethod -Uri "http://localhost:3000/api/dashboards/uid/audit-overview" -Headers @{"Authorization"="Bearer YOUR_API_KEY"} | ConvertTo-Json -Depth 100 > audit_overview_backup.json
```

---

## Troubleshooting

### No Data Displayed

1. **Check Prometheus datasource:**
   ```bash
   curl http://localhost:9090/api/v1/query?query=audit_actions_total
   ```

2. **Verify metrics endpoint:**
   ```bash
   curl http://localhost:8080/metrics | grep audit_
   ```

3. **Check time range:** Ensure dashboard time range matches data availability

4. **Verify service running:** API must be running and processing requests

### Query Errors

- **"No data points"**: No metrics collected yet, generate traffic
- **"Parse error"**: Invalid PromQL syntax, check query
- **"Timeout"**: Query too complex, reduce time range or add filters

### Dashboard Not Loading

1. Check Grafana logs:
   ```bash
   docker logs grafana
   ```

2. Verify JSON syntax:
   ```powershell
   Get-Content audit_overview.json | ConvertFrom-Json
   ```

3. Re-import dashboard with valid JSON

---

## Best Practices

1. **Regular Reviews:**
   - Check Security dashboard daily
   - Review Performance dashboard weekly
   - Export reports monthly

2. **Alert Configuration:**
   - Set up critical alerts first
   - Avoid alert fatigue (tune thresholds)
   - Test alert delivery

3. **Dashboard Maintenance:**
   - Keep dashboards updated with new metrics
   - Remove unused panels
   - Document custom queries

4. **Data Retention:**
   - Configure Prometheus retention (default 15 days)
   - Export historical data for long-term analysis
   - Archive important dashboards

---

## Integration with Other Tools

### Jaeger Integration

Audit logs include `trace_id` and `span_id`:
1. Click action in dashboard
2. Copy trace_id from logs
3. Open Jaeger UI: http://localhost:16686
4. Search by trace_id
5. View full distributed trace

### Log Aggregation (Loki)

Query audit logs in Loki:
```logql
{job="dashtrack-api"} |= "audit" | json | action="DELETE"
```

### Alertmanager

Configure alerts to send to Alertmanager:
```yaml
receivers:
  - name: 'slack'
    slack_configs:
      - channel: '#security-alerts'
        text: 'Failed logins: {{ $value }}'
```

---

## Support

For issues or questions:
- Check documentation: `docs/AUDIT_ARCHITECTURE.md`
- Review metrics: `internal/metrics/audit.go`
- API reference: `internal/handlers/audit.go`

---

**Version:** 1.0  
**Last Updated:** 2024  
**Maintainer:** DevOps Team
