# Workflow: Order Synchronization

This document explains the technical flow of synchronizing order data from Shopify into the GST Invoice Manager.

## 🧭 Overview

Order synchronization ensures that all sales data, customer information, and fulfillment statuses are accurately reflected in the local database for GST reporting and automated communication.

## ⚡ Sync Triggers

1.  **Manual Sync**: Triggered via `/api/shopify/sync`. Fetches historical data using the Shopify GraphQL Admin API.
2.  **Real-time Webhooks**: Triggered by Shopify events (`orders/create`, `orders/fulfilled`, etc.). Uses the Shopify REST Webhook format.

## 🛠 Technical Flow

### 1. Data Ingestion
- **GraphQL (Manual)**: Uses `BulkOperation` or paginated queries to fetch `Order` nodes.
- **REST (Webhook)**: Receives a JSON payload. Validates authenticity using `X-Shopify-Hmac-Sha256`.

### 2. Mapping & Transformation
The `internal/mapper` package converts Shopify DTOs into internal `entity.Order` objects:
- **Phone Normalization**: Converts various phone formats to a standard E.164-like format for WhatsApp.
- **Status Mapping**:
    - `DisplayFinancialStatus: PAID` -> `status: paid`.
    - `DisplayFulfillmentStatus: FULFILLED` -> `status: fulfilled`.
    - `CancelledAt` present -> `status: CANCELLED`.
- **Source Identification**: Identifies if the order originated from `Shopify Online`, `Amazon`, or `POS`.

### 3. Persistence
- **Batch Upsert**: Uses GORM's `OnConflict` logic to insert or update orders and their line items.
- **Duplicate Prevention**: Order ID (`external_order_id`) serves as the unique constraint.

### 4. Downstream Side Effects
- **Customer Profiles**: Updates `total_spent` and `total_orders` in the `customers` table.
- **Automation Triggers**: Passes the event topic (e.g., `orders/paid`) to the `WebhookMappingService` to evaluate if a WhatsApp message should be sent.

---
> [!NOTE]
> The system treats Shopify as the "Source of Truth" for financial totals. If a discrepancy exists, a manual **Reset and Sync** will overwrite local data with Shopify's state.
