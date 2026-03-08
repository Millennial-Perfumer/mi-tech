package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ReportsHandler struct {
	db *sql.DB
}

func NewReportsHandler(db *sql.DB) *ReportsHandler {
	return &ReportsHandler{db: db}
}

type DocumentIssuedRow struct {
	DocumentType string `json:"document_type"`
	FromSerial   string `json:"from_serial"`
	ToSerial     string `json:"to_serial"`
	TotalIssued  int    `json:"total_issued"`
	Cancelled    int    `json:"cancelled"`
	NetIssued    int    `json:"net_issued"`
}

type GSTSummary struct {
	TotalOrders       int     `json:"total_orders"`
	CancelledOrders   int     `json:"cancelled_orders"`
	InvoicesGenerated int     `json:"invoices_generated"`
	TotalRevenue      float64 `json:"total_revenue"`
	TotalTaxableValue float64 `json:"total_taxable_value"`
	TotalGSTCollected float64 `json:"total_gst_collected"`
	TotalIGST         float64 `json:"total_igst"`
	TotalCGST         float64 `json:"total_cgst"`
	TotalSGST         float64 `json:"total_sgst"`
	FulfilledOrders   int     `json:"fulfilled_orders"`
	UnfulfilledOrders int     `json:"unfulfilled_orders"`
	PaidOrders        int     `json:"paid_orders"`
	RefundedOrders    int     `json:"refunded_orders"`
}

func (h *ReportsHandler) GetGSTSummary(w http.ResponseWriter, r *http.Request) {
	startDate, endDate := parseDates(r)

	query := `
		SELECT 
			COUNT(id) as total_orders,
			COUNT(id) FILTER (WHERE LOWER(status) = 'cancelled') as cancelled_orders,
			COUNT(id) FILTER (WHERE LOWER(status) = 'fulfilled') as fulfilled_orders,
			COUNT(id) FILTER (WHERE LOWER(status) = 'unfulfilled') as unfulfilled_orders,
			COUNT(id) FILTER (WHERE LOWER(status) = 'paid') as paid_orders,
			COALESCE(SUM(total_price) FILTER (WHERE LOWER(status) != 'cancelled'), 0) as total_revenue,
			COALESCE(SUM(subtotal_price) FILTER (WHERE LOWER(status) != 'cancelled'), 0) as total_taxable,
			COALESCE(SUM(total_tax) FILTER (WHERE LOWER(status) != 'cancelled'), 0) as total_tax
		FROM shopify_orders 
		WHERE created_at >= $1 AND created_at <= $2
	`

	var s GSTSummary
	err := h.db.QueryRow(query, startDate, endDate).Scan(
		&s.TotalOrders, &s.CancelledOrders, &s.FulfilledOrders, &s.UnfulfilledOrders, &s.PaidOrders,
		&s.TotalRevenue, &s.TotalTaxableValue, &s.TotalGSTCollected,
	)
	if err != nil {
		http.Error(w, "Failed to fetch summary: "+err.Error(), http.StatusInternalServerError)
		return
	}

	s.InvoicesGenerated = s.TotalOrders - s.CancelledOrders

	// Simplified tax split for overall summary based on state logic
	stateQuery := `
		SELECT 
			customer_state,
			SUM(total_tax) as sum_tax
		FROM shopify_orders
		WHERE created_at >= $1 AND created_at <= $2 AND LOWER(status) != 'cancelled'
		GROUP BY customer_state
	`
	rows, _ := h.db.Query(stateQuery, startDate, endDate)
	defer rows.Close()
	for rows.Next() {
		var state string
		var tax float64
		rows.Scan(&state, &tax)
		if isTamilNadu(state) {
			s.TotalCGST += tax / 2
			s.TotalSGST += tax / 2
		} else {
			s.TotalIGST += tax
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "summary": s})
}

type StateSummary struct {
	State        string  `json:"state"`
	Orders       int     `json:"orders"`
	TaxableValue float64 `json:"taxable_value"`
	IGST         float64 `json:"igst"`
	CGST         float64 `json:"cgst"`
	SGST         float64 `json:"sgst"`
	TotalGST     float64 `json:"total_gst"`
	Revenue      float64 `json:"revenue"`
}

func (h *ReportsHandler) GetStateSummary(w http.ResponseWriter, r *http.Request) {
	startDate, endDate := parseDates(r)

	query := `
		SELECT 
			customer_state,
			COUNT(id) as orders,
			COALESCE(SUM(subtotal_price), 0) as taxable,
			COALESCE(SUM(total_tax), 0) as gst,
			COALESCE(SUM(total_price), 0) as revenue
		FROM shopify_orders
		WHERE created_at >= $1 AND created_at <= $2 AND LOWER(status) != 'cancelled'
		GROUP BY customer_state
		ORDER BY revenue DESC
	`

	rows, err := h.db.Query(query, startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var reports []StateSummary
	for rows.Next() {
		var s StateSummary
		rows.Scan(&s.State, &s.Orders, &s.TaxableValue, &s.TotalGST, &s.Revenue)

		if isTamilNadu(s.State) {
			s.CGST = s.TotalGST / 2
			s.SGST = s.TotalGST / 2
		} else {
			s.IGST = s.TotalGST
		}
		reports = append(reports, s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": reports})
}

type HSNSummary struct {
	HSNCode      string  `json:"hsn_code"`
	ProductCount int     `json:"product_count"`
	QtySold      int     `json:"qty_sold"`
	TaxableValue float64 `json:"taxable_value"`
	IGST         float64 `json:"igst"`
	CGST         float64 `json:"cgst"`
	SGST         float64 `json:"sgst"`
	TotalGST     float64 `json:"total_gst"`
	Revenue      float64 `json:"revenue"`
}

func (h *ReportsHandler) GetHSNSummary(w http.ResponseWriter, r *http.Request) {
	startDate, endDate := parseDates(r)

	// Note: We compute tax split per line item by looking at the order's state
	query := `
		SELECT 
			li.hs_code,
			COUNT(DISTINCT li.id) as product_count,
			SUM(li.quantity) as qty,
			SUM((li.price * li.quantity - li.discount) / 1.18) as taxable,
			SUM((li.price * li.quantity - li.discount) - (li.price * li.quantity - li.discount) / 1.18) as gst,
			SUM(li.price * li.quantity - li.discount) as revenue,
			o.customer_state
		FROM shopify_order_line_items li
		JOIN shopify_orders o ON li.order_id = o.id
		WHERE o.created_at >= $1 AND o.created_at <= $2 AND LOWER(o.status) != 'cancelled'
		GROUP BY li.hs_code, o.customer_state
	`

	rows, err := h.db.Query(query, startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	hsnMap := make(map[string]*HSNSummary)
	for rows.Next() {
		var hsn, state string
		var pc, qty int
		var tax, gst, rev float64
		rows.Scan(&hsn, &pc, &qty, &tax, &gst, &rev, &state)

		if hsn == "" {
			hsn = "33029019"
		}

		if _, ok := hsnMap[hsn]; !ok {
			hsnMap[hsn] = &HSNSummary{HSNCode: hsn}
		}
		s := hsnMap[hsn]
		s.ProductCount += pc
		s.QtySold += qty
		s.TaxableValue += tax
		s.TotalGST += gst
		s.Revenue += rev

		if isTamilNadu(state) {
			s.CGST += gst / 2
			s.SGST += gst / 2
		} else {
			s.IGST += gst
		}
	}

	var reports []HSNSummary
	for _, v := range hsnMap {
		reports = append(reports, *v)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": reports})
}

func (h *ReportsHandler) GetDocumentsIssued(w http.ResponseWriter, r *http.Request) {
	startDate, endDate := parseDates(r)

	// We use regexp_replace to strip any non-numeric characters (like #) to ensure integer casting works
	query := `
		SELECT 
			MIN(NULLIF(regexp_replace(order_number, '[^0-9]', '', 'g'), '')::bigint) as min_val,
			MAX(NULLIF(regexp_replace(order_number, '[^0-9]', '', 'g'), '')::bigint) as max_val,
			COUNT(id) as total,
			COUNT(id) FILTER (WHERE LOWER(status) = 'cancelled') as cancelled
		FROM shopify_orders
		WHERE created_at >= $1 AND created_at <= $2
	`

	var minOrder, maxOrder sql.NullInt64
	var total, cancelled int
	err := h.db.QueryRow(query, startDate, endDate).Scan(&minOrder, &maxOrder, &total, &cancelled)

	rows := []DocumentIssuedRow{}
	if err == nil && total > 0 {
		fromS := ""
		toS := ""
		if minOrder.Valid {
			fromS = fmt.Sprintf("INV-%d", minOrder.Int64)
		}
		if maxOrder.Valid {
			toS = fmt.Sprintf("INV-%d", maxOrder.Int64)
		}

		rows = append(rows, DocumentIssuedRow{
			DocumentType: "Tax Invoice",
			FromSerial:   fromS,
			ToSerial:     toS,
			TotalIssued:  total,
			Cancelled:    cancelled,
			NetIssued:    total - cancelled,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": rows})
}

func parseDates(r *http.Request) (time.Time, time.Time) {
	s := r.URL.Query().Get("start_date")
	e := r.URL.Query().Get("end_date")

	start, _ := time.Parse(time.RFC3339, s)
	end, _ := time.Parse(time.RFC3339, e)

	if start.IsZero() {
		now := time.Now()
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}
	if end.IsZero() {
		end = time.Now()
	}
	return start, end
}

func isTamilNadu(state string) bool {
	s := fmt.Sprintf("%v", state)
	return (len(s) > 0 && (s == "Tamil Nadu" || s == "TN" || s == "tamil nadu"))
}
