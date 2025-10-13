# Team Management API - Automated Test Script
# Run all 21 tests and report results

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$AdminEmail = "admin@dashtrack.com",
    [string]$AdminPassword = "Master@123"
)

# Colors for output
$ErrorColor = "Red"
$SuccessColor = "Green"
$InfoColor = "Yellow"
$HeaderColor = "Cyan"

# Test results tracking
$totalTests = 0
$passedTests = 0
$failedTests = 0
$testResults = @()

function Write-TestHeader {
    param([string]$Message)
    Write-Host ""
    Write-Host "========================================" -ForegroundColor $HeaderColor
    Write-Host $Message -ForegroundColor $HeaderColor
    Write-Host "========================================" -ForegroundColor $HeaderColor
    Write-Host ""
}

function Write-TestStep {
    param([string]$Step, [string]$Message)
    Write-Host "[$Step] $Message" -ForegroundColor $InfoColor
}

function Write-TestPass {
    param([string]$Message)
    Write-Host "✓ $Message" -ForegroundColor $SuccessColor
}

function Write-TestFail {
    param([string]$Message)
    Write-Host "✗ $Message" -ForegroundColor $ErrorColor
}

function Test-Endpoint {
    param(
        [string]$Name,
        [string]$Uri,
        [string]$Method,
        [hashtable]$Headers,
        [string]$Body = $null,
        [int]$ExpectedStatus = 200
    )
    
    $script:totalTests++
    
    try {
        $params = @{
            Uri = $Uri
            Method = $Method
            Headers = $Headers
        }
        
        if ($Body) {
            $params.Body = $Body
        }
        
        $response = Invoke-RestMethod @params
        
        if ($response.success -eq $true) {
            $script:passedTests++
            Write-TestPass "$Name"
            $script:testResults += @{
                Test = $Name
                Status = "PASS"
                Details = $response.message
            }
            return $response
        } else {
            $script:failedTests++
            Write-TestFail "$Name - Response not successful"
            $script:testResults += @{
                Test = $Name
                Status = "FAIL"
                Details = $response.message
            }
            return $null
        }
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        
        if ($statusCode -eq $ExpectedStatus) {
            $script:passedTests++
            Write-TestPass "$Name (Expected $ExpectedStatus)"
            $script:testResults += @{
                Test = $Name
                Status = "PASS"
                Details = "Returned expected status $ExpectedStatus"
            }
        } else {
            $script:failedTests++
            Write-TestFail "$Name - Status: $statusCode (Expected: $ExpectedStatus)"
            $script:testResults += @{
                Test = $Name
                Status = "FAIL"
                Details = "Status $statusCode, Expected $ExpectedStatus"
            }
        }
        return $null
    }
}

# ============================================================================
# START TESTING
# ============================================================================

Write-TestHeader "Team Management API Test Suite"
Write-Host "Base URL: $BaseUrl" -ForegroundColor $InfoColor
Write-Host "Admin Email: $AdminEmail" -ForegroundColor $InfoColor
Write-Host ""

# ============================================================================
# TEST 1: Authentication
# ============================================================================

Write-TestStep "1/21" "Authenticating..."

try {
    $loginBody = @{
        email = $AdminEmail
        password = $AdminPassword
    } | ConvertTo-Json

    $loginResponse = Invoke-RestMethod -Uri "$BaseUrl/api/v1/auth/login" `
        -Method POST `
        -Body $loginBody `
        -ContentType "application/json"

    if ($loginResponse.success -and $loginResponse.data.access_token) {
        $token = $loginResponse.data.access_token
        $headers = @{
            "Authorization" = "Bearer $token"
            "Content-Type" = "application/json"
        }
        $totalTests++
        $passedTests++
        Write-TestPass "Authentication successful"
        $testResults += @{
            Test = "Authentication"
            Status = "PASS"
            Details = "Token obtained"
        }
    } else {
        $totalTests++
        $failedTests++
        Write-TestFail "Authentication failed"
        $testResults += @{
            Test = "Authentication"
            Status = "FAIL"
            Details = "No token received"
        }
        Write-Host ""
        Write-Host "Cannot continue without authentication. Exiting..." -ForegroundColor $ErrorColor
        exit 1
    }
} catch {
    $totalTests++
    $failedTests++
    Write-TestFail "Authentication error: $($_.Exception.Message)"
    $testResults += @{
        Test = "Authentication"
        Status = "FAIL"
        Details = $_.Exception.Message
    }
    Write-Host ""
    Write-Host "Cannot continue without authentication. Exiting..." -ForegroundColor $ErrorColor
    exit 1
}

# ============================================================================
# TEST 2: Create Team
# ============================================================================

Write-TestStep "2/21" "Creating team..."

$teamName = "Test Team $(Get-Date -Format 'HHmmss')"
$createTeamBody = @{
    name = $teamName
    description = "Automated test team - $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
    status = "active"
} | ConvertTo-Json

$createResponse = Test-Endpoint -Name "Create Team" `
    -Uri "$BaseUrl/api/v1/company-admin/teams" `
    -Method POST `
    -Headers $headers `
    -Body $createTeamBody `
    -ExpectedStatus 201

if ($createResponse -and $createResponse.data.id) {
    $teamId = $createResponse.data.id
    Write-Host "  Team ID: $teamId" -ForegroundColor Gray
} else {
    Write-Host "  WARNING: Could not create team. Some tests may fail." -ForegroundColor $ErrorColor
    $teamId = [System.Guid]::NewGuid().ToString()
}

# ============================================================================
# TEST 3: List Teams
# ============================================================================

Write-TestStep "3/21" "Listing teams..."

$listUri = $BaseUrl + '/api/v1/company-admin/teams?limit=10&offset=0'
$listResponse = Test-Endpoint -Name "List Teams" `
    -Uri $listUri `
    -Method GET `
    -Headers $headers

if ($listResponse) {
    Write-Host "  Teams found: $($listResponse.data.count)" -ForegroundColor Gray
}

# ============================================================================
# TEST 4: Get Team Details
# ============================================================================

Write-TestStep "4/21" "Getting team details..."

$teamResponse = Test-Endpoint -Name "Get Team Details" `
    -Uri "$BaseUrl/api/v1/company-admin/teams/$teamId" `
    -Method GET `
    -Headers $headers

# ============================================================================
# TEST 5: Update Team
# ============================================================================

Write-TestStep "5/21" "Updating team..."

$updateTeamBody = @{
    name = "$teamName - Updated"
    description = "Updated by automated test"
    status = "active"
} | ConvertTo-Json

$updateResponse = Test-Endpoint -Name "Update Team" `
    -Uri "$BaseUrl/api/v1/company-admin/teams/$teamId" `
    -Method PUT `
    -Headers $headers `
    -Body $updateTeamBody

# ============================================================================
# TEST 6: Get User for Member Tests
# ============================================================================

Write-TestStep "6/21" "Getting test user..."

try {
    $usersUri = $BaseUrl + '/api/v1/company-admin/users?limit=1'
    $usersResponse = Invoke-RestMethod -Uri $usersUri `
        -Method GET `
        -Headers $headers

    if ($usersResponse.data.users -and $usersResponse.data.users.Count -gt 0) {
        $userId = $usersResponse.data.users[0].id
        Write-TestPass "User found for testing"
        Write-Host "  User ID: $userId" -ForegroundColor Gray
    } else {
        $userId = [System.Guid]::NewGuid().ToString()
        Write-Host "  WARNING: No users found. Using fake ID." -ForegroundColor $ErrorColor
    }
} catch {
    $userId = [System.Guid]::NewGuid().ToString()
    Write-Host "  WARNING: Could not get user. Using fake ID." -ForegroundColor $ErrorColor
}

# ============================================================================
# TEST 7: Add Team Member
# ============================================================================

Write-TestStep "7/21" "Adding team member..."

$addMemberBody = @{
    user_id = $userId
    role_in_team = "driver"
} | ConvertTo-Json

$memberResponse = Test-Endpoint -Name "Add Team Member" `
    -Uri "$BaseUrl/api/v1/company-admin/teams/$teamId/members" `
    -Method POST `
    -Headers $headers `
    -Body $addMemberBody `
    -ExpectedStatus 201

# ============================================================================
# TEST 8: List Team Members
# ============================================================================

Write-TestStep "8/21" "Listing team members..."

$membersResponse = Test-Endpoint -Name "List Team Members" `
    -Uri "$BaseUrl/api/v1/company-admin/teams/$teamId/members" `
    -Method GET `
    -Headers $headers

if ($membersResponse) {
    Write-Host "  Members found: $($membersResponse.data.count)" -ForegroundColor Gray
}

# ============================================================================
# TEST 9: Update Member Role
# ============================================================================

Write-TestStep "9/21" "Updating member role..."

$updateRoleBody = @{
    role_in_team = "manager"
} | ConvertTo-Json

$roleResponse = Test-Endpoint -Name "Update Member Role" `
    -Uri "$BaseUrl/api/v1/company-admin/teams/$teamId/members/$userId/role" `
    -Method PUT `
    -Headers $headers `
    -Body $updateRoleBody

# ============================================================================
# TEST 10: Get Team Statistics (Before Vehicles)
# ============================================================================

Write-TestStep "10/21" "Getting team statistics (before vehicles)..."

$statsResponse = Test-Endpoint -Name "Get Team Statistics" `
    -Uri "$BaseUrl/api/v1/company-admin/teams/$teamId/stats" `
    -Method GET `
    -Headers $headers

if ($statsResponse -and $statsResponse.data) {
    Write-Host "  Members: $($statsResponse.data.member_count)" -ForegroundColor Gray
    Write-Host "  Vehicles: $($statsResponse.data.vehicle_count)" -ForegroundColor Gray
    Write-Host "  Active Vehicles: $($statsResponse.data.active_vehicles)" -ForegroundColor Gray
}

# ============================================================================
# TEST 11: Get Team Vehicles (Should be empty)
# ============================================================================

Write-TestStep "11/21" "Getting team vehicles..."

$vehiclesResponse = Test-Endpoint -Name "Get Team Vehicles" `
    -Uri "$BaseUrl/api/v1/company-admin/teams/$teamId/vehicles" `
    -Method GET `
    -Headers $headers

if ($vehiclesResponse) {
    Write-Host "  Vehicles count: $($vehiclesResponse.data.count)" -ForegroundColor Gray
}

# ============================================================================
# TEST 12: Create Test Vehicle
# ============================================================================

Write-TestStep "12/21" "Creating test vehicle..."

$createVehicleBody = @{
    license_plate = "TEST-$(Get-Date -Format 'HHmm')"
    brand = "Ford"
    model = "Transit"
    year = 2023
    vehicle_type = "van"
    fuel_type = "diesel"
    status = "active"
} | ConvertTo-Json

try {
    $vehicleCreateResponse = Invoke-RestMethod -Uri "$BaseUrl/api/v1/company-admin/vehicles" `
        -Method POST `
        -Headers $headers `
        -Body $createVehicleBody

    if ($vehicleCreateResponse.data.id) {
        $vehicleId = $vehicleCreateResponse.data.id
        $totalTests++
        $passedTests++
        Write-TestPass "Create Test Vehicle"
        Write-Host "  Vehicle ID: $vehicleId" -ForegroundColor Gray
        $testResults += @{
            Test = "Create Test Vehicle"
            Status = "PASS"
            Details = "Vehicle created"
        }
    } else {
        $vehicleId = [System.Guid]::NewGuid().ToString()
        $totalTests++
        $failedTests++
        Write-TestFail "Create Test Vehicle"
        Write-Host "  WARNING: Using fake vehicle ID" -ForegroundColor $ErrorColor
        $testResults += @{
            Test = "Create Test Vehicle"
            Status = "FAIL"
            Details = "No vehicle ID received"
        }
    }
} catch {
    $vehicleId = [System.Guid]::NewGuid().ToString()
    $totalTests++
    $failedTests++
    Write-TestFail "Create Test Vehicle - $($_.Exception.Message)"
    Write-Host "  WARNING: Using fake vehicle ID" -ForegroundColor $ErrorColor
    $testResults += @{
        Test = "Create Test Vehicle"
        Status = "FAIL"
        Details = $_.Exception.Message
    }
}

# ============================================================================
# TEST 13: Assign Vehicle to Team
# ============================================================================

Write-TestStep "13/21" "Assigning vehicle to team..."

$assignResponse = Test-Endpoint -Name "Assign Vehicle to Team" `
    -Uri "$BaseUrl/api/v1/company-admin/teams/$teamId/vehicles/$vehicleId" `
    -Method POST `
    -Headers $headers

# ============================================================================
# TEST 14: Get Team Vehicles (After Assignment)
# ============================================================================

Write-TestStep "14/21" "Getting team vehicles (after assignment)..."

$vehiclesResponse2 = Test-Endpoint -Name "Get Team Vehicles After Assignment" `
    -Uri "$BaseUrl/api/v1/company-admin/teams/$teamId/vehicles" `
    -Method GET `
    -Headers $headers

if ($vehiclesResponse2) {
    Write-Host "  Vehicles count: $($vehiclesResponse2.data.count)" -ForegroundColor Gray
}

# ============================================================================
# TEST 15: Get Team Statistics (After Vehicles)
# ============================================================================

Write-TestStep "15/21" "Getting team statistics (after vehicles)..."

$statsResponse2 = Test-Endpoint -Name "Get Team Statistics After Assignment" `
    -Uri "$BaseUrl/api/v1/company-admin/teams/$teamId/stats" `
    -Method GET `
    -Headers $headers

if ($statsResponse2 -and $statsResponse2.data) {
    Write-Host "  Members: $($statsResponse2.data.member_count)" -ForegroundColor Gray
    Write-Host "  Vehicles: $($statsResponse2.data.vehicle_count)" -ForegroundColor Gray
    Write-Host "  Active Vehicles: $($statsResponse2.data.active_vehicles)" -ForegroundColor Gray
}

# ============================================================================
# TEST 16: Unassign Vehicle from Team
# ============================================================================

Write-TestStep "16/21" "Unassigning vehicle from team..."

$unassignResponse = Test-Endpoint -Name "Unassign Vehicle from Team" `
    -Uri "$BaseUrl/api/v1/company-admin/teams/$teamId/vehicles/$vehicleId" `
    -Method DELETE `
    -Headers $headers

# ============================================================================
# TEST 17: Remove Team Member
# ============================================================================

Write-TestStep "17/21" "Removing team member..."

$removeResponse = Test-Endpoint -Name "Remove Team Member" `
    -Uri "$BaseUrl/api/v1/company-admin/teams/$teamId/members/$userId" `
    -Method DELETE `
    -Headers $headers

# ============================================================================
# TEST 18: Get My Teams
# ============================================================================

Write-TestStep "18/21" "Getting my teams..."

$myTeamsResponse = Test-Endpoint -Name "Get My Teams" `
    -Uri "$BaseUrl/api/v1/teams/my-teams" `
    -Method GET `
    -Headers $headers

# ============================================================================
# TEST 19: Invalid Team ID
# ============================================================================

Write-TestStep "19/21" "Testing invalid team ID..."

$null = Test-Endpoint -Name "Invalid Team ID" `
    -Uri "$BaseUrl/api/v1/company-admin/teams/invalid-uuid" `
    -Method GET `
    -Headers $headers `
    -ExpectedStatus 400

# ============================================================================
# TEST 20: Team Not Found
# ============================================================================

Write-TestStep "20/21" "Testing non-existent team..."

$fakeTeamId = [System.Guid]::NewGuid().ToString()
$null = Test-Endpoint -Name "Team Not Found" `
    -Uri "$BaseUrl/api/v1/company-admin/teams/$fakeTeamId" `
    -Method GET `
    -Headers $headers `
    -ExpectedStatus 404

# ============================================================================
# TEST 21: Delete Team
# ============================================================================

Write-TestStep "21/21" "Deleting team (cleanup)..."

$deleteResponse = Test-Endpoint -Name "Delete Team" `
    -Uri "$BaseUrl/api/v1/company-admin/teams/$teamId" `
    -Method DELETE `
    -Headers $headers

# ============================================================================
# RESULTS SUMMARY
# ============================================================================

Write-Host ""
Write-TestHeader "Test Results Summary"

Write-Host "Total Tests:  $totalTests" -ForegroundColor $InfoColor
Write-Host "Passed:       $passedTests" -ForegroundColor $SuccessColor
Write-Host "Failed:       $failedTests" -ForegroundColor $ErrorColor
Write-Host "Success Rate: $([math]::Round(($passedTests / $totalTests) * 100, 2))%" -ForegroundColor $InfoColor

Write-Host ""
Write-Host "Detailed Results:" -ForegroundColor $HeaderColor
Write-Host "----------------------------------------"

foreach ($result in $testResults) {
    $statusColor = if ($result.Status -eq "PASS") { $SuccessColor } else { $ErrorColor }
    Write-Host "[$($result.Status)]" -NoNewline -ForegroundColor $statusColor
    Write-Host " $($result.Test)" -ForegroundColor White
    Write-Host "   -> $($result.Details)" -ForegroundColor Gray
}

Write-Host ""
Write-TestHeader "Testing Complete!"

if ($failedTests -eq 0) {
    Write-Host "[SUCCESS] All tests passed!" -ForegroundColor $SuccessColor
    exit 0
} else {
    Write-Host "[WARNING] Some tests failed. Please review the results above." -ForegroundColor $ErrorColor
    exit 1
}
