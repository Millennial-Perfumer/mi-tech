package service

import (
	"fmt"
	"io"
	"strings"

	"github.com/jung-kurt/gofpdf"

	"mi-tech/internal/entity"
)

// InvoiceService handles PDF invoice generation.
type InvoiceService struct{}

// NewInvoiceService creates a new InvoiceService.
func NewInvoiceService() *InvoiceService {
	return &InvoiceService{}
}

// GeneratePDF creates a professional GST invoice PDF and writes it to the provided writer.
func (s *InvoiceService) GeneratePDF(order entity.Order, items []entity.LineItem, w io.Writer) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10.6, 15, 10.6)

	pdf.AddUTF8Font("Montserrat", "", "internal/fonts/Montserrat-Regular.ttf")
	pdf.AddUTF8Font("Montserrat", "B", "internal/fonts/Montserrat-Bold.ttf")
	pdf.AddUTF8Font("Montserrat", "I", "internal/fonts/Montserrat-SemiBold.ttf")

	pdf.AddPage()

	// -- Header --
	pdf.SetFont("Montserrat", "B", 13.5)
	pdf.CellFormat(100, 10, "PARFUM TRADERS", "0", 0, "L", false, 0, "")
	pdf.CellFormat(80, 10, "TAX INVOICE", "0", 1, "R", false, 0, "")

	pdf.SetFont("Montserrat", "", 7.5)
	pdf.Ln(2)
	pdf.CellFormat(100, 4, "GSTIN: 33AUSPR1909H1ZC", "0", 1, "L", false, 0, "")
	pdf.Ln(1)
	pdf.CellFormat(100, 4, "No. 9/21, 1st floor, Sadiq Basha Nagar,", "0", 1, "L", false, 0, "")
	pdf.CellFormat(100, 4, "2nd Street, Virugambakkam, Chennai - 600092", "0", 1, "L", false, 0, "")
	pdf.Ln(1)
	pdf.CellFormat(100, 4, "Phone: 7904769823", "0", 1, "L", false, 0, "")
	pdf.Ln(6.5)

	// -- Invoice & Customer Details --
	leftCol := 90.0
	rightCol := 90.0

	pdf.SetFont("Montserrat", "B", 9)
	pdf.CellFormat(leftCol, 6, "INVOICE DETAILS", "B", 0, "L", false, 0, "")
	pdf.CellFormat(rightCol, 6, "BILL TO (CUSTOMER)", "B", 1, "L", false, 0, "")
	pdf.Ln(3)

	metaYStart := pdf.GetY()

	// Left side: Invoice meta
	pdf.SetFont("Montserrat", "B", 7.5)
	pdf.CellFormat(30, 4, "Invoice No:", "0", 0, "L", false, 0, "")
	pdf.SetFont("Montserrat", "", 7.5)
	pdf.CellFormat(60, 4, "INV-"+order.OrderNumber, "0", 1, "L", false, 0, "")

	pdf.SetFont("Montserrat", "B", 7.5)
	pdf.CellFormat(30, 4, "Order No:", "0", 0, "L", false, 0, "")
	pdf.SetFont("Montserrat", "", 7.5)
	pdf.CellFormat(60, 4, "#"+order.OrderNumber, "0", 1, "L", false, 0, "")

	pdf.SetFont("Montserrat", "B", 7.5)
	pdf.CellFormat(30, 4, "Date:", "0", 0, "L", false, 0, "")
	pdf.SetFont("Montserrat", "", 7.5)
	pdf.CellFormat(60, 4, order.CreatedAt.Format("2006-01-02"), "0", 1, "L", false, 0, "")

	// Right side: Customer info
	pdf.SetY(metaYStart)
	pdf.SetX(leftCol + 15)

	displayName := ns(order.CustomerName)
	if strings.TrimSpace(displayName) == "" {
		displayName = "Valued Customer"
	}
	pdf.SetFont("Montserrat", "B", 8.25)
	pdf.SetX(leftCol + 15)
	pdf.CellFormat(rightCol, 4, displayName, "0", 1, "L", false, 0, "")

	pdf.SetFont("Montserrat", "", 7.5)
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
	calcTaxable, calcTax := s.renderItemsTable(pdf, items)

	// -- Totals --
	s.renderTotals(pdf, order, calcTaxable, calcTax)

	// -- Footer Terms --
	s.renderFooter(pdf)

	return pdf.Output(w)
}

func (s *InvoiceService) renderItemsTable(pdf *gofpdf.Fpdf, items []entity.LineItem) (float64, float64) {
	pdf.SetFillColor(245, 245, 245)
	pdf.SetFont("Montserrat", "I", 7.5)

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

	pdf.SetFont("Montserrat", "", 6.75)

	totalTaxable := 0.0
	totalTax := 0.0

	for _, item := range items {
		rawPrice := item.Price
		itemDiscount := item.Discount
		qty := float64(item.Quantity)

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
		pdf.CellFormat(wDiscount, h, fmt.Sprintf("%.2f", itemDiscount), "1", 0, "R", false, 0, "")
		pdf.CellFormat(wTaxable, h, fmt.Sprintf("%.2f", lineTaxable), "1", 0, "R", false, 0, "")
		pdf.CellFormat(wGSTPct, h, "18%", "1", 0, "C", false, 0, "")
		pdf.CellFormat(wGSTAmt, h, fmt.Sprintf("%.2f", lineTax), "1", 1, "R", false, 0, "")

		// Reposition to write MultiCell text in the first column
		pdf.SetXY(curX+2, curY+(h-(numLines*4.5))/2)
		pdf.MultiCell(wName-4, 4.5, title, "0", "L", false)
		pdf.SetXY(curX, curY+h)
	}
	return totalTaxable, totalTax
}

func (s *InvoiceService) renderTotals(pdf *gofpdf.Fpdf, order entity.Order, calcTaxable, calcTax float64) {
	oGrandTotal := order.TotalPrice

	isInterState := true
	custState := ns(order.CustomerState)
	if strings.Contains(strings.ToLower(custState), "tamil nadu") || strings.EqualFold(custState, "TN") {
		isInterState = false
	}

	pdf.Ln(5.8)
	pdf.SetX(120)
	pdf.SetFont("Montserrat", "B", 8.25)
	pdf.CellFormat(40, 5, "Total Taxable:", "0", 0, "R", false, 0, "")
	pdf.SetFont("Montserrat", "", 8.25)
	pdf.CellFormat(30, 5, fmt.Sprintf("%.2f", calcTaxable), "0", 1, "R", false, 0, "")
	pdf.Ln(1)

	if !isInterState {
		pdf.SetX(120)
		pdf.SetFont("Montserrat", "B", 8.25)
		pdf.CellFormat(40, 5, "CGST (9%):", "0", 0, "R", false, 0, "")
		pdf.SetFont("Montserrat", "", 8.25)
		pdf.CellFormat(30, 5, fmt.Sprintf("%.2f", calcTax/2), "0", 1, "R", false, 0, "")
		pdf.Ln(1)
		pdf.SetX(120)
		pdf.SetFont("Montserrat", "B", 8.25)
		pdf.CellFormat(40, 5, "SGST (9%):", "0", 0, "R", false, 0, "")
		pdf.SetFont("Montserrat", "", 8.25)
		pdf.CellFormat(30, 5, fmt.Sprintf("%.2f", calcTax/2), "0", 1, "R", false, 0, "")
	} else {
		pdf.SetX(120)
		pdf.SetFont("Montserrat", "B", 8.25)
		pdf.CellFormat(40, 5, "IGST (18%):", "0", 0, "R", false, 0, "")
		pdf.SetFont("Montserrat", "", 8.25)
		pdf.CellFormat(30, 5, fmt.Sprintf("%.2f", calcTax), "0", 1, "R", false, 0, "")
	}

	pdf.Ln(2.6)
	pdf.SetX(120)
	pdf.CellFormat(70, 0, "", "T", 1, "R", false, 0, "")
	pdf.SetX(120)
	pdf.SetFont("Montserrat", "B", 10.5)
	pdf.CellFormat(40, 10, "GRAND TOTAL:", "0", 0, "R", false, 0, "")
	pdf.CellFormat(30, 10, fmt.Sprintf("%.2f", oGrandTotal), "0", 1, "R", false, 0, "")
	pdf.Ln(8.0)
}

func (s *InvoiceService) renderFooter(pdf *gofpdf.Fpdf) {
	footerSize := 6.4
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
}

// ns extracts the string value from a *string pointer, returning "" if nil.
func ns(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
