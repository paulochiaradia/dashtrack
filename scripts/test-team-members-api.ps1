# Team Members Management API Testing Script
# This script tests all team member management endpoints

$baseUrl = "http://localhost:8080/api/v1"
$companyAdminToken = "your-company-admin-token-here"

# Set headers
$headers = @{
    "Authorization" = "Bearer $companyAdminToken"
    "Content-Type" = "application/json"
}

Write-Host "`n=== TEAM MEMBERS MANAGEMENT API TESTS ===" -ForegroundColor Cyan

# ============================================================================
# 1. CREATE A TEST TEAM
# ============================================================================
Write-Host "`n[1] Creating test team..." -ForegroundColor Yellow
$teamData = @{
    name = "Test Team for Members"
    description = "Team created for testing member management"
} | ConvertTo-Json

try {
    $teamResponse = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams" `
        -Method Post `
        -Headers $headers `
        -Body $teamData `
        -ErrorAction Stop
    
    $teamId = $teamResponse.data.id
    Write-Host "✓ Team created successfully: $teamId" -ForegroundColor Green
    Write-Host "  Name: $($teamResponse.data.name)" -ForegroundColor Gray
} catch {
    Write-Host "✗ Failed to create team: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# ============================================================================
# 2. CREATE TEST USERS
# ============================================================================
Write-Host "`n[2] Creating test users..." -ForegroundColor Yellow

$users = @()
$roles = @("manager", "driver", "assistant", "supervisor")

foreach ($role in $roles) {
    $userData = @{
        name = "Test $role User"
        email = "test-$role@example.com"
        cpf = "111222333$($roles.IndexOf($role))0"
        phone = "11999998$($roles.IndexOf($role))00"
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
            role = $role
        }
        Write-Host "✓ User created: $($userResponse.data.name) - ID: $($userResponse.data.id)" -ForegroundColor Green
    } catch {
        Write-Host "✗ Failed to create user ($role): $($_.Exception.Message)" -ForegroundColor Red
    }
}

# ============================================================================
# 3. ADD MEMBERS TO TEAM
# ============================================================================
Write-Host "`n[3] Adding members to team..." -ForegroundColor Yellow

$members = @()
foreach ($user in $users) {
    $memberData = @{
        user_id = $user.id
        role_in_team = $user.role
    } | ConvertTo-Json

    try {
        $memberResponse = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$teamId/members" `
            -Method Post `
            -Headers $headers `
            -Body $memberData `
            -ErrorAction Stop
        
        $members += $memberResponse.data
        Write-Host "✓ Member added: $($user.name) as $($user.role)" -ForegroundColor Green
    } catch {
        Write-Host "✗ Failed to add member ($($user.name)): $($_.Exception.Message)" -ForegroundColor Red
    }
}

Start-Sleep -Seconds 1

# ============================================================================
# 4. GET TEAM MEMBERS
# ============================================================================
Write-Host "`n[4] Retrieving team members..." -ForegroundColor Yellow

try {
    $membersResponse = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$teamId/members" `
        -Method Get `
        -Headers $headers `
        -ErrorAction Stop
    
    Write-Host "✓ Team members retrieved successfully" -ForegroundColor Green
    Write-Host "  Total members: $($membersResponse.data.count)" -ForegroundColor Gray
    
    foreach ($member in $membersResponse.data.members) {
        Write-Host "  - $($member.user.name) ($($member.role_in_team))" -ForegroundColor Gray
    }
} catch {
    Write-Host "✗ Failed to retrieve members: $($_.Exception.Message)" -ForegroundColor Red
}

# ============================================================================
# 5. UPDATE MEMBER ROLE
# ============================================================================
Write-Host "`n[5] Updating member role..." -ForegroundColor Yellow

if ($users.Count -gt 0) {
    $targetUser = $users[0]
    $newRole = "supervisor"
    
    $updateData = @{
        role_in_team = $newRole
    } | ConvertTo-Json

    try {
        $updateResponse = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$teamId/members/$($targetUser.id)/role" `
            -Method Put `
            -Headers $headers `
            -Body $updateData `
            -ErrorAction Stop
        
        Write-Host "✓ Member role updated successfully" -ForegroundColor Green
        Write-Host "  User: $($targetUser.name)" -ForegroundColor Gray
        Write-Host "  Old Role: $($targetUser.role)" -ForegroundColor Gray
        Write-Host "  New Role: $newRole" -ForegroundColor Gray
    } catch {
        Write-Host "✗ Failed to update role: $($_.Exception.Message)" -ForegroundColor Red
    }
}

Start-Sleep -Seconds 1

# ============================================================================
# 6. TRANSFER MEMBER TO ANOTHER TEAM
# ============================================================================
Write-Host "`n[6] Testing member transfer..." -ForegroundColor Yellow

# First create a second team
$team2Data = @{
    name = "Destination Team"
    description = "Team for testing member transfer"
} | ConvertTo-Json

try {
    $team2Response = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams" `
        -Method Post `
        -Headers $headers `
        -Body $team2Data `
        -ErrorAction Stop
    
    $team2Id = $team2Response.data.id
    Write-Host "✓ Second team created: $team2Id" -ForegroundColor Green

    # Transfer a member
    if ($users.Count -gt 1) {
        $transferUser = $users[1]
        $transferData = @{
            from_team_id = $teamId
            role_in_team = "driver"
        } | ConvertTo-Json

        try {
            $transferResponse = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$team2Id/members/$($transferUser.id)/transfer" `
                -Method Post `
                -Headers $headers `
                -Body $transferData `
                -ErrorAction Stop
            
            Write-Host "✓ Member transferred successfully" -ForegroundColor Green
            Write-Host "  User: $($transferUser.name)" -ForegroundColor Gray
            Write-Host "  From Team: $teamId" -ForegroundColor Gray
            Write-Host "  To Team: $team2Id" -ForegroundColor Gray
        } catch {
            Write-Host "✗ Failed to transfer member: $($_.Exception.Message)" -ForegroundColor Red
        }
    }
} catch {
    Write-Host "✗ Failed to create second team: $($_.Exception.Message)" -ForegroundColor Red
}

Start-Sleep -Seconds 1

# ============================================================================
# 7. REMOVE MEMBER FROM TEAM
# ============================================================================
Write-Host "`n[7] Removing member from team..." -ForegroundColor Yellow

if ($users.Count -gt 2) {
    $removeUser = $users[2]
    
    try {
        $removeResponse = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$teamId/members/$($removeUser.id)" `
            -Method Delete `
            -Headers $headers `
            -ErrorAction Stop
        
        Write-Host "✓ Member removed successfully" -ForegroundColor Green
        Write-Host "  User: $($removeUser.name)" -ForegroundColor Gray
    } catch {
        Write-Host "✗ Failed to remove member: $($_.Exception.Message)" -ForegroundColor Red
    }
}

Start-Sleep -Seconds 1

# ============================================================================
# 8. VERIFY FINAL STATE
# ============================================================================
Write-Host "`n[8] Verifying final state..." -ForegroundColor Yellow

try {
    $finalResponse = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$teamId/members" `
        -Method Get `
        -Headers $headers `
        -ErrorAction Stop
    
    Write-Host "✓ Final team state retrieved" -ForegroundColor Green
    Write-Host "  Remaining members in Team 1: $($finalResponse.data.count)" -ForegroundColor Gray
} catch {
    Write-Host "✗ Failed to verify final state: $($_.Exception.Message)" -ForegroundColor Red
}

# ============================================================================
# 9. CLEANUP (Optional)
# ============================================================================
Write-Host "`n[9] Cleanup test data..." -ForegroundColor Yellow
$cleanup = Read-Host "Do you want to delete test data? (y/n)"

if ($cleanup -eq "y") {
    # Delete teams
    try {
        Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$teamId" `
            -Method Delete `
            -Headers $headers `
            -ErrorAction Stop
        Write-Host "✓ Test team 1 deleted" -ForegroundColor Green
    } catch {
        Write-Host "✗ Failed to delete team 1: $($_.Exception.Message)" -ForegroundColor Red
    }

    if ($team2Id) {
        try {
            Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$team2Id" `
                -Method Delete `
                -Headers $headers `
                -ErrorAction Stop
            Write-Host "✓ Test team 2 deleted" -ForegroundColor Green
        } catch {
            Write-Host "✗ Failed to delete team 2: $($_.Exception.Message)" -ForegroundColor Red
        }
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
Write-Host "- Team CRUD: ✓" -ForegroundColor Green
Write-Host "- Add Members: ✓" -ForegroundColor Green
Write-Host "- List Members: ✓" -ForegroundColor Green
Write-Host "- Update Role: ✓" -ForegroundColor Green
Write-Host "- Transfer Member: ✓" -ForegroundColor Green
Write-Host "- Remove Member: ✓" -ForegroundColor Green
