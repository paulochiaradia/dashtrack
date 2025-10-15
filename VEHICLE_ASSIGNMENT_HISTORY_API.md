# Vehicle Assignment History API

Complete documentation for tracking and querying vehicle assignment changes in the DashTrack system.

## Table of Contents
- [Overview](#overview)
- [Features](#features)
- [Endpoint](#endpoint)
- [Data Model](#data-model)
- [Examples](#examples)
- [Testing](#testing)

## Overview

The Vehicle Assignment History system automatically tracks all changes to vehicle assignments, including:
- Driver assignments and changes
- Helper assignments and changes
- Team assignments and changes
- Combined assignment updates

Every time a vehicle's driver, helper, or team assignment is modified, a history record is automatically created with before/after states and metadata.

## Features

✅ **Automatic Tracking** - All assignment changes are logged automatically  
✅ **Change Type Classification** - Identifies whether change was driver, helper, team, or full assignment  
✅ **Complete Audit Trail** - Tracks previous and new values for all fields  
✅ **User Context** - Records which user made the change (when available)  
✅ **Populated Details** - Returns full user and team objects, not just IDs  
✅ **Flexible Querying** - Supports limit parameter for pagination  
✅ **Non-Blocking** - History logging failures don't block assignment updates  

## Endpoint

### Get Vehicle Assignment History

Retrieve the complete assignment history for a specific vehicle.

**Endpoint:** `GET /api/v1/company-admin/vehicles/:id/assignment-history`

**Alternative:** `GET /api/v1/admin/vehicles/:id/assignment-history` (Admin role)

**Path Parameters:**
- `id` (UUID) - Vehicle ID

**Query Parameters:**
- `limit` (integer, optional) - Maximum number of history entries to return (default: 50, max: 500)

**Authentication:**
- Requires: `company_admin` or `admin` role
- Company context is enforced (only see history for your company's vehicles)

**Success Response:** `200 OK`
```json
{
  "success": true,
  "message": "Vehicle assignment history retrieved successfully",
  "data": {
    "vehicle": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "license_plate": "ABC-1234",
      "brand": "Volvo",
      "model": "FH 540"
    },
    "history": [
      {
        "id": "650e8400-e29b-41d4-a716-446655440001",
        "vehicle_id": "550e8400-e29b-41d4-a716-446655440000",
        "company_id": "450e8400-e29b-41d4-a716-446655440000",
        "previous_driver_id": "750e8400-e29b-41d4-a716-446655440001",
        "previous_helper_id": "750e8400-e29b-41d4-a716-446655440002",
        "previous_team_id": null,
        "new_driver_id": "750e8400-e29b-41d4-a716-446655440001",
        "new_helper_id": "750e8400-e29b-41d4-a716-446655440003",
        "new_team_id": "850e8400-e29b-41d4-a716-446655440001",
        "change_type": "full_assignment",
        "changed_by_user_id": null,
        "change_reason": null,
        "changed_at": "2024-01-15T14:30:00Z",
        "created_at": "2024-01-15T14:30:00Z",
        "previous_driver": {
          "id": "750e8400-e29b-41d4-a716-446655440001",
          "name": "John Driver",
          "email": "john@company.com",
          "role": "user"
        },
        "previous_helper": {
          "id": "750e8400-e29b-41d4-a716-446655440002",
          "name": "Jane Helper",
          "email": "jane@company.com",
          "role": "user"
        },
        "new_helper": {
          "id": "750e8400-e29b-41d4-a716-446655440003",
          "name": "Bob Assistant",
          "email": "bob@company.com",
          "role": "user"
        },
        "new_team": {
          "id": "850e8400-e29b-41d4-a716-446655440001",
          "name": "Logistics Team A",
          "company_id": "450e8400-e29b-41d4-a716-446655440000"
        }
      },
      {
        "id": "650e8400-e29b-41d4-a716-446655440002",
        "vehicle_id": "550e8400-e29b-41d4-a716-446655440000",
        "company_id": "450e8400-e29b-41d4-a716-446655440000",
        "previous_driver_id": null,
        "previous_helper_id": null,
        "previous_team_id": null,
        "new_driver_id": "750e8400-e29b-41d4-a716-446655440001",
        "new_helper_id": "750e8400-e29b-41d4-a716-446655440002",
        "new_team_id": null,
        "change_type": "full_assignment",
        "changed_by_user_id": null,
        "change_reason": null,
        "changed_at": "2024-01-15T10:00:00Z",
        "created_at": "2024-01-15T10:00:00Z",
        "new_driver": {
          "id": "750e8400-e29b-41d4-a716-446655440001",
          "name": "John Driver",
          "email": "john@company.com",
          "role": "user"
        },
        "new_helper": {
          "id": "750e8400-e29b-41d4-a716-446655440002",
          "name": "Jane Helper",
          "email": "jane@company.com",
          "role": "user"
        }
      }
    ],
    "count": 2,
    "limit": 50
  }
}
```

**Error Responses:**
- `400 Bad Request` - Invalid vehicle ID or parameters
- `404 Not Found` - Vehicle not found or doesn't belong to your company
- `500 Internal Server Error` - Failed to retrieve history

## Data Model

### VehicleAssignmentHistory

```go
type VehicleAssignmentHistory struct {
    ID               uuid.UUID  `json:"id"`
    VehicleID        uuid.UUID  `json:"vehicle_id"`
    CompanyID        uuid.UUID  `json:"company_id"`
    
    // Previous state (before change)
    PreviousDriverID *uuid.UUID `json:"previous_driver_id"`
    PreviousHelperID *uuid.UUID `json:"previous_helper_id"`
    PreviousTeamID   *uuid.UUID `json:"previous_team_id"`
    
    // New state (after change)
    NewDriverID      *uuid.UUID `json:"new_driver_id"`
    NewHelperID      *uuid.UUID `json:"new_helper_id"`
    NewTeamID        *uuid.UUID `json:"new_team_id"`
    
    // Metadata
    ChangeType       string     `json:"change_type"`      // "driver", "helper", "team", "full_assignment"
    ChangedByUserID  *uuid.UUID `json:"changed_by_user_id"`
    ChangeReason     *string    `json:"change_reason"`
    ChangedAt        time.Time  `json:"changed_at"`
    CreatedAt        time.Time  `json:"created_at"`
    
    // Populated fields (retrieved via JOINs)
    PreviousDriver   *User      `json:"previous_driver,omitempty"`
    PreviousHelper   *User      `json:"previous_helper,omitempty"`
    PreviousTeam     *Team      `json:"previous_team,omitempty"`
    NewDriver        *User      `json:"new_driver,omitempty"`
    NewHelper        *User      `json:"new_helper,omitempty"`
    NewTeam          *Team      `json:"new_team,omitempty"`
    ChangedByUser    *User      `json:"changed_by_user,omitempty"`
}
```

### Change Types

| Change Type | Description | Example |
|------------|-------------|---------|
| `driver` | Only driver changed | Assigned new driver, helper and team unchanged |
| `helper` | Only helper changed | Changed helper, driver and team unchanged |
| `team` | Only team changed | Moved to different team, driver/helper unchanged |
| `full_assignment` | Multiple fields changed | Driver AND helper changed, or all three changed |

## Examples

### PowerShell

```powershell
# Setup
$baseUrl = "http://localhost:8080/api/v1"
$token = "your-token-here"
$headers = @{
    "Authorization" = "Bearer $token"
}

# Get full history (default limit 50)
$history = Invoke-RestMethod -Uri "$baseUrl/company-admin/vehicles/$vehicleId/assignment-history" `
    -Method Get `
    -Headers $headers

Write-Host "Total history entries: $($history.data.count)"

foreach ($entry in $history.data.history) {
    Write-Host "`nChange at: $($entry.changed_at)"
    Write-Host "Type: $($entry.change_type)"
    
    if ($entry.new_driver) {
        Write-Host "New Driver: $($entry.new_driver.name)"
    }
}

# Get limited history (last 10 changes)
$recentHistory = Invoke-RestMethod -Uri "$baseUrl/company-admin/vehicles/$vehicleId/assignment-history?limit=10" `
    -Method Get `
    -Headers $headers

Write-Host "Recent changes: $($recentHistory.data.count)"
```

### cURL

```bash
# Get full history
curl -X GET "http://localhost:8080/api/v1/company-admin/vehicles/{vehicle_id}/assignment-history" \
  -H "Authorization: Bearer {token}"

# Get limited history
curl -X GET "http://localhost:8080/api/v1/company-admin/vehicles/{vehicle_id}/assignment-history?limit=10" \
  -H "Authorization: Bearer {token}"
```

### JavaScript/Fetch

```javascript
// Get vehicle assignment history
async function getVehicleHistory(vehicleId, limit = 50) {
  const response = await fetch(
    `${API_BASE}/company-admin/vehicles/${vehicleId}/assignment-history?limit=${limit}`,
    {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    }
  );
  
  const data = await response.json();
  
  data.data.history.forEach(entry => {
    console.log(`${entry.changed_at}: ${entry.change_type}`);
    
    if (entry.new_driver) {
      console.log(`  Driver: ${entry.new_driver.name}`);
    }
    
    if (entry.new_helper) {
      console.log(`  Helper: ${entry.new_helper.name}`);
    }
    
    if (entry.new_team) {
      console.log(`  Team: ${entry.new_team.name}`);
    }
  });
}
```

## Testing

### Using the Test Script

A comprehensive PowerShell test script is available at `scripts/test-vehicle-assignment-history.ps1`.

**Usage:**
```powershell
cd scripts
.\test-vehicle-assignment-history.ps1
```

**The script tests:**
1. ✓ Creates test users (driver, 2 helpers)
2. ✓ Creates test team and vehicle
3. ✓ Makes 5 different assignment changes:
   - Assign driver only
   - Add helper
   - Assign to team
   - Change helper
   - Remove all assignments
4. ✓ Retrieves full history with details
5. ✓ Tests limit parameter
6. ✓ Verifies populated user/team details
7. ✓ Optional cleanup

### Manual Testing Checklist

- [ ] Assign driver to vehicle → Check history shows driver change
- [ ] Assign helper to vehicle → Check history shows helper change
- [ ] Assign team to vehicle → Check history shows team change
- [ ] Change driver → Check previous and new driver in history
- [ ] Change multiple fields at once → Check change_type is "full_assignment"
- [ ] Remove all assignments → Check history shows removal
- [ ] Query with different limits → Verify limit is respected
- [ ] Check user details are populated → Verify names/emails present
- [ ] Check team details are populated → Verify team names present
- [ ] Verify ordering → Most recent changes should appear first

## How It Works

### Automatic Logging

The history tracking is fully automatic. When `VehicleRepository.UpdateAssignment()` is called:

1. **Capture Current State** - Query current driver/helper/team before update
2. **Perform Update** - Execute the vehicle assignment update
3. **Determine Change Type** - Analyze what changed (driver/helper/team/multiple)
4. **Log History** - Insert history record with previous and new values
5. **Non-Blocking** - If history logging fails, update still succeeds

### Change Type Logic

```
if only driver changed    → change_type = "driver"
if only helper changed    → change_type = "helper"
if only team changed      → change_type = "team"
if 2+ fields changed      → change_type = "full_assignment"
if nothing changed        → no history record created
```

### Performance Considerations

- History is written asynchronously after vehicle update
- Logging failures don't block the main operation
- Indexes on `vehicle_id`, `changed_at`, and other fields ensure fast queries
- User/team details are populated with separate queries (not JOINs in history table)
- Default limit of 50 prevents excessive data transfer

## Database Schema

```sql
CREATE TABLE vehicle_assignment_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    
    previous_driver_id UUID REFERENCES users(id) ON DELETE SET NULL,
    previous_helper_id UUID REFERENCES users(id) ON DELETE SET NULL,
    previous_team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
    
    new_driver_id UUID REFERENCES users(id) ON DELETE SET NULL,
    new_helper_id UUID REFERENCES users(id) ON DELETE SET NULL,
    new_team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
    
    change_type VARCHAR(50) NOT NULL CHECK (change_type IN ('driver', 'helper', 'team', 'full_assignment')),
    changed_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    change_reason TEXT,
    changed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_vehicle_assignment_history_vehicle ON vehicle_assignment_history(vehicle_id);
CREATE INDEX idx_vehicle_assignment_history_changed_at ON vehicle_assignment_history(changed_at DESC);
```

## Integration Points

### Repository Layer
- `UpdateAssignment()` - Modified to capture state and log changes
- `LogAssignmentChange()` - Inserts history record
- `GetAssignmentHistory()` - Retrieves history with basic data
- `GetAssignmentHistoryWithDetails()` - Retrieves history with populated user/team details

### Handler Layer
- `AssignUsers()` in VehicleHandler - Calls UpdateAssignment (automatic logging)
- `AssignVehicleToTeam()` in TeamHandler - Calls UpdateAssignment (automatic logging)
- `GetVehicleAssignmentHistory()` - NEW endpoint for querying history

### Routes
- `GET /company-admin/vehicles/:id/assignment-history` - Company admin access
- `GET /admin/vehicles/:id/assignment-history` - Admin access

## Future Enhancements

Potential improvements for future versions:

- **User Context** - Pass `changed_by_user_id` from authentication context
- **Change Reasons** - Allow optional reason/comment when making changes
- **Webhooks** - Trigger notifications when assignments change
- **Analytics** - Dashboard showing assignment patterns
- **Export** - CSV/PDF export of history for reporting
- **Filtering** - Filter by date range, change type, user
- **Bulk History** - Get history for multiple vehicles in one request

## Notes

- History records are immutable (insert-only, never updated or deleted)
- CASCADE deletion ensures history is cleaned up when vehicle is deleted
- SET NULL on user/team deletions preserves history but removes references
- Default limit prevents accidentally loading huge datasets
- Most recent changes appear first (ORDER BY changed_at DESC)
