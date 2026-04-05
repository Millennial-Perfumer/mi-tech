# API Reference: Settings & Configs

Manage application parameters, date ranges, and third-party API credentials.

## 🛠 Base Paths: `/api/settings` and `/api/configs`

| Endpoint | Method | Auth | Description |
| :--- | :--- | :--- | :--- |
| `/api/settings` | `GET` | ✅ | List all general application settings. |
| `/api/settings` | `PUT` | 🛡️ Admin | Update a specific application setting. |
| `/api/settings/date-range`| `GET` | ✅ | Get the current global date range for data views. |
| `/api/settings/date-range`| `PUT` | 🛡️ Admin | Update the global date range. |
| `/api/configs` | `GET` | ✅ | List all API configurations (keys/secrets masked). |
| `/api/configs` | `PUT` | 🛡️ Admin | Update a specific API configuration. |
| `/api/configs/reveal` | `POST` | 🛡️ Admin | Retrieve unmasked configuration values. |

## 📖 Endpoint Details

### Application Settings
`GET /api/settings`

Returns a flat object of key-value pairs representing application settings.

**Sample Response:**
```json
{
  "success": true,
  "settings": {
    "currency": "INR",
    "timezone": "Asia/Kolkata",
    "shop_name": "Millennial Perfumer"
  }
}
```

### Date Range Management
`GET /api/settings/date-range`

Returns the global date filter applied to the dashboard and reports.

**Sample Response:**
```json
{
  "success": true,
  "start_date": "2023-10-01",
  "end_date": "2023-10-31"
}
```

### Secure Configurations
`GET /api/configs`

Lists integration settings (Shopify, Meta, WhatsApp). For security, sensitive values are masked (e.g., `********`).

**Sample Response:**
```json
{
  "success": true,
  "configs": [
    { "key": "meta_api_token", "value": "********", "is_secret": true }
  ]
}
```

### Reveal Configs
`POST /api/configs/reveal`

To view the raw unmasked values of your API keys, you must provide your login password again.

**Request Body:**
```json
{
  "password": "yourpassword"
}
```

**Sample Response:**
```json
{
  "success": true,
  "configs": [
    { "key": "meta_api_token", "value": "EAAbp...", "is_secret": true }
  ]
}
```

---
> [!CAUTION]
> Updating configurations (like Shopify API secrets) may temporarily disrupt active integrations. Always verify credentials before saving.
