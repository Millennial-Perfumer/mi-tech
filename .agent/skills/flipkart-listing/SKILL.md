---
name: flipkart-listing
description: Comprehensive workflow and automation engine for listing Millennial Perfumer products on Flipkart Seller Hub. Handles multi-tab attribute injection, image asset mapping (1.png to 8.png), HSN/Cess cleanliness, and QC submission.
---

# Flipkart Listing Skill

This skill provides the end-to-end specifications, field mappings, image handling, and automation scripts for listing catalog products on Flipkart Seller Hub.

---

## 1. Product Catalog & Asset Architecture

- **Catalog Database**: `listings/products_catalog_clean.json` (Structured specs for all 82 products)
- **Photo Directory Structure**:
  `/Users/siddiqs_office/Documents/Business/millennial perfumer/Product Photos/<SKU_CODE> - <PRODUCT_NAME>/`
- **Slot Image Mappings**:
  - `1.png` → **Front View (*)** (Mandatory)
  - `2.png` → **Sales Package**
  - `3.png` → **Size View**
  - `4.png` → **Ingredients View**
  - `5.png` - `8.png` → **Additional Views (Image 5 to 8)**

---

## 2. Standard Business Rules & Constants

| Attribute | Value / Rule | Description |
| :--- | :--- | :--- |
| **Vertical** | `perfume` | Flipkart vertical category |
| **Brand** | `Millennial Perfumer` | Approved Brand name |
| **HSN Code** | `33029019` | GST HSN for non-alcoholic / perfume compounds |
| **GST Tax Code** | `GST_18` | 18% GST Bracket |
| **Luxury Cess** | *Leave Untouched* | Do not trigger React array updates to avoid multi-row spawning |
| **Country of Origin** | `IN` | India |
| **Procurement Type** | `EXPRESS` | Express dispatch |
| **Procurement SLA** | `2` | 2 business days handling time |
| **Shipping Provider** | `FLIPKART` | Flipkart Logistics courier partner |
| **Manufacturing Date** | `2026-08-15` | Required valid manufacturing date |
| **Shelf Life** | `48 Months` (or 36) | Maximum shelf life |
| **Classification** | `Extrait De Parfum` | Perfume concentration |
| **Manufacturer Details**| `Millennial Perfumer, Plot No 42, OMR, Chennai - 600096, India` | Standard Mfr Address |
| **Packer Details** | `Parfum Traders, No. 9/21, 1st floor, Sadiq Basha Nagar, 2nd Street, Virugambakkam, Chennai-600092, Contact: +917904769823, millennialperfumer.cc@gmail.com` | Standard Packer Address |

---

## 3. Automation Flow & Execution

The parallel listing runner lives in `listings/flipkart/` and saves each listing **as a draft** (it does NOT auto-submit to QC — the operator reviews drafts and sends to QC manually).

```bash
# Dry-run plan (touches nothing)
python3 -m listings.flipkart.runner --dry-run

# List remaining products as drafts (2 parallel isolated tabs)
python3 -m listings.flipkart.runner --workers 2

# Chunk / selective / resume
python3 -m listings.flipkart.runner --workers 2 --limit 10
python3 -m listings.flipkart.runner --workers 2 --skus F009 F010
python3 -m listings.flipkart.runner --retry-failed
```

### Execution Steps (per product, one isolated tab):
1. **Connects via Chrome DevTools Protocol (CDP)** on port `9222` and opens a fresh tab.
2. **Navigates to Add Single Listing** and initializes a fresh draft request for `Millennial Perfumer` under `perfume`.
3. **Populates All Form Fields**:
   - *Price, Stock & Shipping:* SKU, MRP, SP, Stock, Dimensions, HSN, Tax, Country, Mfr/Packer, Date. Selects are set via the hidden `checkMarkOption_*` radios (display labels, e.g. `Active`, `Seller`, `express`, `GST_18`, `India`).
   - *Product Description:* Model Name, Quantity (`50` + `ml`), **Ideal For (from catalog — per-product Men/Women/Men & Women)**, Fragrance Classification, Fragrance Family.
   - *Additional Description:* Segments, Gas/Pocket/Organic (No), Shelf Life, Description, plus tag fields (Sales Package, Keywords, Key Features, Notes) committed with Enter.
4. **Compresses & uploads Images (1.png to 8.png)** via `sips` (1200px JPEG, ~150KB) into slot file inputs.
5. **Saves the draft** via **Save & Go Back** — never clicks Send to QC.
6. **Records progress** in `listings/flipkart/results.json` (resumable; `completed` / `failed` / `needs_review`).
   - Displays final verified tab counts.
