package service

import (
	"fmt"
	"strings"

	"mi-tech/internal/domain/dashboard/dto"
	"mi-tech/internal/domain/dashboard/repository"
)

// ReportService handles GST report business logic.
type ReportService struct {
	reportRepo repository.ReportRepository
}

// NewReportService creates a new ReportService.
func NewReportService(reportRepo repository.ReportRepository) *ReportService {
	return &ReportService{reportRepo: reportRepo}
}

// GetGSTSummary computes the full GST summary including CGST/SGST/IGST splits by state.
func (s *ReportService) GetGSTSummary(startDate, endDate string) (dto.GSTSummaryResponse, error) {
	res, err := s.reportRepo.GetGSTSummary(startDate, endDate)
	if err != nil {
		return dto.GSTSummaryResponse{}, err
	}

	summary := dto.GSTSummaryResponse{
		TotalOrders:       res.TotalOrders,
		CancelledOrders:   res.CancelledOrders,
		InvoicesGenerated: res.TotalOrders - res.CancelledOrders,
		TotalRevenue:      res.TotalRevenue,
		TotalTaxableValue: res.TotalTaxable,
		TotalGSTCollected: res.TotalTax,
		TotalCGST:         res.CGST,
		TotalSGST:         res.SGST,
		TotalIGST:         res.IGST,
		FulfilledOrders:   res.FulfilledOrders,
		UnfulfilledOrders: res.UnfulfilledOrders,
		PaidOrders:        res.PaidOrders,
	}

	return summary, nil
}

// GetStateSummary returns per-state revenue and tax breakdown.
func (s *ReportService) GetStateSummary(startDate, endDate string) ([]dto.StateSummaryRow, error) {
	results, err := s.reportRepo.GetStateSummary(startDate, endDate)
	if err != nil {
		return nil, err
	}

	var rows []dto.StateSummaryRow
	for _, r := range results {
		row := dto.StateSummaryRow{
			State:        r.State,
			Orders:       r.Orders,
			TaxableValue: r.TaxableValue,
			TotalGST:     r.TotalGST,
			Revenue:      r.Revenue,
		}
		if isTamilNadu(r.State) {
			row.CGST = r.TotalGST / 2
			row.SGST = r.TotalGST / 2
		} else {
			row.IGST = r.TotalGST
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// GetHSNSummary returns per-HSN code revenue and tax breakdown.
func (s *ReportService) GetHSNSummary(startDate, endDate string) ([]dto.HSNSummaryRow, error) {
	results, err := s.reportRepo.GetHSNSummary(startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Aggregate by HSN code (results are per HSN+State)
	hsnMap := make(map[string]*dto.HSNSummaryRow)
	for _, r := range results {
		hsn := r.HSNCode
		if hsn == "" {
			hsn = "33029019"
		}

		if _, ok := hsnMap[hsn]; !ok {
			hsnMap[hsn] = &dto.HSNSummaryRow{HSNCode: hsn}
		}
		row := hsnMap[hsn]
		row.ProductCount += r.ProductCount
		row.QtySold += r.QtySold
		row.TaxableValue += r.TaxableValue
		row.TotalGST += r.TotalGST
		row.Revenue += r.Revenue

		if isTamilNadu(r.State) {
			row.CGST += r.TotalGST / 2
			row.SGST += r.TotalGST / 2
		} else {
			row.IGST += r.TotalGST
		}
	}

	var rows []dto.HSNSummaryRow
	for _, v := range hsnMap {
		rows = append(rows, *v)
	}
	return rows, nil
}

// GetDocumentsIssued returns the documents issued report.
func (s *ReportService) GetDocumentsIssued(startDate, endDate string) ([]dto.DocumentIssuedRow, error) {
	minOrder, maxOrder, total, cancelled, err := s.reportRepo.GetDocumentsIssued(startDate, endDate)
	if err != nil {
		return nil, err
	}

	var rows []dto.DocumentIssuedRow
	if total > 0 {
		fromS := ""
		toS := ""
		if minOrder != nil {
			fromS = fmt.Sprintf("INV-%d", *minOrder)
		}
		if maxOrder != nil {
			toS = fmt.Sprintf("INV-%d", *maxOrder)
		}
		rows = append(rows, dto.DocumentIssuedRow{
			DocumentType: "Tax Invoice",
			FromSerial:   fromS,
			ToSerial:     toS,
			TotalIssued:  total,
			Cancelled:    cancelled,
			NetIssued:    total - cancelled,
		})
	}
	return rows, nil
}

// isTamilNadu checks if a state string refers to Tamil Nadu for CGST/SGST vs IGST split.
func isTamilNadu(state string) bool {
	s := strings.TrimSpace(state)
	return len(s) > 0 && (s == "Tamil Nadu" || s == "TN" || strings.EqualFold(s, "tamil nadu"))
}
