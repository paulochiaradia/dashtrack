# Audit Logs System - Testing & Validation Guide

## Overview

This document provides comprehensive testing procedures for the audit logs system to ensure production readiness.

**Testing Phases:**
1. Unit Tests (Repository & Service)
2. Integration Tests (API Endpoints)
3. Middleware Tests (Automatic Capture)
4. Metrics Validation (Prometheus)
5. Performance Tests (Load & Stress)
6. End-to-End Tests (Full Flow)

---

## Prerequisites

```powershell
# Ensure services are running
docker-compose up -d

# Verify API is running
curl http://localhost:8080/health

# Verify Prometheus
curl http://localhost:9090/-/healthy

# Verify Jaeger
curl http://localhost:16686
```

---

## Phase 1: Unit Tests

### Repository Tests

Test file: `tests/unit/repository/audit_log_test.go`

```go
package repository_test

import (
    "testing"
    "time"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
)

type AuditLogRepositoryTestSuite struct {
    suite.Suite
    repo *repository.AuditLogRepository
    db   *gorm.DB
}

func (suite *AuditLogRepositoryTestSuite) SetupTest() {
    // Setup test database
}

func (suite *AuditLogRepositoryTestSuite) TestCreate() {
    log := &models.AuditLog{
        UserID:       1,
        CompanyID:    1,
        Action:       "CREATE",
        ResourceType: "users",
        Method:       "POST",
        Path:         "/api/users",
    }
    
    err := suite.repo.Create(log)
    assert.NoError(suite.T(), err)
    assert.NotZero(suite.T(), log.ID)
}

func (suite *AuditLogRepositoryTestSuite) TestList() {
    filters := map[string]interface{}{
        "user_id": 1,
        "action":  "CREATE",
    }
    
    logs, err := suite.repo.List(filters, 10, 0)
    assert.NoError(suite.T(), err)
    assert.NotEmpty(suite.T(), logs)
}

func (suite *AuditLogRepositoryTestSuite) TestGetStats() {
    stats, err := suite.repo.GetStats(time.Now().AddDate(0, 0, -7), time.Now())
    assert.NoError(suite.T(), err)
    assert.NotNil(suite.T(), stats)
}
```

**Run Unit Tests:**
```powershell
go test ./tests/unit/repository/... -v
```

---

## Phase 2: Integration Tests

### Test All API Endpoints

#### 1. List Audit Logs (GET /api/audit/logs)

```powershell
# Basic request
curl http://localhost:8080/api/audit/logs

# With pagination
curl "http://localhost:8080/api/audit/logs?limit=20&offset=0"

# Filter by user
curl "http://localhost:8080/api/audit/logs?user_id=1"

# Filter by action
curl "http://localhost:8080/api/audit/logs?action=DELETE"

# Filter by date range
curl "http://localhost:8080/api/audit/logs?start_date=2024-01-01&end_date=2024-12-31"

# Combined filters
curl "http://localhost:8080/api/audit/logs?user_id=1&action=UPDATE&resource_type=users"
```

**Expected Response:**
```json
{
  "logs": [
    {
      "id": 1,
      "user_id": 1,
      "company_id": 1,
      "action": "CREATE",
      "resource_type": "users",
      "resource_id": "123",
      "method": "POST",
      "path": "/api/users",
      "status_code": 201,
      "duration_ms": 45,
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 100,
  "limit": 10,
  "offset": 0
}
```

#### 2. Get Single Log (GET /api/audit/logs/:id)

```powershell
curl http://localhost:8080/api/audit/logs/1
```

**Expected Response:**
```json
{
  "id": 1,
  "user_id": 1,
  "user_email": "admin@example.com",
  "company_id": 1,
  "action": "CREATE",
  "resource_type": "users",
  "resource_id": "123",
  "method": "POST",
  "path": "/api/users",
  "status_code": 201,
  "ip_address": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "changes": {
    "before": null,
    "after": {"name": "New User", "email": "new@example.com"}
  },
  "metadata": {
    "request_id": "abc123",
    "session_id": "xyz789"
  },
  "trace_id": "a1b2c3d4e5f6",
  "span_id": "123456",
  "duration_ms": 45,
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### 3. Get Statistics (GET /api/audit/stats)

```powershell
# Last 7 days
curl http://localhost:8080/api/audit/stats

# Custom date range
curl "http://localhost:8080/api/audit/stats?start_date=2024-01-01&end_date=2024-01-31"

# By action type
curl "http://localhost:8080/api/audit/stats?group_by=action"

# By resource type
curl "http://localhost:8080/api/audit/stats?group_by=resource_type"
```

**Expected Response:**
```json
{
  "total_actions": 1500,
  "actions_by_type": {
    "CREATE": 400,
    "READ": 800,
    "UPDATE": 250,
    "DELETE": 50
  },
  "actions_by_resource": {
    "users": 500,
    "companies": 300,
    "sensors": 700
  },
  "avg_duration_ms": 125,
  "error_rate": 0.02,
  "start_date": "2024-01-01T00:00:00Z",
  "end_date": "2024-01-31T23:59:59Z"
}
```

#### 4. Get Timeline (GET /api/audit/timeline)

```powershell
# Daily timeline
curl "http://localhost:8080/api/audit/timeline?interval=daily"

# Hourly timeline
curl "http://localhost:8080/api/audit/timeline?interval=hourly"

# With date range
curl "http://localhost:8080/api/audit/timeline?interval=daily&start_date=2024-01-01&end_date=2024-01-31"
```

**Expected Response:**
```json
{
  "timeline": [
    {
      "timestamp": "2024-01-01T00:00:00Z",
      "total_actions": 150,
      "actions_by_type": {
        "CREATE": 40,
        "READ": 80,
        "UPDATE": 25,
        "DELETE": 5
      }
    },
    {
      "timestamp": "2024-01-02T00:00:00Z",
      "total_actions": 180,
      "actions_by_type": {
        "CREATE": 50,
        "READ": 90,
        "UPDATE": 35,
        "DELETE": 5
      }
    }
  ]
}
```

#### 5. Get User Logs (GET /api/audit/users/:id/logs)

```powershell
curl http://localhost:8080/api/audit/users/1/logs

# With pagination
curl "http://localhost:8080/api/audit/users/1/logs?limit=20&offset=0"

# Filter by action
curl "http://localhost:8080/api/audit/users/1/logs?action=DELETE"
```

#### 6. Get Resource Logs (GET /api/audit/resources/:type)

```powershell
curl http://localhost:8080/api/audit/resources/users

# With resource ID
curl "http://localhost:8080/api/audit/resources/users?resource_id=123"

# Filter by action
curl "http://localhost:8080/api/audit/resources/users?action=UPDATE"
```

#### 7. Get Trace Logs (GET /api/audit/traces/:trace_id)

```powershell
curl http://localhost:8080/api/audit/traces/a1b2c3d4e5f6
```

**Expected Response:**
```json
{
  "trace_id": "a1b2c3d4e5f6",
  "logs": [
    {
      "id": 1,
      "span_id": "123456",
      "action": "CREATE",
      "resource_type": "users",
      "duration_ms": 45
    },
    {
      "id": 2,
      "span_id": "123457",
      "action": "CREATE",
      "resource_type": "companies",
      "duration_ms": 30
    }
  ],
  "total_duration_ms": 75
}
```

#### 8. Export Logs (GET /api/audit/export)

```powershell
# Export as JSON
curl "http://localhost:8080/api/audit/export?format=json" -o audit_logs.json

# Export as CSV
curl "http://localhost:8080/api/audit/export?format=csv" -o audit_logs.csv

# With filters
curl "http://localhost:8080/api/audit/export?format=json&action=DELETE&start_date=2024-01-01" -o delete_logs.json
```

**Expected CSV Format:**
```csv
ID,UserID,UserEmail,CompanyID,Action,ResourceType,ResourceID,Method,Path,StatusCode,IPAddress,DurationMS,CreatedAt
1,1,admin@example.com,1,CREATE,users,123,POST,/api/users,201,192.168.1.100,45,2024-01-15T10:30:00Z
```

---

## Phase 3: Middleware Tests

### Test Automatic Audit Capture

Create test requests to verify middleware captures all actions:

```powershell
# Test CREATE action
curl -X POST http://localhost:8080/api/users `
  -H "Content-Type: application/json" `
  -H "Authorization: Bearer YOUR_TOKEN" `
  -d '{"name":"Test User","email":"test@example.com"}'

# Test READ action
curl http://localhost:8080/api/users/1 `
  -H "Authorization: Bearer YOUR_TOKEN"

# Test UPDATE action
curl -X PUT http://localhost:8080/api/users/1 `
  -H "Content-Type: application/json" `
  -H "Authorization: Bearer YOUR_TOKEN" `
  -d '{"name":"Updated Name"}'

# Test DELETE action
curl -X DELETE http://localhost:8080/api/users/1 `
  -H "Authorization: Bearer YOUR_TOKEN"

# Verify logs were created
curl http://localhost:8080/api/audit/logs?limit=10
```

### Test Sensitive Data Sanitization

```powershell
# Create user with password
curl -X POST http://localhost:8080/api/auth/register `
  -H "Content-Type: application/json" `
  -d '{"email":"test@example.com","password":"SecretPassword123"}'

# Check audit log
curl http://localhost:8080/api/audit/logs?action=CREATE&resource_type=users

# Verify password is sanitized in changes field
# Expected: "password": "[REDACTED]"
```

### Test Skip Routes

```powershell
# These should NOT be logged (health checks, metrics)
curl http://localhost:8080/health
curl http://localhost:8080/metrics
curl http://localhost:8080/favicon.ico

# Verify no logs created
curl "http://localhost:8080/api/audit/logs?path=/health"
# Expected: empty results
```

---

## Phase 4: Metrics Validation

### Test Prometheus Metrics

```powershell
# Check all audit metrics
curl http://localhost:8080/metrics | Select-String "audit_"

# Expected metrics:
# audit_actions_total{action="CREATE",resource_type="users"} 10
# audit_action_duration_seconds_bucket{action="CREATE",le="0.1"} 8
# audit_errors_total{action="CREATE",error_type="validation"} 2
# audit_authentication_total{success="true"} 15
# audit_suspicious_activity_total{activity_type="rapid_deletes"} 1
# audit_middleware_processing_duration_seconds_count 100
# audit_database_writes_total 95
# audit_database_write_errors_total 5
# audit_queue_size 0
# audit_slow_requests_total{path="/api/users",method="POST"} 3
```

### Test Metric Increments

```powershell
# Get current metric value
$before = (curl http://localhost:8080/metrics | Select-String "audit_actions_total").ToString()

# Create action
curl -X POST http://localhost:8080/api/users `
  -H "Content-Type: application/json" `
  -H "Authorization: Bearer YOUR_TOKEN" `
  -d '{"name":"Metric Test"}'

# Wait 2 seconds for metric update
Start-Sleep -Seconds 2

# Check metric incremented
$after = (curl http://localhost:8080/metrics | Select-String "audit_actions_total").ToString()

# Verify $after value > $before value
```

### Test in Prometheus UI

1. Open http://localhost:9090
2. Query: `audit_actions_total`
3. Verify data is being collected
4. Query: `rate(audit_actions_total[5m])`
5. View graph showing request rate

---

## Phase 5: Performance Tests

### Load Test with Apache Bench

```powershell
# Install Apache Bench (if not installed)
# Download from: https://httpd.apache.org/download.cgi

# Test 1: Simple GET requests
ab -n 1000 -c 10 http://localhost:8080/api/audit/logs

# Test 2: Authenticated requests
ab -n 1000 -c 10 -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/api/audit/logs

# Test 3: Concurrent POST requests
ab -n 500 -c 20 -p test_data.json -T "application/json" -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/api/users
```

**Expected Results:**
- Requests per second: > 100
- Mean response time: < 200ms
- P95 response time: < 500ms
- Failed requests: < 1%

### Load Test with k6

Create `tests/performance/audit_load_test.js`:

```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  vus: 50, // 50 virtual users
  duration: '2m', // 2 minutes
};

export default function () {
  // Test list endpoint
  let res1 = http.get('http://localhost:8080/api/audit/logs?limit=20');
  check(res1, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });

  sleep(1);

  // Test stats endpoint
  let res2 = http.get('http://localhost:8080/api/audit/stats');
  check(res2, {
    'status is 200': (r) => r.status === 200,
    'response time < 300ms': (r) => r.timings.duration < 300,
  });

  sleep(1);
}
```

**Run:**
```powershell
k6 run tests/performance/audit_load_test.js
```

### Stress Test

```javascript
export let options = {
  stages: [
    { duration: '1m', target: 50 },  // Ramp up to 50 users
    { duration: '3m', target: 50 },  // Stay at 50 users
    { duration: '1m', target: 100 }, // Ramp up to 100 users
    { duration: '3m', target: 100 }, // Stay at 100 users
    { duration: '1m', target: 200 }, // Spike to 200 users
    { duration: '2m', target: 0 },   // Ramp down
  ],
};
```

**Expected Results:**
- System remains stable under 100 concurrent users
- No memory leaks
- Database connection pool doesn't exhaust
- Queue size stays reasonable (<1000)

---

## Phase 6: End-to-End Tests

### Complete User Flow Test

```powershell
# Step 1: Register user
$registerResponse = curl -X POST http://localhost:8080/api/auth/register `
  -H "Content-Type: application/json" `
  -d '{"email":"e2e@example.com","password":"Test123!"}' | ConvertFrom-Json

# Step 2: Login
$loginResponse = curl -X POST http://localhost:8080/api/auth/login `
  -H "Content-Type: application/json" `
  -d '{"email":"e2e@example.com","password":"Test123!"}' | ConvertFrom-Json

$token = $loginResponse.token

# Step 3: Create resource
$createResponse = curl -X POST http://localhost:8080/api/users `
  -H "Content-Type: application/json" `
  -H "Authorization: Bearer $token" `
  -d '{"name":"E2E User"}' | ConvertFrom-Json

$userId = $createResponse.id

# Step 4: Update resource
curl -X PUT "http://localhost:8080/api/users/$userId" `
  -H "Content-Type: application/json" `
  -H "Authorization: Bearer $token" `
  -d '{"name":"Updated E2E User"}'

# Step 5: Delete resource
curl -X DELETE "http://localhost:8080/api/users/$userId" `
  -H "Authorization: Bearer $token"

# Step 6: Verify all logs created
$logs = curl "http://localhost:8080/api/audit/logs?user_email=e2e@example.com" | ConvertFrom-Json

# Expected: 5 logs (REGISTER, LOGIN, CREATE, UPDATE, DELETE)
Write-Host "Total logs created: $($logs.total)"
```

### Jaeger Trace Correlation Test

```powershell
# Make request with trace
$response = curl http://localhost:8080/api/users `
  -H "Authorization: Bearer $token" `
  -H "X-Trace-Id: test-trace-123"

# Get audit log
$logs = curl "http://localhost:8080/api/audit/traces/test-trace-123" | ConvertFrom-Json

# Open Jaeger UI
Start-Process "http://localhost:16686/trace/test-trace-123"

# Verify trace_id matches in both systems
```

---

## Testing Checklist

### ✅ Functionality Tests
- [ ] All 8 API endpoints respond correctly
- [ ] Pagination works (limit, offset)
- [ ] Filters work (user_id, action, resource_type, dates)
- [ ] Export works (JSON and CSV formats)
- [ ] Middleware captures all actions automatically
- [ ] Sensitive data is sanitized
- [ ] Skip routes are not logged

### ✅ Data Integrity Tests
- [ ] All required fields are populated
- [ ] JSONB fields (changes, metadata) parse correctly
- [ ] Timestamps are accurate
- [ ] Foreign keys (user_id, company_id) are valid
- [ ] Trace IDs correlate with Jaeger

### ✅ Performance Tests
- [ ] Response time < 200ms (P95)
- [ ] Throughput > 100 req/s
- [ ] No memory leaks under load
- [ ] Database queries use indexes
- [ ] Async logging doesn't block requests

### ✅ Metrics Tests
- [ ] All 17 metrics are registered
- [ ] Metrics increment correctly
- [ ] Prometheus scrapes successfully
- [ ] Histogram buckets configured correctly
- [ ] Labels are accurate

### ✅ Security Tests
- [ ] Authentication required for endpoints
- [ ] Authorization enforced (user can only see own logs unless admin)
- [ ] SQL injection protection
- [ ] XSS protection in responses
- [ ] Rate limiting works

### ✅ Integration Tests
- [ ] Grafana dashboards display data
- [ ] Jaeger traces correlate
- [ ] Prometheus alerts fire correctly
- [ ] Loki logs are searchable
- [ ] Docker Compose stack runs smoothly

---

## Troubleshooting

### No Logs Appearing

1. **Check middleware is registered:**
   ```powershell
   curl http://localhost:8080/api/users -v
   # Look for X-Request-Id header
   ```

2. **Check database connection:**
   ```powershell
   docker exec -it dashtrack-db psql -U postgres -d dashtrack -c "SELECT COUNT(*) FROM audit_logs;"
   ```

3. **Check service logs:**
   ```powershell
   docker logs dashtrack-api | Select-String "audit"
   ```

### Metrics Not Updating

1. **Check metrics endpoint:**
   ```powershell
   curl http://localhost:8080/metrics | Select-String "audit_"
   ```

2. **Check Prometheus config:**
   ```powershell
   curl http://localhost:9090/api/v1/targets
   ```

3. **Force metric increment:**
   ```powershell
   # Make multiple requests
   1..10 | ForEach-Object { curl http://localhost:8080/api/audit/logs }
   ```

### Performance Issues

1. **Check database indexes:**
   ```sql
   SELECT * FROM pg_indexes WHERE tablename = 'audit_logs';
   ```

2. **Check query performance:**
   ```sql
   EXPLAIN ANALYZE 
   SELECT * FROM audit_logs 
   WHERE user_id = 1 
   ORDER BY created_at DESC 
   LIMIT 10;
   ```

3. **Check async queue:**
   ```powershell
   curl http://localhost:8080/metrics | Select-String "audit_queue_size"
   ```

---

## Automated Test Script

Create `tests/run_all_tests.ps1`:

```powershell
#!/usr/bin/env pwsh

Write-Host "=== Audit Logs System - Full Test Suite ===" -ForegroundColor Cyan

# 1. Unit Tests
Write-Host "`n[1/6] Running unit tests..." -ForegroundColor Yellow
go test ./tests/unit/... -v

# 2. Integration Tests
Write-Host "`n[2/6] Running integration tests..." -ForegroundColor Yellow
go test ./tests/integration/... -v

# 3. API Endpoint Tests
Write-Host "`n[3/6] Testing API endpoints..." -ForegroundColor Yellow
./tests/api_test.ps1

# 4. Metrics Validation
Write-Host "`n[4/6] Validating Prometheus metrics..." -ForegroundColor Yellow
./tests/metrics_test.ps1

# 5. Performance Tests
Write-Host "`n[5/6] Running performance tests..." -ForegroundColor Yellow
k6 run tests/performance/audit_load_test.js

# 6. E2E Tests
Write-Host "`n[6/6] Running end-to-end tests..." -ForegroundColor Yellow
./tests/e2e_test.ps1

Write-Host "`n=== All Tests Complete ===" -ForegroundColor Green
```

**Run all tests:**
```powershell
.\tests\run_all_tests.ps1
```

---

## Success Criteria

✅ **Phase 8 Complete When:**
- [ ] All 8 API endpoints tested and working
- [ ] All filters and pagination tested
- [ ] Export (JSON/CSV) tested
- [ ] Middleware capture verified
- [ ] All 17 metrics validated
- [ ] Performance benchmarks met (>100 req/s, <200ms P95)
- [ ] Load tests pass (100 concurrent users)
- [ ] E2E flows complete successfully
- [ ] Jaeger correlation verified
- [ ] Grafana dashboards display data
- [ ] No memory leaks detected
- [ ] Zero critical bugs

---

**Last Updated:** 2024  
**Version:** 1.0  
**Status:** Ready for Execution
