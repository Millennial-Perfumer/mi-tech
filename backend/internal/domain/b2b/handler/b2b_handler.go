package handler

import (
	"encoding/json"
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
		var inv entity.B2BInvoice
		if err := json.NewDecoder(r.Body).Decode(&inv); err != nil {
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
		var inv entity.B2BInvoice
		if err := json.NewDecoder(r.Body).Decode(&inv); err != nil {
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
