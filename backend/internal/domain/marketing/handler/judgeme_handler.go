package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"mi-tech/internal/domain/marketing/dto"
	"mi-tech/internal/domain/marketing/service"
)

type JudgeMeHandler struct {
	judgeMeService *service.JudgeMeService
}

func NewJudgeMeHandler(judgeMeService *service.JudgeMeService) *JudgeMeHandler {
	return &JudgeMeHandler{
		judgeMeService: judgeMeService,
	}
}

// GenerateReviews handles the API endpoint to generate review drafts (capped at max 10 total).
// @Summary Generate Judge.me Draft Reviews
// @Description Generates realistic human-sounding review drafts with Indian names and typo variations (max 10).
// @Tags marketing
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body dto.GenerateReviewsRequest true "Generate Request"
// @Success 200 {array} dto.GeneratedReviewDTO
// @Router /marketing/judgeme/generate [post]
func (h *JudgeMeHandler) GenerateReviews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.GenerateReviewsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request payload"})
		return
	}

	drafts, err := h.judgeMeService.GenerateReviews(r.Context(), req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(drafts)
}

// SubmitReviews handles submitting generated/edited reviews to Judge.me and saving to DB.
// @Summary Submit Reviews to Judge.me
// @Description Posts review drafts via multipart/form-data to Judge.me API and records them in DB.
// @Tags marketing
// @Security Bearer
// @Accept json
// @Produce json
// @Param request body dto.SubmitReviewsRequest true "Submit Request"
// @Success 200 {object} dto.SubmitReviewsResponse
// @Router /marketing/judgeme/submit [post]
func (h *JudgeMeHandler) SubmitReviews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.SubmitReviewsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request payload"})
		return
	}

	res, err := h.judgeMeService.SubmitReviews(r.Context(), req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

// GetPublishedReviews retrieves paginated historical published reviews stored in PostgreSQL database.
// @Summary Get Published Reviews History
// @Description Fetches published review logs stored in database.
// @Tags marketing
// @Security Bearer
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Limit count"
// @Param product_id query string false "Filter by product ID"
// @Param search query string false "Search pattern"
// @Success 200 {object} dto.PublishedReviewsResponse
// @Router /marketing/judgeme/published [get]
func (h *JudgeMeHandler) GetPublishedReviews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	productID := r.URL.Query().Get("product_id")
	search := r.URL.Query().Get("search")

	res, err := h.judgeMeService.GetPublishedReviews(r.Context(), page, limit, productID, search)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
