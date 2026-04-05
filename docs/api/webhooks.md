# API Reference: Webhooks

Real-time data ingestion and integration health monitoring.

## 🛠 Base Path: `/api/webhooks`

| Endpoint | Method | Auth | Description |
| :--- | :--- | :--- | :--- |
| `/api/webhooks/shopify`| `POST` | 🔐 HMAC | Ingest Shopify webhooks (Orders, Customers, Fulfillments). |
| `/api/webhook/status` | `GET` | ✅ | Check the status of the last received webhook. |

## 📖 Endpoint Details

### Shopify Webhook Receiver
`POST /api/webhooks/shopify`

This endpoint is designed to be called by Shopify. It performs HMAC validation and processes events asynchronously.

**Headers:**
- `X-Shopify-Topic`: e.g., `orders/create`, `orders/updated`.
- `X-Shopify-Hmac-Sha256`: Signature for security validation.
- `X-Shopify-Webhook-Id`: Unique delivery ID.

**Supported Topics:**
- `orders/*`: create, updated, paid, fulfilled, cancelled.
- `fulfillments/*`: create, update.
- `customers/*`: create, update, delete.

**Processing Logic:**
1. Validates HMAC signature against `SHOPIFY_WEBHOOK_SECRET`.
2. Responds with `200 OK` immediately to acknowledge receipt.
3. Maps topic to business logic (e.g., creating an order, updating delivery status).
4. **[WhatsApp Automation]**: Triggers messages based on the topic (e.g., "Order Dispatched").

### Webhook Health Status
`GET /api/webhook/status`

Use this to verify if webhooks are being received and processed correctly.

**Sample Response:**
```json
{
  "topic": "orders/fulfilled",
  "status": "success",
  "last_received": "2023-10-27T14:30:15Z"
}
```

---
> [!IMPORTANT]
> HMAC validation is **stricty enforced**. If the `SHOPIFY_WEBHOOK_SECRET` is missing in settings, the system will log a warning and allow all requests (not recommended for production).
