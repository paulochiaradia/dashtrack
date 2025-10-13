# Team Management API Reference

## Overview

Complete API documentation for Team Management endpoints.

**Base URL:** `/api/v1`  
**Authentication:** Required (JWT Bearer Token)  
**Multi-tenancy:** Company context required

---

## Endpoints Summary

| Method | Endpoint | Role Required | Description |
|--------|----------|---------------|-------------|
| **Company Admin Routes** |
| GET | `/company-admin/teams` | company_admin | List all teams |
| POST | `/company-admin/teams` | company_admin | Create new team |
| GET | `/company-admin/teams/:id` | company_admin | Get team details |
| PUT | `/company-admin/teams/:id` | company_admin | Update team |
| DELETE | `/company-admin/teams/:id` | company_admin | Delete team |
| GET | `/company-admin/teams/:id/members` | company_admin | List team members |
| POST | `/company-admin/teams/:id/members` | company_admin | Add member to team |
| DELETE | `/company-admin/teams/:id/members/:userId` | company_admin | Remove member |
| PUT | `/company-admin/teams/:id/members/:userId/role` | company_admin | Update member role |
| GET | `/company-admin/teams/:id/stats` | company_admin | Get team statistics |
| GET | `/company-admin/teams/:id/vehicles` | company_admin | List team vehicles |
| POST | `/company-admin/teams/:id/vehicles/:vehicleId` | company_admin | Assign vehicle to team |
| DELETE | `/company-admin/teams/:id/vehicles/:vehicleId` | company_admin | Unassign vehicle |
| **Admin Routes** |
| GET | `/admin/teams` | admin | List teams |
| GET | `/admin/teams/:id` | admin | Get team details |
| GET | `/admin/teams/:id/members` | admin | List team members |
| GET | `/admin/teams/:id/stats` | admin | Get team statistics |
| **Manager Routes** |
| GET | `/manager/teams` | manager | List teams |
| GET | `/manager/teams/:id` | manager | Get team details |
| GET | `/manager/teams/:id/members` | manager | List team members |
| **User Routes** |
| GET | `/teams/my-teams` | user | Get current user's teams |

**Total:** 21 endpoints

---

## 1. Create Team

Create a new team within the company.

**Endpoint:** `POST /api/v1/company-admin/teams`  
**Role:** `company_admin`

### Request Body
```json
{
  "name": "Team Alpha",
  "description": "Sales and logistics team",
  "manager_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "active"
}
```

### Response (201 Created)
```json
{
  "success": true,
  "message": "Team created successfully",
  "data": {
    "id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
    "company_id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "Team Alpha",
    "description": "Sales and logistics team",
    "manager_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "active",
    "created_at": "2025-10-13T10:00:00Z",
    "updated_at": "2025-10-13T10:00:00Z"
  }
}
```

---

## 2. List Teams

Get all teams for the company with pagination.

**Endpoint:** `GET /api/v1/company-admin/teams`  
**Role:** `company_admin`, `admin`, `manager`

### Query Parameters
- `limit` (int, optional): Number of teams per page (default: 10)
- `offset` (int, optional): Offset for pagination (default: 0)

### Response (200 OK)
```json
{
  "success": true,
  "message": "Teams retrieved successfully",
  "data": {
    "teams": [
      {
        "id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
        "company_id": "123e4567-e89b-12d3-a456-426614174000",
        "name": "Team Alpha",
        "description": "Sales and logistics team",
        "manager_id": "550e8400-e29b-41d4-a716-446655440000",
        "status": "active",
        "created_at": "2025-10-13T10:00:00Z",
        "updated_at": "2025-10-13T10:00:00Z"
      }
    ],
    "count": 1
  }
}
```

---

## 3. Get Team Details

Get detailed information about a specific team.

**Endpoint:** `GET /api/v1/company-admin/teams/:id`  
**Role:** `company_admin`, `admin`, `manager`

### Response (200 OK)
```json
{
  "success": true,
  "message": "Team retrieved successfully",
  "data": {
    "id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
    "company_id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "Team Alpha",
    "description": "Sales and logistics team",
    "manager_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "active",
    "created_at": "2025-10-13T10:00:00Z",
    "updated_at": "2025-10-13T10:00:00Z"
  }
}
```

---

## 4. Update Team

Update team information.

**Endpoint:** `PUT /api/v1/company-admin/teams/:id`  
**Role:** `company_admin`

### Request Body
```json
{
  "name": "Team Alpha Updated",
  "description": "Updated description",
  "manager_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "active"
}
```

### Response (200 OK)
```json
{
  "success": true,
  "message": "Team updated successfully",
  "data": {
    "id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
    "company_id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "Team Alpha Updated",
    "description": "Updated description",
    "manager_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "active",
    "created_at": "2025-10-13T10:00:00Z",
    "updated_at": "2025-10-13T10:30:00Z"
  }
}
```

---

## 5. Delete Team

Soft delete a team (status changed to 'deleted').

**Endpoint:** `DELETE /api/v1/company-admin/teams/:id`  
**Role:** `company_admin`

### Response (200 OK)
```json
{
  "success": true,
  "message": "Team deleted successfully",
  "data": {
    "team_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7"
  }
}
```

---

## 6. Add Team Member

Add a user to a team with a specific role.

**Endpoint:** `POST /api/v1/company-admin/teams/:id/members`  
**Role:** `company_admin`

### Request Body
```json
{
  "user_id": "660e8400-e29b-41d4-a716-446655440000",
  "role_in_team": "driver"
}
```

### Possible Roles
- `manager` - Team manager
- `driver` - Vehicle driver
- `helper` - Driver's helper
- `member` - General team member

### Response (201 Created)
```json
{
  "success": true,
  "message": "Member added to team successfully",
  "data": {
    "team_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
    "user_id": "660e8400-e29b-41d4-a716-446655440000",
    "role_in_team": "driver"
  }
}
```

---

## 7. List Team Members

Get all members of a team with user details.

**Endpoint:** `GET /api/v1/company-admin/teams/:id/members`  
**Role:** `company_admin`, `admin`, `manager`

### Response (200 OK)
```json
{
  "success": true,
  "message": "Team members retrieved successfully",
  "data": {
    "members": [
      {
        "id": "aa9e6679-7425-40de-944b-e07fc1f90ae7",
        "team_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
        "user_id": "660e8400-e29b-41d4-a716-446655440000",
        "role_in_team": "driver",
        "joined_at": "2025-10-13T10:00:00Z",
        "user": {
          "id": "660e8400-e29b-41d4-a716-446655440000",
          "name": "John Doe",
          "email": "john@example.com",
          "cpf": "12345678900",
          "phone": "+5511999999999"
        }
      }
    ],
    "count": 1
  }
}
```

---

## 8. Remove Team Member

Remove a user from a team.

**Endpoint:** `DELETE /api/v1/company-admin/teams/:id/members/:userId`  
**Role:** `company_admin`

### Response (200 OK)
```json
{
  "success": true,
  "message": "Member removed from team successfully",
  "data": {
    "team_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
    "user_id": "660e8400-e29b-41d4-a716-446655440000"
  }
}
```

---

## 9. Update Member Role

Change a team member's role within the team.

**Endpoint:** `PUT /api/v1/company-admin/teams/:id/members/:userId/role`  
**Role:** `company_admin`

### Request Body
```json
{
  "role_in_team": "manager"
}
```

### Response (200 OK)
```json
{
  "success": true,
  "message": "Member role updated successfully",
  "data": {
    "team_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
    "user_id": "660e8400-e29b-41d4-a716-446655440000",
    "role_in_team": "manager"
  }
}
```

---

## 10. Get Team Statistics

Get comprehensive statistics for a team.

**Endpoint:** `GET /api/v1/company-admin/teams/:id/stats`  
**Role:** `company_admin`, `admin`

### Response (200 OK)
```json
{
  "success": true,
  "message": "Team statistics retrieved successfully",
  "data": {
    "team_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
    "team_name": "Team Alpha",
    "member_count": 5,
    "vehicle_count": 3,
    "active_vehicles": 2,
    "status": "active",
    "created_at": "2025-10-13T10:00:00Z",
    "manager_id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Statistics Included:**
- `member_count` - Total team members
- `vehicle_count` - Total vehicles assigned
- `active_vehicles` - Vehicles with status "active"
- `status` - Team status
- `created_at` - Team creation date
- `manager_id` - Team manager UUID

---

## 11. List Team Vehicles

Get all vehicles assigned to a team.

**Endpoint:** `GET /api/v1/company-admin/teams/:id/vehicles`  
**Role:** `company_admin`

### Response (200 OK)
```json
{
  "success": true,
  "message": "Team vehicles retrieved successfully",
  "data": {
    "team": {
      "id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
      "name": "Team Alpha",
      "status": "active"
    },
    "vehicles": [
      {
        "id": "890e8400-e29b-41d4-a716-446655440000",
        "company_id": "123e4567-e89b-12d3-a456-426614174000",
        "team_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
        "license_plate": "ABC-1234",
        "brand": "Ford",
        "model": "Transit",
        "year": 2023,
        "color": "White",
        "vehicle_type": "van",
        "fuel_type": "diesel",
        "capacity_kg": 1500.0,
        "driver_id": "660e8400-e29b-41d4-a716-446655440000",
        "helper_id": null,
        "status": "active",
        "created_at": "2025-10-13T09:00:00Z",
        "updated_at": "2025-10-13T10:00:00Z"
      }
    ],
    "count": 3
  }
}
```

---

## 12. Assign Vehicle to Team

Assign a vehicle to a team.

**Endpoint:** `POST /api/v1/company-admin/teams/:id/vehicles/:vehicleId`  
**Role:** `company_admin`

### Response (200 OK)
```json
{
  "success": true,
  "message": "Vehicle assigned to team successfully",
  "data": {
    "team_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
    "vehicle_id": "890e8400-e29b-41d4-a716-446655440000"
  }
}
```

### Validations
- Team must exist and belong to company
- Vehicle must exist and belong to company
- Vehicle can be assigned to only one team at a time

---

## 13. Unassign Vehicle from Team

Remove a vehicle from a team.

**Endpoint:** `DELETE /api/v1/company-admin/teams/:id/vehicles/:vehicleId`  
**Role:** `company_admin`

### Response (200 OK)
```json
{
  "success": true,
  "message": "Vehicle unassigned from team successfully",
  "data": {
    "team_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
    "vehicle_id": "890e8400-e29b-41d4-a716-446655440000"
  }
}
```

### Validations
- Team must exist and belong to company
- Vehicle must exist and belong to company
- Vehicle must be assigned to the specified team

---

## 14. Get My Teams

Get all teams the authenticated user belongs to.

**Endpoint:** `GET /api/v1/teams/my-teams`  
**Role:** `user` (any authenticated user)

### Response (200 OK)
```json
{
  "success": true,
  "message": "User teams retrieved successfully",
  "data": {
    "teams": [
      {
        "id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
        "company_id": "123e4567-e89b-12d3-a456-426614174000",
        "name": "Team Alpha",
        "description": "Sales and logistics team",
        "manager_id": "550e8400-e29b-41d4-a716-446655440000",
        "status": "active",
        "role_in_team": "driver",
        "joined_at": "2025-10-13T10:00:00Z"
      }
    ],
    "count": 1
  }
}
```

---

## Error Responses

### 400 Bad Request
```json
{
  "success": false,
  "message": "Invalid team ID"
}
```

### 401 Unauthorized
```json
{
  "success": false,
  "message": "Company context required"
}
```

### 403 Forbidden
```json
{
  "success": false,
  "message": "Insufficient permissions"
}
```

### 404 Not Found
```json
{
  "success": false,
  "message": "Team not found"
}
```

### 500 Internal Server Error
```json
{
  "success": false,
  "message": "Failed to retrieve team"
}
```

---

## Authentication

All endpoints require JWT Bearer token authentication:

```
Authorization: Bearer <jwt_token>
```

The token must contain:
- `user_id` - User UUID
- `company_id` - Company UUID (for multi-tenancy)
- `role` - User role (master, company_admin, admin, manager, user)

---

## Role Permissions

| Role | Permissions |
|------|-------------|
| `company_admin` | Full access to all team operations |
| `admin` | Read access + statistics |
| `manager` | Read-only access to teams |
| `user` | Access only to their own teams |

---

## Multi-Tenancy

All operations are scoped to the user's company context:
- Teams are filtered by `company_id`
- Vehicles are filtered by `company_id`
- Users can only access teams within their company
- Cross-company access is prevented at the database level

---

## OpenTelemetry Tracing

All endpoints are instrumented with OpenTelemetry:
- Span names: `TeamHandler.<MethodName>`
- Attributes: `team.id`, `company.id`, `user.id`, `vehicle.id`, counts, etc.
- Errors are recorded in spans

---

## Testing

See `TESTING_MANUAL.md` for:
- Postman collection
- curl examples
- Test scenarios
- Expected results

---

**API Version:** 1.0  
**Last Updated:** October 13, 2025
