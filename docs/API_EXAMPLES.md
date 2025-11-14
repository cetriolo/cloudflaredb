# API Usage Examples

This guide provides practical examples for using the CloudflareDB API.

## Table of Contents

- [Getting Started](#getting-started)
- [cURL Examples](#curl-examples)
- [Go Client Examples](#go-client-examples)
- [JavaScript/TypeScript Examples](#javascripttypescript-examples)
- [Python Examples](#python-examples)
- [Postman Collection](#postman-collection)

## Getting Started

Make sure the API server is running:

```bash
go run cmd/api/main.go
```

The API will be available at `http://localhost:8080`.

## cURL Examples

### Health Check

```bash
curl http://localhost:8080/health
```

**Response:**
```json
{"status":"healthy"}
```

### Create a User

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "external_id": "user_abc123"
  }'
```

**Response:** (201 Created)
```json
{
  "id": 1,
  "external_id": "user_abc123",
  "created_at": "2025-11-13T10:00:00Z",
  "updated_at": "2025-11-13T10:00:00Z"
}
```

### Get a User by ID

```bash
curl http://localhost:8080/users/1
```

**Response:** (200 OK)
```json
{
  "id": 1,
  "external_id": "user_abc123",
  "created_at": "2025-11-13T10:00:00Z",
  "updated_at": "2025-11-13T10:00:00Z"
}
```

### List Users

```bash
# Get first 10 users
curl http://localhost:8080/users

# With pagination
curl "http://localhost:8080/users?limit=5&offset=10"
```

**Response:** (200 OK)
```json
[
  {
    "id": 1,
    "external_id": "user_abc123",
    "created_at": "2025-11-13T10:00:00Z",
    "updated_at": "2025-11-13T10:00:00Z"
  },
  {
    "id": 2,
    "external_id": "user_def456",
    "created_at": "2025-11-13T10:05:00Z",
    "updated_at": "2025-11-13T10:05:00Z"
  }
]
```

### Update a User

```bash
curl -X PUT http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{
    "external_id": "user_xyz789"
  }'
```

**Response:** (200 OK)
```json
{
  "id": 1,
  "external_id": "user_xyz789",
  "created_at": "2025-11-13T10:00:00Z",
  "updated_at": "2025-11-13T10:30:00Z"
}
```

### Delete a User

```bash
curl -X DELETE http://localhost:8080/users/1
```

**Response:** (204 No Content)

## Go Client Examples

### Basic Client

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const apiURL = "http://localhost:8080"

type User struct {
	ID         int64  `json:"id"`
	ExternalID string `json:"external_id"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type CreateUserRequest struct {
	ExternalID string `json:"external_id"`
}

func main() {
	// Create a user
	user, err := createUser("user_alice_001")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created user: %+v\n", user)

	// Get user by ID
	user, err = getUser(user.ID)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Retrieved user: %+v\n", user)

	// List users
	users, err := listUsers(10, 0)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found %d users\n", len(users))

	// Update user
	user.ExternalID = "user_alice_updated"
	user, err = updateUser(user.ID, user)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Updated user: %+v\n", user)

	// Delete user
	err = deleteUser(user.ID)
	if err != nil {
		panic(err)
	}
	fmt.Println("User deleted")
}

func createUser(externalID string) (*User, error) {
	reqBody := CreateUserRequest{
		ExternalID: externalID,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(apiURL+"/users", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func getUser(id int64) (*User, error) {
	resp, err := http.Get(fmt.Sprintf("%s/users/%d", apiURL, id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func listUsers(limit, offset int) ([]*User, error) {
	url := fmt.Sprintf("%s/users?limit=%d&offset=%d", apiURL, limit, offset)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var users []*User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, err
	}

	return users, nil
}

func updateUser(id int64, user *User) (*User, error) {
	jsonData, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut,
		fmt.Sprintf("%s/users/%d", apiURL, id),
		bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var updatedUser User
	if err := json.NewDecoder(resp.Body).Decode(&updatedUser); err != nil {
		return nil, err
	}

	return &updatedUser, nil
}

func deleteUser(id int64) error {
	req, err := http.NewRequest(http.MethodDelete,
		fmt.Sprintf("%s/users/%d", apiURL, id), nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
```

## JavaScript/TypeScript Examples

### Using Fetch API

```javascript
const API_URL = 'http://localhost:8080';

// Create a user
async function createUser(externalId) {
  const response = await fetch(`${API_URL}/users`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ external_id: externalId }),
  });

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }

  return await response.json();
}

// Get a user
async function getUser(id) {
  const response = await fetch(`${API_URL}/users/${id}`);

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }

  return await response.json();
}

// List users
async function listUsers(limit = 10, offset = 0) {
  const response = await fetch(
    `${API_URL}/users?limit=${limit}&offset=${offset}`
  );

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }

  return await response.json();
}

// Update a user
async function updateUser(id, updates) {
  const response = await fetch(`${API_URL}/users/${id}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(updates),
  });

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }

  return await response.json();
}

// Delete a user
async function deleteUser(id) {
  const response = await fetch(`${API_URL}/users/${id}`, {
    method: 'DELETE',
  });

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }

  return true;
}

// Usage example
(async () => {
  try {
    // Create
    const user = await createUser('user_bob_001');
    console.log('Created:', user);

    // Read
    const fetchedUser = await getUser(user.id);
    console.log('Fetched:', fetchedUser);

    // Update
    const updated = await updateUser(user.id, { external_id: 'user_bob_updated' });
    console.log('Updated:', updated);

    // List
    const users = await listUsers();
    console.log('All users:', users);

    // Delete
    await deleteUser(user.id);
    console.log('Deleted user');
  } catch (error) {
    console.error('Error:', error);
  }
})();
```

### TypeScript with Types

```typescript
interface User {
  id: number;
  external_id: string;
  created_at: string;
  updated_at: string;
}

interface CreateUserRequest {
  external_id: string;
}

interface UpdateUserRequest {
  external_id?: string;
}

class UserAPIClient {
  private baseURL: string;

  constructor(baseURL: string = 'http://localhost:8080') {
    this.baseURL = baseURL;
  }

  async createUser(request: CreateUserRequest): Promise<User> {
    const response = await fetch(`${this.baseURL}/users`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      throw new Error(`Failed to create user: ${response.statusText}`);
    }

    return await response.json();
  }

  async getUser(id: number): Promise<User> {
    const response = await fetch(`${this.baseURL}/users/${id}`);

    if (!response.ok) {
      throw new Error(`Failed to get user: ${response.statusText}`);
    }

    return await response.json();
  }

  async listUsers(limit: number = 10, offset: number = 0): Promise<User[]> {
    const response = await fetch(
      `${this.baseURL}/users?limit=${limit}&offset=${offset}`
    );

    if (!response.ok) {
      throw new Error(`Failed to list users: ${response.statusText}`);
    }

    return await response.json();
  }

  async updateUser(id: number, request: UpdateUserRequest): Promise<User> {
    const response = await fetch(`${this.baseURL}/users/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      throw new Error(`Failed to update user: ${response.statusText}`);
    }

    return await response.json();
  }

  async deleteUser(id: number): Promise<void> {
    const response = await fetch(`${this.baseURL}/users/${id}`, {
      method: 'DELETE',
    });

    if (!response.ok) {
      throw new Error(`Failed to delete user: ${response.statusText}`);
    }
  }
}

// Usage
const client = new UserAPIClient();

async function main() {
  const user = await client.createUser({
    external_id: 'user_typescript_001',
  });

  console.log('Created user:', user);
}

main().catch(console.error);
```

## Python Examples

### Using Requests Library

```python
import requests
from typing import Optional, List, Dict

API_URL = "http://localhost:8080"

class UserAPIClient:
    def __init__(self, base_url: str = API_URL):
        self.base_url = base_url

    def create_user(self, external_id: str) -> Dict:
        """Create a new user"""
        response = requests.post(
            f"{self.base_url}/users",
            json={"external_id": external_id}
        )
        response.raise_for_status()
        return response.json()

    def get_user(self, user_id: int) -> Dict:
        """Get a user by ID"""
        response = requests.get(f"{self.base_url}/users/{user_id}")
        response.raise_for_status()
        return response.json()

    def list_users(self, limit: int = 10, offset: int = 0) -> List[Dict]:
        """List users with pagination"""
        response = requests.get(
            f"{self.base_url}/users",
            params={"limit": limit, "offset": offset}
        )
        response.raise_for_status()
        return response.json()

    def update_user(
        self,
        user_id: int,
        external_id: Optional[str] = None
    ) -> Dict:
        """Update a user"""
        updates = {}
        if external_id is not None:
            updates["external_id"] = external_id

        response = requests.put(
            f"{self.base_url}/users/{user_id}",
            json=updates
        )
        response.raise_for_status()
        return response.json()

    def delete_user(self, user_id: int) -> None:
        """Delete a user"""
        response = requests.delete(f"{self.base_url}/users/{user_id}")
        response.raise_for_status()


# Usage example
if __name__ == "__main__":
    client = UserAPIClient()

    # Create a user
    user = client.create_user("user_python_001")
    print(f"Created user: {user}")

    # Get user
    fetched_user = client.get_user(user["id"])
    print(f"Fetched user: {fetched_user}")

    # List users
    users = client.list_users(limit=5)
    print(f"Found {len(users)} users")

    # Update user
    updated_user = client.update_user(user["id"], external_id="user_python_updated")
    print(f"Updated user: {updated_user}")

    # Delete user
    client.delete_user(user["id"])
    print("User deleted")
```

## Postman Collection

You can import this JSON into Postman for easy API testing:

```json
{
  "info": {
    "name": "CloudflareDB API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Health Check",
      "request": {
        "method": "GET",
        "header": [],
        "url": {
          "raw": "{{base_url}}/health",
          "host": ["{{base_url}}"],
          "path": ["health"]
        }
      }
    },
    {
      "name": "Create User",
      "request": {
        "method": "POST",
        "header": [
          {
            "key": "Content-Type",
            "value": "application/json"
          }
        ],
        "body": {
          "mode": "raw",
          "raw": "{\n  \"external_id\": \"user_test_001\"\n}"
        },
        "url": {
          "raw": "{{base_url}}/users",
          "host": ["{{base_url}}"],
          "path": ["users"]
        }
      }
    },
    {
      "name": "Get User",
      "request": {
        "method": "GET",
        "header": [],
        "url": {
          "raw": "{{base_url}}/users/{{user_id}}",
          "host": ["{{base_url}}"],
          "path": ["users", "{{user_id}}"]
        }
      }
    },
    {
      "name": "List Users",
      "request": {
        "method": "GET",
        "header": [],
        "url": {
          "raw": "{{base_url}}/users?limit=10&offset=0",
          "host": ["{{base_url}}"],
          "path": ["users"],
          "query": [
            {
              "key": "limit",
              "value": "10"
            },
            {
              "key": "offset",
              "value": "0"
            }
          ]
        }
      }
    },
    {
      "name": "Update User",
      "request": {
        "method": "PUT",
        "header": [
          {
            "key": "Content-Type",
            "value": "application/json"
          }
        ],
        "body": {
          "mode": "raw",
          "raw": "{\n  \"external_id\": \"user_updated_001\"\n}"
        },
        "url": {
          "raw": "{{base_url}}/users/{{user_id}}",
          "host": ["{{base_url}}"],
          "path": ["users", "{{user_id}}"]
        }
      }
    },
    {
      "name": "Delete User",
      "request": {
        "method": "DELETE",
        "header": [],
        "url": {
          "raw": "{{base_url}}/users/{{user_id}}",
          "host": ["{{base_url}}"],
          "path": ["users", "{{user_id}}"]
        }
      }
    }
  ],
  "variable": [
    {
      "key": "base_url",
      "value": "http://localhost:8080"
    },
    {
      "key": "user_id",
      "value": "1"
    }
  ]
}
```

Save this as `CloudflareDB_API.postman_collection.json` and import into Postman.

## Error Handling Examples

### Handling Errors in Go

```go
user, err := createUser("user_invalid")
if err != nil {
	// Check HTTP status code
	if strings.Contains(err.Error(), "400") {
		log.Println("Bad request - check input")
	} else if strings.Contains(err.Error(), "409") {
		log.Println("User already exists")
	} else {
		log.Printf("Unexpected error: %v", err)
	}
}
```

### Handling Errors in JavaScript

```javascript
try {
  const user = await createUser('user_invalid');
} catch (error) {
  if (error.message.includes('400')) {
    console.error('Invalid request');
  } else if (error.message.includes('404')) {
    console.error('User not found');
  } else {
    console.error('Unexpected error:', error);
  }
}
```

## Advanced Usage

### Bulk Operations

```bash
# Create multiple users
for i in {1..10}; do
  curl -X POST http://localhost:8080/users \
    -H "Content-Type: application/json" \
    -d "{\"external_id\":\"user_batch_${i}\"}"
done
```

### Rate Limiting Simulation

```bash
# Send requests in parallel
seq 1 100 | xargs -P 10 -I {} curl -s http://localhost:8080/health
```

## Next Steps

- Implement authentication/authorization
- Add more endpoints (e.g., search, filters)
- Set up monitoring and logging
- Deploy to production

For more information, see the [main README](../README.md) and [Database Setup Guide](DATABASE_SETUP.md).
