# Workflow: GST Calculation

This document details how GST (Goods and Services Tax) is calculated and reported within the Mi-Tech GST Invoice Manager.

## 🧭 Principles

GST is calculated at a fixed rate of **18% inclusive** for all products and charges (including shipping). The system automatically determines the specific tax component (**IGST** vs. **CGST+SGST**) based on the location of the customer.

## 🛠 Calculation Logic

### 1. Taxable Value
The system assumes the `total_price` from Shopify is inclusive of 18% GST.
- **Taxable Value Formula**: `Total Price / 1.18`
- **Total Tax Formula**: `Total Price - Taxable Value`

### 2. State-Based Distribution (Place of Supply)

The "Place of Supply" is determined by the `customer_state` field.

| Condition | Tax Components | Rates |
| :--- | :--- | :--- |
| **Intra-State** (Customer in Tamil Nadu) | **CGST** (9%) + **SGST** (9%) | Split 50/50 |
| **Inter-State** (Customer outside Tamil Nadu) | **IGST** (18%) | Full 100% |

> [!NOTE]
> The system identifies Tamil Nadu by checking for variations like `tamil nadu`, `tn`, and `tamilnadu` (case-insensitive).

### 3. Line-Item Level Logic
For accurate HSN-wise reporting, the overall order GST must be distributed across line items.
- **Proportional Share**: To handle order-level discounts (e.g., promo codes), the GST for each item is calculated proportionally based on its share of the total order value.
- **Example**: If an item represents 20% of the order's total value, it is assigned 20% of the order's total calculated GST.

### 4. HSN Mapping
- **Default HSN**: `33029019` (Perfumes/Essential Oils) is applied if no HS Code is found in the Shopify product data.
- **Sync**: The system attempts to fetch the `HarmonizedSystemCode` from Shopify's `InventoryItem` via GraphQL.

## 📊 Reporting Categories

- **B2C Small**: Inter-State supplies with invoice value up to ₹2,50,000/-.
- **B2C Large**: Inter-State supplies with invoice value more than ₹2,50,000/-.
- **B2B**: Supplies to registered GSTIN holders (linked via customer metadata).

---
> [!IMPORTANT]
> All GST calculations use the `ROUND(value, 2)` function to ensure consistency with financial standards and avoid floating-point discrepancies in report summaries.
