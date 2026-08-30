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

### [IDE-006] Automated GST Reconciliation & Mis-match Alerting
* **Added On**: 2024-06-20
* **Target Audience**: Store Admins, Operations Leads, Finance
* **3L Growth Vector**: Reduce Operational & Loss Leakage
* **Customer Value Proposition**: Ensures the brand remains fully compliant with local tax laws, minimizing the risk of audits, and ensuring customers are charged the correct tax rates on their purchases.
* **Business Growth & Profit Impact**: Uncaught GST discrepancies can lead to significant financial penalties or miscalculated profit margins, which directly hinders reaching and sustaining the 3L/month target. This automation actively monitors synchronized orders against local GST settings (especially via the `gst` module), flagging incorrect HSN codes, missing GST numbers on B2B invoices, or discrepancies between Shopify tax collection and MI-Tech reported tax. Catching these early prevents loss leakage and streamlines the end-of-month accounting process.
* **Technical Complexity**: Medium
* **Description**: Extend the `order` and `gst` domains by implementing a background reconciliation worker. When new orders are synchronized (via `webhook/order` or manual sync), the worker compares the order line items' tax lines and HSN codes against the canonical database configurations. If a mismatch is detected (e.g., Shopify charged 18% but the item HSN dictates 12%), it logs a `TaxDiscrepancy` event and pushes an alert to the `dashboard` and `communication` channels for manual review before invoice generation.

### [IDE-007] Predictive WhatsApp Replenishment Prompts
* **Added On**: 2024-06-25
* **Target Audience**: End Customers
* **3L Growth Vector**: Increase Purchase Frequency (Customer Lifetime Value - LTV)
* **Customer Value Proposition**: Customers never run out of their favorite consumable products (like perfumes). They receive a timely, personalized reminder with a quick one-click reorder link just as their supply is estimated to run out, creating a magical and frictionless convenience.
* **Business Growth & Profit Impact**: Dramatically boosts Customer Lifetime Value (LTV) and repeat purchase frequency. Assuming a typical 30-60 day consumption cycle for core products, automating this follow-up creates a passive, recurring revenue stream with zero acquisition cost. Stacking predictable repeat orders on top of new customer acquisition is a highly leveraged mechanism to accelerate from 1L to 3L/month while maximizing profit margins.
* **Technical Complexity**: Medium
* **Description**: Implement a new automated background task (`planner/cron`) that analyzes the `order` module for past purchases of consumable SKUs (defined by an 'expected_lifespan_days' attribute added to the `inventory` module). When a customer approaches their estimated depletion date, the system integrates with the `communication` (SMM) hub to dispatch a personalized WhatsApp template (e.g., "Running low on your favorite scent?") containing a pre-filled, one-click Shopify checkout link for instant replenishment.

### [IDE-008] High-Margin "Complete the Set" Checkout Upsell Widget
* **Added On**: 2024-06-25
* **Target Audience**: End Customers, Store Admins
* **3L Growth Vector**: Increase Average Order Value (AOV)
* **Customer Value Proposition**: Enhances the shopping experience by intuitively recommending perfectly matching accessories or smaller complimentary items (like a matching body lotion to a perfume) exactly when they are most excited to buy, saving them time searching.
* **Business Growth & Profit Impact**: By dynamically injecting a high-margin, low-cost upsell item into the checkout flow or post-purchase WhatsApp confirmation, we capture impulse buys. Even a 20% attach rate on a ₹500 high-margin add-on significantly bumps the AOV, directly accelerating the revenue run-rate toward the 3L/month target without acquiring new customers.
* **Technical Complexity**: Medium
* **Description**: Implement a recommendation engine leveraging the `order` module (`LineItem` tracking) and the `inventory` module (`InventoryItem.Price`). When a customer reaches the checkout step or receives a draft order link via the SMM hub, the system queries their current `LineItems`. Based on predefined `InventoryMapping` rules for hero products, it dynamically surfaces a 1-click upsell offer (e.g., a tester vial set) that automatically appends to the `Order` payload before final financial capture.

### [IDE-009] AI-Powered Perfume Clone Matchmaker via WhatsApp
* **Added On**: 2024-06-25
* **Target Audience**: End Customers, Affiliates
* **3L Growth Vector**: Boost Traffic-to-Customer Conversion
* **Customer Value Proposition**: Allows customers to text the name of a famous designer perfume to the brand's WhatsApp number and instantly receive an AI-recommended match for the closest affordable clone available in the MI-Tech inventory, complete with a one-click checkout link.
* **Business Growth & Profit Impact**: This highly interactive and shareable feature acts as a frictionless lead magnet. By capturing high-intent search traffic directly on WhatsApp, it reduces drop-off compared to a traditional website search, accelerates the path to purchase for first-time buyers, and organically increases conversion rates—a critical driver for reaching the 3L/month revenue target.
* **Technical Complexity**: Medium
* **Description**: Leverage the existing `ai` module (`query_guard.go`) and integrate it with the `communication` (WhatsApp/SMM) and `inventory` modules. When a customer sends a message like "Do you have a clone for Baccarat Rouge?", the SMM webhook handler processes the intent, queries the AI to match the designer scent profile against the local `InventoryItem` data (using tags or AI embeddings), and replies with the highest confidence match and a direct Shopify checkout URL.

### [IDE-010] Post-Purchase Automated Referral & Review Engine via WhatsApp
* **Added On**: 2024-07-28
* **Target Audience**: End Customers, Store Admins
* **3L Growth Vector**: Increase Purchase Frequency (Customer Lifetime Value - LTV)
* **Customer Value Proposition**: Delivers a highly personalized and frictionless post-purchase experience. Customers are easily able to leave reviews and refer friends directly through WhatsApp, earning exclusive discounts for future purchases without navigating away from their preferred messaging app.
* **Business Growth & Profit Impact**: By automating post-purchase feedback loops 7 days after delivery, we proactively capture high-intent positive sentiment. Rewarding this sentiment with a referral discount (e.g., "Give ₹500, Get ₹500") directly fuels viral, low-cost customer acquisition while simultaneously guaranteeing repeat purchases (boosting LTV) from the original buyer to redeem their reward. This creates a compounding growth engine essential for scaling to 3L/month.
* **Technical Complexity**: Medium
* **Description**: Implement a background cron job (`planner/cron`) that queries the `order` module for orders marked as "Delivered" 7 days prior. Integrate this with the `feedback` module to generate a unique review/referral tracking link. The `communication` (SMM) module then automatically dispatches a WhatsApp template (e.g., "How are you loving [Product]?") containing this link. Positive responses automatically trigger a follow-up template offering the referral discount code.
