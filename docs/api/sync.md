# API Reference: Shopify Sync

Manual controls for synchronizing data between Shopify and the local database.

## 🛠 Base Path: `/api/shopify`

| Endpoint | Method | Auth | Description |
| :--- | :--- | :--- | :--- |
| `/api/shopify/sync` | `POST` | 🛡️ Admin | Fetch and sync orders within a specific date range. |
| `/api/shopify/reset`| `POST` | 🛡️ Admin | Wipe all local orders and perform a full re-sync. |

## 📖 Endpoint Details

### Sync Orders
`POST /api/shopify/sync`

Triggers a background synchronization task. You can specify the date range via query parameters or the request body.

**Query Parameters (Optional):**
- `start_date`: e.g., `2023-01-01`
- `end_date`: e.g., `2023-01-31`

**Request Body (Optional):**
```json
{
  "start_date": "2023-01-01",
  "end_date": "2023-01-31"
}
```

**Sample Response:**
```json
{
  "success": true,
  "message": "Sync completed successfully",
  "count": 42
}
```

### Reset and Sync
`POST /api/shopify/reset`

> [!CAUTION]
> This endpoint will delete all orders from the local database before starting a fresh sync from Shopify. Use this only if data is corrupted or out of sync.

**Sample Response:**
```json
{
  "success": true,
  "message": "Reset and Sync completed successfully",
  "count": 1500
}
```

---
> [!NOTE]
> Synchronizing large date ranges may take several minutes. The API will respond once the initial fetch is complete.
