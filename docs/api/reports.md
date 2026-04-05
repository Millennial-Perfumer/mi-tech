# API Reference: Reports

Generate detailed GST-compliant reports and summaries.

## 🛠 Base Path: `/api/reports`

| Endpoint | Method | Auth | Description |
| :--- | :--- | :--- | :--- |
| `/api/reports/summary` | `GET` | ✅ | Overall GST summary (B2B, B2C, Exports). |
| `/api/reports/state-wise` | `GET` | ✅ | State-wise distribution of sales and taxes. |
| `/api/reports/hsn-wise` | `GET` | ✅ | HSN-wise summary of goods and services. |
| `/api/reports/documents-issued`| `GET` | ✅ | Sequence of invoices and documents issued. |

## 📖 Endpoint Details

### GST Summary
`GET /api/reports/summary`

**Query Parameters:**
- `start_date`, `end_date`: YYYY-MM-DD range.

**Sample Response:**
```json
{
  "success": true,
  "summary": {
    "total_taxable_value": 125400.00,
    "total_cgst": 11286.00,
    "total_sgst": 11286.00,
    "total_igst": 5400.00,
    "b2c_small": [...],
    "b2c_large": [...]
  }
}
```

### State-wise Summary
`GET /api/reports/state-wise`

**Sample Response:**
```json
{
  "success": true,
  "data": [
    {
      "state_name": "Maharashtra",
      "state_code": "27",
      "taxable_value": 50000.00,
      "cgst": 4500.00,
      "sgst": 4500.00
    }
  ]
}
```

### HSN-wise Summary
`GET /api/reports/hsn-wise`

**Sample Response:**
```json
{
  "success": true,
  "data": [
    {
      "hsn_code": "3304",
      "description": "Perfumes",
      "uqc": "PCS",
      "total_quantity": 150,
      "taxable_value": 75000.00
    }
  ]
}
```

### Documents Issued
`GET /api/reports/documents-issued`

Provides the range of invoice numbers issued during the period.

**Sample Response:**
```json
{
  "success": true,
  "data": {
    "from_serial": "INV/24-25/001",
    "to_serial": "INV/24-25/045",
    "total_number": 45,
    "cancelled_number": 2
  }
}
```

---
> [!NOTE]
> All report data is derived from the `orders` table. Ensure your Shopify sync is up-to-date before generating reports.
