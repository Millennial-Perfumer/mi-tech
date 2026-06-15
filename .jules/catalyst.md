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
