### 🆕 Proposed Ideas

### [IDE-001] Automated Feedback-Driven VIP Upsell & Re-Engagement Loop
* **Added On**: 2026-06-12
* **Target Audience**: Store Admins & End Customers
* **3L Growth Vector**: Increase Purchase Frequency (Customer Lifetime Value - LTV)
* **Customer Value Proposition**: Customers feel valued after their purchase. Highly satisfied customers receive an exclusive "VIP" reward for their positive feedback, while dissatisfied customers get immediate, proactive support resolution.
* **Business Growth & Profit Impact**: Customers who leave a 4 or 5-star rating are the most likely to buy again. By automatically sending a personalized discount code to these customers via WhatsApp immediately after a positive review, we capitalize on their peak satisfaction to drive a repeat purchase, directly boosting LTV. Additionally, instantly flagging 1 or 2-star reviews for admin intervention reduces returns, chargebacks, and negative word-of-mouth.
* **Technical Complexity**: Medium
* **Description**: Implement an event-driven automation hook triggered when a new `CustomerFeedback` record is created.
  1. **Positive Flow (Rating 4-5):** The system automatically interfaces with the Shopify API to generate a unique, time-limited discount code. It then queues a personalized WhatsApp template message (e.g., "Thanks for your 5-star review! Here is 15% off your next order: [CODE]") to the `CustomerPhone`.
  2. **Negative Flow (Rating 1-2):** The system flags the `OrderID` in a new "Urgent Escalations" widget on the admin dashboard, prompting immediate manual follow-up to save the customer relationship.
  3. **Tracking:** Add a tracking column to measure how many generated codes are redeemed to calculate the exact ROI of the automation.
