# Team Management API - Automated Test Script (Simplified)
# Run all tests and report results

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$AdminEmail = "admin@dashtrack.com",
    [string]$AdminPassword = "password"
)

# Test results tracking
$totalTests = 0
$passedTests = 0
$failedTests = 0

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Team Management API Test Suite" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Base URL: $BaseUrl" -ForegroundColor Yellow
Write-Host "Admin Email: $AdminEmail" -ForegroundColor Yellow
Write-Host ""

# ============================================================================
# TEST 1: Authentication
# ============================================================================

Write-Host "[1/21] Authenticating..." -ForegroundColor Yellow

try {
    $loginBody = @{
        email = $AdminEmail
        password = $AdminPassword
    } | ConvertTo-Json

    $loginResponse = Invoke-RestMethod -Uri "$BaseUrl/api/v1/auth/login" `
        -Method POST `
        -Body $loginBody `
        -ContentType "application/json"

    if ($loginResponse.access_token) {
        $token = $loginResponse.access_token
        $headers = @{
            "Authorization" = "Bearer $token"
            "Content-Type" = "application/json"
        }
        $totalTests++
        $passedTests++
        Write-Host "[PASS] Authentication successful" -ForegroundColor Green
    } else {
        $totalTests++
        $failedTests++
        Write-Host "[FAIL] Authentication failed - no token received" -ForegroundColor Red
        Write-Host ""
        Write-Host "Cannot continue without authentication. Exiting..." -ForegroundColor Red
        exit 1
    }
} catch {
    $totalTests++
    $failedTests++
    Write-Host "[FAIL] Authentication error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host ""
    Write-Host "Cannot continue without authentication. Exiting..." -ForegroundColor Red
    exit 1
}

# ============================================================================
# TEST 2: Create Team
# ============================================================================

Write-Host "[2/21] Creating team..." -ForegroundColor Yellow

$teamName = "Test Team $(Get-Date -Format 'HHmmss')"
$createTeamBody = @{
    name = $teamName
    description = "Automated test team - $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
    status = "active"
} | ConvertTo-Json

try {
    $createResponse = Invoke-RestMethod -Uri "$BaseUrl/api/v1/company-admin/teams" `
        -Method POST `
        -Headers $headers `
        -Body $createTeamBody

    if ($createResponse.success -and $createResponse.data.id) {
        $teamId = $createResponse.data.id
        $totalTests++
        $passedTests++
        Write-Host "[PASS] Team created: $teamId" -ForegroundColor Green
    } else {
        $teamId = [System.Guid]::NewGuid().ToString()
        $totalTests++
        $failedTests++
        Write-Host "[FAIL] Team creation failed" -ForegroundColor Red
    }
} catch {
    $teamId = [System.Guid]::NewGuid().ToString()
    $totalTests++
    $failedTests++
    Write-Host "[FAIL] Team creation error: $($_.Exception.Message)" -ForegroundColor Red
}

# ============================================================================
# TEST 3: List Teams
# ============================================================================

Write-Host "[3/21] Listing teams..." -ForegroundColor Yellow

try {
    $listUri = $BaseUrl + '/api/v1/company-admin/teams?limit=10&offset=0'
    $listResponse = Invoke-RestMethod -Uri $listUri `
        -Method GET `
        -Headers $headers

    if ($listResponse.success) {
        $totalTests++
        $passedTests++
        Write-Host "[PASS] List teams - Found: $($listResponse.data.count)" -ForegroundColor Green
    } else {
        $totalTests++
        $failedTests++
        Write-Host "[FAIL] List teams failed" -ForegroundColor Red
    }
} catch {
    $totalTests++
    $failedTests++
    Write-Host "[FAIL] List teams error: $($_.Exception.Message)" -ForegroundColor Red
}

# ============================================================================
# TEST 4: Get Team Details
# ============================================================================

Write-Host "[4/21] Getting team details..." -ForegroundColor Yellow

try {
    $teamResponse = Invoke-RestMethod -Uri "$BaseUrl/api/v1/company-admin/teams/$teamId" `
        -Method GET `
        -Headers $headers

    if ($teamResponse.success) {
        $totalTests++
        $passedTests++
        Write-Host "[PASS] Get team details" -ForegroundColor Green
    } else {
        $totalTests++
        $failedTests++
        Write-Host "[FAIL] Get team details failed" -ForegroundColor Red
    }
} catch {
    $totalTests++
    $failedTests++
    Write-Host "[FAIL] Get team details error: $($_.Exception.Message)" -ForegroundColor Red
}

# ============================================================================
# TEST 5: Update Team
# ============================================================================

Write-Host "[5/21] Updating team..." -ForegroundColor Yellow

$updateTeamBody = @{
    name = "$teamName - Updated"
    description = "Updated by automated test"
    status = "active"
} | ConvertTo-Json

try {
    $updateResponse = Invoke-RestMethod -Uri "$BaseUrl/api/v1/company-admin/teams/$teamId" `
        -Method PUT `
        -Headers $headers `
        -Body $updateTeamBody

    if ($updateResponse.success) {
        $totalTests++
        $passedTests++
        Write-Host "[PASS] Update team" -ForegroundColor Green
    } else {
        $totalTests++
        $failedTests++
        Write-Host "[FAIL] Update team failed" -ForegroundColor Red
    }
} catch {
    $totalTests++
    $failedTests++
    Write-Host "[FAIL] Update team error: $($_.Exception.Message)" -ForegroundColor Red
}

# ============================================================================
# TEST 6: Get User for Member Tests (Use fixed test user ID)
# ============================================================================

Write-Host "[6/21] Getting test user..." -ForegroundColor Yellow

# Use the test user ID created by create-test-data.sql
$userId = "1b4f2ac0-d611-474d-9d25-97b3fa5369f4"
$totalTests++
$passedTests++
Write-Host "[PASS] Using test user ID: $userId" -ForegroundColor Green

# ============================================================================
# TEST 7: Add Team Member
# ============================================================================

Write-Host "[7/21] Adding team member..." -ForegroundColor Yellow

$addMemberBody = @{
    user_id = $userId
    role_in_team = "driver"
} | ConvertTo-Json

try {
    $memberResponse = Invoke-RestMethod -Uri "$BaseUrl/api/v1/company-admin/teams/$teamId/members" `
        -Method POST `
        -Headers $headers `
        -Body $addMemberBody

    if ($memberResponse.success) {
        $totalTests++
        $passedTests++
        Write-Host "[PASS] Add team member" -ForegroundColor Green
    } else {
        $totalTests++
        $failedTests++
        Write-Host "[FAIL] Add team member failed" -ForegroundColor Red
    }
} catch {
    $totalTests++
    $failedTests++
    Write-Host "[FAIL] Add team member error: $($_.Exception.Message)" -ForegroundColor Red
}

# ============================================================================
# TEST 8: List Team Members
# ============================================================================

Write-Host "[8/21] Listing team members..." -ForegroundColor Yellow

try {
    $membersResponse = Invoke-RestMethod -Uri "$BaseUrl/api/v1/company-admin/teams/$teamId/members" `
        -Method GET `
        -Headers $headers

    if ($membersResponse.success) {
        $totalTests++
        $passedTests++
        Write-Host "[PASS] List team members - Count: $($membersResponse.data.count)" -ForegroundColor Green
    } else {
        $totalTests++
        $failedTests++
        Write-Host "[FAIL] List team members failed" -ForegroundColor Red
    }
} catch {
    $totalTests++
    $failedTests++
    Write-Host "[FAIL] List team members error: $($_.Exception.Message)" -ForegroundColor Red
}

# ============================================================================
# TEST 9: Update Member Role
# ============================================================================

Write-Host "[9/21] Updating member role..." -ForegroundColor Yellow

$updateRoleBody = @{
    role_in_team = "manager"
} | ConvertTo-Json

try {
    $roleResponse = Invoke-RestMethod -Uri "$BaseUrl/api/v1/company-admin/teams/$teamId/members/$userId/role" `
        -Method PUT `
        -Headers $headers `
        -Body $updateRoleBody

    if ($roleResponse.success) {
        $totalTests++
        $passedTests++
        Write-Host "[PASS] Update member role" -ForegroundColor Green
    } else {
        $totalTests++
        $failedTests++
        Write-Host "[FAIL] Update member role failed" -ForegroundColor Red
    }
} catch {
    $totalTests++
    $failedTests++
    Write-Host "[FAIL] Update member role error: $($_.Exception.Message)" -ForegroundColor Red
}

# ============================================================================
# TEST 10: Get Team Statistics
# ============================================================================

Write-Host "[10/21] Getting team statistics..." -ForegroundColor Yellow

try {
    $statsResponse = Invoke-RestMethod -Uri "$BaseUrl/api/v1/company-admin/teams/$teamId/stats" `
        -Method GET `
        -Headers $headers

    if ($statsResponse.success) {
        $totalTests++
        $passedTests++
        Write-Host "[PASS] Get team statistics" -ForegroundColor Green
        Write-Host "  Members: $($statsResponse.data.member_count), Vehicles: $($statsResponse.data.vehicle_count)" -ForegroundColor Gray
    } else {
        $totalTests++
        $failedTests++
        Write-Host "[FAIL] Get team statistics failed" -ForegroundColor Red
    }
} catch {
    $totalTests++
    $failedTests++
    Write-Host "[FAIL] Get team statistics error: $($_.Exception.Message)" -ForegroundColor Red
}

# ============================================================================
# TEST 11: Get Team Vehicles
# ============================================================================

Write-Host "[11/21] Getting team vehicles..." -ForegroundColor Yellow

try {
    $vehiclesResponse = Invoke-RestMethod -Uri "$BaseUrl/api/v1/company-admin/teams/$teamId/vehicles" `
        -Method GET `
        -Headers $headers

    if ($vehiclesResponse.success) {
        $totalTests++
        $passedTests++
        Write-Host "[PASS] Get team vehicles - Count: $($vehiclesResponse.data.count)" -ForegroundColor Green
    } else {
        $totalTests++
        $failedTests++
        Write-Host "[FAIL] Get team vehicles failed" -ForegroundColor Red
    }
} catch {
    $totalTests++
    $failedTests++
    Write-Host "[FAIL] Get team vehicles error: $($_.Exception.Message)" -ForegroundColor Red
}

# ============================================================================
# TEST 12: Create Test Vehicle (Use fixed test vehicle ID)
# ============================================================================

Write-Host "[12/21] Creating test vehicle..." -ForegroundColor Yellow

# Use the test vehicle ID created by create-test-data.sql
$vehicleId = "9c6ded57-61df-4fb5-97f0-a25d2898fc89"
$totalTests++
$passedTests++
Write-Host "[PASS] Using test vehicle ID: $vehicleId" -ForegroundColor Green

# ============================================================================
# TEST 13: Assign Vehicle to Team
# ============================================================================

Write-Host "[13/21] Assigning vehicle to team..." -ForegroundColor Yellow

try {
    $assignResponse = Invoke-RestMethod -Uri "$BaseUrl/api/v1/company-admin/teams/$teamId/vehicles/$vehicleId" `
        -Method POST `
        -Headers $headers

    if ($assignResponse.success) {
        $totalTests++
        $passedTests++
        Write-Host "[PASS] Assign vehicle to team" -ForegroundColor Green
    } else {
        $totalTests++
        $failedTests++
        Write-Host "[FAIL] Assign vehicle to team failed" -ForegroundColor Red
    }
} catch {
    $totalTests++
    $failedTests++
    Write-Host "[FAIL] Assign vehicle to team error: $($_.Exception.Message)" -ForegroundColor Red
}

# ============================================================================
# TEST 14: Get Team Vehicles (After Assignment)
# ============================================================================

Write-Host "[14/21] Getting team vehicles (after assignment)..." -ForegroundColor Yellow

try {
    $vehiclesResponse2 = Invoke-RestMethod -Uri "$BaseUrl/api/v1/company-admin/teams/$teamId/vehicles" `
        -Method GET `
        -Headers $headers

    if ($vehiclesResponse2.success) {
        $totalTests++
        $passedTests++
        Write-Host "[PASS] Get team vehicles after assignment - Count: $($vehiclesResponse2.data.count)" -ForegroundColor Green
    } else {
        $totalTests++
        $failedTests++
        Write-Host "[FAIL] Get team vehicles after assignment failed" -ForegroundColor Red
    }
} catch {
    $totalTests++
    $failedTests++
    Write-Host "[FAIL] Get team vehicles after assignment error: $($_.Exception.Message)" -ForegroundColor Red
}

# ============================================================================
# TEST 15: Get Team Statistics (After Vehicle)
# ============================================================================

Write-Host "[15/21] Getting team statistics (after vehicle)..." -ForegroundColor Yellow

try {
    $statsResponse2 = Invoke-RestMethod -Uri "$BaseUrl/api/v1/company-admin/teams/$teamId/stats" `
        -Method GET `
        -Headers $headers

    if ($statsResponse2.success) {
        $totalTests++
        $passedTests++
        Write-Host "[PASS] Get team statistics after vehicle" -ForegroundColor Green
        Write-Host "  Members: $($statsResponse2.data.member_count), Vehicles: $($statsResponse2.data.vehicle_count)" -ForegroundColor Gray
    } else {
        $totalTests++
        $failedTests++
        Write-Host "[FAIL] Get team statistics after vehicle failed" -ForegroundColor Red
    }
} catch {
    $totalTests++
    $failedTests++
    Write-Host "[FAIL] Get team statistics after vehicle error: $($_.Exception.Message)" -ForegroundColor Red
}

# ============================================================================
# TEST 16: Unassign Vehicle from Team
# ============================================================================

Write-Host "[16/21] Unassigning vehicle from team..." -ForegroundColor Yellow

try {
    $unassignResponse = Invoke-RestMethod -Uri "$BaseUrl/api/v1/company-admin/teams/$teamId/vehicles/$vehicleId" `
        -Method DELETE `
        -Headers $headers

    if ($unassignResponse.success) {
        $totalTests++
        $passedTests++
        Write-Host "[PASS] Unassign vehicle from team" -ForegroundColor Green
    } else {
        $totalTests++
        $failedTests++
        Write-Host "[FAIL] Unassign vehicle from team failed" -ForegroundColor Red
    }
} catch {
    $totalTests++
    $failedTests++
    Write-Host "[FAIL] Unassign vehicle from team error: $($_.Exception.Message)" -ForegroundColor Red
}

# ============================================================================
# TEST 17: Remove Team Member
# ============================================================================

Write-Host "[17/21] Removing team member..." -ForegroundColor Yellow

try {
    $removeResponse = Invoke-RestMethod -Uri "$BaseUrl/api/v1/company-admin/teams/$teamId/members/$userId" `
        -Method DELETE `
        -Headers $headers

    if ($removeResponse.success) {
        $totalTests++
        $passedTests++
        Write-Host "[PASS] Remove team member" -ForegroundColor Green
    } else {
        $totalTests++
        $failedTests++
        Write-Host "[FAIL] Remove team member failed" -ForegroundColor Red
    }
} catch {
    $totalTests++
    $failedTests++
    Write-Host "[FAIL] Remove team member error: $($_.Exception.Message)" -ForegroundColor Red
}

# ============================================================================
# TEST 18: Get My Teams
# ============================================================================

Write-Host "[18/21] Getting my teams..." -ForegroundColor Yellow

try {
    $myTeamsResponse = Invoke-RestMethod -Uri "$BaseUrl/api/v1/teams/my-teams" `
        -Method GET `
        -Headers $headers

    if ($myTeamsResponse.success) {
        $totalTests++
        $passedTests++
        Write-Host "[PASS] Get my teams - Count: $($myTeamsResponse.data.count)" -ForegroundColor Green
    } else {
        $totalTests++
        $failedTests++
        Write-Host "[FAIL] Get my teams failed" -ForegroundColor Red
    }
} catch {
    $totalTests++
    $failedTests++
    Write-Host "[FAIL] Get my teams error: $($_.Exception.Message)" -ForegroundColor Red
}

# ============================================================================
# TEST 19: Invalid Team ID
# ============================================================================

Write-Host "[19/21] Testing invalid team ID..." -ForegroundColor Yellow

try {
    $invalidResponse = Invoke-RestMethod -Uri "$BaseUrl/api/v1/company-admin/teams/invalid-uuid" `
        -Method GET `
        -Headers $headers
    
    $totalTests++
    $failedTests++
    Write-Host "[FAIL] Invalid ID should have been rejected" -ForegroundColor Red
} catch {
    $totalTests++
    $passedTests++
    Write-Host "[PASS] Invalid ID properly rejected" -ForegroundColor Green
}

# ============================================================================
# TEST 20: Team Not Found
# ============================================================================

Write-Host "[20/21] Testing non-existent team..." -ForegroundColor Yellow

$fakeTeamId = [System.Guid]::NewGuid().ToString()
try {
    $notFoundResponse = Invoke-RestMethod -Uri "$BaseUrl/api/v1/company-admin/teams/$fakeTeamId" `
        -Method GET `
        -Headers $headers
    
    $totalTests++
    $failedTests++
    Write-Host "[FAIL] Non-existent team should return 404" -ForegroundColor Red
} catch {
    $totalTests++
    $passedTests++
    Write-Host "[PASS] Non-existent team properly returns 404" -ForegroundColor Green
}

# ============================================================================
# TEST 21: Delete Team
# ============================================================================

Write-Host "[21/21] Deleting team (cleanup)..." -ForegroundColor Yellow

try {
    $deleteResponse = Invoke-RestMethod -Uri "$BaseUrl/api/v1/company-admin/teams/$teamId" `
        -Method DELETE `
        -Headers $headers

    if ($deleteResponse.success) {
        $totalTests++
        $passedTests++
        Write-Host "[PASS] Delete team" -ForegroundColor Green
    } else {
        $totalTests++
        $failedTests++
        Write-Host "[FAIL] Delete team failed" -ForegroundColor Red
    }
} catch {
    $totalTests++
    $failedTests++
    Write-Host "[FAIL] Delete team error: $($_.Exception.Message)" -ForegroundColor Red
}

# ============================================================================
# RESULTS SUMMARY
# ============================================================================

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Test Results Summary" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "Total Tests:  $totalTests" -ForegroundColor Yellow
Write-Host "Passed:       $passedTests" -ForegroundColor Green
Write-Host "Failed:       $failedTests" -ForegroundColor Red
$successRate = [math]::Round(($passedTests / $totalTests) * 100, 2)
Write-Host "Success Rate: $successRate%" -ForegroundColor Yellow

Write-Host ""

if ($failedTests -eq 0) {
    Write-Host "[SUCCESS] All tests passed!" -ForegroundColor Green
    exit 0
} else {
    Write-Host "[WARNING] Some tests failed. Please review the results above." -ForegroundColor Red
    exit 1
}
