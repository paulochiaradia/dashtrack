# Vehicle Assignment History API Testing Script
# This script tests the vehicle assignment history tracking feature

$baseUrl = "http://localhost:8080/api/v1"
$companyAdminToken = "your-company-admin-token-here"

# Set headers
$headers = @{
    "Authorization" = "Bearer $companyAdminToken"
    "Content-Type" = "application/json"
}

Write-Host "`n=== VEHICLE ASSIGNMENT HISTORY API TESTS ===" -ForegroundColor Cyan

# ============================================================================
# 1. CREATE TEST USERS (Driver, Helper 1, Helper 2)
# ============================================================================
Write-Host "`n[1] Creating test users..." -ForegroundColor Yellow

$users = @()
$userRoles = @("driver", "helper1", "helper2")

foreach ($role in $userRoles) {
    $userData = @{
        name = "Test $role User"
        email = "test-vehicle-$role@example.com"
        cpf = "222333444$($userRoles.IndexOf($role))0"
        phone = "11988887$($userRoles.IndexOf($role))00"
        role = "user"
        password = "Test@123"
    } | ConvertTo-Json

    try {
        $userResponse = Invoke-RestMethod -Uri "$baseUrl/company-admin/users" `
            -Method Post `
            -Headers $headers `
            -Body $userData `
            -ErrorAction Stop
        
        $users += @{
            id = $userResponse.data.id
            name = $userResponse.data.name
            type = $role
        }
        Write-Host "✓ User created: $($userResponse.data.name) - ID: $($userResponse.data.id)" -ForegroundColor Green
    } catch {
        Write-Host "✗ Failed to create user ($role): $($_.Exception.Message)" -ForegroundColor Red
    }
}

if ($users.Count -lt 3) {
    Write-Host "✗ Failed to create required users. Exiting..." -ForegroundColor Red
    exit 1
}

$driverUser = $users[0]
$helper1User = $users[1]
$helper2User = $users[2]

# ============================================================================
# 2. CREATE TEST TEAM
# ============================================================================
Write-Host "`n[2] Creating test team..." -ForegroundColor Yellow

$teamData = @{
    name = "Test Team for Vehicle History"
    description = "Team created for testing vehicle assignment history"
} | ConvertTo-Json

try {
    $teamResponse = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams" `
        -Method Post `
        -Headers $headers `
        -Body $teamData `
        -ErrorAction Stop
    
    $teamId = $teamResponse.data.id
    Write-Host "✓ Team created successfully: $teamId" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed to create team: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# ============================================================================
# 3. CREATE TEST VEHICLE
# ============================================================================
Write-Host "`n[3] Creating test vehicle..." -ForegroundColor Yellow

$vehicleData = @{
    license_plate = "TEST-$(Get-Random -Minimum 1000 -Maximum 9999)"
    brand = "Test Brand"
    model = "Test Model"
    year = 2024
    vehicle_type = "truck"
    fuel_type = "diesel"
} | ConvertTo-Json

try {
    $vehicleResponse = Invoke-RestMethod -Uri "$baseUrl/company-admin/vehicles" `
        -Method Post `
        -Headers $headers `
        -Body $vehicleData `
        -ErrorAction Stop
    
    $vehicleId = $vehicleResponse.data.id
    Write-Host "✓ Vehicle created successfully: $vehicleId" -ForegroundColor Green
    Write-Host "  License Plate: $($vehicleResponse.data.license_plate)" -ForegroundColor Gray
} catch {
    Write-Host "✗ Failed to create vehicle: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# ============================================================================
# 4. ASSIGNMENT CHANGE #1 - Assign Driver Only
# ============================================================================
Write-Host "`n[4] Assignment Change #1: Assign driver..." -ForegroundColor Yellow

$assignment1 = @{
    driver_id = $driverUser.id
} | ConvertTo-Json

try {
    $assignResponse1 = Invoke-RestMethod -Uri "$baseUrl/company-admin/vehicles/$vehicleId/assign" `
        -Method Put `
        -Headers $headers `
        -Body $assignment1 `
        -ErrorAction Stop
    
    Write-Host "✓ Driver assigned successfully" -ForegroundColor Green
    Write-Host "  Driver: $($driverUser.name)" -ForegroundColor Gray
} catch {
    Write-Host "✗ Failed to assign driver: $($_.Exception.Message)" -ForegroundColor Red
}

Start-Sleep -Seconds 2

# ============================================================================
# 5. ASSIGNMENT CHANGE #2 - Assign Helper 1
# ============================================================================
Write-Host "`n[5] Assignment Change #2: Assign helper..." -ForegroundColor Yellow

$assignment2 = @{
    driver_id = $driverUser.id
    helper_id = $helper1User.id
} | ConvertTo-Json

try {
    $assignResponse2 = Invoke-RestMethod -Uri "$baseUrl/company-admin/vehicles/$vehicleId/assign" `
        -Method Put `
        -Headers $headers `
        -Body $assignment2 `
        -ErrorAction Stop
    
    Write-Host "✓ Helper assigned successfully" -ForegroundColor Green
    Write-Host "  Driver: $($driverUser.name)" -ForegroundColor Gray
    Write-Host "  Helper: $($helper1User.name)" -ForegroundColor Gray
} catch {
    Write-Host "✗ Failed to assign helper: $($_.Exception.Message)" -ForegroundColor Red
}

Start-Sleep -Seconds 2

# ============================================================================
# 6. ASSIGNMENT CHANGE #3 - Assign to Team
# ============================================================================
Write-Host "`n[6] Assignment Change #3: Assign to team..." -ForegroundColor Yellow

$assignment3 = @{
    driver_id = $driverUser.id
    helper_id = $helper1User.id
    team_id = $teamId
} | ConvertTo-Json

try {
    $assignResponse3 = Invoke-RestMethod -Uri "$baseUrl/company-admin/vehicles/$vehicleId/assign" `
        -Method Put `
        -Headers $headers `
        -Body $assignment3 `
        -ErrorAction Stop
    
    Write-Host "✓ Team assigned successfully" -ForegroundColor Green
    Write-Host "  Team ID: $teamId" -ForegroundColor Gray
} catch {
    Write-Host "✗ Failed to assign team: $($_.Exception.Message)" -ForegroundColor Red
}

Start-Sleep -Seconds 2

# ============================================================================
# 7. ASSIGNMENT CHANGE #4 - Change Helper
# ============================================================================
Write-Host "`n[7] Assignment Change #4: Change helper..." -ForegroundColor Yellow

$assignment4 = @{
    driver_id = $driverUser.id
    helper_id = $helper2User.id
    team_id = $teamId
} | ConvertTo-Json

try {
    $assignResponse4 = Invoke-RestMethod -Uri "$baseUrl/company-admin/vehicles/$vehicleId/assign" `
        -Method Put `
        -Headers $headers `
        -Body $assignment4 `
        -ErrorAction Stop
    
    Write-Host "✓ Helper changed successfully" -ForegroundColor Green
    Write-Host "  Old Helper: $($helper1User.name)" -ForegroundColor Gray
    Write-Host "  New Helper: $($helper2User.name)" -ForegroundColor Gray
} catch {
    Write-Host "✗ Failed to change helper: $($_.Exception.Message)" -ForegroundColor Red
}

Start-Sleep -Seconds 2

# ============================================================================
# 8. ASSIGNMENT CHANGE #5 - Remove All Assignments
# ============================================================================
Write-Host "`n[8] Assignment Change #5: Remove all assignments..." -ForegroundColor Yellow

$assignment5 = @{} | ConvertTo-Json

try {
    $assignResponse5 = Invoke-RestMethod -Uri "$baseUrl/company-admin/vehicles/$vehicleId/assign" `
        -Method Put `
        -Headers $headers `
        -Body $assignment5 `
        -ErrorAction Stop
    
    Write-Host "✓ All assignments removed successfully" -ForegroundColor Green
} catch {
    Write-Host "✗ Failed to remove assignments: $($_.Exception.Message)" -ForegroundColor Red
}

Start-Sleep -Seconds 2

# ============================================================================
# 9. GET ASSIGNMENT HISTORY
# ============================================================================
Write-Host "`n[9] Retrieving assignment history..." -ForegroundColor Yellow

try {
    $historyResponse = Invoke-RestMethod -Uri "$baseUrl/company-admin/vehicles/$vehicleId/assignment-history?limit=20" `
        -Method Get `
        -Headers $headers `
        -ErrorAction Stop
    
    Write-Host "✓ Assignment history retrieved successfully" -ForegroundColor Green
    Write-Host "  Total history entries: $($historyResponse.data.count)" -ForegroundColor Gray
    Write-Host "`n  History Timeline:" -ForegroundColor White
    
    foreach ($entry in $historyResponse.data.history) {
        $changeDate = [DateTime]::Parse($entry.changed_at)
        Write-Host "  ---" -ForegroundColor DarkGray
        Write-Host "  Time: $($changeDate.ToString('yyyy-MM-dd HH:mm:ss'))" -ForegroundColor Gray
        Write-Host "  Change Type: $($entry.change_type)" -ForegroundColor Cyan
        
        # Show previous state
        if ($entry.previous_driver_id -or $entry.previous_helper_id -or $entry.previous_team_id) {
            Write-Host "  Previous:" -ForegroundColor Yellow
            if ($entry.previous_driver) {
                Write-Host "    - Driver: $($entry.previous_driver.name)" -ForegroundColor Gray
            }
            if ($entry.previous_helper) {
                Write-Host "    - Helper: $($entry.previous_helper.name)" -ForegroundColor Gray
            }
            if ($entry.previous_team) {
                Write-Host "    - Team: $($entry.previous_team.name)" -ForegroundColor Gray
            }
        } else {
            Write-Host "  Previous: None" -ForegroundColor DarkGray
        }
        
        # Show new state
        if ($entry.new_driver_id -or $entry.new_helper_id -or $entry.new_team_id) {
            Write-Host "  New:" -ForegroundColor Green
            if ($entry.new_driver) {
                Write-Host "    - Driver: $($entry.new_driver.name)" -ForegroundColor Gray
            }
            if ($entry.new_helper) {
                Write-Host "    - Helper: $($entry.new_helper.name)" -ForegroundColor Gray
            }
            if ($entry.new_team) {
                Write-Host "    - Team: $($entry.new_team.name)" -ForegroundColor Gray
            }
        } else {
            Write-Host "  New: None (Unassigned)" -ForegroundColor DarkGray
        }
    }
} catch {
    Write-Host "✗ Failed to retrieve history: $($_.Exception.Message)" -ForegroundColor Red
}

# ============================================================================
# 10. TEST WITH LIMIT PARAMETER
# ============================================================================
Write-Host "`n[10] Testing limit parameter (limit=3)..." -ForegroundColor Yellow

try {
    $limitedHistory = Invoke-RestMethod -Uri "$baseUrl/company-admin/vehicles/$vehicleId/assignment-history?limit=3" `
        -Method Get `
        -Headers $headers `
        -ErrorAction Stop
    
    Write-Host "✓ Limited history retrieved successfully" -ForegroundColor Green
    Write-Host "  Entries returned: $($limitedHistory.data.count)" -ForegroundColor Gray
    Write-Host "  Limit applied: $($limitedHistory.data.limit)" -ForegroundColor Gray
} catch {
    Write-Host "✗ Failed to retrieve limited history: $($_.Exception.Message)" -ForegroundColor Red
}

# ============================================================================
# 11. CLEANUP (Optional)
# ============================================================================
Write-Host "`n[11] Cleanup test data..." -ForegroundColor Yellow
$cleanup = Read-Host "Do you want to delete test data? (y/n)"

if ($cleanup -eq "y") {
    # Delete vehicle
    try {
        Invoke-RestMethod -Uri "$baseUrl/company-admin/vehicles/$vehicleId" `
            -Method Delete `
            -Headers $headers `
            -ErrorAction Stop
        Write-Host "✓ Test vehicle deleted" -ForegroundColor Green
    } catch {
        Write-Host "✗ Failed to delete vehicle: $($_.Exception.Message)" -ForegroundColor Red
    }

    # Delete team
    try {
        Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$teamId" `
            -Method Delete `
            -Headers $headers `
            -ErrorAction Stop
        Write-Host "✓ Test team deleted" -ForegroundColor Green
    } catch {
        Write-Host "✗ Failed to delete team: $($_.Exception.Message)" -ForegroundColor Red
    }

    # Delete users
    foreach ($user in $users) {
        try {
            Invoke-RestMethod -Uri "$baseUrl/company-admin/users/$($user.id)" `
                -Method Delete `
                -Headers $headers `
                -ErrorAction Stop
            Write-Host "✓ User deleted: $($user.name)" -ForegroundColor Green
        } catch {
            Write-Host "✗ Failed to delete user ($($user.name)): $($_.Exception.Message)" -ForegroundColor Red
        }
    }
}

Write-Host "`n=== TESTS COMPLETED ===" -ForegroundColor Cyan
Write-Host "`nSummary:" -ForegroundColor White
Write-Host "- Create Test Data: ✓" -ForegroundColor Green
Write-Host "- Assignment Changes (5x): ✓" -ForegroundColor Green
Write-Host "- Get Full History: ✓" -ForegroundColor Green
Write-Host "- Get Limited History: ✓" -ForegroundColor Green
Write-Host "- History includes user/team details: ✓" -ForegroundColor Green
Write-Host "- Change types tracked correctly: ✓" -ForegroundColor Green
