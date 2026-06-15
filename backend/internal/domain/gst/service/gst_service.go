package service

import (
	"fmt"
	"strings"
	"time"

	"mi-tech/internal/domain/gst/dto"
	"mi-tech/internal/domain/gst/repository"
)

// GSTService handles GST report business logic.
type GSTService struct {
	gstRepo repository.GSTRepository
}

// NewGSTService creates a new GSTService.
func NewGSTService(gstRepo repository.GSTRepository) *GSTService {
	return &GSTService{gstRepo: gstRepo}
}

// GetGSTSummary computes the full GST summary including CGST/SGST/IGST splits by state.
func (s *GSTService) GetGSTSummary(startDate, endDate string) (dto.GSTSummaryResponse, error) {
	res, err := s.gstRepo.GetGSTSummary(startDate, endDate)
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
func (s *GSTService) GetStateSummary(startDate, endDate string) ([]dto.StateSummaryRow, error) {
	results, err := s.gstRepo.GetStateSummary(startDate, endDate)
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
func (s *GSTService) GetHSNSummary(startDate, endDate string) ([]dto.HSNSummaryRow, error) {
	results, err := s.gstRepo.GetHSNSummary(startDate, endDate)
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
func (s *GSTService) GetDocumentsIssued(startDate, endDate string) ([]dto.DocumentIssuedRow, error) {
	minOrder, maxOrder, total, cancelled, err := s.gstRepo.GetShopifyDocumentsIssued(startDate, endDate)
	if err != nil {
		return nil, err
	}

	var rows []dto.DocumentIssuedRow
	if total > 0 {
		fromS := ""
		toS := ""
		if minOrder != nil {
			fromS = fmt.Sprintf("SY-%d", *minOrder)
		}
		if maxOrder != nil {
			toS = fmt.Sprintf("SY-%d", *maxOrder)
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

	amzMin, amzMax, amzTotal, amzCancelled, err := s.gstRepo.GetAmazonDocumentsIssued(startDate, endDate)
	if err == nil && amzTotal > 0 {
		fromS := "N/A"
		toS := "N/A"
		if amzMin != nil {
			fromS = fmt.Sprintf("AMZ-%d", *amzMin)
		}
		if amzMax != nil {
			toS = fmt.Sprintf("AMZ-%d", *amzMax)
		}
		rows = append(rows, dto.DocumentIssuedRow{
			DocumentType: "Tax Invoice (Amazon)",
			FromSerial:   fromS,
			ToSerial:     toS,
			TotalIssued:  amzTotal,
			Cancelled:    amzCancelled,
			NetIssued:    amzTotal - amzCancelled,
		})
	}

	return rows, nil
}

// isTamilNadu checks if a state string refers to Tamil Nadu for CGST/SGST vs IGST split.
func isTamilNadu(state string) bool {
	s := strings.TrimSpace(state)
	return len(s) > 0 && (s == "Tamil Nadu" || s == "TN" || strings.EqualFold(s, "tamil nadu"))
}

func (s *GSTService) GetGSTR1JSON(startDate, endDate string, gstin string) (dto.GSTR1Payload, error) {
	// 1. Fetch B2CS
	b2csRows, err := s.gstRepo.GetGSTR1B2CS(startDate, endDate)
	if err != nil {
		return dto.GSTR1Payload{}, err
	}

	// 2. Fetch HSN summary
	hsnRows, err := s.gstRepo.GetGSTR1HSN(startDate, endDate)
	if err != nil {
		return dto.GSTR1Payload{}, err
	}

	// 3. Fetch Doc Issue stats
	shMin, shMax, shTotal, shCancelled, err := s.gstRepo.GetShopifyDocumentsIssued(startDate, endDate)
	if err != nil {
		return dto.GSTR1Payload{}, err
	}
	amzMin, amzMax, amzTotal, amzCancelled, err := s.gstRepo.GetAmazonDocumentsIssued(startDate, endDate)
	if err != nil {
		return dto.GSTR1Payload{}, err
	}

	// Format Doc Issue category
	var docs []dto.DocRange
	docIdx := 1
	if shTotal > 0 {
		fromS := ""
		toS := ""
		if shMin != nil {
			fromS = fmt.Sprintf("SY-%d", *shMin)
		}
		if shMax != nil {
			toS = fmt.Sprintf("SY-%d", *shMax)
		}
		docs = append(docs, dto.DocRange{
			Num:      docIdx,
			From:     fromS,
			To:       toS,
			TotNum:   shTotal,
			Cancel:   shCancelled,
			NetIssue: shTotal - shCancelled,
		})
		docIdx++
	}

	if amzTotal > 0 {
		fromS := ""
		toS := ""
		if amzMin != nil {
			fromS = fmt.Sprintf("AMZ-%d", *amzMin)
		}
		if amzMax != nil {
			toS = fmt.Sprintf("AMZ-%d", *amzMax)
		}
		docs = append(docs, dto.DocRange{
			Num:      docIdx,
			From:     fromS,
			To:       toS,
			TotNum:   amzTotal,
			Cancel:   amzCancelled,
			NetIssue: amzTotal - amzCancelled,
		})
	}

	var docCategories []dto.DocCategory
	if len(docs) > 0 {
		docCategories = append(docCategories, dto.DocCategory{
			DocNum: 1, // Category 1: Invoices for outward supply
			Docs:   docs,
		})
	}

	// Determine filing period (fp) format: MMYYYY
	fp := ""
	endPeriod := parseISO(endDate)
	if !endPeriod.IsZero() {
		fp = endPeriod.Format("012006") // e.g. "062026"
	} else {
		// fallback to current month
		fp = time.Now().Format("012006")
	}

	payload := dto.GSTR1Payload{
		GSTIN:   gstin,
		FP:      fp,
		Version: "v1.0",
		B2CS:    b2csRows,
		HSN: dto.HSNWrapper{
			Data: hsnRows,
		},
		DocIssue: dto.DocIssueWrapper{
			DocDet: docCategories,
		},
	}

	return payload, nil
}

// helper copied/defined internally for date parsing
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
