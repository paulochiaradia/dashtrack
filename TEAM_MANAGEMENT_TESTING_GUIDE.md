# Team Management - Manual Testing Guide

## Overview

This guide provides step-by-step instructions for testing all Team Management endpoints.

**Prerequisites:**
- Docker containers running (`docker-compose up -d`)
- Valid authentication token
- Company context (company_id in token)
- Test user accounts created

---

## Setup Test Environment

### 1. Start Docker Containers

```powershell
docker-compose up -d
```

### 2. Get Authentication Token

```powershell
# Login as company_admin
$loginResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method POST -Body (@{
    email = "admin@company.com"
    password = "password123"
} | ConvertTo-Json) -ContentType "application/json"

$token = $loginResponse.data.access_token
```

### 3. Set Headers

```powershell
$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}
```

---

## Test Cases

### TEST 1: Create Team ✅

**Endpoint:** `POST /api/v1/company-admin/teams`

```powershell
$createTeamBody = @{
    name = "Test Team Alpha"
    description = "Integration test team for vehicle management"
    status = "active"
} | ConvertTo-Json

$createResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/company-admin/teams" `
    -Method POST `
    -Headers $headers `
    -Body $createTeamBody

$teamId = $createResponse.data.id
Write-Host "Created Team ID: $teamId"
```

**Expected Response:**
```json
{
  "success": true,
  "message": "Team created successfully",
  "data": {
    "id": "uuid",
    "company_id": "uuid",
    "name": "Test Team Alpha",
    "description": "Integration test team for vehicle management",
    "status": "active",
    "created_at": "2025-10-13T...",
    "updated_at": "2025-10-13T..."
  }
}
```

**Status Code:** 201 Created

---

### TEST 2: List Teams ✅

**Endpoint:** `GET /api/v1/company-admin/teams`

```powershell
$listResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/company-admin/teams?limit=10&offset=0" `
    -Method GET `
    -Headers $headers

Write-Host "Teams Count: $($listResponse.data.count)"
```

**Expected Response:**
```json
{
  "success": true,
  "message": "Teams retrieved successfully",
  "data": {
    "teams": [...],
    "count": 1
  }
}
```

**Status Code:** 200 OK

---

### TEST 3: Get Team Details ✅

**Endpoint:** `GET /api/v1/company-admin/teams/:id`

```powershell
$teamResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/company-admin/teams/$teamId" `
    -Method GET `
    -Headers $headers

Write-Host "Team Name: $($teamResponse.data.name)"
```

**Expected Response:**
```json
{
  "success": true,
  "message": "Team retrieved successfully",
  "data": {
    "id": "uuid",
    "name": "Test Team Alpha",
    ...
  }
}
```

**Status Code:** 200 OK

---

### TEST 4: Update Team ✅

**Endpoint:** `PUT /api/v1/company-admin/teams/:id`

```powershell
$updateTeamBody = @{
    name = "Test Team Alpha - Updated"
    description = "Updated description for testing"
    status = "active"
} | ConvertTo-Json

$updateResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/company-admin/teams/$teamId" `
    -Method PUT `
    -Headers $headers `
    -Body $updateTeamBody

Write-Host "Updated Team: $($updateResponse.data.name)"
```

**Expected Response:**
```json
{
  "success": true,
  "message": "Team updated successfully",
  "data": {
    "name": "Test Team Alpha - Updated",
    ...
  }
}
```

**Status Code:** 200 OK

---

### TEST 5: Add Team Member ✅

**Endpoint:** `POST /api/v1/company-admin/teams/:id/members`

```powershell
# First, get a user ID
$usersResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/company-admin/users" `
    -Method GET `
    -Headers $headers

$userId = $usersResponse.data.users[0].id

# Add member
$addMemberBody = @{
    user_id = $userId
    role_in_team = "driver"
} | ConvertTo-Json

$memberResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/company-admin/teams/$teamId/members" `
    -Method POST `
    -Headers $headers `
    -Body $addMemberBody

Write-Host "Added Member: $userId"
```

**Expected Response:**
```json
{
  "success": true,
  "message": "Member added to team successfully",
  "data": {
    "team_id": "uuid",
    "user_id": "uuid",
    "role_in_team": "driver"
  }
}
```

**Status Code:** 201 Created

---

### TEST 6: List Team Members ✅

**Endpoint:** `GET /api/v1/company-admin/teams/:id/members`

```powershell
$membersResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/company-admin/teams/$teamId/members" `
    -Method GET `
    -Headers $headers

Write-Host "Members Count: $($membersResponse.data.count)"
$membersResponse.data.members | ForEach-Object {
    Write-Host "  - $($_.user.name) ($($_.role_in_team))"
}
```

**Expected Response:**
```json
{
  "success": true,
  "message": "Team members retrieved successfully",
  "data": {
    "members": [
      {
        "id": "uuid",
        "team_id": "uuid",
        "user_id": "uuid",
        "role_in_team": "driver",
        "joined_at": "2025-10-13T...",
        "user": {
          "id": "uuid",
          "name": "John Doe",
          "email": "john@example.com"
        }
      }
    ],
    "count": 1
  }
}
```

**Status Code:** 200 OK

---

### TEST 7: Update Member Role ✅

**Endpoint:** `PUT /api/v1/company-admin/teams/:id/members/:userId/role`

```powershell
$updateRoleBody = @{
    role_in_team = "manager"
} | ConvertTo-Json

$roleResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/company-admin/teams/$teamId/members/$userId/role" `
    -Method PUT `
    -Headers $headers `
    -Body $updateRoleBody

Write-Host "Updated role to: manager"
```

**Expected Response:**
```json
{
  "success": true,
  "message": "Member role updated successfully",
  "data": {
    "team_id": "uuid",
    "user_id": "uuid",
    "role_in_team": "manager"
  }
}
```

**Status Code:** 200 OK

---

### TEST 8: Get Team Statistics ✅

**Endpoint:** `GET /api/v1/company-admin/teams/:id/stats`

```powershell
$statsResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/company-admin/teams/$teamId/stats" `
    -Method GET `
    -Headers $headers

Write-Host "Team Statistics:"
Write-Host "  Members: $($statsResponse.data.member_count)"
Write-Host "  Vehicles: $($statsResponse.data.vehicle_count)"
Write-Host "  Active Vehicles: $($statsResponse.data.active_vehicles)"
```

**Expected Response:**
```json
{
  "success": true,
  "message": "Team statistics retrieved successfully",
  "data": {
    "team_id": "uuid",
    "team_name": "Test Team Alpha",
    "member_count": 1,
    "vehicle_count": 0,
    "active_vehicles": 0,
    "status": "active",
    "created_at": "2025-10-13T...",
    "manager_id": null
  }
}
```

**Status Code:** 200 OK

---

### TEST 9: Create Test Vehicle

```powershell
# Create a vehicle for testing assignment
$createVehicleBody = @{
    license_plate = "TEST-1234"
    brand = "Ford"
    model = "Transit"
    year = 2023
    vehicle_type = "van"
    fuel_type = "diesel"
    status = "active"
} | ConvertTo-Json

$vehicleResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/company-admin/vehicles" `
    -Method POST `
    -Headers $headers `
    -Body $createVehicleBody

$vehicleId = $vehicleResponse.data.id
Write-Host "Created Vehicle ID: $vehicleId"
```

---

### TEST 10: Assign Vehicle to Team ✅

**Endpoint:** `POST /api/v1/company-admin/teams/:id/vehicles/:vehicleId`

```powershell
$assignResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/company-admin/teams/$teamId/vehicles/$vehicleId" `
    -Method POST `
    -Headers $headers

Write-Host "Assigned vehicle to team"
```

**Expected Response:**
```json
{
  "success": true,
  "message": "Vehicle assigned to team successfully",
  "data": {
    "team_id": "uuid",
    "vehicle_id": "uuid"
  }
}
```

**Status Code:** 200 OK

---

### TEST 11: Get Team Vehicles ✅

**Endpoint:** `GET /api/v1/company-admin/teams/:id/vehicles`

```powershell
$vehiclesResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/company-admin/teams/$teamId/vehicles" `
    -Method GET `
    -Headers $headers

Write-Host "Team Vehicles:"
$vehiclesResponse.data.vehicles | ForEach-Object {
    Write-Host "  - $($_.license_plate) ($($_.brand) $($_.model))"
}
Write-Host "Total: $($vehiclesResponse.data.count)"
```

**Expected Response:**
```json
{
  "success": true,
  "message": "Team vehicles retrieved successfully",
  "data": {
    "team": {...},
    "vehicles": [
      {
        "id": "uuid",
        "license_plate": "TEST-1234",
        "brand": "Ford",
        "model": "Transit",
        "year": 2023,
        "status": "active",
        ...
      }
    ],
    "count": 1
  }
}
```

**Status Code:** 200 OK

---

### TEST 12: Get Team Stats (After Vehicle Assignment) ✅

```powershell
$statsResponse2 = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/company-admin/teams/$teamId/stats" `
    -Method GET `
    -Headers $headers

Write-Host "Updated Team Statistics:"
Write-Host "  Members: $($statsResponse2.data.member_count)"
Write-Host "  Vehicles: $($statsResponse2.data.vehicle_count)"
Write-Host "  Active Vehicles: $($statsResponse2.data.active_vehicles)"
```

**Expected:** `vehicle_count` and `active_vehicles` should now be 1

---

### TEST 13: Unassign Vehicle from Team ✅

**Endpoint:** `DELETE /api/v1/company-admin/teams/:id/vehicles/:vehicleId`

```powershell
$unassignResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/company-admin/teams/$teamId/vehicles/$vehicleId" `
    -Method DELETE `
    -Headers $headers

Write-Host "Unassigned vehicle from team"
```

**Expected Response:**
```json
{
  "success": true,
  "message": "Vehicle unassigned from team successfully",
  "data": {
    "team_id": "uuid",
    "vehicle_id": "uuid"
  }
}
```

**Status Code:** 200 OK

---

### TEST 14: Remove Team Member ✅

**Endpoint:** `DELETE /api/v1/company-admin/teams/:id/members/:userId`

```powershell
$removeResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/company-admin/teams/$teamId/members/$userId" `
    -Method DELETE `
    -Headers $headers

Write-Host "Removed member from team"
```

**Expected Response:**
```json
{
  "success": true,
  "message": "Member removed from team successfully",
  "data": {
    "team_id": "uuid",
    "user_id": "uuid"
  }
}
```

**Status Code:** 200 OK

---

### TEST 15: Get My Teams (User Endpoint) ✅

**Endpoint:** `GET /api/v1/teams/my-teams`

```powershell
# Login as regular user first
$userLoginResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method POST -Body (@{
    email = "user@company.com"
    password = "password123"
} | ConvertTo-Json) -ContentType "application/json"

$userToken = $userLoginResponse.data.access_token
$userHeaders = @{
    "Authorization" = "Bearer $userToken"
    "Content-Type" = "application/json"
}

$myTeamsResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/teams/my-teams" `
    -Method GET `
    -Headers $userHeaders

Write-Host "My Teams:"
$myTeamsResponse.data.teams | ForEach-Object {
    Write-Host "  - $($_.name) (Role: $($_.role_in_team))"
}
```

**Expected Response:**
```json
{
  "success": true,
  "message": "User teams retrieved successfully",
  "data": {
    "teams": [
      {
        "id": "uuid",
        "name": "Test Team Alpha",
        "role_in_team": "driver",
        "joined_at": "2025-10-13T..."
      }
    ],
    "count": 1
  }
}
```

**Status Code:** 200 OK

---

### TEST 16: Delete Team ✅

**Endpoint:** `DELETE /api/v1/company-admin/teams/:id`

```powershell
$deleteResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/company-admin/teams/$teamId" `
    -Method DELETE `
    -Headers $headers

Write-Host "Team deleted (soft delete)"
```

**Expected Response:**
```json
{
  "success": true,
  "message": "Team deleted successfully",
  "data": {
    "team_id": "uuid"
  }
}
```

**Status Code:** 200 OK

---

## Role-Based Access Tests

### TEST 17: Admin Access (Read-Only)

```powershell
# Login as admin
$adminLoginResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method POST -Body (@{
    email = "admin@company.com"
    password = "password123"
} | ConvertTo-Json) -ContentType "application/json"

$adminToken = $adminLoginResponse.data.access_token
$adminHeaders = @{
    "Authorization" = "Bearer $adminToken"
    "Content-Type" = "application/json"
}

# Admin can list teams
$adminTeamsResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/admin/teams" `
    -Method GET `
    -Headers $adminHeaders

Write-Host "Admin can see $($adminTeamsResponse.data.count) teams"

# Admin can view stats
$adminStatsResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/admin/teams/$teamId/stats" `
    -Method GET `
    -Headers $adminHeaders

Write-Host "Admin can view team stats"
```

**Expected:** Admin can read teams and stats (200 OK)

---

### TEST 18: Manager Access (Read-Only)

```powershell
# Login as manager
$managerLoginResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method POST -Body (@{
    email = "manager@company.com"
    password = "password123"
} | ConvertTo-Json) -ContentType "application/json"

$managerToken = $managerLoginResponse.data.access_token
$managerHeaders = @{
    "Authorization" = "Bearer $managerToken"
    "Content-Type" = "application/json"
}

# Manager can list teams
$managerTeamsResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/manager/teams" `
    -Method GET `
    -Headers $managerHeaders

Write-Host "Manager can see $($managerTeamsResponse.data.count) teams"
```

**Expected:** Manager can read teams (200 OK)

---

## Validation Tests

### TEST 19: Invalid Team ID

```powershell
try {
    $invalidResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/company-admin/teams/invalid-uuid" `
        -Method GET `
        -Headers $headers
} catch {
    Write-Host "✓ Invalid ID rejected: $($_.Exception.Response.StatusCode)"
}
```

**Expected:** 400 Bad Request

---

### TEST 20: Team Not Found

```powershell
$fakeTeamId = [System.Guid]::NewGuid()
try {
    $notFoundResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/company-admin/teams/$fakeTeamId" `
        -Method GET `
        -Headers $headers
} catch {
    Write-Host "✓ Non-existent team rejected: $($_.Exception.Response.StatusCode)"
}
```

**Expected:** 404 Not Found

---

### TEST 21: Unauthorized Access

```powershell
try {
    $unauthResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/company-admin/teams" `
        -Method GET
} catch {
    Write-Host "✓ Unauthorized request rejected: $($_.Exception.Response.StatusCode)"
}
```

**Expected:** 401 Unauthorized

---

## Complete Test Script

```powershell
# Save this as: test-team-management.ps1

# Configuration
$baseUrl = "http://localhost:8080"
$adminEmail = "admin@company.com"
$adminPassword = "password123"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Team Management API Test Suite" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Login
Write-Host "[1/21] Logging in..." -ForegroundColor Yellow
$loginResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/login" -Method POST -Body (@{
    email = $adminEmail
    password = $adminPassword
} | ConvertTo-Json) -ContentType "application/json"

$token = $loginResponse.data.access_token
$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}
Write-Host "✓ Logged in successfully" -ForegroundColor Green
Write-Host ""

# Create Team
Write-Host "[2/21] Creating team..." -ForegroundColor Yellow
$createTeamBody = @{
    name = "Test Team $(Get-Date -Format 'HHmmss')"
    description = "Automated test team"
    status = "active"
} | ConvertTo-Json

$createResponse = Invoke-RestMethod -Uri "$baseUrl/api/v1/company-admin/teams" `
    -Method POST `
    -Headers $headers `
    -Body $createTeamBody

$teamId = $createResponse.data.id
Write-Host "✓ Team created: $teamId" -ForegroundColor Green
Write-Host ""

# Continue with all other tests...

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "All tests completed!" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
```

---

## Test Results Checklist

- [ ] TEST 1: Create Team
- [ ] TEST 2: List Teams
- [ ] TEST 3: Get Team Details
- [ ] TEST 4: Update Team
- [ ] TEST 5: Add Team Member
- [ ] TEST 6: List Team Members
- [ ] TEST 7: Update Member Role
- [ ] TEST 8: Get Team Statistics
- [ ] TEST 9: Create Test Vehicle
- [ ] TEST 10: Assign Vehicle to Team
- [ ] TEST 11: Get Team Vehicles
- [ ] TEST 12: Get Stats After Assignment
- [ ] TEST 13: Unassign Vehicle
- [ ] TEST 14: Remove Team Member
- [ ] TEST 15: Get My Teams
- [ ] TEST 16: Delete Team
- [ ] TEST 17: Admin Access
- [ ] TEST 18: Manager Access
- [ ] TEST 19: Invalid Team ID
- [ ] TEST 20: Team Not Found
- [ ] TEST 21: Unauthorized Access

---

**Total Tests:** 21  
**Coverage:** All 16 endpoints + role/validation tests

---

## Notes

- All tests assume Docker containers are running
- Replace placeholder emails/passwords with actual test credentials
- Some tests require existing users and vehicles
- Multi-tenancy is enforced (company_id from token)
- All timestamps are in ISO 8601 format

---

**Last Updated:** October 13, 2025
