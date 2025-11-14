

# Room API Documentation

Complete API reference for Room and User-Room relationship management.

## Overview

The Room API provides endpoints for managing rooms and their many-to-many relationships with users:
- **Users** can be assigned to **multiple rooms**
- **Rooms** can have **multiple users** assigned

## Room Endpoints

### Create Room

```http
POST /rooms
Content-Type: application/json

{
  "name": "Conference Room A",
  "description": "Large meeting room on 2nd floor",
  "room_type_id": 1
}
```

**Response:** `201 Created`
```json
{
  "id": 1,
  "name": "Conference Room A",
  "description": "Large meeting room on 2nd floor",
  "room_type_id": 1,
  "created_at": "2025-11-13T10:00:00Z",
  "updated_at": "2025-11-13T10:00:00Z"
}
```

### Get Room

```http
GET /rooms/{id}
```

**Response:** `200 OK`
```json
{
  "id": 1,
  "name": "Conference Room A",
  "description": "Large meeting room on 2nd floor",
  "room_type_id": 1,
  "created_at": "2025-11-13T10:00:00Z",
  "updated_at": "2025-11-13T10:00:00Z"
}
```

### List Rooms

```http
GET /rooms?limit=10&offset=0
```

**Query Parameters:**
- `limit` (optional): Number of rooms to return (default: 10)
- `offset` (optional): Number of rooms to skip (default: 0)

**Response:** `200 OK`
```json
[
  {
    "id": 1,
    "name": "Conference Room A",
    "description": "Large meeting room",
    "room_type_id": 1,
    "created_at": "2025-11-13T10:00:00Z",
    "updated_at": "2025-11-13T10:00:00Z"
  }
]
```

### Update Room

```http
PUT /rooms/{id}
Content-Type: application/json

{
  "name": "Updated Room Name",
  "description": "Updated description",
  "room_type_id": 2
}
```

**Response:** `200 OK`
```json
{
  "id": 1,
  "name": "Updated Room Name",
  "description": "Updated description",
  "room_type_id": 2,
  "created_at": "2025-11-13T10:00:00Z",
  "updated_at": "2025-11-13T10:30:00Z"
}
```

### Delete Room

```http
DELETE /rooms/{id}
```

**Response:** `204 No Content`

**Note:** Deleting a room will also remove all user assignments to that room (CASCADE).

## Room Type Endpoints

### Create Room Type

```http
POST /room_types
Content-Type: application/json

{
  "size": "large",
  "style": "modern"
}
```

**Response:** `201 Created`
```json
{
  "id": 1,
  "size": "large",
  "style": "modern",
  "created_at": "2025-11-13T10:00:00Z",
  "updated_at": "2025-11-13T10:00:00Z"
}
```

### Get Room Type

```http
GET /room_types/{id}
```

**Response:** `200 OK`
```json
{
  "id": 1,
  "size": "large",
  "style": "modern",
  "created_at": "2025-11-13T10:00:00Z",
  "updated_at": "2025-11-13T10:00:00Z"
}
```

### List Room Types

```http
GET /room_types?limit=10&offset=0
```

**Response:** `200 OK`
```json
[
  {
    "id": 1,
    "size": "large",
    "style": "modern",
    "created_at": "2025-11-13T10:00:00Z",
    "updated_at": "2025-11-13T10:00:00Z"
  },
  {
    "id": 2,
    "size": "small",
    "style": "traditional",
    "created_at": "2025-11-13T10:05:00Z",
    "updated_at": "2025-11-13T10:05:00Z"
  }
]
```

### Update Room Type

```http
PUT /room_types/{id}
Content-Type: application/json

{
  "size": "medium",
  "style": "contemporary"
}
```

**Response:** `200 OK`
```json
{
  "id": 1,
  "size": "medium",
  "style": "contemporary",
  "created_at": "2025-11-13T10:00:00Z",
  "updated_at": "2025-11-13T10:30:00Z"
}
```

### Delete Room Type

```http
DELETE /room_types/{id}
```

**Response:** `204 No Content`

**Note:** Deleting a room type will set the `room_type_id` to NULL for all rooms using that type (SET NULL).

## User-Room Relationship Endpoints

### Get Room Users

Get all users assigned to a specific room.

```http
GET /rooms/{id}/users
```

**Response:** `200 OK`
```json
{
  "id": 1,
  "name": "Conference Room A",
  "description": "Large meeting room",
  "room_type_id": 1,
  "created_at": "2025-11-13T10:00:00Z",
  "updated_at": "2025-11-13T10:00:00Z",
  "users": [
    {
      "id": 1,
      "external_id": "user_john_001",
      "created_at": "2025-11-13T09:00:00Z",
      "updated_at": "2025-11-13T09:00:00Z"
    },
    {
      "id": 2,
      "external_id": "user_jane_002",
      "created_at": "2025-11-13T09:30:00Z",
      "updated_at": "2025-11-13T09:30:00Z"
    }
  ]
}
```

### Assign User to Room

Assign a user to a room. Users can be in multiple rooms.

```http
POST /rooms/{roomId}/users
Content-Type: application/json

{
  "user_id": 1
}
```

**Response:** `200 OK`
```json
{
  "message": "User assigned to room successfully"
}
```

**Error Response:** `409 Conflict` (if user already in this room)
```json
{
  "error": "user already assigned to this room"
}
```

### Remove User from Room

Remove a user from a specific room.

```http
DELETE /rooms/{roomId}/users/{userId}
```

**Response:** `204 No Content`

### Get User Rooms

Get all rooms assigned to a specific user.

```http
GET /users/{id}/rooms
```

**Response:** `200 OK`
```json
[
  {
    "id": 1,
    "name": "Conference Room A",
    "description": "Large meeting room",
    "room_type_id": 1,
    "created_at": "2025-11-13T10:00:00Z",
    "updated_at": "2025-11-13T10:00:00Z"
  },
  {
    "id": 2,
    "name": "Meeting Room B",
    "description": "Small meeting room",
    "room_type_id": 2,
    "created_at": "2025-11-13T10:15:00Z",
    "updated_at": "2025-11-13T10:15:00Z"
  }
]
```

## Testing with cURL

### Create and Assign Workflow

```bash
# 1. Create a room type first
curl -X POST http://localhost:8080/room_types \
  -H "Content-Type: application/json" \
  -d '{
    "size": "large",
    "style": "modern"
  }'

# 2. Create a room with room_type_id
curl -X POST http://localhost:8080/rooms \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Dev Team Room",
    "description": "Development team workspace",
    "room_type_id": 1
  }'

# 3. Create users
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"external_id": "user_dev1"}'

curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"external_id": "user_dev2"}'

# 4. Assign users to room
curl -X POST http://localhost:8080/rooms/1/users \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1}'

curl -X POST http://localhost:8080/rooms/1/users \
  -H "Content-Type: application/json" \
  -d '{"user_id": 2}'

# 5. Get room with users
curl http://localhost:8080/rooms/1/users

# 6. Get user's rooms
curl http://localhost:8080/users/1/rooms

# 7. Remove user from room
curl -X DELETE http://localhost:8080/rooms/1/users/1
```

## Error Responses

| Status Code | Description | Example |
|-------------|-------------|---------|
| 400 | Bad Request | Invalid room ID, missing required fields |
| 404 | Not Found | Room or user doesn't exist |
| 409 | Conflict | User already assigned to room |
| 500 | Internal Server Error | Database error |

## Business Rules

1. **Many-to-Many Relationship**
   - A user can be in multiple rooms
   - A room can have multiple users
   - No duplicate assignments (enforced by database constraint)

2. **Cascade Deletion**
   - Deleting a room removes all user assignments
   - Deleting a user removes all their room assignments

3. **Room Types**
   - Rooms can be associated with room types via `room_type_id`
   - Room type is optional (nullable foreign key)
   - Room types define size and style characteristics
   - Deleting a room type sets affected rooms' `room_type_id` to NULL

## Database Schema

### Rooms Table

```sql
CREATE TABLE rooms (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    room_type_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (room_type_id) REFERENCES room_types(id) ON DELETE SET NULL
);
```

### Room Types Table

```sql
CREATE TABLE room_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    size TEXT NOT NULL,
    style TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### User-Rooms Junction Table

```sql
CREATE TABLE user_rooms (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    room_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE,
    UNIQUE(user_id, room_id)
);
```

## Code Examples

### Go

```go
// Create room
type CreateRoomRequest struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    RoomTypeID  *int   `json:"room_type_id,omitempty"`
}

req := CreateRoomRequest{
    Name:        "Conference Room",
    Description: "Main conference room",
    RoomTypeID:  intPtr(1), // Helper function: func intPtr(i int) *int { return &i }
}

// Assign user to room
type AssignRequest struct {
    UserID int64 `json:"user_id"`
}

assignReq := AssignRequest{UserID: 1}
```

### JavaScript

```javascript
// Create room
const room = await fetch('http://localhost:8080/rooms', {
  method: 'POST',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({
    name: 'Conference Room',
    description: 'Main meeting space',
    room_type_id: 1
  })
});

// Assign user to room
await fetch('http://localhost:8080/rooms/1/users', {
  method: 'POST',
  headers: {'Content-Type': 'application/json'},
  body: JSON.stringify({user_id: 1})
});

// Get user's rooms
const rooms = await fetch('http://localhost:8080/users/1/rooms');
const userRooms = await rooms.json();
```

### Python

```python
import requests

# Create room
room = requests.post('http://localhost:8080/rooms', json={
    'name': 'Conference Room',
    'description': 'Main meeting space',
    'room_type_id': 1
})

# Assign user to room
requests.post('http://localhost:8080/rooms/1/users', json={'user_id': 1})

# Get user's rooms
rooms = requests.get('http://localhost:8080/users/1/rooms')
user_rooms = rooms.json()
```

## Testing Scenarios

### Scenario 1: Team Room Setup

```
1. Create a room type (size: "medium", style: "collaborative")
2. Create "Dev Team Room" with room_type_id pointing to the room type
3. Create 5 developer users with external_ids
4. Assign all developers to the room
5. Verify room shows 5 users
6. Verify each user shows "Dev Team Room" in their rooms list
```

### Scenario 2: User in Multiple Rooms

```
1. Create "Morning Standup" room
2. Create "Project X" room
3. Create "All Hands" room
4. Assign user Alice to all three rooms
5. GET /users/{alice_id}/rooms should return 3 rooms
6. Each room should show Alice in their users list
```

### Scenario 3: Remove and Cleanup

```
1. Assign user Bob to "Meeting Room A"
2. Verify assignment
3. Remove Bob from "Meeting Room A"
4. Verify Bob has no rooms
5. Delete "Meeting Room A"
6. Verify all assignments removed
```

## See Also

- [User API Documentation](../README.md#api-endpoints)
- [API Testing Page](API_TESTER.md)
- [Database Setup](DATABASE_SETUP.md)
