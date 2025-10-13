# Complete API Test Suite
# Date: October 13, 2025
# Purpose: Test ALL endpoints systematically

Write-Host "`n╔════════════════════════════════════════════════════════════╗" -ForegroundColor Magenta
Write-Host "║       TESTE COMPLETO DE TODOS OS ENDPOINTS               ║" -ForegroundColor Magenta
Write-Host "╚════════════════════════════════════════════════════════════╝`n" -ForegroundColor Magenta

# Initialize counters
$global:passedTests = 0
$global:failedTests = 0
$global:failedEndpoints = @()

function Test-Endpoint {
    param(
        [string]$Name,
        [string]$Method,
        [string]$Uri,
        [hashtable]$Headers,
        [string]$Body = $null,
        [int]$ExpectedStatus = 200
    )
    
    try {
        $params = @{
            Uri = $Uri
            Method = $Method
            Headers = $Headers
            ContentType = 'application/json'
        }
        
        if ($Body) {
            $params.Body = $Body
        }
        
        $response = Invoke-RestMethod @params
        Write-Host "✅ $Name" -ForegroundColor Green
        $global:passedTests++
        return $response
    }
    catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        $errorMsg = $_.Exception.Message
        Write-Host "❌ $Name" -ForegroundColor Red
        Write-Host "   Status: $statusCode" -ForegroundColor Gray
        Write-Host "   Error: $errorMsg" -ForegroundColor Gray
        $global:failedTests++
        $global:failedEndpoints += $Name
        return $null
    }
}

# ============================================================
# 1. AUTHENTICATION & SETUP
# ============================================================
Write-Host "`n[1] AUTHENTICATION" -ForegroundColor Cyan
Write-Host "═══════════════════" -ForegroundColor Cyan

$loginBody = @{email='company@test.com'; password='password123'} | ConvertTo-Json
$loginResponse = Test-Endpoint `
    -Name "POST /auth/login (Company Admin)" `
    -Method POST `
    -Uri 'http://localhost:8080/api/v1/auth/login' `
    -Headers @{} `
    -Body $loginBody

if ($loginResponse) {
    $global:companyToken = $loginResponse.access_token
    $headers = @{Authorization="Bearer $global:companyToken"}
} else {
    Write-Host "`n❌ CRITICAL: Cannot continue without authentication!" -ForegroundColor Red
    exit 1
}

# ============================================================
# 2. TEAM ENDPOINTS
# ============================================================
Write-Host "`n[2] TEAM MANAGEMENT" -ForegroundColor Cyan
Write-Host "═══════════════════" -ForegroundColor Cyan

# Create Team
$teamBody = @{name='Test Team'; description='Automated test team'} | ConvertTo-Json
$team = Test-Endpoint `
    -Name "POST /company-admin/teams (Create)" `
    -Method POST `
    -Uri 'http://localhost:8080/api/v1/company-admin/teams' `
    -Headers $headers `
    -Body $teamBody

if ($team) { $global:teamId = $team.data.id }

# List Teams
Test-Endpoint `
    -Name "GET /company-admin/teams (List)" `
    -Method GET `
    -Uri 'http://localhost:8080/api/v1/company-admin/teams' `
    -Headers $headers

# Get Team
if ($global:teamId) {
    Test-Endpoint `
        -Name "GET /company-admin/teams/:id (Get One)" `
        -Method GET `
        -Uri "http://localhost:8080/api/v1/company-admin/teams/$global:teamId" `
        -Headers $headers
}

# Update Team
if ($global:teamId) {
    $updateBody = @{name='Test Team Updated'; description='Updated description'} | ConvertTo-Json
    Test-Endpoint `
        -Name "PUT /company-admin/teams/:id (Update)" `
        -Method PUT `
        -Uri "http://localhost:8080/api/v1/company-admin/teams/$global:teamId" `
        -Headers $headers `
        -Body $updateBody
}

# Get Team Stats
if ($global:teamId) {
    Test-Endpoint `
        -Name "GET /company-admin/teams/:id/stats (Stats)" `
        -Method GET `
        -Uri "http://localhost:8080/api/v1/company-admin/teams/$global:teamId/stats" `
        -Headers $headers
}

# Get Team Members
if ($global:teamId) {
    Test-Endpoint `
        -Name "GET /company-admin/teams/:id/members (List Members)" `
        -Method GET `
        -Uri "http://localhost:8080/api/v1/company-admin/teams/$global:teamId/members" `
        -Headers $headers
}

# ============================================================
# 3. VEHICLE ENDPOINTS
# ============================================================
Write-Host "`n[3] VEHICLE MANAGEMENT" -ForegroundColor Cyan
Write-Host "═══════════════════════" -ForegroundColor Cyan

# Create Vehicle
$vehicleBody = @{
    license_plate='TEST-123'
    brand='Toyota'
    model='Corolla'
    year=2024
    vehicle_type='car'
    fuel_type='gasoline'
    cargo_capacity=500.0
} | ConvertTo-Json

$vehicle = Test-Endpoint `
    -Name "POST /company-admin/vehicles (Create)" `
    -Method POST `
    -Uri 'http://localhost:8080/api/v1/company-admin/vehicles' `
    -Headers $headers `
    -Body $vehicleBody

if ($vehicle) { $global:vehicleId = $vehicle.data.id }

# List Vehicles
Test-Endpoint `
    -Name "GET /company-admin/vehicles (List)" `
    -Method GET `
    -Uri 'http://localhost:8080/api/v1/company-admin/vehicles' `
    -Headers $headers

# Get Vehicle
if ($global:vehicleId) {
    Test-Endpoint `
        -Name "GET /company-admin/vehicles/:id (Get One)" `
        -Method GET `
        -Uri "http://localhost:8080/api/v1/company-admin/vehicles/$global:vehicleId" `
        -Headers $headers
}

# Update Vehicle
if ($global:vehicleId) {
    $updateVehicleBody = @{
        license_plate='TEST-123'
        brand='Toyota'
        model='Corolla'
        year=2024
        vehicle_type='car'
        fuel_type='gasoline'
        cargo_capacity=500.0
        color='Blue'
        status='active'
    } | ConvertTo-Json
    
    Test-Endpoint `
        -Name "PUT /company-admin/vehicles/:id (Update)" `
        -Method PUT `
        -Uri "http://localhost:8080/api/v1/company-admin/vehicles/$global:vehicleId" `
        -Headers $headers `
        -Body $updateVehicleBody
}

# ============================================================
# 4. TEAM-VEHICLE INTEGRATION
# ============================================================
Write-Host "`n[4] TEAM-VEHICLE INTEGRATION" -ForegroundColor Cyan
Write-Host "═══════════════════════════════" -ForegroundColor Cyan

# Assign Vehicle to Team
if ($global:teamId -and $global:vehicleId) {
    Test-Endpoint `
        -Name "POST /teams/:id/vehicles/:vehicleId (Assign)" `
        -Method POST `
        -Uri "http://localhost:8080/api/v1/company-admin/teams/$global:teamId/vehicles/$global:vehicleId" `
        -Headers $headers
}

# Get Team Vehicles
if ($global:teamId) {
    Test-Endpoint `
        -Name "GET /teams/:id/vehicles (List Team Vehicles)" `
        -Method GET `
        -Uri "http://localhost:8080/api/v1/company-admin/teams/$global:teamId/vehicles" `
        -Headers $headers
}

# Unassign Vehicle from Team
if ($global:teamId -and $global:vehicleId) {
    Test-Endpoint `
        -Name "DELETE /teams/:id/vehicles/:vehicleId (Unassign)" `
        -Method DELETE `
        -Uri "http://localhost:8080/api/v1/company-admin/teams/$global:teamId/vehicles/$global:vehicleId" `
        -Headers $headers
}

# ============================================================
# 5. USER MANAGEMENT
# ============================================================
Write-Host "`n[5] USER MANAGEMENT" -ForegroundColor Cyan
Write-Host "══════════════════" -ForegroundColor Cyan

# List Users (Admin endpoint)
Test-Endpoint `
    -Name "GET /admin/users (List)" `
    -Method GET `
    -Uri 'http://localhost:8080/api/v1/admin/users' `
    -Headers $headers

# Get Profile
Test-Endpoint `
    -Name "GET /profile (Get Profile)" `
    -Method GET `
    -Uri 'http://localhost:8080/api/v1/profile' `
    -Headers $headers

# ============================================================
# 6. COMPANY MANAGEMENT (Master)
# ============================================================
Write-Host "`n[6] COMPANY MANAGEMENT" -ForegroundColor Cyan
Write-Host "═════════════════════" -ForegroundColor Cyan

# Login as Master
$masterLogin = @{email='admin@test.com'; password='password123'} | ConvertTo-Json
$masterResponse = Test-Endpoint `
    -Name "POST /auth/login (Master)" `
    -Method POST `
    -Uri 'http://localhost:8080/api/v1/auth/login' `
    -Headers @{} `
    -Body $masterLogin

if ($masterResponse) {
    $masterHeaders = @{Authorization="Bearer $($masterResponse.access_token)"}
    
    # List Companies
    Test-Endpoint `
        -Name "GET /master/companies (List)" `
        -Method GET `
        -Uri 'http://localhost:8080/api/v1/master/companies' `
        -Headers $masterHeaders
}

# ============================================================
# 7. AUDIT LOGS
# ============================================================
Write-Host "`n[7] AUDIT LOGS" -ForegroundColor Cyan
Write-Host "═════════════" -ForegroundColor Cyan

Test-Endpoint `
    -Name "GET /audit/logs (List)" `
    -Method GET `
    -Uri 'http://localhost:8080/api/v1/audit/logs' `
    -Headers $headers

Test-Endpoint `
    -Name "GET /audit/stats (Stats)" `
    -Method GET `
    -Uri 'http://localhost:8080/api/v1/audit/stats' `
    -Headers $headers

# ============================================================
# 8. SECURITY AND 2FA
# ============================================================
Write-Host "`n[8] SECURITY AND 2FA" -ForegroundColor Cyan
Write-Host "═══════════════════" -ForegroundColor Cyan

Test-Endpoint `
    -Name "GET /security/2fa/status (Get 2FA Status)" `
    -Method GET `
    -Uri 'http://localhost:8080/api/v1/security/2fa/status' `
    -Headers $headers

# ============================================================
# 9. SESSIONS
# ============================================================
Write-Host "`n[9] SESSION MANAGEMENT" -ForegroundColor Cyan
Write-Host "═════════════════════" -ForegroundColor Cyan

Test-Endpoint `
    -Name "GET /sessions/active (Active Sessions)" `
    -Method GET `
    -Uri 'http://localhost:8080/api/v1/sessions/active' `
    -Headers $headers

Test-Endpoint `
    -Name "GET /sessions/dashboard (Session Dashboard)" `
    -Method GET `
    -Uri 'http://localhost:8080/api/v1/sessions/dashboard' `
    -Headers $headers

# ============================================================
# 10. HEALTH AND METRICS
# ============================================================
Write-Host "`n[10] HEALTH AND METRICS" -ForegroundColor Cyan
Write-Host "══════════════════════" -ForegroundColor Cyan

Test-Endpoint `
    -Name "GET /health (Health Check)" `
    -Method GET `
    -Uri 'http://localhost:8080/health' `
    -Headers @{}

Test-Endpoint `
    -Name "GET /metrics (Prometheus Metrics)" `
    -Method GET `
    -Uri 'http://localhost:8080/metrics' `
    -Headers @{}

# ============================================================
# CLEANUP
# ============================================================
Write-Host "`n[11] CLEANUP" -ForegroundColor Cyan
Write-Host "═══════════" -ForegroundColor Cyan

# Delete Vehicle
if ($global:vehicleId) {
    Test-Endpoint `
        -Name "DELETE /company-admin/vehicles/:id (Delete)" `
        -Method DELETE `
        -Uri "http://localhost:8080/api/v1/company-admin/vehicles/$global:vehicleId" `
        -Headers $headers
}

# Delete Team
if ($global:teamId) {
    Test-Endpoint `
        -Name "DELETE /company-admin/teams/:id (Delete)" `
        -Method DELETE `
        -Uri "http://localhost:8080/api/v1/company-admin/teams/$global:teamId" `
        -Headers $headers
}

# ============================================================
# FINAL REPORT
# ============================================================
Write-Host "`n" -ForegroundColor Yellow
Write-Host "                    RELATORIO FINAL                        " -ForegroundColor Yellow
Write-Host "" -ForegroundColor Yellow

$totalTests = $global:passedTests + $global:failedTests
$successRate = [math]::Round(($global:passedTests / $totalTests) * 100, 2)

Write-Host "`nESTATISTICAS:" -ForegroundColor Cyan
Write-Host "   Total de testes: $totalTests" -ForegroundColor White
Write-Host "   Passou: $global:passedTests" -ForegroundColor Green
Write-Host "   Falhou: $global:failedTests" -ForegroundColor Red
Write-Host "   Taxa de sucesso: $successRate%" -ForegroundColor Cyan

if ($global:failedTests -gt 0) {
    Write-Host "`nENDPOINTS COM FALHA:" -ForegroundColor Red
    foreach ($endpoint in $global:failedEndpoints) {
        Write-Host "   - $endpoint" -ForegroundColor Red
    }
}

Write-Host "`n" -ForegroundColor Yellow

if ($successRate -eq 100) {
    Write-Host "TODOS OS TESTES PASSARAM!" -ForegroundColor Green -BackgroundColor DarkGreen
} elseif ($successRate -ge 80) {
    Write-Host "MAIORIA DOS TESTES PASSOU, MAS HA FALHAS!" -ForegroundColor Yellow -BackgroundColor DarkYellow
} else {
    Write-Host "MUITOS TESTES FALHARAM - ACAO NECESSARIA!" -ForegroundColor Red -BackgroundColor DarkRed
}

Write-Host ""
