### 🆕 Proposed Ideas

### [IDE-001] Automated WhatsApp Review-to-Discount Loop
* **Added On**: 2024-06-14
* **Target Audience**: End Customers
* **3L Growth Vector**: Increase Purchase Frequency (LTV Boost)
* **Customer Value Proposition**: Customers are instantly rewarded for their positive feedback with a personalized, exclusive discount on their next purchase, creating a delightful post-purchase experience and building brand loyalty.
* **Business Growth & Profit Impact**: By systematically converting 4/5-star reviewers into repeat buyers via an automated WhatsApp trigger, we bypass customer acquisition costs (CAC). This zero-CAC channel directly drives higher net profit margins and increases Customer Lifetime Value (LTV), which is critical for scaling from 1L to 3L/month sustainably.
* **Technical Complexity**: Medium
* **Description**: Integrate the `feedback` module with the `communication` (WhatsApp/SMM hub) and automation planner. When a customer submits a positive review (>3 stars), an event is dispatched to trigger an automated workflow. The workflow will programmatically generate a dynamic, single-use Shopify discount code using the Shopify API and send a targeted WhatsApp message via the SMM hub, thanking the customer and offering a high-margin cross-sell or bundle for their next purchase.

### [IDE-002] Abandoned Checkout Sync & WhatsApp Recovery
* **Added On**: 2026-06-14
* **Target Audience**: Store Admins, End Customers
* **3L Growth Vector**: Boost Traffic-to-Customer Conversion
* **Customer Value Proposition**: Provides immediate assistance to customers who left items in their cart, answering potential questions and offering personalized incentives (like a small discount or free shipping) to help them complete their purchase seamlessly via WhatsApp.
* **Business Growth & Profit Impact**: Recovering abandoned carts is one of the highest leverage ways to increase conversion rates and AOV. By capturing draft orders from Shopify and directly pushing them into the MI-Tech SMM hub, we can implement an automated or semi-automated WhatsApp sequence. Even a 10-15% recovery rate on abandoned carts will directly and significantly contribute to hitting the 3 Lakhs/month revenue target with very little additional marginal cost.
* **Technical Complexity**: Medium
* **Description**: Implement a new backend module that listens for Shopify 'checkout/update' and 'checkout/create' webhooks (or periodically polls draft orders). The system will map these to a new "AbandonedCart" entity. Integrate this with the SMM hub to automatically dispatch a WhatsApp template message (e.g., "Hi [Name], you left [Item] in your cart. Need help?") 30-60 minutes after abandonment, complete with a direct checkout link.

### [IDE-003] VIP Customer Segmentation & Automated Retention Loop
* **Added On**: 2024-06-17
* **Target Audience**: Store Admins, End Customers
* **3L Growth Vector**: Increase Purchase Frequency (LTV Boost)
* **Customer Value Proposition**: High-value customers feel recognized and rewarded. They receive exclusive early access, targeted bundles, or special gifts via WhatsApp, cementing their loyalty and enhancing their premium brand experience.
* **Business Growth & Profit Impact**: By identifying and isolating the top 10% of customers (VIPs based on `TotalSpent` and `TotalOrders`), we can run highly targeted retention campaigns via the SMM hub. VIPs are significantly more likely to convert on higher-margin upsells and repeat purchases without relying on paid ads, driving net profit margins and accelerating the path to 3L/month.
* **Technical Complexity**: Medium
* **Description**: Implement a scheduled background job in the backend that scans the `customers` table for individuals exceeding defined thresholds for `TotalSpent` and `TotalOrders`. These customers are automatically tagged as 'VIP'. Integrate this tag with the `communication` (SMM) module to allow admins to broadcast exclusive WhatsApp templates (e.g., new product drops, personalized clone recommendations) specifically to this high-value segment.

### [IDE-005] High-Margin Stockout Risk Alerting Pipeline
* **Added On**: 2024-06-18
* **Target Audience**: Store Admins, Operations Leads
* **3L Growth Vector**: Reduce Operational & Loss Leakage
* **Customer Value Proposition**: Ensures top-selling and high-margin products remain continuously available, preventing customer frustration from out-of-stock experiences and maintaining trust in order fulfillment.
* **Business Growth & Profit Impact**: Stockouts on high-margin items represent direct, unrecoverable revenue leakage. By identifying top-performing SKUs (via sales velocity metrics from the order/report repository) and correlating them with current inventory levels, the system can proactively alert admins 7-14 days before a stockout occurs. This prevents loss of sales momentum on hero products, directly safeguarding the revenue trajectory needed to hit 3L/month while optimizing inventory turnover and capital allocation.
* **Technical Complexity**: Low
* **Description**: Implement an automated weekly (or daily) background task (`planner/cron`) that queries the `report_repository` to determine the top 20% of items by sales volume and margin over the last 30 days. It then checks the `inventory` module for their `CurrentStock`. If the projected run-rate suggests a stockout within 14 days, it triggers a notification via the `communication` hub to the store admin's WhatsApp/Dashboard, providing an actionable restocking manifest.

### [IDE-006] Automated GST Reconciliation & Mismatch Alerting
* **Added On**: 2024-06-21
* **Target Audience**: Store Admins, Operations Leads, Accounting
* **3L Growth Vector**: Reduce Operational & Loss Leakage
* **Customer Value Proposition**: Ensures accurate tax collection and invoicing, providing customers with compliant and error-free tax invoices which builds trust and professional credibility.
* **Business Growth & Profit Impact**: Manually reconciling GST across synced Shopify/Amazon orders against internal inventory tax mappings is highly error-prone. Mismatched HSN codes or tax discrepancies lead to compliance fines and margin leakage during tax filings. By automating this matching pipeline to flag discrepancies immediately upon order sync, we prevent compounding errors and save significant accounting hours, directly preserving net profit margins as order volume scales to 3 Lakhs/month.
* **Technical Complexity**: Medium
* **Description**: Create a background job or event listener in the `gst` module that triggers whenever an order is synced from Shopify/Amazon. It will compare the tax amount and rates reported by the sales channel against the expected tax calculated using the MI-Tech `inventory` module's HSN codes and pricing. Discrepancies exceeding a defined tolerance (e.g., ₹1) will be flagged in a new "GST Reconciliation Alerts" dashboard view, and notifications will be sent to admins via the `communication` hub.
