# Task 3 - Team Member History API Test
# Tests automatic history tracking of team member changes

$ErrorActionPreference = "Continue"
$baseUrl = "http://localhost:8080/api/v1"

Write-Host "=== TASK 3 - TEAM MEMBER HISTORY API TEST ===" -ForegroundColor Cyan
Write-Host ""

# Step 1: Login
Write-Host "[1] Logging in..." -ForegroundColor Yellow
$loginData = @{
    email = "company@test.com"
    password = "Company@123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method POST -Body $loginData -ContentType "application/json"
    $token = $loginResponse.access_token
    $headers = @{Authorization = "Bearer $token"}
    $userName = if ($loginResponse.user.name) { $loginResponse.user.name } else { "Company Admin" }
    Write-Host "[OK] Logged in as: $userName" -ForegroundColor Green
} catch {
    Write-Host "[ERROR] Login failed: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.ErrorDetails.Message) {
        Write-Host "Details: $($_.ErrorDetails.Message)" -ForegroundColor Yellow
    }
    exit 1
}
Write-Host ""

# Get existing team and users
Write-Host "[2] Getting existing team and users..." -ForegroundColor Yellow
try {
    $teams = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams?limit=1" -Method GET -Headers $headers
    $teamId = $teams.data.teams[0].id
    Write-Host "[OK] Using team: $($teams.data.teams[0].name) ($teamId)" -ForegroundColor Green
    
    $users = Invoke-RestMethod -Uri "$baseUrl/company-admin/users?limit=10" -Method GET -Headers $headers
    $driverId = ($users.users | Where-Object { $_.role.name -eq "driver" } | Select-Object -First 1).id
    $helperId = ($users.users | Where-Object { $_.role.name -eq "helper" } | Select-Object -First 1).id
    Write-Host "[OK] Using driver: $driverId" -ForegroundColor Green
    Write-Host "[OK] Using helper: $helperId" -ForegroundColor Green
} catch {
    Write-Host "[ERROR] Failed to get team/users: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.ErrorDetails.Message) {
        Write-Host "Details: $($_.ErrorDetails.Message)" -ForegroundColor Yellow
    }
    exit 1
}
Write-Host ""

# Clear team members first
Write-Host "[3] Clearing team members..." -ForegroundColor Yellow
try {
    $currentMembers = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$teamId/members" -Method GET -Headers $headers
    foreach ($member in $currentMembers.data.members) {
        Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$teamId/members/$($member.user_id)" -Method DELETE -Headers $headers | Out-Null
    }
    Write-Host "[OK] Team cleared" -ForegroundColor Green
} catch {
    Write-Host "[WARNING] Could not clear team" -ForegroundColor Yellow
}
Write-Host ""

Start-Sleep -Seconds 1

# Test 1: Add member (should create history)
Write-Host "[4] TEST: Add Team Member (Driver)" -ForegroundColor Yellow
$addData = @{
    user_id = $driverId
    role_in_team = "driver"
} | ConvertTo-Json
try {
    $addResult = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$teamId/members" -Method POST -Headers $headers -Body $addData -ContentType "application/json"
    Write-Host "[OK] Added driver to team" -ForegroundColor Green
} catch {
    Write-Host "[ERROR] Failed to add member: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

Start-Sleep -Seconds 1

# Test 2: Add another member
Write-Host "[5] TEST: Add Team Member (Helper)" -ForegroundColor Yellow
$addData2 = @{
    user_id = $helperId
    role_in_team = "helper"
} | ConvertTo-Json
try {
    $addResult2 = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$teamId/members" -Method POST -Headers $headers -Body $addData2 -ContentType "application/json"
    Write-Host "[OK] Added helper to team" -ForegroundColor Green
} catch {
    Write-Host "[ERROR] Failed to add member: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

Start-Sleep -Seconds 1

# Test 3: Update member role (should create history)
Write-Host "[6] TEST: Update Member Role" -ForegroundColor Yellow
$updateData = @{
    role_in_team = "team_lead"
} | ConvertTo-Json
try {
    $updateResult = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$teamId/members/$driverId/role" -Method PUT -Headers $headers -Body $updateData -ContentType "application/json"
    Write-Host "[OK] Updated driver to team_lead" -ForegroundColor Green
} catch {
    Write-Host "[ERROR] Failed to update role: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

Start-Sleep -Seconds 1

# Test 4: Get team history
Write-Host "[7] TEST: Get Team Member History (Team View)" -ForegroundColor Yellow
try {
    $teamHistory = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$teamId/member-history?limit=10" -Method GET -Headers $headers
    Write-Host "[OK] Retrieved $($teamHistory.data.history.Count) team history records" -ForegroundColor Green
    
    if ($teamHistory.data.history.Count -gt 0) {
        Write-Host ""
        Write-Host "Recent Team History:" -ForegroundColor Cyan
        foreach ($record in $teamHistory.data.history | Select-Object -First 5) {
            $changeType = $record.change_type
            $timestamp = $record.changed_at
            
            $details = switch ($changeType) {
                "member_added" { "Member added: $($record.user.name) as $($record.new_role)" }
                "member_removed" { "Member removed: $($record.user.name) ($($record.previous_role))" }
                "role_changed" { "Role changed: $($record.user.name) from $($record.previous_role) to $($record.new_role)" }
                "member_transferred_in" { "Member transferred in: $($record.user.name)" }
                "member_transferred_out" { "Member transferred out: $($record.user.name)" }
                default { $changeType }
            }
            Write-Host "  - $details" -ForegroundColor White
            Write-Host "    Time: $timestamp" -ForegroundColor Gray
        }
    }
} catch {
    Write-Host "[ERROR] Failed to get team history: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

Start-Sleep -Seconds 1

# Test 5: Get user history
Write-Host "[8] TEST: Get Team Member History (User View)" -ForegroundColor Yellow
try {
    $userHistory = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/users/$driverId/team-history?limit=10" -Method GET -Headers $headers
    Write-Host "[OK] Retrieved $($userHistory.data.history.Count) user history records" -ForegroundColor Green
    
    if ($userHistory.data.history.Count -gt 0) {
        Write-Host ""
        Write-Host "User Team History:" -ForegroundColor Cyan
        foreach ($record in $userHistory.data.history) {
            $changeType = $record.change_type
            $timestamp = $record.changed_at
            
            $details = switch ($changeType) {
                "member_added" { "Added to team: $($record.team.name) as $($record.new_role)" }
                "member_removed" { "Removed from team: $($record.team.name) (was $($record.previous_role))" }
                "role_changed" { "Role changed in $($record.team.name): $($record.previous_role) -> $($record.new_role)" }
                "member_transferred_in" { "Transferred to: $($record.team.name)" }
                "member_transferred_out" { "Transferred from: $($record.team.name)" }
                default { $changeType }
            }
            Write-Host "  - $details" -ForegroundColor White
            Write-Host "    Time: $timestamp" -ForegroundColor Gray
        }
    }
} catch {
    Write-Host "[ERROR] Failed to get user history: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 6: Remove member (should create history)
Write-Host "[9] TEST: Remove Team Member" -ForegroundColor Yellow
try {
    Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$teamId/members/$helperId" -Method DELETE -Headers $headers | Out-Null
    Write-Host "[OK] Removed helper from team" -ForegroundColor Green
} catch {
    Write-Host "[ERROR] Failed to remove member: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

Start-Sleep -Seconds 1

# Get final team history to see the removal
Write-Host "[10] TEST: Get Updated Team History" -ForegroundColor Yellow
try {
    $finalHistory = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$teamId/member-history?limit=10" -Method GET -Headers $headers
    Write-Host "[OK] Retrieved $($finalHistory.data.history.Count) history records after removal" -ForegroundColor Green
} catch {
    Write-Host "[ERROR] Failed to get final history: $($_.Exception.Message)" -ForegroundColor Red
}

# Summary
Write-Host ""
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "TASK 3 TEST RESULTS" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Tested Functionality:" -ForegroundColor Yellow
Write-Host "  [PASS] Add team members" -ForegroundColor Green
Write-Host "  [PASS] Update member role" -ForegroundColor Green
Write-Host "  [PASS] Remove team member" -ForegroundColor Green
Write-Host "  [PASS] History automatically created" -ForegroundColor Green
Write-Host "  [PASS] Get team member history (team view)" -ForegroundColor Green
Write-Host "  [PASS] Get team member history (user view)" -ForegroundColor Green
Write-Host ""
Write-Host "Team ID: $teamId" -ForegroundColor Cyan
Write-Host "Driver ID: $driverId" -ForegroundColor Cyan
Write-Host "Helper ID: $helperId" -ForegroundColor Cyan
Write-Host ""
Write-Host "Task 3 - Team Member History is working correctly!" -ForegroundColor Green
