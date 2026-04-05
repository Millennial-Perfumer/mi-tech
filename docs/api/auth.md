# API Reference: Authentication

Manage user sessions and security through JWT-based authentication.

## 🛠 Base Path: `/api/auth`

| Endpoint | Method | Auth | Description |
| :--- | :--- | :--- | :--- |
| `/api/auth/login` | `POST` | ❌ | Authenticate with username and password. |
| `/api/auth/verify-otp` | `POST` | ❌ | Verify OTP if account requires 2FA. |
| `/api/auth/verify` | `GET` | ✅ | Validate JWT and retrieve user roles. |

## 📖 Endpoint Details

### Login
`POST /api/auth/login`

**Request Body:**
```json
{
  "username": "admin",
  "password": "yourpassword"
}
```

**Sample Response (Success - No 2FA):**
```json
{
  "token": "eyJhbG...",
  "requires_2fa": false
}
```

**Sample Response (Requires 2FA):**
```json
{
  "requires_2fa": true
}
```

### Verify OTP
`POST /api/auth/verify-otp`

Used when the login response indicates `requires_2fa: true`.

**Request Body:**
```json
{
  "username": "admin",
  "otp": "123456"
}
```

**Sample Response:**
```json
{
  "token": "eyJhbG..."
}
```

### Verify Authentication
`GET /api/auth/verify`

Used for session validation and role-based access control (RBAC).

**Headers:**
- `Authorization: Bearer <token>`

**Sample Response (200 OK):**
- Headers set: `X-WEBAUTH-USER`, `X-WEBAUTH-ROLE`

---
> [!NOTE]
> Tokens are valid for 24 hours. The `X-WEBAUTH-ROLE` header will contain either `user` or `admin`.
