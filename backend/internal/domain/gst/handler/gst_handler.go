package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	_ "mi-tech/internal/domain/gst/dto"
	"mi-tech/internal/domain/gst/service"
)

// GSTHandler is a thin HTTP adapter for GST report endpoints.
type GSTHandler struct {
	gstService *service.GSTService
}

// NewGSTHandler creates a new GSTHandler.
func NewGSTHandler(gstService *service.GSTService) *GSTHandler {
	return &GSTHandler{gstService: gstService}
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
func (h *GSTHandler) GetGSTSummary(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	summary, err := h.gstService.GetGSTSummary(startDate, endDate)
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
func (h *GSTHandler) GetStateSummary(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	data, err := h.gstService.GetStateSummary(startDate, endDate)
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
func (h *GSTHandler) GetHSNSummary(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	data, err := h.gstService.GetHSNSummary(startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": data})
}

// GetDocumentsIssued handles GET /api/reports/documents-issued.
func (h *GSTHandler) GetDocumentsIssued(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	data, err := h.gstService.GetDocumentsIssued(startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "data": data})
}

// GetGSTR1JSON handles GET /api/reports/gstr1-json.
// @Summary Export GSTR-1 JSON
// @Description Compiles and downloads GSTR-1 offline utility-compliant JSON containing Table 7 (B2CS), Table 12 (HSN), and Table 13 (Documents).
// @Tags reports
// @Security Bearer
// @Produce json
// @Param start_date query string false "Start date"
// @Param end_date query string false "End date"
// @Param gstin query string false "GSTIN override"
// @Success 200 {object} dto.GSTR1Payload
// @Router /reports/gstr1-json [get]
func (h *GSTHandler) GetGSTR1JSON(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	gstin := r.URL.Query().Get("gstin")
	if gstin == "" {
		gstin = "33AUSPR1909H1ZC" // Default fallback
	}

	payload, err := h.gstService.GetGSTR1JSON(startDate, endDate, gstin)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=GSTR1_%s_%s.json", gstin, payload.FP))
	json.NewEncoder(w).Encode(payload)
}
