# API Reference: Orders

Comprehensive documentation for managing orders via the REST API.

## 🛠 Base Path: `/api/orders`

| Endpoint | Method | Auth | Description |
| :--- | :--- | :--- | :--- |
| `/api/orders` | `GET` | ✅ | List orders with pagination and filters. |
| `/api/orders` | `GET?id=...`| ✅ | Get details for a specific order. |
| `/api/orders` | `PUT?id=...`| 🛡️ Admin | Update order and sync with Shopify. |
| `/api/orders/status` | `PUT` | ✅ | Update internal workflow status. |
| `/api/orders/invoice`| `GET?id=...`| ✅ | Stream/Download GST invoice PDF. |
| `/api/sources` | `GET` | ✅ | List all order sources (Shopify, Amazon, etc.) |

## 📖 Endpoint Details

### List Orders
`GET /api/orders`

**Query Parameters:**
- `start_date` / `end_date`: Filter by creation date (YYYY-MM-DD).
- `page` / `limit`: Pagination controls.
- `search`: Filter by order number or customer name.
- `financial_status`: e.g., "paid", "pending".

**Sample Response:**
```json
{
  "success": true,
  "orders": [...],
  "total_count": 142,
  "page": 1
}
```

### Update Order Status
`PUT /api/orders/status?id={id}`

**Request Body:**
```json
{
  "status": "Ready to Ship"
}
```

### Generate Invoice
`GET /api/orders/invoice?id={id}`

Returns a binary stream with `Content-Type: application/pdf`.
- **Filename Format**: `invoice-{order_number}.pdf`

---
> [!NOTE]
> Authentication is JWT-based. Include the token in the `Authorization` header.
