package repository

import (
	"database/sql"
	"fmt"
	"time"

	"mi-tech/internal/domain/gst/dto"

	"gorm.io/gorm"
)

type gormGSTRepository struct {
	db *gorm.DB
}

// NewGSTRepository creates a new GORM-backed GSTRepository.
func NewGSTRepository(db *gorm.DB) GSTRepository {
	return &gormGSTRepository{db: db}
}

func (r *gormGSTRepository) GetGSTSummary(startDate, endDate string) (GSTSummaryResult, error) {
	start, end := parseDateRange(startDate, endDate)

	query := `
		SELECT 
			COUNT(t.transaction_id),
			COUNT(t.transaction_id) FILTER (WHERE LOWER(t.order_status) IN ('cancelled', 'canceled') OR LOWER(COALESCE(t.fulfillment_status, '')) IN ('cancelled', 'canceled')),
			COUNT(t.transaction_id) FILTER (WHERE LOWER(t.fulfillment_status) = 'fulfilled'),
			COUNT(t.transaction_id) FILTER (WHERE LOWER(COALESCE(t.fulfillment_status, '')) != 'fulfilled' AND NOT (LOWER(COALESCE(t.order_status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(t.fulfillment_status, '')) IN ('cancelled', 'canceled'))),
			COUNT(t.transaction_id) FILTER (WHERE LOWER(t.payment_status) = 'paid'),
			COALESCE(SUM(t.total_price) FILTER (WHERE NOT (LOWER(COALESCE(t.order_status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(t.fulfillment_status, '')) IN ('cancelled', 'canceled'))), 0) as revenue,
			COALESCE(SUM(ROUND(t.total_price / 1.18, 2)) FILTER (WHERE NOT (LOWER(COALESCE(t.order_status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(t.fulfillment_status, '')) IN ('cancelled', 'canceled'))), 0) as taxable,
			COALESCE(SUM(t.total_price - ROUND(t.total_price / 1.18, 2)) FILTER (WHERE NOT (LOWER(COALESCE(t.order_status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(t.fulfillment_status, '')) IN ('cancelled', 'canceled'))), 0) as tax,
			COALESCE(SUM(CASE WHEN COALESCE(s.code, '33') = '33' THEN (t.total_price - ROUND(t.total_price / 1.18, 2)) / 2 ELSE 0 END) FILTER (WHERE NOT (LOWER(COALESCE(t.order_status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(t.fulfillment_status, '')) IN ('cancelled', 'canceled'))), 0) as cgst,
			COALESCE(SUM(CASE WHEN COALESCE(s.code, '33') = '33' THEN (t.total_price - ROUND(t.total_price / 1.18, 2)) / 2 ELSE 0 END) FILTER (WHERE NOT (LOWER(COALESCE(t.order_status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(t.fulfillment_status, '')) IN ('cancelled', 'canceled'))), 0) as sgst,
			COALESCE(SUM(CASE WHEN COALESCE(s.code, '33') != '33' THEN (t.total_price - ROUND(t.total_price / 1.18, 2)) ELSE 0 END) FILTER (WHERE NOT (LOWER(COALESCE(t.order_status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(t.fulfillment_status, '')) IN ('cancelled', 'canceled'))), 0) as igst
		FROM unified_revenue_transactions t
		LEFT JOIN gst_state_codes s ON LOWER(TRIM(t.state)) = ANY(s.aliases)
		WHERE t.tx_date >= ? AND t.tx_date <= ?
	`

	var result GSTSummaryResult
	row := r.db.Raw(query, start, end).Row()
	err := row.Scan(
		&result.TotalOrders, &result.CancelledOrders, &result.FulfilledOrders,
		&result.UnfulfilledOrders, &result.PaidOrders,
		&result.TotalRevenue, &result.TotalTaxable, &result.TotalTax,
		&result.CGST, &result.SGST, &result.IGST,
	)
	if err != nil {
		return GSTSummaryResult{}, err
	}

	return result, nil
}

func (r *gormGSTRepository) GetStateSummary(startDate, endDate string) ([]StateSummaryResult, error) {
	start, end := parseDateRange(startDate, endDate)

	query := `
		SELECT 
			INITCAP(COALESCE(state, 'N/A')) as state,
			COUNT(transaction_id) as orders,
			COALESCE(SUM(ROUND(total_price / 1.18, 2)), 0) as taxable_value,
			COALESCE(SUM(total_price - ROUND(total_price / 1.18, 2)), 0) as total_gst,
			COALESCE(SUM(total_price), 0) as revenue
		FROM unified_revenue_transactions
		WHERE tx_date >= ? AND tx_date <= ? AND LOWER(COALESCE(source_id, '')) NOT IN ('b2b') AND NOT (LOWER(COALESCE(order_status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(fulfillment_status, '')) IN ('cancelled', 'canceled'))
		GROUP BY INITCAP(COALESCE(state, 'N/A'))
		ORDER BY revenue DESC
	`

	var results []StateSummaryResult
	if err := r.db.Raw(query, start, end).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to query state summary: %w", err)
	}
	return results, nil
}

func (r *gormGSTRepository) GetHSNSummary(startDate, endDate string) ([]HSNSummaryResult, error) {
	start, end := parseDateRange(startDate, endDate)

	query := `
		WITH LineItemShares AS (
			SELECT 
				li.order_id,
				COALESCE(li.hs_code, '33029019') as hs_code,
				li.quantity,
				(li.price * li.quantity - li.discount - COALESCE(li.order_discount, 0)) as line_val,
				SUM(li.price * li.quantity - li.discount - COALESCE(li.order_discount, 0)) OVER (PARTITION BY li.order_id) as line_sum,
				SUM(li.quantity) OVER (PARTITION BY li.order_id) as qty_sum,
				COUNT(li.id) OVER (PARTITION BY li.order_id) as item_count,
				o.total_price,
				ROUND(o.total_price / 1.18, 2) as order_taxable,
				(o.total_price - ROUND(o.total_price / 1.18, 2)) as order_tax,
				INITCAP(COALESCE(o.customer_state, 'N/A')) as state
			FROM order_line_items li
			JOIN orders o ON li.order_id = o.id
			WHERE o.created_at >= ? AND o.created_at <= ?
			  AND (
			    (LOWER(COALESCE(o.source_id, '')) = 'b2b' AND LOWER(COALESCE(o.status, '')) = 'issued')
			    OR 
			    (LOWER(COALESCE(o.source_id, '')) != 'b2b' AND NOT (LOWER(COALESCE(o.status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(o.fulfillment_status, '')) IN ('cancelled', 'canceled')))
			  )
		),
		CalculatedShares AS (
			SELECT
				order_id,
				hs_code,
				quantity,
				total_price,
				order_taxable,
				order_tax,
				state,
				CASE 
					WHEN line_sum > 0 THEN (line_val / line_sum)
					WHEN qty_sum > 0 THEN (quantity::numeric / qty_sum)
					ELSE (1.0::numeric / item_count)
				END as share
			FROM LineItemShares
		)
		SELECT 
			hs_code as hsn_code,
			COUNT(DISTINCT order_id) as product_count,
			SUM(quantity) as qty_sold,
			ROUND(SUM(share * order_taxable), 2) as taxable_value,
			ROUND(SUM(share * order_tax), 2) as total_gst,
			ROUND(SUM(share * total_price), 2) as revenue,
			state
		FROM CalculatedShares
		GROUP BY hs_code, state
		ORDER BY revenue DESC
	`

	var results []HSNSummaryResult
	if err := r.db.Raw(query, start, end).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to query HSN summary: %w", err)
	}
	return results, nil
}

func (r *gormGSTRepository) GetShopifyDocumentsIssued(startDate, endDate string) (minOrder, maxOrder *int64, total, cancelled int, err error) {
	start, end := parseDateRange(startDate, endDate)

	query := `
		SELECT 
			MIN(NULLIF(regexp_replace(invoice_number, '[^0-9]', '', 'g'), '')::bigint) as min_val,
			MAX(NULLIF(regexp_replace(invoice_number, '[^0-9]', '', 'g'), '')::bigint) as max_val,
			COUNT(id) as total,
			COUNT(id) FILTER (WHERE LOWER(status) IN ('cancelled', 'canceled') OR LOWER(fulfillment_status) IN ('cancelled', 'canceled')) as cancelled
		FROM orders
		WHERE created_at >= ? AND created_at <= ? AND source_id = 'shopify'
	`

	var minV, maxV sql.NullInt64
	row := r.db.Raw(query, start, end).Row()
	err = row.Scan(&minV, &maxV, &total, &cancelled)
	if err != nil {
		return
	}
	if minV.Valid {
		minOrder = &minV.Int64
	}
	if maxV.Valid {
		maxOrder = &maxV.Int64
	}
	return
}

func (r *gormGSTRepository) GetAmazonDocumentsIssued(startDate, endDate string) (minOrder, maxOrder *int64, total, cancelled int, err error) {
	start, end := parseDateRange(startDate, endDate)

	query := `
		SELECT 
			MIN(NULLIF(regexp_replace(invoice_number, '[^0-9]', '', 'g'), '')::bigint) as min_val,
			MAX(NULLIF(regexp_replace(invoice_number, '[^0-9]', '', 'g'), '')::bigint) as max_val,
			COUNT(id) as total,
			COUNT(id) FILTER (WHERE LOWER(status) IN ('cancelled', 'canceled') OR LOWER(fulfillment_status) IN ('cancelled', 'canceled')) as cancelled
		FROM orders
		WHERE created_at >= ? AND created_at <= ? AND source_id = 'amazon'
	`

	var minV, maxV sql.NullInt64
	row := r.db.Raw(query, start, end).Row()
	err = row.Scan(&minV, &maxV, &total, &cancelled)
	if err != nil {
		return
	}
	if minV.Valid {
		minOrder = &minV.Int64
	}
	if maxV.Valid {
		maxOrder = &maxV.Int64
	}
	return
}

func (r *gormGSTRepository) GetGSTR1B2CS(startDate, endDate string) ([]dto.B2CSRow, error) {
	start, end := parseDateRange(startDate, endDate)

	query := `
		SELECT 
			COALESCE(s.code, '33') as pos_code,
			(o.source_id = 'amazon') as is_amazon,
			COALESCE(SUM(ROUND(o.total_price / 1.18, 2)), 0) as taxable_value,
			COALESCE(SUM(o.total_price - ROUND(o.total_price / 1.18, 2)), 0) as total_gst
		FROM orders o
		LEFT JOIN gst_state_codes s ON LOWER(TRIM(o.customer_state)) = ANY(s.aliases)
		WHERE o.created_at >= ? AND o.created_at <= ? AND COALESCE(o.source_id, '') != 'b2b' AND NOT (LOWER(COALESCE(o.status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(o.fulfillment_status, '')) IN ('cancelled', 'canceled'))
		GROUP BY COALESCE(s.code, '33'), (o.source_id = 'amazon')
	`

	type b2cQueryResult struct {
		POSCode      string  `gorm:"column:pos_code"`
		IsAmazon     bool    `gorm:"column:is_amazon"`
		TaxableValue float64 `gorm:"column:taxable_value"`
		TotalGST     float64 `gorm:"column:total_gst"`
	}

	var results []b2cQueryResult
	if err := r.db.Raw(query, start, end).Scan(&results).Error; err != nil {
		return nil, err
	}

	var rows []dto.B2CSRow
	for _, res := range results {
		splyTy := "INTER"
		var iamt, camt, samt float64
		if res.POSCode == "33" {
			splyTy = "INTRA"
			camt = res.TotalGST / 2
			samt = res.TotalGST / 2
		} else {
			iamt = res.TotalGST
		}

		typ := "OE"
		if res.IsAmazon {
			typ = "E"
		}

		rows = append(rows, dto.B2CSRow{
			SplyTy: splyTy,
			POS:    res.POSCode,
			Rt:     18.0,
			TxVal:  res.TaxableValue,
			Iamt:   iamt,
			Camt:   camt,
			Samt:   samt,
			Typ:    typ,
		})
	}

	return rows, nil
}

func (r *gormGSTRepository) GetGSTR1HSN(startDate, endDate string) ([]dto.HSNRow, error) {
	start, end := parseDateRange(startDate, endDate)

	query := `
		WITH LineItemShares AS (
			SELECT 
				li.order_id,
				COALESCE(li.hs_code, '33029019') as hs_code,
				COALESCE(li.title, 'Products') as title,
				li.quantity,
				(li.price * li.quantity - li.discount - COALESCE(li.order_discount, 0)) as line_val,
				SUM(li.price * li.quantity - li.discount - COALESCE(li.order_discount, 0)) OVER (PARTITION BY li.order_id) as line_sum,
				SUM(li.quantity) OVER (PARTITION BY li.order_id) as qty_sum,
				COUNT(li.id) OVER (PARTITION BY li.order_id) as item_count,
				o.total_price,
				ROUND(o.total_price / 1.18, 2) as order_taxable,
				(o.total_price - ROUND(o.total_price / 1.18, 2)) as order_tax,
				COALESCE(s.code, '33') as pos_code,
				CASE WHEN LOWER(COALESCE(o.source_id, '')) = 'b2b' THEN true ELSE false END as is_b2b
			FROM order_line_items li
			JOIN orders o ON li.order_id = o.id
			LEFT JOIN gst_state_codes s ON LOWER(TRIM(o.customer_state)) = ANY(s.aliases)
			WHERE o.created_at >= ? AND o.created_at <= ?
			  AND (
			    (LOWER(COALESCE(o.source_id, '')) = 'b2b' AND LOWER(COALESCE(o.status, '')) = 'issued')
			    OR 
			    (LOWER(COALESCE(o.source_id, '')) != 'b2b' AND NOT (LOWER(COALESCE(o.status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(o.fulfillment_status, '')) IN ('cancelled', 'canceled')))
			  )
		),
		CalculatedShares AS (
			SELECT
				order_id,
				hs_code,
				title,
				quantity,
				total_price,
				order_taxable,
				order_tax,
				pos_code,
				is_b2b,
				CASE 
					WHEN line_sum > 0 THEN (line_val / line_sum)
					WHEN qty_sum > 0 THEN (quantity::numeric / qty_sum)
					ELSE (1.0::numeric / item_count)
				END as share
			FROM LineItemShares
		)
		SELECT 
			hs_code as hsn_code,
			MAX(title) as description,
			SUM(quantity) as qty,
			ROUND(SUM(share * total_price), 2) as gross_val,
			ROUND(SUM(share * order_taxable), 2) as taxable_val,
			ROUND(SUM(CASE WHEN pos_code != '33' THEN share * order_tax ELSE 0 END), 2) as igst,
			ROUND(SUM(CASE WHEN pos_code = '33' THEN (share * order_tax) / 2 ELSE 0 END), 2) as cgst,
			ROUND(SUM(CASE WHEN pos_code = '33' THEN (share * order_tax) / 2 ELSE 0 END), 2) as sgst,
			is_b2b
		FROM CalculatedShares
		GROUP BY hs_code, is_b2b
	`

	type hsnQueryResult struct {
		HsnCode     string
		Description string
		Qty         float64
		GrossVal    float64
		TaxableVal  float64
		Igst        float64
		Cgst        float64
		Sgst        float64
		IsB2b       bool
	}

	var results []hsnQueryResult
	if err := r.db.Raw(query, start, end).Scan(&results).Error; err != nil {
		return nil, err
	}

	var rows []dto.HSNRow
	for idx, res := range results {
		desc := res.Description
		if len(desc) > 30 {
			desc = desc[:30]
		}
		rows = append(rows, dto.HSNRow{
			Num:   idx + 1,
			HsnSc: res.HsnCode,
			Desc:  desc,
			Uqc:   "PCS",
			Qty:   res.Qty,
			Val:   res.GrossVal,
			TxVal: res.TaxableVal,
			Iamt:  res.Igst,
			Camt:  res.Cgst,
			Samt:  res.Sgst,
			IsB2B: res.IsB2b,
		})
	}

	return rows, nil
}

func (r *gormGSTRepository) GetGSTR1B2B(startDate, endDate string) ([]dto.GSTR1B2B, error) {
	start, end := parseDateRange(startDate, endDate)

	query := `
		SELECT 
			i.customer_gstin,
			i.invoice_number,
			i.invoice_date,
			i.total_price,
			i.customer_state_code,
			ROUND(i.total_price / 1.18, 2) as taxable_value,
			CASE WHEN i.seller_state_code = i.customer_state_code THEN ROUND((i.total_price - ROUND(i.total_price / 1.18, 2)) / 2, 2) ELSE 0 END as cgst,
			CASE WHEN i.seller_state_code = i.customer_state_code THEN ROUND((i.total_price - ROUND(i.total_price / 1.18, 2)) / 2, 2) ELSE 0 END as sgst,
			CASE WHEN i.seller_state_code != i.customer_state_code THEN (i.total_price - ROUND(i.total_price / 1.18, 2)) ELSE 0 END as igst
		FROM b2b_invoices i
		WHERE i.invoice_date >= ? AND i.invoice_date <= ? AND i.status = 'ISSUED'
	`

	type b2bQueryResult struct {
		CustomerGSTIN     string    `gorm:"column:customer_gstin"`
		InvoiceNumber     string    `gorm:"column:invoice_number"`
		InvoiceDate       time.Time `gorm:"column:invoice_date"`
		TotalPrice        float64   `gorm:"column:total_price"`
		CustomerStateCode string    `gorm:"column:customer_state_code"`
		TaxableValue      float64   `gorm:"column:taxable_value"`
		CGST              float64   `gorm:"column:cgst"`
		SGST              float64   `gorm:"column:sgst"`
		IGST              float64   `gorm:"column:igst"`
	}

	var results []b2bQueryResult
	if err := r.db.Raw(query, start.Format("2006-01-02"), end.Format("2006-01-02")).Scan(&results).Error; err != nil {
		return nil, err
	}

	// Group by customer GSTIN
	b2bMap := make(map[string]*dto.GSTR1B2B)
	for _, res := range results {
		gstin := res.CustomerGSTIN
		if _, exists := b2bMap[gstin]; !exists {
			b2bMap[gstin] = &dto.GSTR1B2B{
				Ctin: gstin,
				Inv:  []dto.GSTR1B2BInv{},
			}
		}

		inv := dto.GSTR1B2BInv{
			Inum:   res.InvoiceNumber,
			Idt:    res.InvoiceDate.Format("02-01-2006"),
			Val:    res.TotalPrice,
			Pos:    res.CustomerStateCode,
			Rchrg:  "N",
			InvTyp: "R",
			Itms: []dto.GSTR1TaxItem{
				{
					Num: 1,
					ItmDet: dto.GSTR1TaxDetails{
						Rt:    18.0,
						TxVal: res.TaxableValue,
						Iamt:  res.IGST,
						Camt:  res.CGST,
						Samt:  res.SGST,
					},
				},
			},
		}
		b2bMap[gstin].Inv = append(b2bMap[gstin].Inv, inv)
	}

	var list []dto.GSTR1B2B
	for _, v := range b2bMap {
		list = append(list, *v)
	}
	return list, nil
}

func (r *gormGSTRepository) GetGSTR1CDNR(startDate, endDate string) ([]dto.GSTR1CDNR, error) {
	start, end := parseDateRange(startDate, endDate)

	query := `
		SELECT 
			customer_gstin,
			note_number,
			note_date,
			invoice_number,
			total_price,
			customer_state_code,
			note_type,
			COALESCE(SUM(item_amount), 0) as taxable_value,
			COALESCE(SUM(cgst), 0) as cgst,
			COALESCE(SUM(sgst), 0) as sgst,
			COALESCE(SUM(igst), 0) as igst
		FROM (
			SELECT 
				c.customer_gstin,
				c.credit_note_number as note_number,
				c.note_date,
				c.invoice_number,
				c.total_price,
				c.customer_state_code,
				'C' as note_type,
				ci.amount as item_amount,
				CASE WHEN c.seller_state_code = c.customer_state_code THEN ci.amount * 0.09 ELSE 0 END as cgst,
				CASE WHEN c.seller_state_code = c.customer_state_code THEN ci.amount * 0.09 ELSE 0 END as sgst,
				CASE WHEN c.seller_state_code != c.customer_state_code THEN ci.amount * 0.18 ELSE 0 END as igst
			FROM b2b_credit_notes c
			JOIN b2b_credit_note_items ci ON c.id = ci.credit_note_id
			WHERE c.note_date >= ? AND c.note_date <= ? AND c.status = 'ISSUED'
			
			UNION ALL
			
			SELECT 
				d.customer_gstin,
				d.debit_note_number as note_number,
				d.note_date,
				d.invoice_number,
				d.total_price,
				d.customer_state_code,
				'D' as note_type,
				di.amount as item_amount,
				CASE WHEN d.seller_state_code = d.customer_state_code THEN di.amount * 0.09 ELSE 0 END as cgst,
				CASE WHEN d.seller_state_code = d.customer_state_code THEN di.amount * 0.09 ELSE 0 END as sgst,
				CASE WHEN d.seller_state_code != d.customer_state_code THEN di.amount * 0.18 ELSE 0 END as igst
			FROM b2b_debit_notes d
			JOIN b2b_debit_note_items di ON d.id = di.debit_note_id
			WHERE d.note_date >= ? AND d.note_date <= ? AND d.status = 'ISSUED'
		) combined
		GROUP BY customer_gstin, note_number, note_date, invoice_number, total_price, customer_state_code, note_type
	`

	type cdnrQueryResult struct {
		CustomerGSTIN     string    `gorm:"column:customer_gstin"`
		NoteNumber        string    `gorm:"column:note_number"`
		NoteDate          time.Time `gorm:"column:note_date"`
		InvoiceNumber     string    `gorm:"column:invoice_number"`
		TotalPrice        float64   `gorm:"column:total_price"`
		CustomerStateCode string    `gorm:"column:customer_state_code"`
		NoteType          string    `gorm:"column:note_type"`
		TaxableValue      float64   `gorm:"column:taxable_value"`
		CGST              float64   `gorm:"column:cgst"`
		SGST              float64   `gorm:"column:sgst"`
		IGST              float64   `gorm:"column:igst"`
	}

	var results []cdnrQueryResult
	dtStr := start.Format("2006-01-02")
	dtEndStr := end.Format("2006-01-02")
	if err := r.db.Raw(query, dtStr, dtEndStr, dtStr, dtEndStr).Scan(&results).Error; err != nil {
		return nil, err
	}

	cdnrMap := make(map[string]*dto.GSTR1CDNR)
	for _, res := range results {
		gstin := res.CustomerGSTIN
		if _, exists := cdnrMap[gstin]; !exists {
			cdnrMap[gstin] = &dto.GSTR1CDNR{
				Ctin: gsinDeref(gstin),
				Nt:   []dto.GSTR1CDNRNt{},
			}
		}

		nt := dto.GSTR1CDNRNt{
			Ntty:  res.NoteType,
			NtNum: res.NoteNumber,
			NtDt:  res.NoteDate.Format("02-01-2006"),
			Inum:  res.InvoiceNumber,
			Idt:   res.NoteDate.Format("02-01-2006"), // fallback same date
			Val:   res.TotalPrice,
			Pos:   res.CustomerStateCode,
			Itms: []dto.GSTR1TaxItem{
				{
					Num: 1,
					ItmDet: dto.GSTR1TaxDetails{
						Rt:    18.0,
						TxVal: res.TaxableValue,
						Iamt:  res.IGST,
						Camt:  res.CGST,
						Samt:  res.SGST,
					},
				},
			},
		}
		cdnrMap[gstin].Nt = append(cdnrMap[gstin].Nt, nt)
	}

	var list []dto.GSTR1CDNR
	for _, v := range cdnrMap {
		list = append(list, *v)
	}
	return list, nil
}

func gsinDeref(s string) string {
	return s
}

func (r *gormGSTRepository) GetB2BDocumentsIssued(startDate, endDate string) (minInvoice, maxInvoice *string, total, cancelled int, err error) {
	start, end := parseDateRange(startDate, endDate)

	var minSeq, maxSeq sql.NullInt64
	query := `
		SELECT 
			MIN(invoice_sequence) as min_val,
			MAX(invoice_sequence) as max_val,
			COUNT(id) as total,
			COUNT(id) FILTER (WHERE status = 'CANCELLED') as cancelled
		FROM b2b_invoices
		WHERE invoice_date >= ? AND invoice_date <= ? AND status IN ('ISSUED', 'CANCELLED')
	`
	row := r.db.Raw(query, start, end).Row()
	err = row.Scan(&minSeq, &maxSeq, &total, &cancelled)
	if err != nil {
		return
	}

	if minSeq.Valid {
		var minNum string
		if err := r.db.Table("b2b_invoices").
			Where("invoice_date >= ? AND invoice_date <= ? AND invoice_sequence = ?", start, end, minSeq.Int64).
			Pluck("invoice_number", &minNum).Error; err == nil && minNum != "" {
			minInvoice = &minNum
		}
	}
	if maxSeq.Valid {
		var maxNum string
		if err := r.db.Table("b2b_invoices").
			Where("invoice_date >= ? AND invoice_date <= ? AND invoice_sequence = ?", start, end, maxSeq.Int64).
			Pluck("invoice_number", &maxNum).Error; err == nil && maxNum != "" {
			maxInvoice = &maxNum
		}
	}

	return
}

func (r *gormGSTRepository) GetB2BCreditNotesIssued(startDate, endDate string) (minNote, maxNote *string, total, cancelled int, err error) {
	start, end := parseDateRange(startDate, endDate)
	var minSeq, maxSeq sql.NullInt64
	query := `
		SELECT 
			MIN(credit_note_sequence) as min_val,
			MAX(credit_note_sequence) as max_val,
			COUNT(id) as total,
			COUNT(id) FILTER (WHERE status = 'CANCELLED') as cancelled
		FROM b2b_credit_notes
		WHERE note_date >= ? AND note_date <= ? AND status IN ('ISSUED', 'CANCELLED')
	`
	row := r.db.Raw(query, start, end).Row()
	err = row.Scan(&minSeq, &maxSeq, &total, &cancelled)
	if err != nil {
		return
	}
	if minSeq.Valid {
		var minNum string
		if err := r.db.Table("b2b_credit_notes").
			Where("note_date >= ? AND note_date <= ? AND credit_note_sequence = ?", start, end, minSeq.Int64).
			Pluck("credit_note_number", &minNum).Error; err == nil && minNum != "" {
			minNote = &minNum
		}
	}
	if maxSeq.Valid {
		var maxNum string
		if err := r.db.Table("b2b_credit_notes").
			Where("note_date >= ? AND note_date <= ? AND credit_note_sequence = ?", start, end, maxSeq.Int64).
			Pluck("credit_note_number", &maxNum).Error; err == nil && maxNum != "" {
			maxNote = &maxNum
		}
	}
	return
}

func (r *gormGSTRepository) GetB2BDebitNotesIssued(startDate, endDate string) (minNote, maxNote *string, total, cancelled int, err error) {
	start, end := parseDateRange(startDate, endDate)
	var minSeq, maxSeq sql.NullInt64
	query := `
		SELECT 
			MIN(debit_note_sequence) as min_val,
			MAX(debit_note_sequence) as max_val,
			COUNT(id) as total,
			COUNT(id) FILTER (WHERE status = 'CANCELLED') as cancelled
		FROM b2b_debit_notes
		WHERE note_date >= ? AND note_date <= ? AND status IN ('ISSUED', 'CANCELLED')
	`
	row := r.db.Raw(query, start, end).Row()
	err = row.Scan(&minSeq, &maxSeq, &total, &cancelled)
	if err != nil {
		return
	}
	if minSeq.Valid {
		var minNum string
		if err := r.db.Table("b2b_debit_notes").
			Where("note_date >= ? AND note_date <= ? AND debit_note_sequence = ?", start, end, minSeq.Int64).
			Pluck("debit_note_number", &minNum).Error; err == nil && minNum != "" {
			minNote = &minNum
		}
	}
	if maxSeq.Valid {
		var maxNum string
		if err := r.db.Table("b2b_debit_notes").
			Where("note_date >= ? AND note_date <= ? AND debit_note_sequence = ?", start, end, maxSeq.Int64).
			Pluck("debit_note_number", &maxNum).Error; err == nil && maxNum != "" {
			maxNote = &maxNum
		}
	}
	return
}

func parseDateRange(startDate, endDate string) (time.Time, time.Time) {
	start := parseISO(startDate)
	end := parseISO(endDate)

	if start.IsZero() {
		now := time.Now()
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}
	if end.IsZero() {
		end = time.Now()
	}
	return start, end
}

func parseISO(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return t
	}
	t, err = time.Parse("2006-01-02T15:04:05.000Z", s)
	if err == nil {
		return t
	}
	return time.Time{}
}
