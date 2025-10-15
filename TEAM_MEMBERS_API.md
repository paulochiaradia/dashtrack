# Team Members Management API

Complete API documentation for managing team members in the DashTrack system.

## Table of Contents
- [Overview](#overview)
- [Endpoints](#endpoints)
- [Data Models](#data-models)
- [Examples](#examples)
- [Testing](#testing)

## Overview

The Team Members Management API provides complete CRUD operations for managing team memberships. This includes:
- Adding users to teams
- Listing team members with user details
- Updating member roles
- Removing members from teams
- Transferring members between teams

### Available Roles
- `manager` - Team leader with management responsibilities
- `driver` - Vehicle operator
- `assistant` - Support role
- `supervisor` - Overseer role

### Authentication
All endpoints require authentication and company admin role.

**Headers:**
```http
Authorization: Bearer <token>
Content-Type: application/json
```

## Endpoints

### 1. Add Member to Team

Add a user to a team with a specific role.

**Endpoint:** `POST /api/v1/company-admin/teams/:id/members`

**Path Parameters:**
- `id` (UUID) - Team ID

**Request Body:**
```json
{
  "user_id": "uuid",
  "role_in_team": "manager|driver|assistant|supervisor"
}
```

**Success Response:** `201 Created`
```json
{
  "success": true,
  "message": "Team member added successfully",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "team_id": "550e8400-e29b-41d4-a716-446655440001",
    "user_id": "550e8400-e29b-41d4-a716-446655440002",
    "role_in_team": "driver",
    "joined_at": "2024-01-15T10:30:00Z"
  }
}
```

**Error Responses:**
- `400 Bad Request` - Invalid parameters or user already a member
- `404 Not Found` - Team or user not found
- `409 Conflict` - User already member of team

**Business Rules:**
- User must belong to same company as team
- User cannot be added twice to same team
- Role must be one of the valid options

---

### 2. Get Team Members

Retrieve all members of a team with populated user details.

**Endpoint:** `GET /api/v1/company-admin/teams/:id/members`

**Path Parameters:**
- `id` (UUID) - Team ID

**Success Response:** `200 OK`
```json
{
  "success": true,
  "message": "Team members retrieved successfully",
  "data": {
    "team": {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "name": "Logistics Team",
      "description": "Main logistics team",
      "status": "active"
    },
    "members": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "team_id": "550e8400-e29b-41d4-a716-446655440001",
        "user_id": "550e8400-e29b-41d4-a716-446655440002",
        "role_in_team": "manager",
        "joined_at": "2024-01-15T10:30:00Z",
        "user": {
          "id": "550e8400-e29b-41d4-a716-446655440002",
          "name": "John Manager",
          "email": "john@company.com",
          "role": "user",
          "status": "active"
        }
      },
      {
        "id": "550e8400-e29b-41d4-a716-446655440003",
        "team_id": "550e8400-e29b-41d4-a716-446655440001",
        "user_id": "550e8400-e29b-41d4-a716-446655440004",
        "role_in_team": "driver",
        "joined_at": "2024-01-15T11:00:00Z",
        "user": {
          "id": "550e8400-e29b-41d4-a716-446655440004",
          "name": "Jane Driver",
          "email": "jane@company.com",
          "role": "user",
          "status": "active"
        }
      }
    ],
    "count": 2
  }
}
```

**Error Responses:**
- `400 Bad Request` - Invalid team ID
- `404 Not Found` - Team not found

---

### 3. Update Member Role

Update a team member's role within the team.

**Endpoint:** `PUT /api/v1/company-admin/teams/:id/members/:userId/role`

**Path Parameters:**
- `id` (UUID) - Team ID
- `userId` (UUID) - User ID

**Request Body:**
```json
{
  "role_in_team": "manager|driver|assistant|supervisor"
}
```

**Success Response:** `200 OK`
```json
{
  "success": true,
  "message": "Member role updated successfully",
  "data": {
    "team_id": "550e8400-e29b-41d4-a716-446655440001",
    "user_id": "550e8400-e29b-41d4-a716-446655440002",
    "role_in_team": "supervisor"
  }
}
```

**Error Responses:**
- `400 Bad Request` - Invalid parameters or role
- `404 Not Found` - Team not found or user not a member

**Business Rules:**
- User must already be a member of the team
- New role must be valid

---

### 4. Remove Member from Team

Remove a user from a team.

**Endpoint:** `DELETE /api/v1/company-admin/teams/:id/members/:userId`

**Path Parameters:**
- `id` (UUID) - Team ID
- `userId` (UUID) - User ID

**Success Response:** `200 OK`
```json
{
  "success": true,
  "message": "Team member removed successfully",
  "data": {
    "team_id": "550e8400-e29b-41d4-a716-446655440001",
    "user_id": "550e8400-e29b-41d4-a716-446655440002"
  }
}
```

**Error Responses:**
- `400 Bad Request` - Invalid parameters
- `404 Not Found` - Team not found or user not a member

**Business Rules:**
- User must be a member of the team

---

### 5. Transfer Member to Another Team

Transfer a user from one team to another atomically.

**Endpoint:** `POST /api/v1/company-admin/teams/:id/members/:userId/transfer`

**Path Parameters:**
- `id` (UUID) - Destination Team ID
- `userId` (UUID) - User ID to transfer

**Request Body:**
```json
{
  "from_team_id": "uuid",
  "role_in_team": "manager|driver|assistant|supervisor"
}
```

**Success Response:** `200 OK`
```json
{
  "success": true,
  "message": "Team member transferred successfully",
  "data": {
    "from_team_id": "550e8400-e29b-41d4-a716-446655440001",
    "to_team_id": "550e8400-e29b-41d4-a716-446655440005",
    "user_id": "550e8400-e29b-41d4-a716-446655440002",
    "role": "driver"
  }
}
```

**Error Responses:**
- `400 Bad Request` - Invalid parameters, user not in source team, or already in destination team
- `404 Not Found` - Source or destination team not found

**Business Rules:**
- Both teams must exist and belong to same company
- User must be member of source team
- User cannot already be member of destination team
- If transfer fails, attempts to rollback by re-adding to source team

---

## Data Models

### TeamMember
```go
type TeamMember struct {
    ID         uuid.UUID `json:"id"`
    TeamID     uuid.UUID `json:"team_id"`
    UserID     uuid.UUID `json:"user_id"`
    RoleInTeam string    `json:"role_in_team"`
    JoinedAt   time.Time `json:"joined_at"`
    User       *User     `json:"user,omitempty"`
    Team       *Team     `json:"team,omitempty"`
}
```

### Request Models

**AssignTeamMemberRequest**
```go
type AssignTeamMemberRequest struct {
    UserID     uuid.UUID `json:"user_id" binding:"required"`
    RoleInTeam string    `json:"role_in_team" binding:"required,oneof=manager driver assistant supervisor"`
}
```

**TransferTeamMemberRequest**
```go
type TransferTeamMemberRequest struct {
    FromTeamID uuid.UUID `json:"from_team_id" binding:"required"`
    RoleInTeam string    `json:"role_in_team" binding:"required,oneof=manager driver assistant supervisor"`
}
```

## Examples

### PowerShell Example

```powershell
# Setup
$baseUrl = "http://localhost:8080/api/v1"
$token = "your-token-here"
$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

# Add member to team
$memberData = @{
    user_id = "550e8400-e29b-41d4-a716-446655440002"
    role_in_team = "driver"
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$teamId/members" `
    -Method Post `
    -Headers $headers `
    -Body $memberData

# List team members
$members = Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$teamId/members" `
    -Method Get `
    -Headers $headers

# Update role
$updateData = @{
    role_in_team = "supervisor"
} | ConvertTo-Json

Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$teamId/members/$userId/role" `
    -Method Put `
    -Headers $headers `
    -Body $updateData

# Transfer member
$transferData = @{
    from_team_id = "550e8400-e29b-41d4-a716-446655440001"
    role_in_team = "driver"
} | ConvertTo-Json

Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$newTeamId/members/$userId/transfer" `
    -Method Post `
    -Headers $headers `
    -Body $transferData

# Remove member
Invoke-RestMethod -Uri "$baseUrl/company-admin/teams/$teamId/members/$userId" `
    -Method Delete `
    -Headers $headers
```

### cURL Example

```bash
# Add member
curl -X POST http://localhost:8080/api/v1/company-admin/teams/{team_id}/members \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "550e8400-e29b-41d4-a716-446655440002",
    "role_in_team": "driver"
  }'

# List members
curl -X GET http://localhost:8080/api/v1/company-admin/teams/{team_id}/members \
  -H "Authorization: Bearer {token}"

# Update role
curl -X PUT http://localhost:8080/api/v1/company-admin/teams/{team_id}/members/{user_id}/role \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "role_in_team": "supervisor"
  }'

# Transfer member
curl -X POST http://localhost:8080/api/v1/company-admin/teams/{new_team_id}/members/{user_id}/transfer \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "from_team_id": "550e8400-e29b-41d4-a716-446655440001",
    "role_in_team": "driver"
  }'

# Remove member
curl -X DELETE http://localhost:8080/api/v1/company-admin/teams/{team_id}/members/{user_id} \
  -H "Authorization: Bearer {token}"
```

## Testing

### Using the Test Script

A comprehensive PowerShell test script is available at `scripts/test-team-members-api.ps1`.

**Usage:**
```powershell
cd scripts
.\test-team-members-api.ps1
```

**The script tests:**
1. ✓ Team creation
2. ✓ User creation with all roles
3. ✓ Adding members to team
4. ✓ Retrieving team members
5. ✓ Updating member roles
6. ✓ Transferring members between teams
7. ✓ Removing members from team
8. ✓ Final state verification
9. ✓ Optional cleanup

### Manual Testing Checklist

- [ ] Add member with valid role
- [ ] Try to add same member twice (should fail)
- [ ] Add member from different company (should fail)
- [ ] List members and verify user details populated
- [ ] Update member role to all valid roles
- [ ] Update role for non-member (should fail)
- [ ] Transfer member between teams
- [ ] Transfer non-member (should fail)
- [ ] Transfer to team where already member (should fail)
- [ ] Remove member from team
- [ ] Remove non-member (should fail)

## Architecture

### Repository Layer
Located in `internal/repository/team.go`:
- `AddMember(ctx, teamMember)` - Insert team member
- `RemoveMember(ctx, teamID, userID)` - Delete team member
- `GetMembers(ctx, teamID)` - Get all members with user JOIN
- `UpdateMemberRole(ctx, teamID, userID, role)` - Update role
- `CheckMemberExists(ctx, teamID, userID)` - Check membership
- `GetTeamsByUser(ctx, userID)` - Get user's teams

### Handler Layer
Located in `internal/handlers/team.go`:
- `AddMember` - HTTP handler for adding members
- `GetMembers` - HTTP handler for listing members
- `UpdateMemberRole` - HTTP handler for role updates
- `RemoveMember` - HTTP handler for removal
- `TransferMemberToTeam` - HTTP handler for transfers (NEW)

### Routes Layer
Located in `internal/routes/team.go`:
- Configures all team member management routes
- Enforces company_admin role requirement
- Provides RESTful URL structure

## Features

### Implemented
✅ Add member with role validation  
✅ List members with user details (JOIN)  
✅ Update member role  
✅ Remove member  
✅ Transfer member between teams (NEW)  
✅ Duplicate member prevention  
✅ Company boundary enforcement  
✅ OpenTelemetry tracing  
✅ Comprehensive error handling  
✅ Rollback on transfer failure  

### Next Steps (Future Tasks)
⏳ Team member history tracking (Task 3)  
⏳ Vehicle assignment history (Task 2)  
⏳ Audit logging for all operations  
⏳ Bulk member operations  
⏳ Member invitation system  

## Database Schema

```sql
CREATE TABLE team_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_in_team VARCHAR(50) NOT NULL CHECK (role_in_team IN ('manager', 'driver', 'assistant', 'supervisor')),
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(team_id, user_id)
);

CREATE INDEX idx_team_members_team ON team_members(team_id);
CREATE INDEX idx_team_members_user ON team_members(user_id);
CREATE INDEX idx_team_members_role ON team_members(role_in_team);
```

## Notes

- All operations are scoped to company context
- User details are populated via JOIN for efficiency
- Transfer operation is atomic (rollback on failure)
- OpenTelemetry spans track all operations
- Comprehensive validation at all levels
