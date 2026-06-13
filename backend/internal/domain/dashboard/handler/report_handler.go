package handler

import (
	"encoding/json"
	"net/http"

	"mi-tech/internal/domain/dashboard/service"
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
// GetGSTSummary handles GET /api/reports/gst.
// @Summary GST Summary Report
// @Description Aggregate revenue, taxable values, and tax splits (CGST/SGST/IGST) for a date range.
// @Tags reports
// @Security Bearer
// @Produce json
// @Param start_date query string false "Start date"
// @Param end_date query string false "End date"
// @Success 200 {object} map[string]interface{}
// @Router /reports/gst [get]
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
// GetStateSummary handles GET /api/reports/state.
// @Summary State-wise Summary Report
// @Description Breakdown of revenue and taxes grouped by customer state.
// @Tags reports
// @Security Bearer
// @Produce json
// @Param start_date query string false "Start date"
// @Param end_date query string false "End date"
// @Success 200 {object} map[string]interface{}
// @Router /reports/state [get]
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
// GetHSNSummary handles GET /api/reports/hsn.
// @Summary HSN-wise Summary Report
// @Description Detailed breakdown of taxes and sales grouped by HSN code.
// @Tags reports
// @Security Bearer
// @Produce json
// @Param start_date query string false "Start date"
// @Param end_date query string false "End date"
// @Success 200 {object} map[string]interface{}
// @Router /reports/hsn [get]
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
