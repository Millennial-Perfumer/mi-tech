# API Reference: Users

Manage administrator and standard user accounts.

## 🛠 Base Path: `/api/users`

| Endpoint | Method | Auth | Description |
| :--- | :--- | :--- | :--- |
| `/api/users` | `GET` | 🛡️ Admin | Retrieve a list of all registered users. |
| `/api/users` | `POST` | 🛡️ Admin | Create a new user account. |

## 📖 Endpoint Details

### List Users
`GET /api/users`

**Sample Response (Success):**
```json
{
  "success": true,
  "users": [
    {
      "id": "...",
      "username": "admin",
      "role": "admin",
      "created_at": "..."
    }
  ]
}
```

### Create User
`POST /api/users`

**Request Body:**
```json
{
  "username": "newuser",
  "password": "securepassword",
  "role": "user"
}
```
> [!NOTE]
> Roles are restricted to either `user` or `admin`.

**Sample Response (Success):**
```json
{
  "success": true,
  "message": "user created successfully"
}
```

---
> [!IMPORTANT]
> User management endpoints are strictly for administrators. Standard users will receive a `403 Forbidden` response.
