package repository

import (
	"mi-tech/internal/domain/dashboard/dto"
)

type GSTSummaryResult struct {
	TotalOrders       int
	CancelledOrders   int
	FulfilledOrders   int
	UnfulfilledOrders int
	PaidOrders        int
	TotalRevenue      float64
	TotalTaxable      float64
	TotalTax          float64
	CGST              float64
	SGST              float64
	IGST              float64
}

type StateSummaryResult struct {
	State        string
	Orders       int
	TaxableValue float64
	TotalGST     float64
	Revenue      float64
}

type HSNSummaryResult struct {
	HSNCode      string
	State        string
	ProductCount int
	QtySold      int
	TaxableValue float64
	TotalGST     float64
	Revenue      float64
}

// MetricsRepository defines data access for dashboard metric queries.
type MetricsRepository interface {
	GetDashboardMetrics(startDate, endDate string, sourceIDs []string) (dto.DashboardMetrics, error)
	GetTopProducts(startDate, endDate string, sourceIDs []string, limit int) ([]dto.TopProductRow, error)
	GetRevenueTrend(startDate, endDate string, sourceIDs []string) ([]dto.RevenueTrendRow, error)
	GetGeoDistribution(startDate, endDate string, sourceIDs []string, limit int) ([]dto.GeoDistributionRow, error)
}

// ReportRepository defines data access for GST report queries.
type ReportRepository interface {
	GetGSTSummary(startDate, endDate string) (GSTSummaryResult, error)
	GetStateSummary(startDate, endDate string) (results []StateSummaryResult, err error)
	GetHSNSummary(startDate, endDate string) (results []HSNSummaryResult, err error)
	GetShopifyDocumentsIssued(startDate, endDate string) (minOrder, maxOrder *int64, total, cancelled int, err error)
	GetAmazonDocumentsIssued(startDate, endDate string) (minOrder, maxOrder *int64, total, cancelled int, err error)
}
