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
			COUNT(o.id),
			COUNT(o.id) FILTER (WHERE LOWER(o.status) IN ('cancelled', 'canceled') OR LOWER(COALESCE(o.fulfillment_status, '')) IN ('cancelled', 'canceled')),
			COUNT(o.id) FILTER (WHERE LOWER(o.fulfillment_status) = 'fulfilled'),
			COUNT(o.id) FILTER (WHERE LOWER(COALESCE(o.fulfillment_status, '')) != 'fulfilled' AND NOT (LOWER(COALESCE(o.status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(o.fulfillment_status, '')) IN ('cancelled', 'canceled'))),
			COUNT(o.id) FILTER (WHERE LOWER(o.financial_status) = 'paid'),
			COALESCE(SUM(o.total_price) FILTER (WHERE NOT (LOWER(COALESCE(o.status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(o.fulfillment_status, '')) IN ('cancelled', 'canceled'))), 0) as revenue,
			COALESCE(SUM(ROUND(o.total_price / 1.18, 2)) FILTER (WHERE NOT (LOWER(COALESCE(o.status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(o.fulfillment_status, '')) IN ('cancelled', 'canceled'))), 0) as taxable,
			COALESCE(SUM(o.total_price - ROUND(o.total_price / 1.18, 2)) FILTER (WHERE NOT (LOWER(COALESCE(o.status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(o.fulfillment_status, '')) IN ('cancelled', 'canceled'))), 0) as tax,
			COALESCE(SUM(CASE WHEN COALESCE(s.code, '33') = '33' THEN (o.total_price - ROUND(o.total_price / 1.18, 2)) / 2 ELSE 0 END) FILTER (WHERE NOT (LOWER(COALESCE(o.status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(o.fulfillment_status, '')) IN ('cancelled', 'canceled'))), 0) as cgst,
			COALESCE(SUM(CASE WHEN COALESCE(s.code, '33') = '33' THEN (o.total_price - ROUND(o.total_price / 1.18, 2)) / 2 ELSE 0 END) FILTER (WHERE NOT (LOWER(COALESCE(o.status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(o.fulfillment_status, '')) IN ('cancelled', 'canceled'))), 0) as sgst,
			COALESCE(SUM(CASE WHEN COALESCE(s.code, '33') != '33' THEN (o.total_price - ROUND(o.total_price / 1.18, 2)) ELSE 0 END) FILTER (WHERE NOT (LOWER(COALESCE(o.status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(o.fulfillment_status, '')) IN ('cancelled', 'canceled'))), 0) as igst
		FROM orders o
		LEFT JOIN gst_state_codes s ON LOWER(TRIM(o.customer_state)) = ANY(s.aliases)
		WHERE o.created_at >= ? AND o.created_at <= ?
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
			INITCAP(COALESCE(customer_state, 'N/A')) as state,
			COUNT(id) as orders,
			COALESCE(SUM(ROUND(total_price / 1.18, 2)), 0) as taxable_value,
			COALESCE(SUM(total_price - ROUND(total_price / 1.18, 2)), 0) as total_gst,
			COALESCE(SUM(total_price), 0) as revenue
		FROM orders
		WHERE created_at >= ? AND created_at <= ? AND NOT (LOWER(COALESCE(status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(fulfillment_status, '')) IN ('cancelled', 'canceled'))
		GROUP BY INITCAP(COALESCE(customer_state, 'N/A'))
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
			WHERE o.created_at >= ? AND o.created_at <= ? AND NOT (LOWER(COALESCE(o.status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(o.fulfillment_status, '')) IN ('cancelled', 'canceled'))
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
		WHERE o.created_at >= ? AND o.created_at <= ? AND NOT (LOWER(COALESCE(o.status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(o.fulfillment_status, '')) IN ('cancelled', 'canceled'))
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
				COALESCE(s.code, '33') as pos_code
			FROM order_line_items li
			JOIN orders o ON li.order_id = o.id
			LEFT JOIN gst_state_codes s ON LOWER(TRIM(o.customer_state)) = ANY(s.aliases)
			WHERE o.created_at >= ? AND o.created_at <= ? AND NOT (LOWER(COALESCE(o.status, '')) IN ('cancelled', 'canceled') OR LOWER(COALESCE(o.fulfillment_status, '')) IN ('cancelled', 'canceled'))
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
			ROUND(SUM(CASE WHEN pos_code = '33' THEN (share * order_tax) / 2 ELSE 0 END), 2) as sgst
		FROM CalculatedShares
		GROUP BY hs_code
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
		})
	}

	return rows, nil
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
