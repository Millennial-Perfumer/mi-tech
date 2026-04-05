# API Reference: Dashboard & Metrics

Analyze business performance, GST collections, and system health.

## 🛠 Base Path: `/api`

| Endpoint | Method | Auth | Description |
| :--- | :--- | :--- | :--- |
| `/api/dashboard/metrics`| `GET` | ✅ | Retrieve summary cards for the main dashboard. |
| `/api/metrics` | `GET` | ❌ | Prometheus-formatted metrics for system monitoring. |

## 📖 Endpoint Details

### Dashboard Summary
`GET /api/dashboard/metrics`

Returns aggregated financial and operational data.

**Query Parameters:**
- `start_date`: e.g., `2023-01-01`
- `end_date`: e.g., `2023-01-31`

**Sample Response:**
```json
{
  "success": true,
  "metrics": {
    "total_revenue": 154200.50,
    "total_invoices": 45,
    "total_gst_collected": 27756.00,
    "cgst_collected": 13878.00,
    "sgst_collected": 13878.00,
    "igst_collected": 0.00,
    "total_orders": 45,
    "cancelled_orders": 2,
    "fulfilled_orders": 40,
    "unfulfilled_orders": 3
  }
}
```

### Prometheus Metrics
`GET /api/metrics`

This endpoint exposes application-level metrics (request counts, latency, memory usage) in a format suitable for Prometheus scraping.

**Sample Output Fragment:**
```text
# HELP http_requests_total Total number of HTTP requests.
# TYPE http_requests_total counter
http_requests_total{code="200",method="GET",path="/api/orders"} 142
```

---
> [!NOTE]
> Dashboard metrics are calculated in real-time from the database. For historical reports, see the **[Reports](file:///Users/siddiqs_office/Documents/Personal%20Dev/GST%20Invoice%20Manager/docs/api/reports.md)** section.
