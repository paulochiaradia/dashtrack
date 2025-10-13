# ============================================
# DASHTRACK API - COMPREHENSIVE ENDPOINT TEST
# ============================================

$baseUrl = "http://localhost:8080/api/v1"
$passed = 0
$failed = 0
$results = @()

function Test-Endpoint {
    param(
        [string]$Method,
        [string]$Endpoint,
        [hashtable]$Headers = @{},
        [string]$Body = $null,
        [string]$Description
    )
    
    $fullUrl = "$baseUrl$Endpoint"
    
    try {
        $params = @{
            Uri = $fullUrl
            Method = $Method
            Headers = $Headers
            ContentType = "application/json"
        }
        
        if ($Body) {
            $params.Body = $Body
        }
        
        $response = Invoke-RestMethod @params
        $script:passed++
        $script:results += [PSCustomObject]@{
            Status = "PASS"
            Method = $Method
            Endpoint = $Endpoint
            Description = $Description
            StatusCode = "200"
        }
        Write-Host "[PASS] $Method $Endpoint - $Description" -ForegroundColor Green
        return $response
    }
    catch {
        $statusCode = $_.Exception.Response.StatusCode.Value__
        $script:failed++
        $script:results += [PSCustomObject]@{
            Status = "FAIL"
            Method = $Method
            Endpoint = $Endpoint
            Description = $Description
            StatusCode = $statusCode
            Error = $_.Exception.Message
        }
        Write-Host "[FAIL] $Method $Endpoint - Status: $statusCode - $Description" -ForegroundColor Red
        return $null
    }
}

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "STARTING COMPREHENSIVE ENDPOINT TESTS" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

# ============================================
# 1. AUTHENTICATION ENDPOINTS
# ============================================
Write-Host "`n--- AUTHENTICATION TESTS ---`n" -ForegroundColor Yellow

# 1.1 Login as Master
$loginBody = '{"email":"admin@test.com","password":"Admin@123"}'
$masterLogin = Test-Endpoint -Method "POST" -Endpoint "/auth/login" -Body $loginBody -Description "Login as Master"
$masterToken = $masterLogin.access_token
$masterHeaders = @{"Authorization" = "Bearer $masterToken"}

# 1.2 Login as Company Admin
$companyLoginBody = '{"email":"company@test.com","password":"Company@123"}'
$companyLogin = Test-Endpoint -Method "POST" -Endpoint "/auth/login" -Body $companyLoginBody -Description "Login as Company Admin"
$companyToken = $companyLogin.access_token
$companyHeaders = @{"Authorization" = "Bearer $companyToken"}

# 1.3 Login as Driver
$driverLoginBody = '{"email":"driver@test.com","password":"Driver@123"}'
$driverLogin = Test-Endpoint -Method "POST" -Endpoint "/auth/login" -Body $driverLoginBody -Description "Login as Driver"
$driverToken = $driverLogin.access_token
$driverHeaders = @{"Authorization" = "Bearer $driverToken"}

# ============================================
# 2. PROFILE ENDPOINTS
# ============================================
Write-Host "`n--- PROFILE TESTS ---`n" -ForegroundColor Yellow

Test-Endpoint -Method "GET" -Endpoint "/profile" -Headers $masterHeaders -Description "Get Master Profile"
Test-Endpoint -Method "GET" -Endpoint "/profile" -Headers $companyHeaders -Description "Get Company Admin Profile"
Test-Endpoint -Method "GET" -Endpoint "/profile" -Headers $driverHeaders -Description "Get Driver Profile"

# ============================================
# 3. ADMIN ENDPOINTS (Master & Admin Only)
# ============================================
Write-Host "`n--- ADMIN ENDPOINTS ---`n" -ForegroundColor Yellow

Test-Endpoint -Method "GET" -Endpoint "/admin/users" -Headers $masterHeaders -Description "List Users (Master)"
Test-Endpoint -Method "GET" -Endpoint "/admin/users" -Headers $companyHeaders -Description "List Users (Company Admin - Should Fail)"

# ============================================
# 4. COMPANY ADMIN - TEAMS
# ============================================
Write-Host "`n--- TEAMS ENDPOINTS ---`n" -ForegroundColor Yellow

Test-Endpoint -Method "GET" -Endpoint "/company-admin/teams" -Headers $companyHeaders -Description "List Teams (Company Admin)"
Test-Endpoint -Method "GET" -Endpoint "/company-admin/teams" -Headers $masterHeaders -Description "List Teams (Master)"
Test-Endpoint -Method "GET" -Endpoint "/company-admin/teams" -Headers $driverHeaders -Description "List Teams (Driver - Should Fail)"

# ============================================
# 5. COMPANY ADMIN - VEHICLES
# ============================================
Write-Host "`n--- VEHICLES ENDPOINTS ---`n" -ForegroundColor Yellow

Test-Endpoint -Method "GET" -Endpoint "/company-admin/vehicles" -Headers $companyHeaders -Description "List Vehicles (Company Admin)"
Test-Endpoint -Method "GET" -Endpoint "/company-admin/vehicles" -Headers $masterHeaders -Description "List Vehicles (Master)"
Test-Endpoint -Method "GET" -Endpoint "/company-admin/vehicles" -Headers $driverHeaders -Description "List Vehicles (Driver - Should Fail)"

# ============================================
# 6. TEAM-VEHICLE INTEGRATION
# ============================================
Write-Host "`n--- TEAM-VEHICLE INTEGRATION ---`n" -ForegroundColor Yellow

# Get first team and vehicle for testing
$teams = Test-Endpoint -Method "GET" -Endpoint "/company-admin/teams" -Headers $companyHeaders -Description "Get Teams for Integration Test"
$vehicles = Test-Endpoint -Method "GET" -Endpoint "/company-admin/vehicles" -Headers $companyHeaders -Description "Get Vehicles for Integration Test"

if ($teams -and $teams.data -and $teams.data.Count -gt 0 -and $vehicles -and $vehicles.data -and $vehicles.data.Count -gt 0) {
    $teamId = $teams.data[0].id
    $vehicleId = $vehicles.data[0].id
    
    # Assign vehicle to team
    $assignBody = "{`"vehicle_id`":`"$vehicleId`"}"
    Test-Endpoint -Method "POST" -Endpoint "/company-admin/teams/$teamId/vehicles" -Headers $companyHeaders -Body $assignBody -Description "Assign Vehicle to Team"
    
    # List team vehicles
    Test-Endpoint -Method "GET" -Endpoint "/company-admin/teams/$teamId/vehicles" -Headers $companyHeaders -Description "List Team Vehicles"
    
    # Get team stats
    Test-Endpoint -Method "GET" -Endpoint "/company-admin/teams/$teamId/stats" -Headers $companyHeaders -Description "Get Team Stats"
    
    # Unassign vehicle from team
    Test-Endpoint -Method "DELETE" -Endpoint "/company-admin/teams/$teamId/vehicles/$vehicleId" -Headers $companyHeaders -Description "Unassign Vehicle from Team"
}

# ============================================
# 7. AUDIT ENDPOINTS
# ============================================
Write-Host "`n--- AUDIT ENDPOINTS ---`n" -ForegroundColor Yellow

Test-Endpoint -Method "GET" -Endpoint "/audit/logs" -Headers $masterHeaders -Description "Get Audit Logs (Master)"
Test-Endpoint -Method "GET" -Endpoint "/audit/logs" -Headers $companyHeaders -Description "Get Audit Logs (Company Admin)"
Test-Endpoint -Method "GET" -Endpoint "/audit/stats" -Headers $masterHeaders -Description "Get Audit Stats (Master)"
Test-Endpoint -Method "GET" -Endpoint "/audit/stats" -Headers $companyHeaders -Description "Get Audit Stats (Company Admin)"

# ============================================
# 8. SESSION ENDPOINTS
# ============================================
Write-Host "`n--- SESSION ENDPOINTS ---`n" -ForegroundColor Yellow

Test-Endpoint -Method "GET" -Endpoint "/sessions/active" -Headers $masterHeaders -Description "Get Active Sessions (Master)"
Test-Endpoint -Method "GET" -Endpoint "/sessions/active" -Headers $companyHeaders -Description "Get Active Sessions (Company Admin)"
Test-Endpoint -Method "GET" -Endpoint "/sessions/dashboard" -Headers $masterHeaders -Description "Get Session Dashboard (Master)"

# ============================================
# 9. SECURITY ENDPOINTS
# ============================================
Write-Host "`n--- SECURITY ENDPOINTS ---`n" -ForegroundColor Yellow

Test-Endpoint -Method "GET" -Endpoint "/security/2fa/status" -Headers $masterHeaders -Description "Get 2FA Status (Master)"
Test-Endpoint -Method "GET" -Endpoint "/security/2fa/status" -Headers $companyHeaders -Description "Get 2FA Status (Company Admin)"

# ============================================
# 10. HEALTH & METRICS (Without /api/v1 prefix)
# ============================================
Write-Host "`n--- HEALTH AND METRICS ---`n" -ForegroundColor Yellow

# These endpoints are not under /api/v1
$healthUrl = "http://localhost:8080/health"
$metricsUrl = "http://localhost:8080/metrics"

try {
    $health = Invoke-RestMethod -Uri $healthUrl
    $script:passed++
    $script:results += [PSCustomObject]@{
        Status = "PASS"
        Method = "GET"
        Endpoint = "/health"
        Description = "Health Check (No Auth)"
        StatusCode = "200"
    }
    Write-Host "[PASS] GET /health - Health Check (No Auth)" -ForegroundColor Green
} catch {
    $script:failed++
    $script:results += [PSCustomObject]@{
        Status = "FAIL"
        Method = "GET"
        Endpoint = "/health"
        Description = "Health Check (No Auth)"
        StatusCode = $_.Exception.Response.StatusCode.Value__
        Error = $_.Exception.Message
    }
    Write-Host "[FAIL] GET /health - Status: $($_.Exception.Response.StatusCode.Value__) - Health Check (No Auth)" -ForegroundColor Red
}

try {
    $metrics = Invoke-RestMethod -Uri $metricsUrl
    $script:passed++
    $script:results += [PSCustomObject]@{
        Status = "PASS"
        Method = "GET"
        Endpoint = "/metrics"
        Description = "Prometheus Metrics (No Auth)"
        StatusCode = "200"
    }
    Write-Host "[PASS] GET /metrics - Prometheus Metrics (No Auth)" -ForegroundColor Green
} catch {
    $script:failed++
    $script:results += [PSCustomObject]@{
        Status = "FAIL"
        Method = "GET"
        Endpoint = "/metrics"
        Description = "Prometheus Metrics (No Auth)"
        StatusCode = $_.Exception.Response.StatusCode.Value__
        Error = $_.Exception.Message
    }
    Write-Host "[FAIL] GET /metrics - Status: $($_.Exception.Response.StatusCode.Value__) - Prometheus Metrics (No Auth)" -ForegroundColor Red
}

# ============================================
# FINAL REPORT
# ============================================
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "TEST RESULTS SUMMARY" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

Write-Host "Total Tests: $($passed + $failed)" -ForegroundColor White
Write-Host "Passed: $passed" -ForegroundColor Green
Write-Host "Failed: $failed" -ForegroundColor Red
$successRate = [math]::Round(($passed / ($passed + $failed)) * 100, 2)
Write-Host "Success Rate: $successRate%" -ForegroundColor $(if ($successRate -ge 80) { "Green" } else { "Yellow" })

# Show failed tests
if ($failed -gt 0) {
    Write-Host "`nFailed Tests:" -ForegroundColor Red
    $results | Where-Object { $_.Status -eq "FAIL" } | Format-Table -AutoSize
}

Write-Host "`n========================================`n" -ForegroundColor Cyan
