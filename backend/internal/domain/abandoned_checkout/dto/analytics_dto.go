package dto

type WhatsappStats struct {
	Sent      int64 `json:"sent"`
	Delivered int64 `json:"delivered"`
	Read      int64 `json:"read"`
	Clicked   int64 `json:"clicked"`
	Failed    int64 `json:"failed"`
}

type RevenueTimelineItem struct {
	Date            string  `json:"date"`
	AbandonedAmount float64 `json:"abandonedAmount"`
	RecoveredAmount float64 `json:"recoveredAmount"`
}

type StatusBreakdownItem struct {
	Status string  `json:"status"`
	Count  int64   `json:"count"`
	Amount float64 `json:"amount"`
}

type TopLostCartItem struct {
	CustomerName   string  `json:"customer_name"`
	Phone          string  `json:"phone"`
	TotalPrice     float64 `json:"total_price"`
	Currency       string  `json:"currency"`
	AbandonedAt    string  `json:"abandoned_at"`
	RecoveryStatus string  `json:"recovery_status"`
	Attempts       int     `json:"recovery_attempts"`
}

type AbandonedCheckoutAnalyticsResponse struct {
	TotalAbandonedRevenue float64               `json:"totalAbandonedRevenue"`
	RecoveredRevenue      float64               `json:"recoveredRevenue"`
	PendingRevenue        float64               `json:"pendingRevenue"`
	AbandonedCartCount    int64                 `json:"abandonedCartCount"`
	RecoveredCartCount    int64                 `json:"recoveredCartCount"`
	RecoveryRate          float64               `json:"recoveryRate"`
	CartsCreatedCount     int64                 `json:"cartsCreatedCount"`
	AddCartToCheckoutRate float64               `json:"addCartToCheckoutRate"`
	AddCartToOrderRate    float64               `json:"addCartToOrderRate"`
	WhatsappStats         WhatsappStats         `json:"whatsappStats"`
	RevenueTimeline       []RevenueTimelineItem `json:"revenueTimeline"`
	StatusBreakdown       []StatusBreakdownItem `json:"statusBreakdown"`
	TopLostCarts          []TopLostCartItem     `json:"topLostCarts"`
}
