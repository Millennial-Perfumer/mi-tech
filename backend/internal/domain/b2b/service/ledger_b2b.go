package service

import (
	"fmt"
	"sort"
	"time"

	"mi-tech/internal/domain/b2b/entity"
)

type CustomerLedgerRow struct {
	Date           time.Time `json:"date"`
	Type           string    `json:"type"` // INVOICE, PAYMENT, CREDIT_NOTE, DEBIT_NOTE
	Reference      string    `json:"reference"`
	Debit          float64   `json:"debit"`
	Credit         float64   `json:"credit"`
	RunningBalance float64   `json:"running_balance"`
}

type CustomerLedger struct {
	CustomerID     int64               `json:"customer_id"`
	CustomerName   string              `json:"customer_name"`
	OpeningBalance float64             `json:"opening_balance"`
	ClosingBalance float64             `json:"closing_balance"`
	Transactions   []CustomerLedgerRow `json:"transactions"`
}

type AgingSummary struct {
	CustomerID   int64   `json:"customer_id"`
	CustomerName string  `json:"customer_name"`
	GSTIN        string  `json:"gstin"`
	Days0to30    float64 `json:"days_0_30"`
	Days31to60   float64 `json:"days_31_60"`
	Days60Plus   float64 `json:"days_60_plus"`
	TotalDue     float64 `json:"total_due"`
}

func (s *B2BService) GetCustomerLedger(customerID int64) (*CustomerLedger, error) {
	cust, err := s.repo.GetCustomerByID(customerID)
	if err != nil {
		return nil, err
	}

	var rows []CustomerLedgerRow

	// 1. Fetch Invoices (and emit Debit invoice + Credit payment transactions)
	var invoices []entity.B2BInvoice
	err = s.db.Preload("Items").Where("customer_id = ? AND status = 'ISSUED'", customerID).Find(&invoices).Error
	if err != nil {
		return nil, err
	}

	for _, inv := range invoices {
		// Debit row (invoice issuance)
		rows = append(rows, CustomerLedgerRow{
			Date:      inv.InvoiceDate,
			Type:      "INVOICE",
			Reference: *inv.InvoiceNumber,
			Debit:     inv.TotalPrice,
			Credit:    0,
		})

		// Credit row (invoice payment)
		if inv.PaidAmount > 0 && inv.PaymentDate != nil {
			rows = append(rows, CustomerLedgerRow{
				Date:      *inv.PaymentDate,
				Type:      "PAYMENT",
				Reference: fmt.Sprintf("Payment for %s", *inv.InvoiceNumber),
				Debit:     0,
				Credit:    inv.PaidAmount,
			})
		}
	}

	// 2. Fetch Credit Notes
	var creditNotes []entity.B2BCreditNote
	err = s.db.Where("customer_id = ? AND status = 'ISSUED'", customerID).Find(&creditNotes).Error
	if err == nil {
		for _, cn := range creditNotes {
			rows = append(rows, CustomerLedgerRow{
				Date:      cn.NoteDate,
				Type:      "CREDIT_NOTE",
				Reference: *cn.CreditNoteNumber,
				Debit:     0,
				Credit:    cn.TotalPrice,
			})
		}
	}

	// 3. Fetch Debit Notes
	var debitNotes []entity.B2BDebitNote
	err = s.db.Where("customer_id = ? AND status = 'ISSUED'", customerID).Find(&debitNotes).Error
	if err == nil {
		for _, dn := range debitNotes {
			rows = append(rows, CustomerLedgerRow{
				Date:      dn.NoteDate,
				Type:      "DEBIT_NOTE",
				Reference: *dn.DebitNoteNumber,
				Debit:     dn.TotalPrice,
				Credit:    0,
			})
		}
	}

	// Sort chronologically
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Date.Before(rows[j].Date)
	})

	// Calculate running balance
	var balance float64
	for i := range rows {
		balance += rows[i].Debit
		balance -= rows[i].Credit
		rows[i].RunningBalance = balance
	}

	return &CustomerLedger{
		CustomerID:     customerID,
		CustomerName:   cust.LegalName,
		OpeningBalance: 0.00,
		ClosingBalance: balance,
		Transactions:   rows,
	}, nil
}

func (s *B2BService) GetOutstandingAgingReport() ([]AgingSummary, error) {
	var invoices []entity.B2BInvoice
	err := s.db.Where("status = 'ISSUED' AND balance_amount > 0").Find(&invoices).Error
	if err != nil {
		return nil, err
	}

	agingMap := make(map[int64]*AgingSummary)
	now := time.Now()

	for _, inv := range invoices {
		if inv.CustomerID == nil {
			continue
		}
		cid := *inv.CustomerID

		if _, exists := agingMap[cid]; !exists {
			agingMap[cid] = &AgingSummary{
				CustomerID:   cid,
				CustomerName: inv.CustomerName,
				GSTIN:        inv.CustomerGSTIN,
			}
		}

		summary := agingMap[cid]
		daysOverdue := 0
		if inv.DueDate != nil {
			daysOverdue = int(now.Sub(*inv.DueDate).Hours() / 24)
		} else {
			daysOverdue = int(now.Sub(inv.InvoiceDate).Hours() / 24)
		}

		amt := inv.BalanceAmount
		if daysOverdue <= 30 {
			summary.Days0to30 += amt
		} else if daysOverdue <= 60 {
			summary.Days31to60 += amt
		} else {
			summary.Days60Plus += amt
		}
		summary.TotalDue += amt
	}

	var list []AgingSummary
	for _, v := range agingMap {
		list = append(list, *v)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].TotalDue > list[j].TotalDue
	})

	return list, nil
}
