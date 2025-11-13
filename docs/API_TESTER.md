# API Testing Page Guide

The CloudflareDB API includes a built-in web interface for testing all API endpoints interactively.

## Quick Start

1. **Start the server:**
```bash
go run cmd/api/main.go
# Or
make run
```

2. **Open your browser:**
```
http://localhost:8080
```

3. **Start testing!**

## Features

### Visual Interface

- **Modern Design**: Clean, gradient background with card-based layout
- **Color-Coded Methods**: Each HTTP method has a distinct color
  - 🟢 GET - Green
  - 🔵 POST - Blue
  - 🟠 PUT - Orange
  - 🔴 DELETE - Red

### Real-Time Testing

- Interactive forms for each endpoint
- Instant feedback with formatted responses
- Loading indicators during requests
- Auto-clearing forms after successful operations

### Response Viewer

- Console-style response display
- JSON syntax highlighting
- Timestamps for each request
- Status badges (Success/Error)
- Scrollable response area
- Clear button to reset display

## Using the Testing Page

### 1. Health Check

**Purpose**: Verify the API is running

**Steps**:
1. Click "Check Health" button
2. View response showing `{"status":"healthy"}`

**Expected Response**:
```json
{
  "status": "healthy"
}
```

### 2. Create User

**Purpose**: Add a new user to the database

**Steps**:
1. Enter email address (e.g., `john@example.com`)
2. Enter full name (e.g., `John Doe`)
3. Click "Create User"

**Expected Response**:
```json
{
  "id": 1,
  "email": "john@example.com",
  "name": "John Doe",
  "created_at": "2025-11-13T10:00:00Z",
  "updated_at": "2025-11-13T10:00:00Z"
}
```

**Note**: The form automatically clears after successful creation.

### 3. Get User

**Purpose**: Retrieve a specific user by ID

**Steps**:
1. Enter user ID (e.g., `1`)
2. Click "Get User"

**Expected Response**:
```json
{
  "id": 1,
  "email": "john@example.com",
  "name": "John Doe",
  "created_at": "2025-11-13T10:00:00Z",
  "updated_at": "2025-11-13T10:00:00Z"
}
```

**Error Response** (if user doesn't exist):
```json
{
  "error": "User not found"
}
```

### 4. List Users

**Purpose**: Retrieve multiple users with pagination

**Steps**:
1. Set limit (default: 10)
2. Set offset (default: 0)
3. Click "List Users"

**Expected Response**:
```json
[
  {
    "id": 1,
    "email": "john@example.com",
    "name": "John Doe",
    "created_at": "2025-11-13T10:00:00Z",
    "updated_at": "2025-11-13T10:00:00Z"
  },
  {
    "id": 2,
    "email": "jane@example.com",
    "name": "Jane Smith",
    "created_at": "2025-11-13T10:05:00Z",
    "updated_at": "2025-11-13T10:05:00Z"
  }
]
```

**Pagination Examples**:
- First 10 users: `limit=10, offset=0`
- Next 10 users: `limit=10, offset=10`
- First 5 users: `limit=5, offset=0`

### 5. Update User

**Purpose**: Modify user information

**Steps**:
1. Enter user ID to update
2. Enter new email (optional)
3. Enter new name (optional)
4. Click "Update User"

**Note**: You can update just email, just name, or both.

**Expected Response**:
```json
{
  "id": 1,
  "email": "newemail@example.com",
  "name": "Updated Name",
  "created_at": "2025-11-13T10:00:00Z",
  "updated_at": "2025-11-13T10:30:00Z"
}
```

### 6. Delete User

**Purpose**: Remove a user from the database

**Steps**:
1. Enter user ID to delete
2. Click "Delete User"
3. Confirm deletion in the popup dialog

**Expected Response**: Empty response with 204 status code

**Note**: A confirmation dialog prevents accidental deletions.

## Tips and Best Practices

### Testing Workflow

1. **Start with Health Check**: Verify API is running
2. **Create Test Users**: Add a few users to work with
3. **Test Retrieval**: Get users by ID and list them
4. **Test Updates**: Modify user information
5. **Test Deletion**: Remove users (be careful!)

### Common Testing Scenarios

**Scenario 1: Full CRUD Operations**
```
1. Create user "Alice"
2. Get Alice by ID
3. List all users (verify Alice is there)
4. Update Alice's email
5. Delete Alice
6. Try to get Alice (should return 404)
```

**Scenario 2: Pagination Testing**
```
1. Create 25 users
2. List first 10 (limit=10, offset=0)
3. List next 10 (limit=10, offset=10)
4. List last 5 (limit=10, offset=20)
```

**Scenario 3: Error Handling**
```
1. Try to get user with ID 9999 (should fail)
2. Try to create user with duplicate email (should fail)
3. Try to update non-existent user (should fail)
4. Try to delete non-existent user (should fail)
```

## Response Codes

The response viewer shows status badges:

| Status Code | Badge Color | Meaning |
|-------------|-------------|---------|
| 200 | Green | Success - OK |
| 201 | Green | Success - Created |
| 204 | Green | Success - No Content |
| 400 | Red | Error - Bad Request |
| 404 | Red | Error - Not Found |
| 409 | Red | Error - Conflict (duplicate) |
| 500 | Red | Error - Internal Server Error |

## Keyboard Shortcuts

- **Enter**: Submit focused form
- **Tab**: Navigate between fields
- **Esc**: Close confirmation dialogs

## Browser Compatibility

The testing page works in all modern browsers:
- ✅ Chrome/Edge (latest)
- ✅ Firefox (latest)
- ✅ Safari (latest)
- ✅ Opera (latest)

**Note**: Internet Explorer is not supported.

## Troubleshooting

### Page Doesn't Load

**Problem**: Browser shows "Can't reach this page"

**Solution**:
1. Verify server is running: Check terminal for "Server starting on port 8080"
2. Check port: Ensure nothing else is using port 8080
3. Check URL: Should be `http://localhost:8080` (not https)

### API Calls Fail

**Problem**: All requests show errors

**Solution**:
1. Check server logs for errors
2. Verify database connection
3. Check CORS settings (if using from different domain)
4. Clear browser cache and reload page

### Response Not Displayed

**Problem**: No response shown after clicking button

**Solution**:
1. Check browser console for JavaScript errors (F12)
2. Verify network requests in browser dev tools
3. Try refreshing the page

### Database Errors

**Problem**: "Database is locked" or connection errors

**Solution**:
1. Stop any other instances of the application
2. Remove lock files: `rm *.db-shm *.db-wal`
3. Restart the server

## Advanced Usage

### Testing from Multiple Tabs

You can open the testing page in multiple browser tabs to:
- Test concurrent operations
- Simulate multiple users
- Verify data consistency

### Copying Request/Response

1. Click in the response viewer
2. Select the JSON you want to copy
3. Use Ctrl+C (Cmd+C on Mac) to copy

### Custom API Base URL

If running on a different port or host, the page automatically adapts. The JavaScript uses `window.location.origin` to determine the API base URL.

### Saving Test Data

To export test data:
1. Use List Users endpoint
2. Copy the JSON response
3. Save to a file for later use

## Integration with Development Workflow

### During Development

1. Make code changes
2. Restart server
3. Refresh testing page
4. Test your changes immediately

### For Demonstrations

1. Open testing page in fullscreen
2. Create sample data
3. Demonstrate CRUD operations live
4. Show real-time API responses

### For Documentation

1. Screenshot the testing page
2. Show request/response examples
3. Demonstrate error handling
4. Create video tutorials

## Security Notes

⚠️ **Important Security Considerations**:

- This testing page is intended for **development use only**
- Do not expose this page in production without authentication
- The page has full access to all API endpoints
- No rate limiting is applied to requests from the testing page

### Production Deployment

If deploying to production:

1. **Remove the testing page**:
```go
// Comment out or remove this line in cmd/api/main.go
// fs := http.FileServer(http.Dir("web/static"))
// mux.Handle("/", fs)
```

2. **Or add authentication**:
```go
mux.Handle("/", authMiddleware(fs))
```

3. **Or restrict by environment**:
```go
if cfg.Environment == "development" {
    fs := http.FileServer(http.Dir("web/static"))
    mux.Handle("/", fs)
}
```

## Additional Resources

- [API Examples](API_EXAMPLES.md) - Code examples in multiple languages
- [Database Setup](DATABASE_SETUP.md) - Database configuration guide
- [Troubleshooting](TROUBLESHOOTING.md) - Common issues and solutions
- [Main README](../README.md) - Project overview

## Feedback

The testing page is a simple tool to help with development. If you need additional features or find bugs, please:
1. Check existing issues
2. Create a new issue with details
3. Contribute improvements via pull request

Enjoy testing your API! 🚀
