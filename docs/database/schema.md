# Database Schema Reference

This document provides a detailed reference for the PostgreSQL database schema used by the GST Invoice Manager.

## 📊 Entity Relationship Diagram

```mermaid
erDiagram
    sources ||--o{ orders : "provides"
    orders ||--o{ order_line_items : "contains"
    orders ||--o{ webhook_events : "logged in"
    customers ||--o{ orders : "places"
    automation_templates ||--o{ automation_triggers : "used by"
    automation_templates ||--o{ automation_messages : "sent as"
    automation_triggers ||--o{ automation_messages : "triggers"

    sources {
        string id PK
        string name
        boolean enabled
        timestamp created_at
    }

    orders {
        string id PK
        string source_id FK
        string external_order_id
        string order_number
        decimal total_price
        timestamp created_at
        timestamp updated_at
        string customer_name
        string status
        string financial_status
        string fulfillment_status
    }

    order_line_items {
        string id PK
        string order_id FK
        string title
        string sku
        string hs_code
        integer quantity
        decimal price
        decimal discount
    }

    customers {
        int64 id PK
        string phone_number UK
        string first_name
        string last_name
        string email
        decimal total_spent
        int total_orders
    }

    users {
        uint id PK
        string username UK
        string role
        boolean two_factor_enabled
    }
```

## 🗄️ Core Tables

### `orders`
Primary storage for transaction data.
- `id`: Internal UUID/String identifier.
- `external_order_id`: ID from the source platform (e.g., Shopify ID).
- `order_number`: Human-readable number (e.g., #1001).
- `status`: Internal workflow status.

### `order_line_items`
Individual items within an order.
- `hs_code`: Harmonized System code used for GST calculation.
- `order_discount`: Proportionate discount applied to this item from order-level discounts.

### `customers`
Aggregated customer profiles.
- Identifies customers primarily by `phone_number` for WhatsApp continuity.

### `automation_templates`
WhatsApp message templates approved by Meta.
- `body`: The message text with `{{n}}` placeholders.
- `variable_mappings`: JSON mapping placeholders to order/customer fields.

### `app_configs`
Centralized encrypted storage for API keys and environment-specific settings.

---
> [!NOTE]
> Schema is derived from `backend/internal/database/migrations` and verified against GORM tags in `backend/internal/entity`.
