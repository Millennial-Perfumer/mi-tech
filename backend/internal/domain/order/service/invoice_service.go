package service

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jung-kurt/gofpdf"

	"mi-tech/internal/domain/order/entity"
	configRepoPkg "mi-tech/internal/shared/config/repository"
)

const (
	DefaultBusinessName   = "PARFUM TRADERS"
	DefaultBusinessGSTIN  = "33AUSPR1909H1ZC"
	DefaultAddressLine1   = "No. 9/21, 1st floor, Sadiq Basha Nagar,"
	DefaultAddressLine2   = "2nd Street, Virugambakkam, Chennai - 600092"
	DefaultBusinessPhone  = "7904769823"
	DefaultFooterBusiness = "PARFUM TRADERS"
)

// InvoiceTotals holds the calculated breakdown for an invoice.
type InvoiceTotals struct {
	GrossSubtotal float64
	OrderDiscount float64
	TaxableValue  float64
	TotalTax      float64
	GrandTotal    float64
}

// CalculateInvoiceTotals performs the shared calculation logic for all order totals.
func (s *InvoiceService) CalculateInvoiceTotals(items []entity.LineItem) InvoiceTotals {
	var t InvoiceTotals
	for _, item := range items {
		qty := float64(item.Quantity)
		lineGross := (item.Price * qty) - item.Discount
		t.GrossSubtotal += lineGross
		t.OrderDiscount += item.OrderDiscount

		lineNet := lineGross - item.OrderDiscount
		if lineNet < 0 {
			lineNet = 0
		}

		lineTaxable := lineNet / 1.18
		lineTax := lineNet - lineTaxable

		t.TaxableValue += lineTaxable
		t.TotalTax += lineTax
	}
	t.GrandTotal = t.TaxableValue + t.TotalTax
	return t
}

// InvoiceService handles PDF invoice generation.
type InvoiceService struct {
	settingsRepo configRepoPkg.SettingsRepository
}

// NewInvoiceService creates a new InvoiceService.
func NewInvoiceService(settingsRepo configRepoPkg.SettingsRepository) *InvoiceService {
	return &InvoiceService{settingsRepo: settingsRepo}
}

func (s *InvoiceService) getSetting(key, defaultValue string) string {
	val, _ := s.settingsRepo.Get(key)
	if val == "" {
		return defaultValue
	}
	return val
}

// safeSetFont attempts to set a custom font (like Montserrat) and falls back to Arial if it's not available.
func (s *InvoiceService) safeSetFont(pdf *gofpdf.Fpdf, family, style string, size float64, hasMontserrat bool) {
	if family == "Montserrat" && !hasMontserrat {
		pdf.SetFont("Arial", style, size)
		return
	}
	pdf.SetFont(family, style, size)
}

// GeneratePDF creates a professional GST invoice PDF and writes it to the provided writer.
func (s *InvoiceService) GeneratePDF(order entity.Order, items []entity.LineItem, w io.Writer) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// -- Font Path Discovery & Fallback --
	// Montserrat is our premium font, but it requires local files.
	// We try multiple paths (useful for tests) and fallback to Arial if missing.
	fontPaths := []string{
		"assets/fonts/",
		"../../assets/fonts/",
		"backend/assets/fonts/",
	}

	regularName := "Montserrat-Regular.ttf"
	boldName := "Montserrat-Bold.ttf"
	semiBoldName := "Montserrat-SemiBold.ttf"

	hasMontserrat := false
	for _, p := range fontPaths {
		if _, err := os.Stat(p + regularName); err == nil {
			pdf.AddUTF8Font("Montserrat", "", p+regularName)
			pdf.AddUTF8Font("Montserrat", "B", p+boldName)
			pdf.AddUTF8Font("Montserrat", "I", p+semiBoldName)
			hasMontserrat = true
			break
		}
	}

	// -- Header --
	bizName := s.getSetting("business_name", DefaultBusinessName)
	gstin := s.getSetting("business_gstin", DefaultBusinessGSTIN)
	addr1 := s.getSetting("business_address_line1", DefaultAddressLine1)
	addr2 := s.getSetting("business_address_line2", DefaultAddressLine2)
	phone := s.getSetting("business_phone", DefaultBusinessPhone)

	s.safeSetFont(pdf, "Montserrat", "B", 13.5, hasMontserrat)
	pdf.CellFormat(100, 10, bizName, "0", 0, "L", false, 0, "")
	pdf.CellFormat(80, 10, "TAX INVOICE", "0", 1, "R", false, 0, "")

	s.safeSetFont(pdf, "Montserrat", "", 7.5, hasMontserrat)
	pdf.Ln(2)
	pdf.CellFormat(100, 4, "GSTIN: "+gstin, "0", 1, "L", false, 0, "")
	pdf.Ln(1)
	pdf.CellFormat(100, 4, addr1, "0", 1, "L", false, 0, "")
	pdf.CellFormat(100, 4, addr2, "0", 1, "L", false, 0, "")
	pdf.Ln(1)
	pdf.CellFormat(100, 4, "Phone: "+phone, "0", 1, "L", false, 0, "")
	pdf.Ln(6.5)

	// -- Invoice & Customer Details --
	leftCol := 90.0
	rightCol := 90.0

	s.safeSetFont(pdf, "Arial", "B", 9, hasMontserrat)
	pdf.CellFormat(leftCol, 6, "INVOICE DETAILS", "B", 0, "L", false, 0, "")
	pdf.CellFormat(rightCol, 6, "BILL TO (CUSTOMER)", "B", 1, "L", false, 0, "")
	pdf.Ln(3)

	metaYStart := pdf.GetY()

	// Left side: Invoice meta
	s.safeSetFont(pdf, "Montserrat", "B", 7.5, hasMontserrat)
	pdf.CellFormat(30, 4, "Invoice No:", "0", 0, "L", false, 0, "")
	s.safeSetFont(pdf, "Montserrat", "", 7.5, hasMontserrat)
	pdf.CellFormat(60, 4, "INV-"+order.OrderNumber, "0", 1, "L", false, 0, "")

	s.safeSetFont(pdf, "Montserrat", "B", 7.5, hasMontserrat)
	pdf.CellFormat(30, 4, "Order No:", "0", 0, "L", false, 0, "")
	s.safeSetFont(pdf, "Montserrat", "", 7.5, hasMontserrat)
	pdf.CellFormat(60, 4, order.OrderNumber, "0", 1, "L", false, 0, "")

	s.safeSetFont(pdf, "Montserrat", "B", 7.5, hasMontserrat)
	pdf.CellFormat(30, 4, "Date:", "0", 0, "L", false, 0, "")
	s.safeSetFont(pdf, "Montserrat", "", 7.5, hasMontserrat)
	pdf.CellFormat(60, 4, order.CreatedAt.Format("2006-01-02"), "0", 1, "L", false, 0, "")

	// Right side: Customer info
	pdf.SetY(metaYStart)
	pdf.SetX(leftCol + 15)

	displayName := ns(order.CustomerName)
	if strings.TrimSpace(displayName) == "" {
		displayName = "Valued Customer"
	}
	s.safeSetFont(pdf, "Montserrat", "B", 8.25, hasMontserrat)
	pdf.SetX(leftCol + 15)
	pdf.CellFormat(rightCol, 4, displayName, "0", 1, "L", false, 0, "")

	s.safeSetFont(pdf, "Montserrat", "", 7.5, hasMontserrat)
	if email := ns(order.CustomerEmail); email != "" {
		pdf.SetX(leftCol + 15)
		pdf.CellFormat(rightCol, 4, email, "0", 1, "L", false, 0, "")
	}
	if phone := ns(order.CustomerPhone); phone != "" {
		pdf.SetX(leftCol + 15)
		pdf.CellFormat(rightCol, 4, "WhatsApp: "+phone, "0", 1, "L", false, 0, "")
	}

	addressLine := ""
	if city := ns(order.CustomerCity); city != "" {
		addressLine += city + ", "
	}
	if state := ns(order.CustomerState); state != "" {
		addressLine += state + ", "
	}
	if country := ns(order.CustomerCountry); country != "" {
		addressLine += country
	}
	pdf.SetX(leftCol + 15)
	pdf.CellFormat(rightCol, 4, addressLine, "0", 1, "L", false, 0, "")
	pdf.Ln(4.8)

	// -- Items Table --
	totals := s.CalculateInvoiceTotals(items)
	s.renderItemsTable(pdf, items, hasMontserrat)

	// -- Totals --
	s.renderTotals(pdf, order, items, totals, hasMontserrat)

	// -- Footer Terms --
	s.renderFooter(pdf, hasMontserrat)

	return pdf.Output(w)
}

func (s *InvoiceService) renderItemsTable(pdf *gofpdf.Fpdf, items []entity.LineItem, hasMontserrat bool) {
	pdf.SetFillColor(245, 245, 245)
	s.safeSetFont(pdf, "Montserrat", "I", 7.5, hasMontserrat)

	wName := 76.2
	wSKU := 14.5
	wHSN := 16.3
	wQty := 9.1
	wPrice := 16.3
	wDiscount := 12.7
	wTaxable := 16.3
	wGSTPct := 10.9
	wGSTAmt := 16.3

	hHeader := 9.0
	pdf.CellFormat(wName, hHeader, "Product Name", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wSKU, hHeader, "SKU", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wHSN, hHeader, "HSN", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wQty, hHeader, "Qty", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wPrice, hHeader, "Price", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wDiscount, hHeader, "Disc", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wTaxable, hHeader, "Taxable", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wGSTPct, hHeader, "GST %", "1", 0, "C", true, 0, "")
	pdf.CellFormat(wGSTAmt, hHeader, "GST Amt", "1", 1, "C", true, 0, "")

	s.safeSetFont(pdf, "Montserrat", "", 6.75, hasMontserrat)

	totalTaxable := 0.0
	totalTax := 0.0

	for _, item := range items {
		rawPrice := item.Price
		itemDiscount := item.Discount
		orderDiscount := item.OrderDiscount
		qty := float64(item.Quantity)

		lineTotal := (rawPrice * qty) - itemDiscount - orderDiscount
		if lineTotal < 0 {
			lineTotal = 0
		}
		displayPrice := rawPrice
		if itemDiscount >= (rawPrice * qty) {
			displayPrice = 0.00
		}

		lineTaxable := lineTotal / 1.18
		lineTax := lineTotal - lineTaxable

		totalTaxable += lineTaxable
		totalTax += lineTax

		hsCode := ns(item.HSCode)
		if hsCode == "" {
			hsCode = "33029019"
		}
		title := ns(item.Title)
		sku := ns(item.SKU)

		curX := pdf.GetX()
		curY := pdf.GetY()

		// Calculate height for wrapping text
		textLines := pdf.SplitLines([]byte(title), wName-4.0)
		numLines := float64(len(textLines))
		if numLines < 1 {
			numLines = 1
		}
		h := numLines * 4.5
		if h < 8.5 {
			h = 8.5
		}

		// Draw background/borders for the whole row height h
		pdf.CellFormat(wName, h, "", "1", 0, "L", false, 0, "")
		pdf.CellFormat(wSKU, h, sku, "1", 0, "C", false, 0, "")
		pdf.CellFormat(wHSN, h, hsCode, "1", 0, "C", false, 0, "")
		pdf.CellFormat(wQty, h, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(wPrice, h, fmt.Sprintf("%.2f", displayPrice), "1", 0, "R", false, 0, "")

		totalItemDiscountForDisplay := itemDiscount + orderDiscount
		pdf.CellFormat(wDiscount, h, fmt.Sprintf("%.2f", totalItemDiscountForDisplay), "1", 0, "R", false, 0, "")

		pdf.CellFormat(wTaxable, h, fmt.Sprintf("%.2f", lineTaxable), "1", 0, "R", false, 0, "")
		pdf.CellFormat(wGSTPct, h, "18%", "1", 0, "C", false, 0, "")
		pdf.CellFormat(wGSTAmt, h, fmt.Sprintf("%.2f", lineTax), "1", 1, "R", false, 0, "")

		// Reposition to write MultiCell text in the first column
		pdf.SetXY(curX+2, curY+(h-(numLines*4.5))/2)
		pdf.MultiCell(wName-4, 4.5, title, "0", "L", false)
		pdf.SetXY(curX, curY+h)
	}
}

func (s *InvoiceService) renderTotals(pdf *gofpdf.Fpdf, order entity.Order, items []entity.LineItem, totals InvoiceTotals, hasMontserrat bool) {
	oGrandTotal := totals.GrandTotal

	isInterState := true
	custState := ns(order.CustomerState)
	if strings.Contains(strings.ToLower(custState), "tamil nadu") || strings.EqualFold(custState, "TN") {
		isInterState = false
	}

	totalGrossSubtotal := totals.GrossSubtotal
	totalOrderDiscount := totals.OrderDiscount

	pdf.Ln(5.8)
	pdf.SetX(120)
	s.safeSetFont(pdf, "Montserrat", "B", 8.25, hasMontserrat)
	pdf.CellFormat(40, 5, "Subtotal (Gross):", "0", 0, "R", false, 0, "")
	s.safeSetFont(pdf, "Montserrat", "", 8.25, hasMontserrat)
	pdf.CellFormat(30, 5, fmt.Sprintf("%.2f", totalGrossSubtotal), "0", 1, "R", false, 0, "")

	discountLabel := "Discount"
	if order.RawPayload != nil && len(*order.RawPayload) > 0 {
		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(*order.RawPayload), &raw); err == nil {
			if codes, ok := raw["discount_codes"].([]interface{}); ok && len(codes) > 0 {
				if codeObj, ok := codes[0].(map[string]interface{}); ok {
					if code, ok := codeObj["code"].(string); ok && code != "" {
						discountLabel = "Discount (" + code + ")"
					}
				}
			} else if apps, ok := raw["discount_applications"].([]interface{}); ok && len(apps) > 0 {
				if app, ok := apps[0].(map[string]interface{}); ok {
					if code, ok := app["code"].(string); ok && code != "" {
						discountLabel = "Discount (" + code + ")"
					}
				}
			}
		}
	}

	if totalOrderDiscount > 0 {
		pdf.Ln(1)
		pdf.SetX(120)
		s.safeSetFont(pdf, "Montserrat", "B", 8.25, hasMontserrat)
		pdf.CellFormat(40, 5, discountLabel+":", "0", 0, "R", false, 0, "")
		s.safeSetFont(pdf, "Montserrat", "", 8.25, hasMontserrat)
		pdf.CellFormat(30, 5, fmt.Sprintf("-%.2f", totalOrderDiscount), "0", 1, "R", false, 0, "")
	}

	pdf.Ln(1)
	pdf.SetX(120)
	s.safeSetFont(pdf, "Montserrat", "B", 8.25, hasMontserrat)
	pdf.CellFormat(40, 5, "Net Taxable:", "0", 0, "R", false, 0, "")
	s.safeSetFont(pdf, "Montserrat", "", 8.25, hasMontserrat)
	pdf.CellFormat(30, 5, fmt.Sprintf("%.2f", totals.TaxableValue), "0", 1, "R", false, 0, "")
	pdf.Ln(1)

	if !isInterState {
		pdf.SetX(120)
		s.safeSetFont(pdf, "Montserrat", "B", 8.25, hasMontserrat)
		pdf.CellFormat(40, 5, "CGST (9%):", "0", 0, "R", false, 0, "")
		s.safeSetFont(pdf, "Montserrat", "", 8.25, hasMontserrat)
		pdf.CellFormat(30, 5, fmt.Sprintf("%.2f", totals.TotalTax/2), "0", 1, "R", false, 0, "")
		pdf.Ln(1)
		pdf.SetX(120)
		s.safeSetFont(pdf, "Montserrat", "B", 8.25, hasMontserrat)
		pdf.CellFormat(40, 5, "SGST (9%):", "0", 0, "R", false, 0, "")
		s.safeSetFont(pdf, "Montserrat", "", 8.25, hasMontserrat)
		pdf.CellFormat(30, 5, fmt.Sprintf("%.2f", totals.TotalTax/2), "0", 1, "R", false, 0, "")
	} else {
		pdf.SetX(120)
		s.safeSetFont(pdf, "Montserrat", "B", 8.25, hasMontserrat)
		pdf.CellFormat(40, 5, "IGST (18%):", "0", 0, "R", false, 0, "")
		s.safeSetFont(pdf, "Montserrat", "", 8.25, hasMontserrat)
		pdf.CellFormat(30, 5, fmt.Sprintf("%.2f", totals.TotalTax), "0", 1, "R", false, 0, "")
	}

	pdf.Ln(2.6)
	pdf.SetX(120)
	pdf.CellFormat(70, 0, "", "T", 1, "R", false, 0, "")
	pdf.SetX(120)
	s.safeSetFont(pdf, "Montserrat", "B", 10.5, hasMontserrat)
	pdf.CellFormat(40, 10, "GRAND TOTAL:", "0", 0, "R", false, 0, "")
	pdf.CellFormat(30, 10, fmt.Sprintf("%.2f", oGrandTotal), "0", 1, "R", false, 0, "")
	pdf.Ln(8.0)
}

func (s *InvoiceService) renderFooter(pdf *gofpdf.Fpdf, hasMontserrat bool) {
	footerSize := 6.4
	s.safeSetFont(pdf, "Montserrat", "B", footerSize, hasMontserrat)
	pdf.CellFormat(0, 4, "Payment Terms:", "0", 1, "L", false, 0, "")
	s.safeSetFont(pdf, "Montserrat", "", footerSize, hasMontserrat)
	pdf.MultiCell(0, 3.5, "Full payment is required before the due date mentioned on the invoice.", "0", "L", false)
	pdf.Ln(2)

	s.safeSetFont(pdf, "Montserrat", "B", footerSize, hasMontserrat)
	pdf.CellFormat(0, 4, "No Refunds & Returns:", "0", 1, "L", false, 0, "")
	s.safeSetFont(pdf, "Montserrat", "", footerSize, hasMontserrat)
	pdf.MultiCell(0, 3.5, "Due to the nature of our products, we do not accept returns or provide refunds once the item has been opened or used.", "0", "L", false)
	pdf.Ln(2)

	s.safeSetFont(pdf, "Montserrat", "B", footerSize, hasMontserrat)
	pdf.CellFormat(0, 4, "Intellectual Property:", "0", 1, "L", false, 0, "")
	s.safeSetFont(pdf, "Montserrat", "", footerSize, hasMontserrat)
	bizName := s.getSetting("business_name", DefaultFooterBusiness)
	pdf.MultiCell(0, 3.5, fmt.Sprintf("All branding and product names are trademarks of %s and may not be reproduced without permission.", bizName), "0", "L", false)
}

// ns extracts the string value from a *string pointer, returning "" if nil.
func ns(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
