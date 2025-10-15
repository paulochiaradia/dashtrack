# Team Member History API Documentation

## Overview

The Team Member History API provides complete audit tracking for team membership changes. Every time a member is added, removed, has their role changed, or is transferred between teams, a history record is automatically created. This enables comprehensive team analytics, compliance auditing, and change tracking.

## Features

- **Automatic History Tracking**: All member operations (add, remove, role updates, transfers) automatically log to history
- **Non-Blocking Logging**: History failures don't prevent the main operation from succeeding
- **Change Type Classification**: Clear categorization of all changes (added, removed, role_changed)
- **Transfer Detection**: Special handling for member transfers between teams
- **Dual Query Capability**: Query by team (team-centric) or by user (user-centric)
- **Populated Details**: History records include full user and team object details
- **Multi-Tenant Support**: All queries are scoped to the authenticated user's company
- **Pagination**: Configurable result limits (default 50, max 500)

## Change Types

| Change Type | Description | Fields Populated |
|------------|-------------|------------------|
| `added` | Member added to team | `new_role_in_team` |
| `removed` | Member removed from team | `previous_role_in_team` |
| `role_changed` | Member's role updated | `previous_role_in_team`, `new_role_in_team` |
| `transferred_in` | *(Reserved for future use)* | Transfer fields |
| `transferred_out` | *(Reserved for future use)* | Transfer fields |

**Note**: Currently, transfers use two separate records (`removed` from old team, `added` to new team). Future versions may use `transferred_in` and `transferred_out` with linked records.

## Endpoints

### 1. Get Team Member History

Retrieve the complete history of membership changes for a specific team.

**Endpoint**: `GET /api/v1/company-admin/teams/:id/member-history`

**Authentication**: Required (company_admin role)

**URL Parameters**:
- `id` (required): Team UUID

**Query Parameters**:
- `limit` (optional): Maximum number of records to return (default: 50, max: 500)

**Response**:
```json
{
  "status": "success",
  "message": "Team member history retrieved successfully",
  "data": {
    "team": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Delivery Team Alpha"
    },
    "history": [
      {
        "id": "123e4567-e89b-12d3-a456-426614174000",
        "team_id": "550e8400-e29b-41d4-a716-446655440000",
        "user_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
        "company_id": "c9a38d9f-9d6e-4b6e-8c1a-3f5e8d6e9b1a",
        "change_type": "role_changed",
        "previous_role_in_team": "driver",
        "new_role_in_team": "team_lead",
        "previous_team_id": null,
        "new_team_id": null,
        "changed_at": "2024-01-15T14:30:00Z",
        "changed_by_user_id": "admin-user-uuid",
        "notes": null,
        "user": {
          "id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
          "name": "John Doe",
          "email": "john@example.com",
          "role": "user"
        },
        "team": {
          "id": "550e8400-e29b-41d4-a716-446655440000",
          "name": "Delivery Team Alpha",
          "description": "Main delivery operations team"
        },
        "changed_by_user": {
          "id": "admin-user-uuid",
          "name": "Admin User",
          "email": "admin@company.com"
        }
      }
    ],
    "count": 1,
    "limit": 50
  }
}
```

### 2. Get User Team History

Retrieve the complete history of team memberships for a specific user.

**Endpoint**: `GET /api/v1/company-admin/teams/users/:userId/team-history`

**Authentication**: Required (company_admin role)

**URL Parameters**:
- `userId` (required): User UUID

**Query Parameters**:
- `limit` (optional): Maximum number of records to return (default: 50, max: 500)

**Response**:
```json
{
  "status": "success",
  "message": "User team history retrieved successfully",
  "data": {
    "user_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
    "history": [
      {
        "id": "123e4567-e89b-12d3-a456-426614174000",
        "team_id": "550e8400-e29b-41d4-a716-446655440000",
        "user_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
        "company_id": "c9a38d9f-9d6e-4b6e-8c1a-3f5e8d6e9b1a",
        "change_type": "added",
        "previous_role_in_team": null,
        "new_role_in_team": "driver",
        "previous_team_id": null,
        "new_team_id": null,
        "changed_at": "2024-01-10T09:00:00Z",
        "changed_by_user_id": "admin-user-uuid",
        "notes": null,
        "team": {
          "id": "550e8400-e29b-41d4-a716-446655440000",
          "name": "Delivery Team Alpha",
          "description": "Main delivery operations team"
        },
        "user": {
          "id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
          "name": "John Doe",
          "email": "john@example.com"
        }
      }
    ],
    "count": 1,
    "limit": 50
  }
}
```

## Data Model

### TeamMemberHistory

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Unique identifier for the history record |
| `team_id` | UUID | ID of the team (FK to teams) |
| `user_id` | UUID | ID of the user (FK to users) |
| `company_id` | UUID | ID of the company (for multi-tenancy) |
| `change_type` | VARCHAR(50) | Type of change (added, removed, role_changed, etc.) |
| `previous_role_in_team` | VARCHAR(50) | Role before change (null for additions) |
| `new_role_in_team` | VARCHAR(50) | Role after change (null for removals) |
| `previous_team_id` | UUID | Previous team (for transfers) |
| `new_team_id` | UUID | New team (for transfers) |
| `changed_at` | TIMESTAMP | When the change occurred |
| `changed_by_user_id` | UUID | User who made the change (FK to users) |
| `notes` | TEXT | Optional notes about the change |

### Populated Objects

When using the "WithDetails" endpoints, the following objects are populated:

- **user**: Full user object (id, name, email, role, etc.)
- **team**: Full team object (id, name, description, etc.)
- **previous_team**: Full team object (for transfers)
- **new_team**: Full team object (for transfers)
- **changed_by_user**: Full user object of who made the change

## Example Usage

### PowerShell

```powershell
# Get team member history
$token = "your-jwt-token"
$teamId = "550e8400-e29b-41d4-a716-446655440000"

$response = Invoke-RestMethod `
    -Method GET `
    -Uri "http://localhost:8080/api/v1/company-admin/teams/$teamId/member-history?limit=20" `
    -Headers @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }

# Display history
foreach ($record in $response.data.history) {
    Write-Host "$($record.user.name) - $($record.change_type) - $($record.changed_at)"
}

# Get user team history
$userId = "7c9e6679-7425-40de-944b-e07fc1f90ae7"

$userHistory = Invoke-RestMethod `
    -Method GET `
    -Uri "http://localhost:8080/api/v1/company-admin/teams/users/$userId/team-history" `
    -Headers @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }

foreach ($record in $userHistory.data.history) {
    Write-Host "$($record.team.name) - $($record.change_type) - $($record.changed_at)"
}
```

### cURL

```bash
# Get team member history
curl -X GET \
  "http://localhost:8080/api/v1/company-admin/teams/550e8400-e29b-41d4-a716-446655440000/member-history?limit=20" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json"

# Get user team history
curl -X GET \
  "http://localhost:8080/api/v1/company-admin/teams/users/7c9e6679-7425-40de-944b-e07fc1f90ae7/team-history" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json"
```

### JavaScript (Fetch API)

```javascript
// Get team member history
async function getTeamMemberHistory(teamId, limit = 50) {
  const response = await fetch(
    `http://localhost:8080/api/v1/company-admin/teams/${teamId}/member-history?limit=${limit}`,
    {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    }
  );
  
  const data = await response.json();
  return data.data.history;
}

// Get user team history
async function getUserTeamHistory(userId, limit = 50) {
  const response = await fetch(
    `http://localhost:8080/api/v1/company-admin/teams/users/${userId}/team-history?limit=${limit}`,
    {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    }
  );
  
  const data = await response.json();
  return data.data.history;
}

// Example usage
const teamHistory = await getTeamMemberHistory('550e8400-e29b-41d4-a716-446655440000');
console.log('Team history:', teamHistory);

const userHistory = await getUserTeamHistory('7c9e6679-7425-40de-944b-e07fc1f90ae7');
console.log('User history:', userHistory);
```

## Common Use Cases

### 1. Audit Trail

Track who made changes to team composition and when:

```javascript
const history = await getTeamMemberHistory(teamId);
history.forEach(record => {
  console.log(`${record.changed_at}: ${record.change_type} by ${record.changed_by_user.name}`);
});
```

### 2. User Career Timeline

View a user's progression through different teams and roles:

```javascript
const userHistory = await getUserTeamHistory(userId);
userHistory.forEach(record => {
  if (record.change_type === 'role_changed') {
    console.log(`Promoted from ${record.previous_role_in_team} to ${record.new_role_in_team}`);
  }
});
```

### 3. Team Turnover Analysis

Analyze team stability by counting additions and removals:

```javascript
const history = await getTeamMemberHistory(teamId);
const additions = history.filter(r => r.change_type === 'added').length;
const removals = history.filter(r => r.change_type === 'removed').length;
console.log(`Turnover: ${additions} additions, ${removals} removals`);
```

### 4. Role Distribution Tracking

Track how roles have evolved over time:

```javascript
const roleChanges = history.filter(r => r.change_type === 'role_changed');
roleChanges.forEach(change => {
  console.log(`${change.user.name}: ${change.previous_role_in_team} → ${change.new_role_in_team}`);
});
```

## Testing

A comprehensive test script is provided to verify all functionality:

```powershell
# Run all tests
.\scripts\test-team-member-history.ps1

# Run with verbose output
.\scripts\test-team-member-history.ps1 -Verbose

# Run with automatic cleanup
.\scripts\test-team-member-history.ps1 -Cleanup

# Custom API base URL
.\scripts\test-team-member-history.ps1 -BaseURL "http://localhost:8080/api/v1"
```

The test script verifies:
- ✅ Member additions (change_type: added)
- ✅ Member removals (change_type: removed)
- ✅ Role changes (change_type: role_changed)
- ✅ Member transfers (removed + added records)
- ✅ Team history queries with populated details
- ✅ User history queries with populated details
- ✅ Pagination and limit parameters

## Error Handling

### Common Error Responses

**404 Not Found** - Team or user doesn't exist:
```json
{
  "status": "error",
  "message": "Team not found"
}
```

**400 Bad Request** - Invalid parameters:
```json
{
  "status": "error",
  "message": "Invalid team ID"
}
```

**401 Unauthorized** - Missing or invalid token:
```json
{
  "status": "error",
  "message": "Unauthorized"
}
```

**403 Forbidden** - Insufficient permissions:
```json
{
  "status": "error",
  "message": "Access denied"
}
```

## Database Schema

```sql
CREATE TABLE team_member_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    change_type VARCHAR(50) NOT NULL,
    previous_role_in_team VARCHAR(50),
    new_role_in_team VARCHAR(50),
    previous_team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
    new_team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
    changed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    changed_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    notes TEXT
);

-- Indexes for efficient queries
CREATE INDEX idx_team_member_history_team_id ON team_member_history(team_id);
CREATE INDEX idx_team_member_history_user_id ON team_member_history(user_id);
CREATE INDEX idx_team_member_history_company_id ON team_member_history(company_id);
CREATE INDEX idx_team_member_history_changed_at ON team_member_history(changed_at DESC);
CREATE INDEX idx_team_member_history_change_type ON team_member_history(change_type);
CREATE INDEX idx_team_member_history_previous_team ON team_member_history(previous_team_id);
CREATE INDEX idx_team_member_history_new_team ON team_member_history(new_team_id);
```

## Integration Points

### Automatic History Logging

History is automatically created by these operations:

1. **AddMember** (repository method)
   - Creates `added` record with `new_role_in_team`

2. **RemoveMember** (repository method)
   - Captures current role before deletion
   - Creates `removed` record with `previous_role_in_team`

3. **UpdateMemberRole** (repository method)
   - Captures old role before update
   - Creates `role_changed` record (only if role actually changed)
   - Includes both `previous_role_in_team` and `new_role_in_team`

4. **TransferMemberToTeam** (handler method)
   - Calls `RemoveMember` (logs removal from old team)
   - Calls `AddMember` (logs addition to new team)
   - Results in two separate history records

### Non-Blocking Logging

All history logging is non-blocking:
- History insertion failures are logged but don't prevent the main operation
- This ensures team operations always succeed even if history tracking fails
- Failed history operations are logged to application logs for monitoring

## Performance Considerations

1. **Indexes**: Seven indexes optimize common query patterns
2. **Pagination**: Default limit of 50, max 500 to prevent large result sets
3. **WithDetails Queries**: Use separate queries + population rather than complex JOINs
4. **Company Scoping**: All queries filtered by company_id for security and performance
5. **Timestamp Index**: Descending index on `changed_at` for recent history queries

## Best Practices

1. **Always use pagination**: Specify appropriate limits for your use case
2. **Choose the right endpoint**: Use team history for team analytics, user history for employee records
3. **Cache frequently accessed history**: Consider caching recent team history
4. **Monitor history table growth**: Implement archival strategies for old history
5. **Use change types for filtering**: Query specific change types when doing analytics

## Related Documentation

- [Team Management API](./TEAM_MANAGEMENT_API.md)
- [Team Members API](./TEAM_MEMBERS_API.md)
- [Vehicle Assignment History API](./VEHICLE_ASSIGNMENT_HISTORY_API.md)

## Support

For issues or questions, please refer to the main API documentation or contact the development team.
