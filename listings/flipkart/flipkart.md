# Flipkart Perfume Listing: Complete Field Specifications & Default Values

This document contains the exact field specifications extracted directly from your active Flipkart Seller session for the **Perfume** vertical under **Millennial Perfumer**, with your confirmed business defaults applied.

---

## 1. Global Default Product Configuration (Base for all 82 Products)

| Field Name | Configured Base Value |
|:---|:---|
| **Brand** | `Millennial Perfumer` |
| **MRP** | `₹1299` |
| **Selling Price** | `₹499` |
| **Procurement SLA** | `1` or `2` Days *(Time taken to pack & dispatch order once received)* |
| **Stock** | `10` (or as per inventory) |
| **Package Dimensions** | Length: `12 cm` \| Breadth: `10 cm` \| Height: `19 cm` |
| **Package Weight** | `0.25 kg` (250g) |
| **HSN Code** | `33029019` |
| **Tax Code / GST Rate** | `GST_18` (18%) |
| **Fragrance Classification** | `Extrait De Parfum` |
| **Quantity / Volume** | `50 ml` |
| **Country of Origin** | `India` |
| **Manufacturer Details** | `Parfum Traders, No. 9/21, 1st floor, Sadiq Basha Nagar, 2nd Street, Virugambakkam, Chennai-600092, Contact: +917904769823, millennialperfumer.cc@gmail.com` |
| **Packer Details** | `Parfum Traders, No. 9/21, 1st floor, Sadiq Basha Nagar, 2nd Street, Virugambakkam, Chennai-600092, Contact: +917904769823, millennialperfumer.cc@gmail.com` |
| **Gas / Pocket Perfume** | Gas: `No` \| Pocket Perfume: `No` |
| **Maximum Shelf Life** | `36 Months` |

---

## 2. What is SLA (Procurement SLA)?

> **Procurement SLA (Service Level Agreement)** is the time in **DAYS** you take to pack the order and hand it over to Flipkart's logistics courier partner after a buyer places an order.
> - **Recommended value**: `1` or `2` Days.
> - **1 Day SLA** gives higher search visibility & badge (Fast Delivery).
> - **2 Days SLA** gives you more buffer time to pack and print shipping labels.

---

## 3. Product Catalog Schema & Image Mapping

- **Asset Root**: `/Users/siddiqs_office/Documents/Business/millennial perfumer/Product Photos/`
- **Total Folders**: 82 variants (`F001 - OCEAN DRIFT` to `F085 - THE ONE`)
- **Image Slots per Product**:
  1. `1.png` → **Front View (*)** (Mandatory)
  2. `2.png` → **Sales Package**
  3. `3.png` → **Size View**
  4. `4.png` → **Ingredients View**
  5. `5.png` - `8.png` → **Additional Views (5 to 8)**

---

## 4. Automation Workflow

1. **Master Catalog Generation**: `listings/products_catalog.json` contains the structured attributes for all 82 items.
2. **Batch Script Execution**: Automated CDP/Playwright script attaches to your open Chrome browser (`localhost:9222`), navigates to `Add Single Listing`, enters all fields, uploads images, generates the title, and saves/submits to QC.
