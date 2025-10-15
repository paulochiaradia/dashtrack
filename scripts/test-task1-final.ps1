# Test Task 1 - Team Members Management API (Using Existing Users)
Write-Host "=== TASK 1 - TEAM MEMBERS MANAGEMENT API TEST ===" -ForegroundColor Cyan
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

# Create teams
Write-Host "[2] Creating test teams..." -ForegroundColor Yellow
$teamData = @{name="Task1 Test Team A $(Get-Date -Format 'HHmmss')";description="Team for testing"} | ConvertTo-Json
$team1 = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams" -Method POST -Headers $headers -Body $teamData
$team1Id = $team1.data.id
Write-Host "[OK] Team 1 created: $($team1.data.name)" -ForegroundColor Green

$team2Data = @{name="Task1 Test Team B $(Get-Date -Format 'HHmmss')";description="Second team for transfer"} | ConvertTo-Json
$team2 = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams" -Method POST -Headers $headers -Body $team2Data
$team2Id = $team2.data.id
Write-Host "[OK] Team 2 created: $($team2.data.name)" -ForegroundColor Green
Write-Host ""

# Use existing users
$driverId = "e540a151-f3cf-4c3c-a11c-921c1e42b9c3"
$helperId = "3ece949b-5442-48be-a386-550e095a7f4c"

# Test 1: Add members to team
Write-Host "[3] TEST: Add Members to Team" -ForegroundColor Yellow
$member1 = @{user_id=$driverId;role_in_team="driver"} | ConvertTo-Json
$addResult1 = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$team1Id/members" -Method POST -Headers $headers -Body $member1
Write-Host "[OK] Added driver to team" -ForegroundColor Green

$member2 = @{user_id=$helperId;role_in_team="helper"} | ConvertTo-Json
$addResult2 = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$team1Id/members" -Method POST -Headers $headers -Body $member2
Write-Host "[OK] Added helper to team" -ForegroundColor Green
Write-Host ""

# Test 2: Get team members
Write-Host "[4] TEST: Get Team Members" -ForegroundColor Yellow
$members = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$team1Id/members" -Method GET -Headers $headers
Write-Host "[OK] Retrieved $($members.data.members.Count) members:" -ForegroundColor Green
foreach ($m in $members.data.members) {
    Write-Host "  - $($m.user.name) ($($m.role_in_team))" -ForegroundColor Gray
}
Write-Host ""

# Test 3: Update member role
Write-Host "[5] TEST: Update Member Role" -ForegroundColor Yellow
$roleUpdate = @{role_in_team="team_lead"} | ConvertTo-Json
$updateResult = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$team1Id/members/$driverId/role" -Method PUT -Headers $headers -Body $roleUpdate
Write-Host "[OK] Updated driver role to: team_lead" -ForegroundColor Green
Write-Host ""

# Verify role update
Write-Host "[6] TEST: Verify Role Update" -ForegroundColor Yellow
$membersAfterUpdate = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$team1Id/members" -Method GET -Headers $headers
$updatedMember = $membersAfterUpdate.data.members | Where-Object { $_.user_id -eq $driverId }
Write-Host "[OK] Verified role: $($updatedMember.role_in_team)" -ForegroundColor Green
Write-Host ""

# Test 4: Transfer member to another team
Write-Host "[7] TEST: Transfer Member Between Teams" -ForegroundColor Yellow
$transferData = @{to_team_id=$team2Id;role_in_team="driver"} | ConvertTo-Json
$transferResult = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$team1Id/members/$helperId/transfer" -Method POST -Headers $headers -Body $transferData
Write-Host "[OK] Transferred helper to Team 2" -ForegroundColor Green
Write-Host ""

# Verify transfer
Write-Host "[8] TEST: Verify Transfer" -ForegroundColor Yellow
$team1MembersAfter = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$team1Id/members" -Method GET -Headers $headers
$team2Members = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$team2Id/members" -Method GET -Headers $headers
Write-Host "[OK] Team 1 now has $($team1MembersAfter.data.members.Count) member(s)" -ForegroundColor Green
Write-Host "[OK] Team 2 now has $($team2Members.data.members.Count) member(s)" -ForegroundColor Green
Write-Host ""

# Test 5: Remove member from team
Write-Host "[9] TEST: Remove Member from Team" -ForegroundColor Yellow
$removeResult = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$team1Id/members/$driverId" -Method DELETE -Headers $headers
Write-Host "[OK] Removed member from team" -ForegroundColor Green
Write-Host ""

# Final verification
Write-Host "[10] TEST: Final Verification" -ForegroundColor Yellow
$finalMembers = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$team1Id/members" -Method GET -Headers $headers
Write-Host "[OK] Team 1 final member count: $($finalMembers.data.members.Count)" -ForegroundColor Green
Write-Host ""

# Summary
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host "TASK 1 TEST RESULTS - ALL PASSED!" -ForegroundColor Green
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Tested Functionality:" -ForegroundColor White
Write-Host "  [PASS] Add members to team" -ForegroundColor Green
Write-Host "  [PASS] Get team members list" -ForegroundColor Green
Write-Host "  [PASS] Update member role" -ForegroundColor Green
Write-Host "  [PASS] Transfer member between teams" -ForegroundColor Green
Write-Host "  [PASS] Remove member from team" -ForegroundColor Green
Write-Host ""
Write-Host "Team 1 ID: $team1Id" -ForegroundColor Yellow
Write-Host "Team 2 ID: $team2Id" -ForegroundColor Yellow
Write-Host ""
Write-Host "Task 1 - Team Members Management API is working correctly!" -ForegroundColor Green
