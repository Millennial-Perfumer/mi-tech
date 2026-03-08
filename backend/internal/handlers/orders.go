package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"

	"shopify-gst-app/internal/models"
	"shopify-gst-app/internal/shopify"
)

type OrdersHandler struct {
	db          *sql.DB
	syncService *shopify.SyncService
}

func NewOrdersHandler(db *sql.DB, syncService *shopify.SyncService) *OrdersHandler {
	return &OrdersHandler{
		db:          db,
		syncService: syncService,
	}
}

// SyncOrders triggers a proactive sync with Shopify
func (h *OrdersHandler) SyncOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	count, err := h.syncService.Sync()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Sync completed successfully",
		"count":   count,
	})
}

// ResetOrders wipes the local DB and performs a full fresh sync
func (h *OrdersHandler) ResetOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	count, err := h.syncService.ResetAndSync()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Reset and Sync completed successfully",
		"count":   count,
	})
}

// GetOrders retrieves stored orders from PostgreSQL with pagination support
func (h *OrdersHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 25
	}
	offset := (page - 1) * limit

	// 1. Get total count for the given filters
	countQuery := "SELECT COUNT(*) FROM shopify_orders WHERE 1=1"
	countArgs := []interface{}{}
	countArgIndex := 1

	if startDateStr != "" {
		countQuery += fmt.Sprintf(" AND created_at >= $%d", countArgIndex)
		countArgs = append(countArgs, startDateStr)
		countArgIndex++
	}
	if endDateStr != "" {
		countQuery += fmt.Sprintf(" AND created_at <= $%d", countArgIndex)
		countArgs = append(countArgs, endDateStr)
		countArgIndex++
	}

	var totalCount int
	err := h.db.QueryRow(countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		http.Error(w, "Failed to get total count", http.StatusInternalServerError)
		return
	}

	// 2. Get paginated orders
	query := `
		SELECT id, order_number, total_price, created_at, customer_name, customer_city, customer_state, customer_country, status 
		FROM shopify_orders 
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	if startDateStr != "" {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, startDateStr)
		argIndex++
	}
	if endDateStr != "" {
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, endDateStr)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT %d OFFSET %d", limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		http.Error(w, "Failed to query database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.OrderNumber, &o.TotalPrice, &o.CreatedAt, &o.CustomerName, &o.CustomerCity, &o.CustomerState, &o.CustomerCountry, &o.Status); err != nil {
			http.Error(w, "Failed to parse database rows", http.StatusInternalServerError)
			return
		}
		orders = append(orders, o)
	}

	if orders == nil {
		orders = []models.Order{} // return empty list instead of null
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"orders":      orders,
		"total_count": totalCount,
		"page":        page,
		"limit":       limit,
	})
}

// UpdateOrderStatus handles manual status updates from the frontend
func (h *OrdersHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodOptions {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing order id", http.StatusBadRequest)
		return
	}

	var reqBody struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	res, err := h.db.Exec(`UPDATE shopify_orders SET status = $1 WHERE id = $2`, reqBody.Status, id)
	if err != nil {
		http.Error(w, "Failed to update database", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Status updated successfully",
	})
}

// GenerateInvoice creates and streams a professional PDF GST invoice
func (h *OrdersHandler) GenerateInvoice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing order id", http.StatusBadRequest)
		return
	}

	// 1. Fetch Order from Database
	var o models.Order
	err := h.db.QueryRow(`
		SELECT order_number, total_price, subtotal_price, total_tax, created_at, 
		       customer_name, customer_email, customer_phone, 
		       customer_city, customer_state, customer_country
		FROM shopify_orders WHERE id = $1
	`, id).Scan(&o.OrderNumber, &o.TotalPrice, &o.SubtotalPrice, &o.TotalTax, &o.CreatedAt,
		&o.CustomerName, &o.CustomerEmail, &o.CustomerPhone,
		&o.CustomerCity, &o.CustomerState, &o.CustomerCountry)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Order not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// 2. Fetch Line Items
	rows, err := h.db.Query(`
		SELECT id, title, sku, hs_code, quantity, price, discount
		FROM shopify_order_line_items WHERE order_id = $1
	`, id)
	if err != nil {
		http.Error(w, "Database error fetching items", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []models.LineItem
	for rows.Next() {
		var li models.LineItem
		if err := rows.Scan(&li.ID, &li.Title, &li.SKU, &li.HSCode, &li.Quantity, &li.Price, &li.Discount); err != nil {
			log.Printf("Error scanning item: %v", err)
			continue
		}
		items = append(items, li)
	}

	// 3. Setup PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10.6, 15, 10.6) // ~40px margins

	// Load Montserrat Fonts with absolute paths if possible, but relative should work inside backend
	pdf.AddUTF8Font("Montserrat", "", "internal/fonts/Montserrat-Regular.ttf")
	pdf.AddUTF8Font("Montserrat", "B", "internal/fonts/Montserrat-Bold.ttf")
	// Use Bold as a fallback for SemiBold if needed
	pdf.AddUTF8Font("Montserrat", "I", "internal/fonts/Montserrat-SemiBold.ttf")

	pdf.AddPage()

	// -- Header: Shipper Details --
	pdf.SetFont("Montserrat", "B", 13.5) // ~18px
	pdf.CellFormat(100, 10, "PARFUM TRADERS", "0", 0, "L", false, 0, "")
	pdf.SetFont("Montserrat", "B", 13.5) // ~18px
	pdf.CellFormat(80, 10, "TAX INVOICE", "0", 1, "R", false, 0, "")

	pdf.SetFont("Montserrat", "", 7.5) // ~10px
	pdf.Ln(2)
	pdf.CellFormat(100, 4, "GSTIN: 33AUSPR1909H1ZC", "0", 1, "L", false, 0, "")
	pdf.Ln(1)
	pdf.CellFormat(100, 4, "No. 9/21, 1st floor, Sadiq Basha Nagar,", "0", 1, "L", false, 0, "")
	pdf.CellFormat(100, 4, "2nd Street, Virugambakkam, Chennai - 600092", "0", 1, "L", false, 0, "")
	pdf.Ln(1)
	pdf.CellFormat(100, 4, "Phone: 7904769823", "0", 1, "L", false, 0, "")
	pdf.Ln(6.5) // Spacing before Details

	// -- Order & Customer Info Grid --
	leftCol := 90.0
	rightCol := 90.0

	// Row 1: Invoice Meta | Bill To
	pdf.SetFont("Montserrat", "B", 9) // ~12px Section Titles
	pdf.CellFormat(leftCol, 6, "INVOICE DETAILS", "B", 0, "L", false, 0, "")
	pdf.CellFormat(rightCol, 6, "BILL TO (CUSTOMER)", "B", 1, "L", false, 0, "")
	pdf.Ln(3)

	// Meta Content
	pdf.SetFont("Montserrat", "", 10)
	metaYStart := pdf.GetY()

	// Left: Invoice Meta
	pdf.SetFont("Montserrat", "B", 7.5) // ~10px headers
	pdf.CellFormat(30, 4, "Invoice No:", "0", 0, "L", false, 0, "")
	pdf.SetFont("Montserrat", "", 7.5)
	pdf.CellFormat(60, 4, "INV-"+o.OrderNumber, "0", 1, "L", false, 0, "")

	pdf.SetFont("Montserrat", "B", 7.5)
	pdf.CellFormat(30, 4, "Order No:", "0", 0, "L", false, 0, "")
	pdf.SetFont("Montserrat", "", 7.5)
	pdf.CellFormat(60, 4, "#"+o.OrderNumber, "0", 1, "L", false, 0, "")

	dateStr := o.CreatedAt
	if dateT, err := time.Parse(time.RFC3339, dateStr); err == nil {
		dateStr = dateT.Format("2006-01-02")
	} else if len(dateStr) >= 10 {
		dateStr = dateStr[:10]
	}

	pdf.SetFont("Montserrat", "B", 7.5)
	pdf.CellFormat(30, 4, "Date:", "0", 0, "L", false, 0, "")
	pdf.SetFont("Montserrat", "", 7.5)
	pdf.CellFormat(60, 4, dateStr, "0", 1, "L", false, 0, "")

	pdf.SetY(metaYStart)
	pdf.SetX(leftCol + 15)

	displayName := o.CustomerName
	if strings.TrimSpace(displayName) == "" {
		displayName = "Valued Customer"
	}
	pdf.SetFont("Montserrat", "B", 8.25) // ~11pt for customer name
	pdf.SetX(leftCol + 15)
	pdf.CellFormat(rightCol, 4, displayName, "0", 1, "L", false, 0, "")

	pdf.SetFont("Montserrat", "", 7.5) // ~10px
	if o.CustomerEmail != "" {
		pdf.SetX(leftCol + 15)
		pdf.CellFormat(rightCol, 4, o.CustomerEmail, "0", 1, "L", false, 0, "")
	}

	if o.CustomerPhone != "" {
		pdf.SetX(leftCol + 15)
		pdf.CellFormat(rightCol, 4, "WhatsApp: "+o.CustomerPhone, "0", 1, "L", false, 0, "")
	}

	addressLine := ""
	if o.CustomerCity != "" {
		addressLine += o.CustomerCity + ", "
	}
	if o.CustomerState != "" {
		addressLine += o.CustomerState + ", "
	}
	if o.CustomerCountry != "" {
		addressLine += o.CustomerCountry
	}
	pdf.SetX(leftCol + 15)
	pdf.CellFormat(rightCol, 4, addressLine, "0", 1, "L", false, 0, "")

	pdf.Ln(4.8) // Spacing before Table

	// -- Items Table --
	pdf.SetFillColor(245, 245, 245)
	pdf.SetFont("Montserrat", "I", 7.5) // Montserrat SemiBold at 10px (~7.5pt)

	// Column Widths (total 188.8mm printable width)
	// User percentages: 42, 8, 9, 5, 9, 7, 9, 6, 9 (Total 104%)
	// Scaled to 188.8mm:
	wName := 76.2
	wSKU := 14.5
	wHSN := 16.3
	wQty := 9.1
	wPrice := 16.3
	wDiscount := 12.7
	wTaxable := 16.3
	wGSTPct := 10.9
	wGSTAmt := 16.3

	// Header
	hHeader := 9.0 // ~32px height equivalent
	pdf.CellFormat(wName, hHeader, "Product Name", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wSKU, hHeader, "SKU", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wHSN, hHeader, "HSN", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wQty, hHeader, "Qty", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wPrice, hHeader, "Price", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wDiscount, hHeader, "Disc", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wTaxable, hHeader, "Taxable", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wGSTPct, hHeader, "GST %", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wGSTAmt, hHeader, "GST Amt", "1", 1, "C", true, 0, "")

	pdf.SetFont("Montserrat", "", 6.75) // ~9px Table Content

	isInterState := true
	if strings.Contains(strings.ToLower(o.CustomerState), "tamil nadu") || strings.EqualFold(o.CustomerState, "TN") {
		isInterState = false
	}

	for _, item := range items {
		rawPrice, _ := strconv.ParseFloat(item.Price, 64)
		itemDiscount, _ := strconv.ParseFloat(item.Discount, 64)
		qty := float64(item.Quantity)

		// Final Row Total (Inclusive of tax)
		lineTotal := (rawPrice * qty) - itemDiscount
		if lineTotal < 0 {
			lineTotal = 0
		}

		displayPrice := rawPrice
		if itemDiscount >= (rawPrice * qty) {
			displayPrice = 0.00
		}

		lineTaxable := lineTotal / 1.18
		lineTax := lineTotal - lineTaxable

		hsCode := item.HSCode
		if hsCode == "" {
			hsCode = "33029019"
		}

		// Calculate height for multi-line title
		curX := pdf.GetX()
		curY := pdf.GetY()

		// Professional padding: ~2.6mm Top/Bottom, ~3.2mm Left/Right (12px)
		// Estimate line count for dynamic height
		avgCharWidth := 1.35
		charsPerLine := (wName - 6.4) / avgCharWidth // Subtract horizontal padding (3.2 * 2)
		titleLen := float64(len(item.Title))
		numLines := math.Ceil(titleLen / charsPerLine)
		if numLines < 1 {
			numLines = 1
		}

		h := numLines * 4.5
		if h < 8.5 { // Minimum row height ~32px
			h = 8.5
		}

		// Print wrapped Product Name with padding
		pdf.MultiCell(wName, 4.5, item.Title, "1", "L", false)

		// Fill remaining columns in the same row with strict alignment
		pdf.SetXY(curX+wName, curY)
		pdf.CellFormat(wSKU, h, item.SKU, "1", 0, "C", false, 0, "")
		pdf.CellFormat(wHSN, h, hsCode, "1", 0, "C", false, 0, "")
		pdf.CellFormat(wQty, h, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(wPrice, h, fmt.Sprintf("%.2f", displayPrice), "1", 0, "R", false, 0, "")
		pdf.CellFormat(wDiscount, h, fmt.Sprintf("%.2f", itemDiscount), "1", 0, "R", false, 0, "")
		pdf.CellFormat(wTaxable, h, fmt.Sprintf("%.2f", lineTaxable), "1", 0, "R", false, 0, "")
		pdf.CellFormat(wGSTPct, h, "18%", "1", 0, "C", false, 0, "")
		pdf.CellFormat(wGSTAmt, h, fmt.Sprintf("%.2f", lineTax), "1", 1, "R", false, 0, "")
	}

	// -- Totals Section --
	oSubtotal, _ := strconv.ParseFloat(o.SubtotalPrice, 64)
	oTax, _ := strconv.ParseFloat(o.TotalTax, 64)
	oGrandTotal, _ := strconv.ParseFloat(o.TotalPrice, 64)

	pdf.Ln(5.8) // Spacing before Totals
	pdf.SetX(120)
	pdf.SetFont("Montserrat", "B", 8.25) // ~11px
	pdf.CellFormat(40, 5, "Total Taxable:", "0", 0, "R", false, 0, "")
	pdf.SetFont("Montserrat", "", 8.25)
	pdf.CellFormat(30, 5, fmt.Sprintf("%.2f", oSubtotal), "0", 1, "R", false, 0, "")
	pdf.Ln(1) // Row spacing

	if !isInterState {
		pdf.SetX(120)
		pdf.SetFont("Montserrat", "B", 8.25)
		pdf.CellFormat(40, 5, "CGST (9%):", "0", 0, "R", false, 0, "")
		pdf.SetFont("Montserrat", "", 8.25)
		pdf.CellFormat(30, 5, fmt.Sprintf("%.2f", oTax/2), "0", 1, "R", false, 0, "")
		pdf.Ln(1)

		pdf.SetX(120)
		pdf.SetFont("Montserrat", "B", 8.25)
		pdf.CellFormat(40, 5, "SGST (9%):", "0", 0, "R", false, 0, "")
		pdf.SetFont("Montserrat", "", 8.25)
		pdf.CellFormat(30, 5, fmt.Sprintf("%.2f", oTax/2), "0", 1, "R", false, 0, "")
	} else {
		pdf.SetX(120)
		pdf.SetFont("Montserrat", "B", 8.25)
		pdf.CellFormat(40, 5, "IGST (18%):", "0", 0, "R", false, 0, "")
		pdf.SetFont("Montserrat", "", 8.25)
		pdf.CellFormat(30, 5, fmt.Sprintf("%.2f", oTax), "0", 1, "R", false, 0, "")
	}

	pdf.Ln(2.6) // Extra padding above Grand Total
	pdf.SetX(120)
	pdf.CellFormat(70, 0, "", "T", 1, "R", false, 0, "")
	pdf.SetX(120)
	pdf.SetFont("Montserrat", "B", 10.5) // ~14px
	pdf.CellFormat(40, 10, "GRAND TOTAL:", "0", 0, "R", false, 0, "")
	pdf.CellFormat(30, 10, fmt.Sprintf("%.2f", oGrandTotal), "0", 1, "R", false, 0, "")

	pdf.Ln(8.0) // Spacing before Footer Terms

	// -- Footer: Terms --
	footerSize := 6.4 // ~8.5px
	pdf.SetFont("Montserrat", "B", footerSize)
	pdf.CellFormat(0, 4, "Payment Terms:", "0", 1, "L", false, 0, "")
	pdf.SetFont("Montserrat", "", footerSize)
	pdf.MultiCell(0, 3.5, "Full payment is required before the due date mentioned on the invoice.", "0", "L", false)
	pdf.Ln(2)

	pdf.SetFont("Montserrat", "B", footerSize)
	pdf.CellFormat(0, 4, "No Refunds & Returns:", "0", 1, "L", false, 0, "")
	pdf.SetFont("Montserrat", "", footerSize)
	pdf.MultiCell(0, 3.5, "Due to the nature of our products, we do not accept returns or provide refunds once the item has been opened or used.", "0", "L", false)
	pdf.Ln(2)

	pdf.SetFont("Montserrat", "B", footerSize)
	pdf.CellFormat(0, 4, "Intellectual Property:", "0", 1, "L", false, 0, "")
	pdf.SetFont("Montserrat", "", footerSize)
	pdf.MultiCell(0, 3.5, "All branding and product names are trademarks of Parfum Traders and may not be reproduced without permission.", "0", "L", false)

	// Stream PDF
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=invoice-%s.pdf", o.OrderNumber))

	if err := pdf.Output(w); err != nil {
		log.Printf("PDF Error: %v", err)
	}
}
