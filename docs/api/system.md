# API Reference: System & Utilities

Health monitoring, system metrics, and tracking redirects.

## 🛠 Endpoints

| Endpoint | Method | Auth | Description |
| :--- | :--- | :--- | :--- |
| `/api/health` | `GET` | ❌ | Check system availability and dependency health. |
| `/api/metrics`| `GET` | ❌ | Prometheus scraping endpoint. |
| `/t/{id}` | `GET` | ❌ | Tracking redirect for customer order status. |

## 📖 Endpoint Details

### Health Check
`GET /api/health`

Used by monitoring systems (e.g., UptimeRobot, Kubernetes liveness probes) to ensure the API is responsive.

**Sample Response:**
```json
{
  "status": "ok",
  "message": "mi-tech API is running"
}
```

### Tracking Redirects
`GET /t/{order_id}`

A short URL for customers to track their parcels. This service automatically resolves the correct carrier tracking page or falls back to the store homepage.

**Logic:**
1. Accepts an internal ID or external Order Number (e.g., `#1234`).
2. Fetches the `tracking_url` from the database.
3. Performs a `307 Temporary Redirect` to the destination.
4. Fallback: `https://millennialperfumer.com`.

---
> [!NOTE]
> The `/t/` prefix is optimized for inclusion in SMS and WhatsApp automation messages to maximize delivery success and minimize character count.
