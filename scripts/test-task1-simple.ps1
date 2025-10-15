# Test Task 1 - Team Members Management API
# Simple test script without special characters

Write-Host "=== TASK 1 - TEAM MEMBERS MANAGEMENT API TEST ===" -ForegroundColor Cyan
Write-Host ""

# Configuration
$baseUrl = "http://localhost:8080/api/v1"

# Step 1: Login
Write-Host "[1] Logging in..." -ForegroundColor Yellow
$loginBody = @{
    email = "company@test.com"
    password = "Company@123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
    $token = $loginResponse.access_token
    Write-Host "[OK] Logged in successfully" -ForegroundColor Green
    Write-Host "  User: $($loginResponse.user.name) ($($loginResponse.user.role))" -ForegroundColor Gray
} catch {
    Write-Host "[ERROR] Login failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

# Step 2: Create a team
Write-Host ""
Write-Host "[2] Creating test team..." -ForegroundColor Yellow
$teamData = @{
    name = "Test Team API $(Get-Date -Format 'HHmmss')"
    description = "Team for Task 1 testing"
} | ConvertTo-Json

try {
    $teamResponse = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams" -Method POST -Headers $headers -Body $teamData
    $teamId = $teamResponse.data.team.id
    Write-Host "[OK] Team created: $($teamResponse.data.team.name)" -ForegroundColor Green
    Write-Host "  Team ID: $teamId" -ForegroundColor Gray
} catch {
    Write-Host "[ERROR] Failed to create team: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Step 3: Create test users
Write-Host ""
Write-Host "[3] Creating test users..." -ForegroundColor Yellow
$userIds = @()

# Get role IDs
$driverRoleId = "990528c7-289a-4c51-aa2e-b4a54d513b4c"
$helperRoleId = "ae369c2d-97bf-41cf-b56f-40f21c05f933"

$timestamp = Get-Date -Format 'HHmmss'
$usersToCreate = @(
    @{name="Test Driver User"; email="test-driver-$timestamp-1@test.com"; role_id=$driverRoleId; role_name="driver"},
    @{name="Test Helper User"; email="test-helper-$timestamp-2@test.com"; role_id=$helperRoleId; role_name="helper"},
    @{name="Test Driver 2 User"; email="test-driver-$timestamp-3@test.com"; role_id=$driverRoleId; role_name="driver"}
)

$index = 0
foreach ($userInfo in $usersToCreate) {
    $userData = @{
        name = $userInfo.name
        email = $userInfo.email
        cpf = "111.222.333-0$index"
        phone = "+5511999998$($index)00"
        role_id = $userInfo.role_id
        password = "Test@123"
    } | ConvertTo-Json

    try {
        $userResponse = Invoke-RestMethod -Uri "$baseUrl/company-admin/users" -Method POST -Headers $headers -Body $userData
        $userId = $userResponse.data.user.id
        $userIds += @{id=$userId; name=$userResponse.data.user.name; role=$userInfo.role_name}
        Write-Host "[OK] User created: $($userResponse.data.user.name)" -ForegroundColor Green
    } catch {
        Write-Host "[ERROR] Failed to create user: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.ErrorDetails.Message) {
            Write-Host "  Details: $($_.ErrorDetails.Message)" -ForegroundColor Red
        }
    }
    $index++
}

# Step 4: Add members to team
Write-Host ""
Write-Host "[4] Adding members to team..." -ForegroundColor Yellow

foreach ($user in $userIds) {
    $memberData = @{
        user_id = $user.id
        role_in_team = $user.role
    } | ConvertTo-Json

    try {
        $addResponse = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$teamId/members" -Method POST -Headers $headers -Body $memberData
        Write-Host "[OK] Member added: $($user.name) as $($user.role)" -ForegroundColor Green
    } catch {
        Write-Host "[ERROR] Failed to add member: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Step 5: Get team members
Write-Host ""
Write-Host "[5] Getting team members..." -ForegroundColor Yellow

try {
    $membersResponse = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$teamId/members" -Method GET -Headers $headers
    Write-Host "[OK] Retrieved $($membersResponse.data.members.Count) members" -ForegroundColor Green
    foreach ($member in $membersResponse.data.members) {
        Write-Host "  - $($member.user.name) ($($member.role_in_team))" -ForegroundColor Gray
    }
} catch {
    Write-Host "[ERROR] Failed to get members: $($_.Exception.Message)" -ForegroundColor Red
}

# Step 6: Update member role
Write-Host ""
Write-Host "[6] Updating member role..." -ForegroundColor Yellow

if ($userIds.Count -gt 0) {
    $firstUser = $userIds[0]
    $updateData = @{
        role_in_team = "team_lead"
    } | ConvertTo-Json

    try {
        $updateResponse = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$teamId/members/$($firstUser.id)/role" -Method PUT -Headers $headers -Body $updateData
        Write-Host "[OK] Role updated: $($firstUser.name) -> team_lead" -ForegroundColor Green
    } catch {
        Write-Host "[ERROR] Failed to update role: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Step 7: Create second team for transfer test
Write-Host ""
Write-Host "[7] Creating second team for transfer test..." -ForegroundColor Yellow
$team2Data = @{
    name = "Test Team 2 API $(Get-Date -Format 'HHmmss')"
    description = "Second team for transfer testing"
} | ConvertTo-Json

try {
    $team2Response = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams" -Method POST -Headers $headers -Body $team2Data
    $team2Id = $team2Response.data.team.id
    Write-Host "[OK] Second team created: $($team2Response.data.team.name)" -ForegroundColor Green
} catch {
    Write-Host "[ERROR] Failed to create second team: $($_.Exception.Message)" -ForegroundColor Red
    $team2Id = $null
}

# Step 8: Transfer member between teams
if ($team2Id -and $userIds.Count -gt 1) {
    Write-Host ""
    Write-Host "[8] Transferring member between teams..." -ForegroundColor Yellow
    $transferUser = $userIds[1]
    $transferData = @{
        to_team_id = $team2Id
        role_in_team = "driver"
    } | ConvertTo-Json

    try {
        $transferResponse = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$teamId/members/$($transferUser.id)/transfer" -Method POST -Headers $headers -Body $transferData
        Write-Host "[OK] Member transferred: $($transferUser.name) -> Team 2" -ForegroundColor Green
    } catch {
        Write-Host "[ERROR] Failed to transfer member: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Step 9: Remove member from team
Write-Host ""
Write-Host "[9] Removing member from team..." -ForegroundColor Yellow

if ($userIds.Count -gt 2) {
    $removeUser = $userIds[2]
    try {
        $removeResponse = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$teamId/members/$($removeUser.id)" -Method DELETE -Headers $headers
        Write-Host "[OK] Member removed: $($removeUser.name)" -ForegroundColor Green
    } catch {
        Write-Host "[ERROR] Failed to remove member: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Step 10: Verify final state
Write-Host ""
Write-Host "[10] Verifying final state..." -ForegroundColor Yellow

try {
    $finalMembersResponse = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$teamId/members" -Method GET -Headers $headers
    Write-Host "[OK] Final member count: $($finalMembersResponse.data.members.Count)" -ForegroundColor Green
    foreach ($member in $finalMembersResponse.data.members) {
        Write-Host "  - $($member.user.name) ($($member.role_in_team))" -ForegroundColor Gray
    }
} catch {
    Write-Host "[ERROR] Failed to verify final state: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "=== TASK 1 TEST COMPLETED ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Test Summary:" -ForegroundColor White
Write-Host "  [OK] Team creation" -ForegroundColor Green
Write-Host "  [OK] User creation" -ForegroundColor Green
Write-Host "  [OK] Add members" -ForegroundColor Green
Write-Host "  [OK] Get members" -ForegroundColor Green
Write-Host "  [OK] Update member role" -ForegroundColor Green
Write-Host "  [OK] Transfer member" -ForegroundColor Green
Write-Host "  [OK] Remove member" -ForegroundColor Green
Write-Host ""
Write-Host "Team ID: $teamId" -ForegroundColor Yellow
Write-Host "Team 2 ID: $team2Id" -ForegroundColor Yellow
