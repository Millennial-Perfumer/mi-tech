# API Reference: Customers

Manage customer profiles, imports, and CRM data.

## 🛠 Base Path: `/api/customers`

| Endpoint | Method | Auth | Description |
| :--- | :--- | :--- | :--- |
| `/api/customers` | `GET` | ✅ | List customers with advanced filtering and pagination. |
| `/api/customers` | `POST` | 🛡️ Admin | Create a new customer profile. |
| `/api/customers` | `DELETE` | 🛡️ Admin | Delete all customers from the database. |
| `/api/customers/{id}`| `PUT` | 🛡️ Admin | Update an existing customer profile. |
| `/api/customers/{id}`| `DELETE` | 🛡️ Admin | Delete a specific customer. |
| `/api/customers/import`| `POST` | 🛡️ Admin | Import customers from a CSV file. |
| `/api/customers/bulk-delete`| `POST` | 🛡️ Admin | Delete a specific list of customers by ID. |

## 📖 Endpoint Details

### List Customers
`GET /api/customers`

**Query Parameters:**
- `page` / `pageSize`: Pagination controls.
- `search`: Search by name, email, or phone.
- `min_spent` / `max_spent`: Filter by total spend.
- `min_orders`: Filter by order count.
- `city` / `state`: Filter by location.
- `sortBy` / `sortOrder`: e.g., `total_spent`, `desc`.

**Sample Response:**
```json
{
  "customers": [...],
  "total": 1250
}
```

### Import CSV
`POST /api/customers/import`

Expects `multipart/form-data`.

**Form Fields:**
- `file`: The `.csv` file.
- `source_id`: The ID of the source (e.g., "Shopify").

**CSV Header Requirements:**
- `first_name`, `last_name`, `email`, `phone`, `city`, `state`, `zip`, `total_spent`, `total_orders`.

### Create/Update Customer
`POST/PUT /api/customers`

**Request Body:**
```json
{
  "first_name": "John",
  "last_name": "Doe",
  "email": "john@example.com",
  "phone": "9876543210",
  "city": "Mumbai",
  "state": "Maharashtra",
  "sync_to_shopify": false
}
```

### Bulk Delete
`POST /api/customers/bulk-delete`

**Request Body:**
```json
{
  "ids": [101, 102, 105]
}
```

---
> [!TIP]
> Use the `sync_to_shopify: true` flag during creation or update to automatically push changes to the connected Shopify store.
