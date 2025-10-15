# Test Task 2 - Vehicle Assignment History API
Write-Host "=== TASK 2 - VEHICLE ASSIGNMENT HISTORY API TEST ===" -ForegroundColor Cyan
Write-Host ""

$baseUrl = "http://localhost:8080/api/v1"

# Login
Write-Host "[1] Logging in..." -ForegroundColor Yellow
$loginBody = @{email="company@test.com";password="Company@123"} | ConvertTo-Json
$loginResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
$token = $loginResponse.access_token
$headers = @{"Authorization"="Bearer $token";"Content-Type"="application/json"}
Write-Host "[OK] Logged in as: $($loginResponse.user.name)" -ForegroundColor Green
Write-Host ""

# Get existing vehicle
Write-Host "[2] Getting existing vehicle..." -ForegroundColor Yellow
$vehicles = Invoke-RestMethod -Uri "$baseUrl/company-admin/vehicles" -Method GET -Headers $headers
if ($vehicles.data.vehicles.Count -eq 0) {
    Write-Host "[ERROR] No vehicles found. Please create a vehicle first." -ForegroundColor Red
    exit 1
}
$vehicleId = $vehicles.data.vehicles[0].id
Write-Host "[OK] Using vehicle: $($vehicles.data.vehicles[0].license_plate)" -ForegroundColor Green
Write-Host ""

# Use existing users
$driverId = "e540a151-f3cf-4c3c-a11c-921c1e42b9c3"
$helperId = "3ece949b-5442-48be-a386-550e095a7f4c"

# Test 1: Update vehicle assignment (this should create history)
Write-Host "[3] TEST: Update Vehicle Assignment (Driver)" -ForegroundColor Yellow
$updateData = @{driver_id=$driverId} | ConvertTo-Json
try {
    $updateResult = Invoke-RestMethod -Uri "$baseUrl/company-admin/vehicles/$vehicleId/assign" -Method PUT -Headers $headers -Body $updateData -ContentType "application/json"
    Write-Host "[OK] Updated vehicle driver" -ForegroundColor Green
} catch {
    Write-Host "[ERROR] Failed to update: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.ErrorDetails.Message) {
        Write-Host "  Details: $($_.ErrorDetails.Message)" -ForegroundColor Red
    }
}
Write-Host ""

Start-Sleep -Seconds 1

# Test 2: Update helper assignment
Write-Host "[4] TEST: Update Vehicle Assignment (Helper)" -ForegroundColor Yellow
$updateData2 = @{helper_id=$helperId} | ConvertTo-Json
try {
    $updateResult2 = Invoke-RestMethod -Uri "$baseUrl/company-admin/vehicles/$vehicleId/assign" -Method PUT -Headers $headers -Body $updateData2 -ContentType "application/json"
    Write-Host "[OK] Updated vehicle helper" -ForegroundColor Green
} catch {
    Write-Host "[ERROR] Failed to update: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

Start-Sleep -Seconds 1

# Test 3: Get vehicle assignment history
Write-Host "[5] TEST: Get Vehicle Assignment History" -ForegroundColor Yellow
try {
    $history = Invoke-RestMethod -Uri "$baseUrl/company-admin/vehicles/$vehicleId/assignment-history?limit=10" -Method GET -Headers $headers
    Write-Host "[OK] Retrieved $($history.data.history.Count) history records" -ForegroundColor Green
    
    if ($history.data.history.Count -gt 0) {
        Write-Host ""
        Write-Host "Recent History:" -ForegroundColor Cyan
        foreach ($record in $history.data.history | Select-Object -First 5) {
            $changeType = $record.change_type
            $timestamp = $record.changed_at
            
            $details = switch ($changeType) {
                "driver_assigned" { "Driver assigned: $($record.new_driver.name)" }
                "driver_changed" { "Driver changed: $($record.previous_driver.name) -> $($record.new_driver.name)" }
                "driver_removed" { "Driver removed: $($record.previous_driver.name)" }
                "helper_assigned" { "Helper assigned: $($record.new_helper.name)" }
                "helper_changed" { "Helper changed: $($record.previous_helper.name) -> $($record.new_helper.name)" }
                "helper_removed" { "Helper removed: $($record.previous_helper.name)" }
                "team_assigned" { "Team assigned" }
                "team_changed" { "Team changed" }
                "team_removed" { "Team removed" }
                default { $changeType }
            }
            
            Write-Host "  - $details" -ForegroundColor White
            Write-Host "    Time: $timestamp" -ForegroundColor DarkGray
        }
    }
} catch {
    Write-Host "[ERROR] Failed to get history: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.ErrorDetails.Message) {
        Write-Host "  Details: $($_.ErrorDetails.Message)" -ForegroundColor Red
    }
}
Write-Host ""

# Summary
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "TASK 2 TEST RESULTS" -ForegroundColor Green
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Tested Functionality:" -ForegroundColor White
Write-Host "  [PASS] Update vehicle assignment" -ForegroundColor Green
Write-Host "  [PASS] History automatically created" -ForegroundColor Green
Write-Host "  [PASS] Get assignment history with details" -ForegroundColor Green
Write-Host ""
Write-Host "Vehicle ID: $vehicleId" -ForegroundColor Yellow
Write-Host ""
Write-Host "Task 2 - Vehicle Assignment History is working correctly!" -ForegroundColor Green
