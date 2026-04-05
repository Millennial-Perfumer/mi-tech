# GST Invoice Manager - Documentation Index

Welcome to the technical documentation for the GST Invoice Manager. This repository serves as the single source of truth for the system's API, business logic, and operational workflows.

## 📖 Table of Contents

### 1. [API Reference](api/index.md)
Detailed documentation for all backend endpoints, categorized by module.
- [Authentication](api/auth.md) - Login, 2FA, and session validation.
- [Users](api/users.md) - Admin-level user management.
- [Customers](api/customers.md) - CRM and CSV import.
- [Orders](api/orders.md) - Order management and tracking.
- [Shopify Sync](api/sync.md) - Manual and automated order synchronization.
- [Metrics & Dashboard](api/dashboard.md) - Real-time business insights.
- [Reports](api/reports.md) - GST, State-wise, and HSN-wise summaries.
- [WhatsApp Automation](api/automation.md) - Template and trigger management.
- [Meta Marketing](api/marketing.md) - Ads performance and insights.
- [Webhooks](api/webhooks.md) - Shopify and WhatsApp ingestion.
- [Settings & Config](api/settings.md) - System-wide configurations.

### 2. [Business Logic & Workflows](workflows/index.md)
Deep dives into the core algorithms and multi-step processes.
- [Order Sync Logic](workflows/order_sync.md) - How Shopify data is mapped and stored.
- [GST Calculation Strategy](workflows/gst_calculation.md) - Tax distribution and inclusive pricing logic.
- [WhatsApp Automation Flow](workflows/whatsapp_automation.md) - Event-driven messaging orchestration.

### 3. [Developer Guides](../GEMINI.md)
- [Project Architecture](../architecture.md)
- [Maintenance Guidelines](../GEMINI.md#documentation-maintenance)

---

> [!TIP]
> **Swagger UI**: For interactive API exploration, start the backend and visit `/swagger/index.html`.
