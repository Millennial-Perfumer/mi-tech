package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"mi-tech/internal/domain/b2b/entity"
	"mi-tech/internal/domain/b2b/service"
)

type B2BHandler struct {
	srv *service.B2BService
}

func NewB2BHandler(srv *service.B2BService) *B2BHandler {
	return &B2BHandler{srv: srv}
}

func (h *B2BHandler) HandleCustomers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		search := r.URL.Query().Get("search")
		custs, err := h.srv.ListCustomers(search)
		if err != nil {
			log.Printf("B2BHandler.ListCustomers error: %v", err)
			http.Error(w, "Failed to load customers", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(custs)

	case http.MethodPost:
		var cust entity.B2BCustomer
		if err := json.NewDecoder(r.Body).Decode(&cust); err != nil {
			http.Error(w, "Invalid customer body", http.StatusBadRequest)
			return
		}
		if err := h.srv.CreateCustomer(&cust); err != nil {
			log.Printf("B2BHandler.CreateCustomer error: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(cust)

	case http.MethodPut:
		var cust entity.B2BCustomer
		if err := json.NewDecoder(r.Body).Decode(&cust); err != nil {
			http.Error(w, "Invalid customer body", http.StatusBadRequest)
			return
		}
		if err := h.srv.UpdateCustomer(&cust); err != nil {
			log.Printf("B2BHandler.UpdateCustomer error: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(cust)

	case http.MethodDelete:
		idStr := r.URL.Query().Get("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			http.Error(w, "Invalid customer ID", http.StatusBadRequest)
			return
		}
		if err := h.srv.DeleteCustomer(id); err != nil {
			log.Printf("B2BHandler.DeleteCustomer error: %v", err)
			http.Error(w, "Failed to delete customer", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *B2BHandler) HandleInvoices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		startDate := r.URL.Query().Get("startDate")
		endDate := r.URL.Query().Get("endDate")
		status := r.URL.Query().Get("status")
		invs, err := h.srv.ListInvoices(startDate, endDate, status)
		if err != nil {
			log.Printf("B2BHandler.ListInvoices error: %v", err)
			http.Error(w, "Failed to load invoices", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(invs)

	case http.MethodPost:
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("B2BHandler.CreateInvoice read body error: %v", err)
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			return
		}
		var inv entity.B2BInvoice
		if err := json.Unmarshal(bodyBytes, &inv); err != nil {
			log.Printf("B2BHandler.CreateInvoice JSON decode error: %v. Body: %s", err, string(bodyBytes))
			http.Error(w, "Invalid invoice body", http.StatusBadRequest)
			return
		}
		if err := h.srv.CreateInvoice(&inv); err != nil {
			log.Printf("B2BHandler.CreateInvoice error: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(inv)

	case http.MethodPut:
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("B2BHandler.UpdateInvoice read body error: %v", err)
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			return
		}
		var inv entity.B2BInvoice
		if err := json.Unmarshal(bodyBytes, &inv); err != nil {
			log.Printf("B2BHandler.UpdateInvoice JSON decode error: %v. Body: %s", err, string(bodyBytes))
			http.Error(w, "Invalid invoice body", http.StatusBadRequest)
			return
		}
		if err := h.srv.UpdateInvoice(&inv); err != nil {
			log.Printf("B2BHandler.UpdateInvoice error: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(inv)

	case http.MethodDelete:
		idStr := r.URL.Query().Get("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
			return
		}
		if err := h.srv.DeleteInvoice(id); err != nil {
			log.Printf("B2BHandler.DeleteInvoice error: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *B2BHandler) GetInvoiceByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	inv, err := h.srv.GetInvoiceByID(id)
	if err != nil {
		log.Printf("B2BHandler.GetInvoiceByID error: %v", err)
		http.Error(w, "Invoice not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(inv)
}

func (h *B2BHandler) GetNextInvoiceNumber(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	date := r.URL.Query().Get("date")
	nextNum, err := h.srv.GetNextInvoiceNumber(date)
	if err != nil {
		log.Printf("B2BHandler.GetNextInvoiceNumber error: %v", err)
		http.Error(w, "Failed to compute next invoice number", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"next_invoice_number": nextNum})
}

func (h *B2BHandler) IssueInvoice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	inv, err := h.srv.IssueInvoice(id)
	if err != nil {
		log.Printf("B2BHandler.IssueInvoice error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(inv)
}

func (h *B2BHandler) CancelInvoice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	if err := h.srv.CancelInvoice(id); err != nil {
		log.Printf("B2BHandler.CancelInvoice error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *B2BHandler) DeductInventory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	if err := h.srv.DeductInventory(id); err != nil {
		log.Printf("B2BHandler.DeductInventory error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *B2BHandler) RevertInventory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	if err := h.srv.RevertInventory(id); err != nil {
		log.Printf("B2BHandler.RevertInventory error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *B2BHandler) UpdatePayment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID            int64   `json:"id"`
		PaidAmount    float64 `json:"paid_amount"`
		PaymentMethod string  `json:"payment_method"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	inv, err := h.srv.UpdatePayment(req.ID, req.PaidAmount, req.PaymentMethod)
	if err != nil {
		log.Printf("B2BHandler.UpdatePayment error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(inv)
}

func (h *B2BHandler) HandlePaymentTerms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		terms, err := h.srv.ListPaymentTerms()
		if err != nil {
			log.Printf("B2BHandler.ListPaymentTerms error: %v", err)
			http.Error(w, "Failed to load payment terms", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(terms)

	case http.MethodPost:
		var term entity.B2BPaymentTerm
		if err := json.NewDecoder(r.Body).Decode(&term); err != nil {
			http.Error(w, "Invalid payment term body", http.StatusBadRequest)
			return
		}
		if err := h.srv.CreatePaymentTerm(&term); err != nil {
			log.Printf("B2BHandler.CreatePaymentTerm error: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(term)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *B2BHandler) HandleCreditNotes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		invoiceIDStr := r.URL.Query().Get("invoice_id")
		var invoiceID int64
		if invoiceIDStr != "" {
			invoiceID, _ = strconv.ParseInt(invoiceIDStr, 10, 64)
		}
		notes, err := h.srv.ListCreditNotes(invoiceID)
		if err != nil {
			http.Error(w, "Failed to load credit notes", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(notes)

	case http.MethodPost:
		var note entity.B2BCreditNote
		if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
			http.Error(w, "Invalid credit note body", http.StatusBadRequest)
			return
		}
		if err := h.srv.CreateCreditNote(&note); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(note)

	case http.MethodPut:
		var note entity.B2BCreditNote
		if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
			http.Error(w, "Invalid credit note body", http.StatusBadRequest)
			return
		}
		if err := h.srv.UpdateCreditNote(&note); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(note)

	case http.MethodDelete:
		idStr := r.URL.Query().Get("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}
		if err := h.srv.DeleteCreditNote(id); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *B2BHandler) IssueCreditNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	note, err := h.srv.IssueCreditNote(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(note)
}

func (h *B2BHandler) CancelCreditNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	if err := h.srv.CancelCreditNote(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *B2BHandler) HandleDebitNotes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		invoiceIDStr := r.URL.Query().Get("invoice_id")
		var invoiceID int64
		if invoiceIDStr != "" {
			invoiceID, _ = strconv.ParseInt(invoiceIDStr, 10, 64)
		}
		notes, err := h.srv.ListDebitNotes(invoiceID)
		if err != nil {
			http.Error(w, "Failed to load debit notes", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(notes)

	case http.MethodPost:
		var note entity.B2BDebitNote
		if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
			http.Error(w, "Invalid debit note body", http.StatusBadRequest)
			return
		}
		if err := h.srv.CreateDebitNote(&note); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(note)

	case http.MethodPut:
		var note entity.B2BDebitNote
		if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
			http.Error(w, "Invalid debit note body", http.StatusBadRequest)
			return
		}
		if err := h.srv.UpdateDebitNote(&note); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(note)

	case http.MethodDelete:
		idStr := r.URL.Query().Get("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}
		if err := h.srv.DeleteDebitNote(id); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *B2BHandler) IssueDebitNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	note, err := h.srv.IssueDebitNote(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(note)
}

func (h *B2BHandler) CancelDebitNote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	if err := h.srv.CancelDebitNote(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *B2BHandler) GetCustomerLedger(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	customerIDStr := r.URL.Query().Get("customer_id")
	customerID, err := strconv.ParseInt(customerIDStr, 10, 64)
	if err != nil || customerID <= 0 {
		http.Error(w, "Invalid customer ID", http.StatusBadRequest)
		return
	}

	ledger, err := h.srv.GetCustomerLedger(customerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(ledger)
}

func (h *B2BHandler) GetOutstandingAgingReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	report, err := h.srv.GetOutstandingAgingReport()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(report)
}

func (h *B2BHandler) HandleGSTPeriods(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		periods, err := h.srv.ListGSTPeriods()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(periods)

	case http.MethodPost:
		var req entity.GSTPeriod
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid body", http.StatusBadRequest)
			return
		}
		if err := h.srv.SaveGSTPeriod(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(req)
	}
}

func (h *B2BHandler) HandleProformas(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		startDate := r.URL.Query().Get("startDate")
		endDate := r.URL.Query().Get("endDate")
		status := r.URL.Query().Get("status")
		pfs, err := h.srv.ListProformas(startDate, endDate, status)
		if err != nil {
			log.Printf("B2BHandler.ListProformas error: %v", err)
			http.Error(w, "Failed to load proforma invoices", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(pfs)

	case http.MethodPost:
		var pf entity.B2BProformaInvoice
		if err := json.NewDecoder(r.Body).Decode(&pf); err != nil {
			http.Error(w, "Invalid proforma body", http.StatusBadRequest)
			return
		}
		log.Printf("[B2BHandler] CreateProforma (POST) called with body ID: %d, note_date: %v", pf.ID, pf.NoteDate)
		if err := h.srv.CreateProforma(&pf); err != nil {
			log.Printf("B2BHandler.CreateProforma error: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(pf)

	case http.MethodPut:
		var pf entity.B2BProformaInvoice
		if err := json.NewDecoder(r.Body).Decode(&pf); err != nil {
			http.Error(w, "Invalid proforma body", http.StatusBadRequest)
			return
		}
		log.Printf("[B2BHandler] UpdateProforma (PUT) called with body ID: %d, note_date: %v", pf.ID, pf.NoteDate)
		if err := h.srv.UpdateProforma(&pf); err != nil {
			log.Printf("B2BHandler.UpdateProforma error: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(pf)

	case http.MethodDelete:
		idStr := r.URL.Query().Get("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			http.Error(w, "Invalid proforma ID", http.StatusBadRequest)
			return
		}
		if err := h.srv.DeleteProforma(id); err != nil {
			log.Printf("B2BHandler.DeleteProforma error: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *B2BHandler) GetProformaByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid proforma ID", http.StatusBadRequest)
		return
	}

	pf, err := h.srv.GetProformaByID(id)
	if err != nil {
		log.Printf("B2BHandler.GetProformaByID error: %v", err)
		http.Error(w, "Proforma invoice not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(pf)
}

func (h *B2BHandler) GetNextProformaNumber(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	date := r.URL.Query().Get("date")
	nextNum, err := h.srv.GetNextProformaNumber(date)
	if err != nil {
		log.Printf("B2BHandler.GetNextProformaNumber error: %v", err)
		http.Error(w, "Failed to compute next proforma number", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"next_proforma_number": nextNum})
}

func (h *B2BHandler) IssueProforma(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid proforma ID", http.StatusBadRequest)
		return
	}

	pf, err := h.srv.IssueProforma(id)
	if err != nil {
		log.Printf("B2BHandler.IssueProforma error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(pf)
}

func (h *B2BHandler) AcceptProforma(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid proforma ID", http.StatusBadRequest)
		return
	}

	if err := h.srv.AcceptProforma(id); err != nil {
		log.Printf("B2BHandler.AcceptProforma error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *B2BHandler) RejectProforma(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid proforma ID", http.StatusBadRequest)
		return
	}

	if err := h.srv.RejectProforma(id); err != nil {
		log.Printf("B2BHandler.RejectProforma error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *B2BHandler) CancelProforma(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid proforma ID", http.StatusBadRequest)
		return
	}

	if err := h.srv.CancelProforma(id); err != nil {
		log.Printf("B2BHandler.CancelProforma error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *B2BHandler) CreateRevision(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid proforma ID", http.StatusBadRequest)
		return
	}

	pf, err := h.srv.CreateRevision(id)
	if err != nil {
		log.Printf("B2BHandler.CreateRevision error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(pf)
}

func (h *B2BHandler) ConvertToTaxInvoice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid proforma ID", http.StatusBadRequest)
		return
	}

	inv, err := h.srv.ConvertToTaxInvoice(id)
	if err != nil {
		log.Printf("B2BHandler.ConvertToTaxInvoice error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(inv)
}

func (h *B2BHandler) CheckExpiredProformas(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	count, err := h.srv.MarkExpiredProformas()
	if err != nil {
		log.Printf("B2BHandler.CheckExpiredProformas error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "expired_count": count})
}

