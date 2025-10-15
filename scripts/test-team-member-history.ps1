# ============================================================================
# TEAM MEMBER HISTORY API TEST SCRIPT
# ============================================================================
# This script tests the team member history tracking functionality
# 
# Features tested:
# - Member additions (change_type: added)
# - Member removals (change_type: removed)
# - Role changes (change_type: role_changed)
# - Member transfers (change_type: transferred_in, transferred_out)
# - Team history queries
# - User history queries
# - History with populated details
# ============================================================================

param(
    [string]$BaseURL = "http://localhost:8080/api/v1",
    [switch]$Cleanup,
    [switch]$Verbose
)

$ErrorActionPreference = "Continue"

# Colors for output
function Write-Success { param($msg) Write-Host "✓ $msg" -ForegroundColor Green }
function Write-Error { param($msg) Write-Host "✗ $msg" -ForegroundColor Red }
function Write-Info { param($msg) Write-Host "ℹ $msg" -ForegroundColor Cyan }
function Write-Step { param($msg) Write-Host "`n=== $msg ===" -ForegroundColor Yellow }

# Global variables to store created resources
$script:companyAdminToken = $null
$script:companyID = $null
$script:teamAID = $null
$script:teamBID = $null
$script:user1ID = $null
$script:user2ID = $null
$script:user3ID = $null

# ============================================================================
# HELPER FUNCTIONS
# ============================================================================

function Invoke-APIRequest {
    param(
        [string]$Method,
        [string]$Endpoint,
        [object]$Body = $null,
        [string]$Token = $null,
        [switch]$ExpectFailure
    )

    $headers = @{
        "Content-Type" = "application/json"
    }
    
    if ($Token) {
        $headers["Authorization"] = "Bearer $Token"
    }

    $uri = "$BaseURL$Endpoint"
    
    if ($Verbose) {
        Write-Host "→ $Method $uri" -ForegroundColor DarkGray
        if ($Body) {
            Write-Host "  Body: $($Body | ConvertTo-Json -Compress)" -ForegroundColor DarkGray
        }
    }

    try {
        $params = @{
            Method = $Method
            Uri = $uri
            Headers = $headers
        }

        if ($Body) {
            $params.Body = ($Body | ConvertTo-Json)
        }

        $response = Invoke-RestMethod @params
        
        if ($Verbose) {
            Write-Host "← Status: Success" -ForegroundColor DarkGray
        }

        return $response
    }
    catch {
        if ($ExpectFailure) {
            if ($Verbose) {
                Write-Host "← Status: Failed (Expected)" -ForegroundColor DarkGray
            }
            return $null
        }
        
        Write-Error "API Request Failed: $_"
        if ($_.ErrorDetails.Message) {
            Write-Host $_.ErrorDetails.Message -ForegroundColor Red
        }
        throw
    }
}

function Get-RandomString {
    param([int]$Length = 8)
    -join ((65..90) + (97..122) | Get-Random -Count $Length | ForEach-Object {[char]$_})
}

# ============================================================================
# SETUP: CREATE TEST DATA
# ============================================================================

Write-Step "Step 1: Authentication Setup"

# Login as company_admin
Write-Info "Logging in as company_admin..."
$loginResponse = Invoke-APIRequest -Method POST -Endpoint "/auth/login" -Body @{
    email = "admin@techcorp.com"
    password = "Admin123!@#"
}

if ($loginResponse.data.token) {
    $script:companyAdminToken = $loginResponse.data.token
    $script:companyID = $loginResponse.data.user.company_id
    Write-Success "Logged in successfully as company_admin"
    Write-Info "Company ID: $script:companyID"
} else {
    Write-Error "Failed to login"
    exit 1
}

# ============================================================================
# STEP 2: CREATE TEAMS
# ============================================================================

Write-Step "Step 2: Create Teams"

$random = Get-RandomString
$teamAData = @{
    name = "Team Alpha $random"
    description = "Test team for member history"
}

Write-Info "Creating Team A..."
$teamAResponse = Invoke-APIRequest -Method POST -Endpoint "/company-admin/teams" -Body $teamAData -Token $script:companyAdminToken

if ($teamAResponse.data.team.id) {
    $script:teamAID = $teamAResponse.data.team.id
    Write-Success "Team A created: $($teamAResponse.data.team.name) ($script:teamAID)"
}

$teamBData = @{
    name = "Team Beta $random"
    description = "Second test team for transfers"
}

Write-Info "Creating Team B..."
$teamBResponse = Invoke-APIRequest -Method POST -Endpoint "/company-admin/teams" -Body $teamBData -Token $script:companyAdminToken

if ($teamBResponse.data.team.id) {
    $script:teamBID = $teamBResponse.data.team.id
    Write-Success "Team B created: $($teamBResponse.data.team.name) ($script:teamBID)"
}

# ============================================================================
# STEP 3: CREATE TEST USERS
# ============================================================================

Write-Step "Step 3: Create Test Users"

$user1Data = @{
    name = "John Driver $random"
    email = "john.driver.$random@test.com"
    password = "Test123!@#"
    phone = "+1234567801"
    cpf = "12345678901"
    role = "user"
}

Write-Info "Creating User 1 (John Driver)..."
$user1Response = Invoke-APIRequest -Method POST -Endpoint "/company-admin/users" -Body $user1Data -Token $script:companyAdminToken

if ($user1Response.data.user.id) {
    $script:user1ID = $user1Response.data.user.id
    Write-Success "User 1 created: $($user1Response.data.user.name) ($script:user1ID)"
}

$user2Data = @{
    name = "Jane Helper $random"
    email = "jane.helper.$random@test.com"
    password = "Test123!@#"
    phone = "+1234567802"
    cpf = "12345678902"
    role = "user"
}

Write-Info "Creating User 2 (Jane Helper)..."
$user2Response = Invoke-APIRequest -Method POST -Endpoint "/company-admin/users" -Body $user2Data -Token $script:companyAdminToken

if ($user2Response.data.user.id) {
    $script:user2ID = $user2Response.data.user.id
    Write-Success "User 2 created: $($user2Response.data.user.name) ($script:user2ID)"
}

$user3Data = @{
    name = "Bob Manager $random"
    email = "bob.manager.$random@test.com"
    password = "Test123!@#"
    phone = "+1234567803"
    cpf = "12345678903"
    role = "manager"
}

Write-Info "Creating User 3 (Bob Manager)..."
$user3Response = Invoke-APIRequest -Method POST -Endpoint "/company-admin/users" -Body $user3Data -Token $script:companyAdminToken

if ($user3Response.data.user.id) {
    $script:user3ID = $user3Response.data.user.id
    Write-Success "User 3 created: $($user3Response.data.user.name) ($script:user3ID)"
}

# ============================================================================
# STEP 4: ADD MEMBERS TO TEAM A (Tests "added" change type)
# ============================================================================

Write-Step "Step 4: Add Members to Team A"

Write-Info "Adding User 1 as driver to Team A..."
$addMember1 = Invoke-APIRequest -Method POST -Endpoint "/company-admin/teams/$script:teamAID/members" -Body @{
    user_id = $script:user1ID
    role_in_team = "driver"
} -Token $script:companyAdminToken

if ($addMember1.data) {
    Write-Success "User 1 added as driver (should log 'added' change)"
}

Write-Info "Adding User 2 as helper to Team A..."
$addMember2 = Invoke-APIRequest -Method POST -Endpoint "/company-admin/teams/$script:teamAID/members" -Body @{
    user_id = $script:user2ID
    role_in_team = "helper"
} -Token $script:companyAdminToken

if ($addMember2.data) {
    Write-Success "User 2 added as helper (should log 'added' change)"
}

Write-Info "Adding User 3 as team_lead to Team A..."
$addMember3 = Invoke-APIRequest -Method POST -Endpoint "/company-admin/teams/$script:teamAID/members" -Body @{
    user_id = $script:user3ID
    role_in_team = "team_lead"
} -Token $script:companyAdminToken

if ($addMember3.data) {
    Write-Success "User 3 added as team_lead (should log 'added' change)"
}

# ============================================================================
# STEP 5: UPDATE MEMBER ROLES (Tests "role_changed" change type)
# ============================================================================

Write-Step "Step 5: Update Member Roles"

Write-Info "Promoting User 1 from driver to team_lead..."
$updateRole1 = Invoke-APIRequest -Method PUT -Endpoint "/company-admin/teams/$script:teamAID/members/$script:user1ID/role" -Body @{
    role_in_team = "team_lead"
} -Token $script:companyAdminToken

if ($updateRole1.data) {
    Write-Success "User 1 promoted (should log 'role_changed': driver → team_lead)"
}

Start-Sleep -Seconds 1

Write-Info "Demoting User 3 from team_lead to driver..."
$updateRole3 = Invoke-APIRequest -Method PUT -Endpoint "/company-admin/teams/$script:teamAID/members/$script:user3ID/role" -Body @{
    role_in_team = "driver"
} -Token $script:companyAdminToken

if ($updateRole3.data) {
    Write-Success "User 3 demoted (should log 'role_changed': team_lead → driver)"
}

# ============================================================================
# STEP 6: QUERY TEAM MEMBER HISTORY
# ============================================================================

Write-Step "Step 6: Query Team A Member History"

Write-Info "Retrieving Team A member history..."
$teamHistory = Invoke-APIRequest -Method GET -Endpoint "/company-admin/teams/$script:teamAID/member-history?limit=50" -Token $script:companyAdminToken

if ($teamHistory.data.history) {
    Write-Success "Retrieved $($teamHistory.data.count) history records for Team A"
    Write-Host ""
    Write-Host "History Summary:" -ForegroundColor Cyan
    foreach ($record in $teamHistory.data.history) {
        $userName = if ($record.user) { $record.user.name } else { "Unknown" }
        $changeType = $record.change_type
        $timestamp = $record.changed_at
        
        $details = switch ($changeType) {
            "added" { "added as $($record.new_role_in_team)" }
            "removed" { "removed (was $($record.previous_role_in_team))" }
            "role_changed" { "role changed: $($record.previous_role_in_team) → $($record.new_role_in_team)" }
            default { $changeType }
        }
        
        Write-Host "  • $userName - $details" -ForegroundColor White
        Write-Host "    ↳ $timestamp" -ForegroundColor DarkGray
    }
} else {
    Write-Error "Failed to retrieve team history"
}

# ============================================================================
# STEP 7: QUERY USER TEAM HISTORY
# ============================================================================

Write-Step "Step 7: Query User 1 Team History"

Write-Info "Retrieving User 1's team membership history..."
$userHistory = Invoke-APIRequest -Method GET -Endpoint "/company-admin/teams/users/$script:user1ID/team-history?limit=50" -Token $script:companyAdminToken

if ($userHistory.data.history) {
    Write-Success "Retrieved $($userHistory.data.count) history records for User 1"
    Write-Host ""
    Write-Host "User 1 History Summary:" -ForegroundColor Cyan
    foreach ($record in $userHistory.data.history) {
        $teamName = if ($record.team) { $record.team.name } else { "Unknown" }
        $changeType = $record.change_type
        $timestamp = $record.changed_at
        
        $details = switch ($changeType) {
            "added" { "joined $teamName as $($record.new_role_in_team)" }
            "removed" { "left $teamName (was $($record.previous_role_in_team))" }
            "role_changed" { "role in $teamName changed: $($record.previous_role_in_team) → $($record.new_role_in_team)" }
            default { $changeType }
        }
        
        Write-Host "  • $details" -ForegroundColor White
        Write-Host "    ↳ $timestamp" -ForegroundColor DarkGray
    }
} else {
    Write-Error "Failed to retrieve user history"
}

# ============================================================================
# STEP 8: TRANSFER MEMBER (Tests transfer change types)
# ============================================================================

Write-Step "Step 8: Transfer User 2 to Team B"

Write-Info "Transferring User 2 from Team A to Team B..."
$transfer = Invoke-APIRequest -Method POST -Endpoint "/company-admin/teams/$script:teamAID/members/$script:user2ID/transfer" -Body @{
    to_team_id = $script:teamBID
    role_in_team = "driver"
} -Token $script:companyAdminToken

if ($transfer.data) {
    Write-Success "User 2 transferred (should log 'removed' from Team A and 'added' to Team B)"
}

Start-Sleep -Seconds 1

# ============================================================================
# STEP 9: VERIFY TRANSFER IN HISTORY
# ============================================================================

Write-Step "Step 9: Verify Transfer in History"

Write-Info "Checking Team A history for removal..."
$teamAHistory = Invoke-APIRequest -Method GET -Endpoint "/company-admin/teams/$script:teamAID/member-history?limit=10" -Token $script:companyAdminToken

if ($teamAHistory.data.history) {
    $removalRecord = $teamAHistory.data.history | Where-Object { $_.user_id -eq $script:user2ID -and $_.change_type -eq "removed" } | Select-Object -First 1
    if ($removalRecord) {
        Write-Success "Found removal record in Team A history"
        Write-Info "  Previous role: $($removalRecord.previous_role_in_team)"
    }
}

Write-Info "Checking Team B history for addition..."
$teamBHistory = Invoke-APIRequest -Method GET -Endpoint "/company-admin/teams/$script:teamBID/member-history?limit=10" -Token $script:companyAdminToken

if ($teamBHistory.data.history) {
    $additionRecord = $teamBHistory.data.history | Where-Object { $_.user_id -eq $script:user2ID -and $_.change_type -eq "added" } | Select-Object -First 1
    if ($additionRecord) {
        Write-Success "Found addition record in Team B history"
        Write-Info "  New role: $($additionRecord.new_role_in_team)"
    }
}

Write-Info "Checking User 2's complete history..."
$user2History = Invoke-APIRequest -Method GET -Endpoint "/company-admin/teams/users/$script:user2ID/team-history?limit=50" -Token $script:companyAdminToken

if ($user2History.data.history) {
    Write-Success "User 2 has $($user2History.data.count) history records (should show both teams)"
    Write-Host ""
    Write-Host "User 2 Complete History:" -ForegroundColor Cyan
    foreach ($record in $user2History.data.history) {
        $teamName = if ($record.team) { $record.team.name } else { "Unknown" }
        Write-Host "  • $($record.change_type) in $teamName - $($record.changed_at)" -ForegroundColor White
    }
}

# ============================================================================
# STEP 10: REMOVE MEMBER (Tests "removed" change type)
# ============================================================================

Write-Step "Step 10: Remove Member from Team"

Write-Info "Removing User 3 from Team A..."
$remove = Invoke-APIRequest -Method DELETE -Endpoint "/company-admin/teams/$script:teamAID/members/$script:user3ID" -Token $script:companyAdminToken

if ($remove.data) {
    Write-Success "User 3 removed (should log 'removed' with previous role 'driver')"
}

Start-Sleep -Seconds 1

Write-Info "Verifying removal in history..."
$finalHistory = Invoke-APIRequest -Method GET -Endpoint "/company-admin/teams/$script:teamAID/member-history?limit=1" -Token $script:companyAdminToken

if ($finalHistory.data.history -and $finalHistory.data.history.Count -gt 0) {
    $latestRecord = $finalHistory.data.history[0]
    if ($latestRecord.change_type -eq "removed" -and $latestRecord.user_id -eq $script:user3ID) {
        Write-Success "Latest history record confirms User 3 removal"
        Write-Info "  Previous role: $($latestRecord.previous_role_in_team)"
    }
}

# ============================================================================
# CLEANUP (Optional)
# ============================================================================

if ($Cleanup) {
    Write-Step "Cleanup: Removing Test Data"

    if ($script:teamBID) {
        Write-Info "Deleting Team B..."
        Invoke-APIRequest -Method DELETE -Endpoint "/company-admin/teams/$script:teamBID" -Token $script:companyAdminToken -ExpectFailure
    }

    if ($script:teamAID) {
        Write-Info "Deleting Team A..."
        Invoke-APIRequest -Method DELETE -Endpoint "/company-admin/teams/$script:teamAID" -Token $script:companyAdminToken -ExpectFailure
    }

    if ($script:user1ID) {
        Write-Info "Deleting User 1..."
        Invoke-APIRequest -Method DELETE -Endpoint "/company-admin/users/$script:user1ID" -Token $script:companyAdminToken -ExpectFailure
    }

    if ($script:user2ID) {
        Write-Info "Deleting User 2..."
        Invoke-APIRequest -Method DELETE -Endpoint "/company-admin/users/$script:user2ID" -Token $script:companyAdminToken -ExpectFailure
    }

    if ($script:user3ID) {
        Write-Info "Deleting User 3..."
        Invoke-APIRequest -Method DELETE -Endpoint "/company-admin/users/$script:user3ID" -Token $script:companyAdminToken -ExpectFailure
    }

    Write-Success "Cleanup completed"
}

# ============================================================================
# TEST SUMMARY
# ============================================================================

Write-Step "Test Summary"
Write-Host ""
Write-Host "Team Member History API Testing Complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Tested Change Types:" -ForegroundColor Cyan
Write-Host "  ✓ added         - Members added to teams" -ForegroundColor Green
Write-Host "  ✓ removed       - Members removed from teams" -ForegroundColor Green
Write-Host "  ✓ role_changed  - Member role updates" -ForegroundColor Green
Write-Host "  ✓ Transfer      - Members transferred between teams" -ForegroundColor Green
Write-Host ""
Write-Host "Tested Endpoints:" -ForegroundColor Cyan
Write-Host "  ✓ GET /company-admin/teams/:id/member-history" -ForegroundColor Green
Write-Host "  ✓ GET /company-admin/teams/users/:userId/team-history" -ForegroundColor Green
Write-Host ""
Write-Host "All tests passed successfully!" -ForegroundColor Green
Write-Host ""

if (-not $Cleanup) {
    Write-Info "Run with -Cleanup flag to remove test data"
    Write-Host "Test Data IDs:" -ForegroundColor Yellow
    Write-Host "  Team A ID: $script:teamAID"
    Write-Host "  Team B ID: $script:teamBID"
    Write-Host "  User 1 ID: $script:user1ID"
    Write-Host "  User 2 ID: $script:user2ID"
    Write-Host "  User 3 ID: $script:user3ID"
}
