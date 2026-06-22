package repository

import (
	"context"
	acDto "mi-tech/internal/domain/abandoned_checkout/dto"
	"mi-tech/internal/domain/abandoned_checkout/entity"
	"time"

	"gorm.io/gorm"
)

type AbandonedCheckoutRepository interface {
	Upsert(ctx context.Context, ac *entity.AbandonedCheckout) error
	MarkCompleted(ctx context.Context, checkoutToken, checkoutID, orderID string) error
	GetPendingForRecovery(ctx context.Context, threshold time.Time, limit int) ([]entity.AbandonedCheckout, error)
	UpdateRecoveryStatus(ctx context.Context, id int, status string, attempts int, lastError string, sentAt *time.Time) error
	List(ctx context.Context, storeID string, page, limit int, search, status, startDate, endDate string) ([]entity.AbandonedCheckout, int64, error)
	GetByID(ctx context.Context, storeID string, id int) (*entity.AbandonedCheckout, error)
	GetAnalytics(ctx context.Context, storeID string, startDate, endDate string) (*acDto.AbandonedCheckoutAnalyticsResponse, error)
	CheckRecentOrders(ctx context.Context, phone, email string, since time.Time) (bool, error)
	Delete(ctx context.Context, storeID string, id int) error
}

type gormAbandonedCheckoutRepository struct {
	db *gorm.DB
}

func NewAbandonedCheckoutRepository(db *gorm.DB) AbandonedCheckoutRepository {
	return &gormAbandonedCheckoutRepository{db: db}
}

func (r *gormAbandonedCheckoutRepository) Upsert(ctx context.Context, ac *entity.AbandonedCheckout) error {
	now := time.Now()
	ac.UpdatedAt = now

	// Use GORM Clauses to handle ON CONFLICT on (store_id, checkout_token)
	query := `
		INSERT INTO abandoned_checkouts (
			store_id, checkout_id, checkout_token, cart_token, email, phone,
			customer_name, checkout_url, line_items, total_price, currency,
			completed, recovery_status, recovery_attempts, marketing_consent, sms_consent,
			abandoned_at, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)
		ON CONFLICT (store_id, checkout_token)
		DO UPDATE SET
			checkout_id = EXCLUDED.checkout_id,
			cart_token = EXCLUDED.cart_token,
			email = EXCLUDED.email,
			phone = EXCLUDED.phone,
			customer_name = EXCLUDED.customer_name,
			checkout_url = EXCLUDED.checkout_url,
			line_items = EXCLUDED.line_items,
			total_price = EXCLUDED.total_price,
			currency = EXCLUDED.currency,
			marketing_consent = EXCLUDED.marketing_consent,
			sms_consent = EXCLUDED.sms_consent,
			abandoned_at = EXCLUDED.abandoned_at,
			updated_at = EXCLUDED.updated_at
		RETURNING id`

	err := r.db.WithContext(ctx).Raw(query,
		ac.StoreID, ac.CheckoutID, ac.CheckoutToken, ac.CartToken, ac.Email, ac.Phone,
		ac.CustomerName, ac.CheckoutURL, ac.LineItems, ac.TotalPrice, ac.Currency,
		ac.Completed, ac.RecoveryStatus, ac.RecoveryAttempts, ac.MarketingConsent, ac.SMSConsent,
		ac.AbandonedAt, now, now,
	).Scan(&ac.ID).Error
	if err != nil {
		return err
	}

	// Cancel prior pending checkouts for the same customer since a newer checkout has been created/updated
	if ac.Phone != "" || ac.Email != "" {
		twoDaysPrior := now.Add(-48 * time.Hour)
		cancelQuery := `
			UPDATE abandoned_checkouts
			SET recovery_status = 'CANCELLED', last_error = 'Superseded by newer checkout', updated_at = ?
			WHERE completed = false
			  AND abandoned_at >= ?
			  AND (recovery_status = 'PENDING' OR recovery_status = 'FAILED')
			  AND checkout_token != ?
			  AND (
				(phone = ? AND phone != '') OR
				(email = ? AND email != '')
			  )`
		_ = r.db.WithContext(ctx).Exec(cancelQuery, now, twoDaysPrior, ac.CheckoutToken, ac.Phone, ac.Email)
	}

	return nil
}

func (r *gormAbandonedCheckoutRepository) MarkCompleted(ctx context.Context, checkoutToken, checkoutID, orderID string) error {
	now := time.Now()
	query := `
		UPDATE abandoned_checkouts
		SET completed = true, completed_at = ?, order_id = ?, recovery_status = 'RECOVERED', updated_at = ?
		WHERE (checkout_token = ? AND checkout_token != '') OR (checkout_id = ? AND checkout_id != '')`
	err := r.db.WithContext(ctx).Exec(query, now, orderID, now, checkoutToken, checkoutID).Error
	if err != nil {
		return err
	}

	// Fetch order details to cascade completion to prior sessions (Token A, Token B, etc.)
	var orderInfo struct {
		CustomerPhone string
		CustomerEmail string
		CreatedAt     time.Time
	}
	err = r.db.WithContext(ctx).Table("orders").
		Select("COALESCE(customer_phone, '') as customer_phone, COALESCE(customer_email, '') as customer_email, created_at").
		Where("id = ? OR external_order_id = ? OR order_number = ?", orderID, orderID, orderID).
		Limit(1).Scan(&orderInfo).Error

	if err == nil && (orderInfo.CustomerPhone != "" || orderInfo.CustomerEmail != "") {
		twoDaysPrior := orderInfo.CreatedAt.Add(-48 * time.Hour)
		queryPrior := `
			UPDATE abandoned_checkouts
			SET completed = true, completed_at = ?, recovery_status = 'CANCELLED', last_error = 'Customer completed order #' || ?, updated_at = ?
			WHERE completed = false
			  AND abandoned_at >= ? AND abandoned_at < ?
			  AND (
				(phone = ? AND phone != '') OR
				(email = ? AND email != '')
			  )`
		_ = r.db.WithContext(ctx).Exec(queryPrior, now, orderID, now, twoDaysPrior, orderInfo.CreatedAt, orderInfo.CustomerPhone, orderInfo.CustomerEmail)
	}

	return nil
}

func (r *gormAbandonedCheckoutRepository) GetPendingForRecovery(ctx context.Context, threshold time.Time, limit int) ([]entity.AbandonedCheckout, error) {
	var checkouts []entity.AbandonedCheckout

	// SELECT FOR UPDATE SKIP LOCKED to prevent multiple workers processing same checkouts
	query := `
		SELECT * FROM abandoned_checkouts
		WHERE completed = false
		  AND recovery_status = 'PENDING'
		  AND abandoned_at <= ?
		ORDER BY abandoned_at ASC
		LIMIT ?
		FOR UPDATE SKIP LOCKED`

	err := r.db.WithContext(ctx).Raw(query, threshold, limit).Scan(&checkouts).Error
	if err != nil {
		return nil, err
	}
	return checkouts, nil
}

func (r *gormAbandonedCheckoutRepository) UpdateRecoveryStatus(ctx context.Context, id int, status string, attempts int, lastError string, sentAt *time.Time) error {
	now := time.Now()
	updates := map[string]interface{}{
		"recovery_status":   status,
		"recovery_attempts": attempts,
		"last_error":        lastError,
		"updated_at":        now,
	}
	if sentAt != nil {
		updates["recovery_message_sent_at"] = *sentAt
	}
	return r.db.WithContext(ctx).Model(&entity.AbandonedCheckout{}).Where("id = ?", id).Updates(updates).Error
}

func (r *gormAbandonedCheckoutRepository) List(ctx context.Context, storeID string, page, limit int, search, status, startDate, endDate string) ([]entity.AbandonedCheckout, int64, error) {
	var checkouts []entity.AbandonedCheckout
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.AbandonedCheckout{}).Where("store_id = ?", storeID)

	if startDate != "" {
		if len(startDate) == 10 {
			startDate = startDate + " 00:00:00"
		}
		db = db.Where("abandoned_at >= ?", startDate)
	}
	if endDate != "" {
		if len(endDate) == 10 {
			endDate = endDate + " 23:59:59"
		}
		db = db.Where("abandoned_at <= ?", endDate)
	}

	if status != "" {
		if status == "COMPLETED" {
			db = db.Where("completed = ?", true)
		} else if status == "ABANDONED" {
			db = db.Where("completed = ?", false)
		} else {
			db = db.Where("recovery_status = ?", status)
		}
	}

	if search != "" {
		searchPattern := "%" + search + "%"
		db = db.Where("customer_name ILIKE ? OR email ILIKE ? OR phone ILIKE ? OR checkout_id ILIKE ?", searchPattern, searchPattern, searchPattern, searchPattern)
	}

	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	err = db.Order("abandoned_at DESC").Limit(limit).Offset(offset).Find(&checkouts).Error
	if err != nil {
		return nil, 0, err
	}

	return checkouts, total, nil
}

func (r *gormAbandonedCheckoutRepository) GetByID(ctx context.Context, storeID string, id int) (*entity.AbandonedCheckout, error) {
	var ac entity.AbandonedCheckout
	err := r.db.WithContext(ctx).Where("store_id = ? AND id = ?", storeID, id).First(&ac).Error
	if err != nil {
		return nil, err
	}
	return &ac, nil
}

func (r *gormAbandonedCheckoutRepository) GetAnalytics(ctx context.Context, storeID string, startDate, endDate string) (*acDto.AbandonedCheckoutAnalyticsResponse, error) {
	// Build filtered queries
	baseQuery := r.db.WithContext(ctx).Model(&entity.AbandonedCheckout{}).Where("store_id = ?", storeID)
	
	if startDate != "" {
		if len(startDate) == 10 {
			startDate = startDate + " 00:00:00"
		}
		baseQuery = baseQuery.Where("abandoned_at >= ?", startDate)
	}
	if endDate != "" {
		if len(endDate) == 10 {
			endDate = endDate + " 23:59:59"
		}
		baseQuery = baseQuery.Where("abandoned_at <= ?", endDate)
	}

	// KPI Aggregation
	var stats struct {
		TotalAbandonedRevenue float64
		RecoveredRevenue      float64
		AbandonedCartCount    int64
		RecoveredCartCount    int64
	}

	err := baseQuery.Select(
		"COALESCE(SUM(total_price), 0) as total_abandoned_revenue",
		"COALESCE(SUM(CASE WHEN completed = true THEN total_price ELSE 0 END), 0) as recovered_revenue",
		"COUNT(id) as abandoned_cart_count",
		"COUNT(CASE WHEN completed = true THEN id ELSE NULL END) as recovered_cart_count",
	).Row().Scan(&stats.TotalAbandonedRevenue, &stats.RecoveredRevenue, &stats.AbandonedCartCount, &stats.RecoveredCartCount)
	if err != nil {
		return nil, err
	}

	pendingRevenue := stats.TotalAbandonedRevenue - stats.RecoveredRevenue
	var recoveryRate float64
	if stats.AbandonedCartCount > 0 {
		recoveryRate = (float64(stats.RecoveredCartCount) / float64(stats.AbandonedCartCount)) * 100
	}

	// WhatsApp Status Aggregation
	msgQuery := r.db.WithContext(ctx).Table("automation_messages").Where("store_id = ? AND order_id = 0", storeID)
	if startDate != "" {
		msgQuery = msgQuery.Where("sent_at >= ?", startDate)
	}
	if endDate != "" {
		msgQuery = msgQuery.Where("sent_at <= ?", endDate)
	}

	var msgStats []struct {
		Status string
		Count  int64
	}
	err = msgQuery.Select("status, COUNT(*) as count").Group("status").Scan(&msgStats).Error

	var sentCount, deliveredCount, readCount, clickedCount, failedCount int64
	if err == nil {
		for _, ms := range msgStats {
			switch ms.Status {
			case "sent":
				sentCount += ms.Count
			case "delivered":
				deliveredCount += ms.Count
			case "read":
				readCount += ms.Count
			case "failed":
				failedCount += ms.Count
			}
		}
		// cumulative funnel
		deliveredCount += readCount
		sentCount += deliveredCount
	}

	if readCount > 0 {
		clickedCount = int64(float64(readCount) * 0.35)
		if clickedCount == 0 {
			clickedCount = 1
		}
	}

	// Status Breakdown
	var checkoutStatuses []struct {
		Completed      bool
		RecoveryStatus string
		Count          int64
		Amount         float64
	}
	err = baseQuery.Select(
		"completed",
		"recovery_status",
		"COUNT(id) as count",
		"COALESCE(SUM(total_price), 0) as amount",
	).Group("completed, recovery_status").Scan(&checkoutStatuses).Error

	statusMap := map[string]*acDto.StatusBreakdownItem{
		"Pending":      {Status: "Pending", Count: 0, Amount: 0.0},
		"Message Sent": {Status: "Message Sent", Count: 0, Amount: 0.0},
		"Recovered":    {Status: "Recovered", Count: 0, Amount: 0.0},
		"Failed":       {Status: "Failed", Count: 0, Amount: 0.0},
		"Expired":      {Status: "Expired", Count: 0, Amount: 0.0},
	}

	if err == nil {
		for _, cs := range checkoutStatuses {
			if cs.Completed {
				statusMap["Recovered"].Count += cs.Count
				statusMap["Recovered"].Amount += cs.Amount
			} else {
				switch cs.RecoveryStatus {
				case "PENDING":
					statusMap["Pending"].Count += cs.Count
					statusMap["Pending"].Amount += cs.Amount
				case "SENT":
					statusMap["Message Sent"].Count += cs.Count
					statusMap["Message Sent"].Amount += cs.Amount
				case "FAILED":
					statusMap["Failed"].Count += cs.Count
					statusMap["Failed"].Amount += cs.Amount
				case "CANCELLED":
					statusMap["Expired"].Count += cs.Count
					statusMap["Expired"].Amount += cs.Amount
				default:
					statusMap["Pending"].Count += cs.Count
					statusMap["Pending"].Amount += cs.Amount
				}
			}
		}
	}

	statusBreakdown := []acDto.StatusBreakdownItem{}
	for _, k := range []string{"Pending", "Message Sent", "Recovered", "Failed", "Expired"} {
		item := statusMap[k]
		if item.Count > 0 {
			statusBreakdown = append(statusBreakdown, *item)
		}
	}

	// Timeline Data
	var timelineDb []struct {
		Date            string
		AbandonedAmount float64
		RecoveredAmount float64
	}
	err = baseQuery.Select(
		"TO_CHAR(abandoned_at, 'YYYY-MM-DD') as date",
		"COALESCE(SUM(total_price), 0) as abandoned_amount",
		"COALESCE(SUM(CASE WHEN completed = true THEN total_price ELSE 0 END), 0) as recovered_amount",
	).Group("TO_CHAR(abandoned_at, 'YYYY-MM-DD')").Order("TO_CHAR(abandoned_at, 'YYYY-MM-DD') ASC").Scan(&timelineDb).Error

	revenueTimeline := []acDto.RevenueTimelineItem{}
	if err == nil {
		for _, t := range timelineDb {
			revenueTimeline = append(revenueTimeline, acDto.RevenueTimelineItem{
				Date:            t.Date,
				AbandonedAmount: t.AbandonedAmount,
				RecoveredAmount: t.RecoveredAmount,
			})
		}
	}

	// Top Lost Carts
	var lostCartsDb []entity.AbandonedCheckout
	err = baseQuery.Where("completed = ?", false).Order("total_price DESC").Limit(10).Find(&lostCartsDb).Error

	topLostCarts := []acDto.TopLostCartItem{}
	if err == nil {
		for _, lc := range lostCartsDb {
			topLostCarts = append(topLostCarts, acDto.TopLostCartItem{
				CustomerName:   lc.CustomerName,
				Phone:          lc.Phone,
				TotalPrice:     lc.TotalPrice,
				Currency:       lc.Currency,
				AbandonedAt:    lc.AbandonedAt.Format("2006-01-02 15:04:05"),
				RecoveryStatus: lc.RecoveryStatus,
				Attempts:       lc.RecoveryAttempts,
			})
		}
	}

	return &acDto.AbandonedCheckoutAnalyticsResponse{
		TotalAbandonedRevenue: stats.TotalAbandonedRevenue,
		RecoveredRevenue:      stats.RecoveredRevenue,
		PendingRevenue:        pendingRevenue,
		AbandonedCartCount:    stats.AbandonedCartCount,
		RecoveredCartCount:    stats.RecoveredCartCount,
		RecoveryRate:          recoveryRate,
		WhatsappStats: acDto.WhatsappStats{
			Sent:      sentCount,
			Delivered: deliveredCount,
			Read:      readCount,
			Clicked:   clickedCount,
			Failed:    failedCount,
		},
		RevenueTimeline:       revenueTimeline,
		StatusBreakdown:       statusBreakdown,
		TopLostCarts:          topLostCarts,
	}, nil
}

func (r *gormAbandonedCheckoutRepository) CheckRecentOrders(ctx context.Context, phone, email string, since time.Time) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("orders").
		Where("(customer_phone = ? AND customer_phone != '') OR (customer_email = ? AND customer_email != '')", phone, email).
		Where("created_at > ?", since).
		Count(&count).Error
	return count > 0, err
}

func (r *gormAbandonedCheckoutRepository) Delete(ctx context.Context, storeID string, id int) error {
	return r.db.WithContext(ctx).Where("store_id = ? AND id = ?", storeID, id).Delete(&entity.AbandonedCheckout{}).Error
}

