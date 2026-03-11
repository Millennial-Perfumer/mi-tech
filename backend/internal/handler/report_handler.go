package handler

import (
	"encoding/json"
	"net/http"

	"shopify-gst-app/internal/service"
)

// ReportHandler is a thin HTTP adapter for GST report endpoints.
type ReportHandler struct {
	reportService *service.ReportService
}

// NewReportHandler creates a new ReportHandler.
func NewReportHandler(reportService *service.ReportService) *ReportHandler {
	return &ReportHandler{reportService: reportService}
}

// GetGSTSummary handles GET /api/reports/summary.
func (h *ReportHandler) GetGSTSummary(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	summary, err := h.reportService.GetGSTSummary(startDate, endDate)
	if err != nil {
		http.Error(w, "Failed to fetch summary: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "summary": summary})
}

// GetStateSummary handles GET /api/reports/state-wise.
func (h *ReportHandler) GetStateSummary(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	data, err := h.reportService.GetStateSummary(startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": data})
}

// GetHSNSummary handles GET /api/reports/hsn-wise.
func (h *ReportHandler) GetHSNSummary(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	data, err := h.reportService.GetHSNSummary(startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": data})
}

// GetDocumentsIssued handles GET /api/reports/documents-issued.
func (h *ReportHandler) GetDocumentsIssued(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	data, err := h.reportService.GetDocumentsIssued(startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": data})
}
