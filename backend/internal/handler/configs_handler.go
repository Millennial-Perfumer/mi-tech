package handler

import (
	"encoding/json"
	"net/http"

	"mi-tech/internal/repository"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ConfigsHandler handles the /api/configs endpoints.
type ConfigsHandler struct {
	configsRepo *repository.ConfigsRepository
	db          *gorm.DB
}

// NewConfigsHandler creates a new ConfigsHandler.
func NewConfigsHandler(configsRepo *repository.ConfigsRepository, db *gorm.DB) *ConfigsHandler {
	return &ConfigsHandler{configsRepo: configsRepo, db: db}
}

// GetAllConfigs returns all configs with secret values masked.
// @Summary List API configurations
// @Description Retrieve a list of all API keys and secrets. Sensitive values are masked (********).
// @Tags configurations
// @Security Bearer
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /configs [get]
func (h *ConfigsHandler) GetAllConfigs(w http.ResponseWriter, r *http.Request) {
	configs, err := h.configsRepo.GetAll()
	if err != nil {
		http.Error(w, "Failed to get configs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"configs": configs,
	})
}

// RevealConfigs returns all configs with values unmasked after password verification.
// RevealConfigs handles POST /api/configs/reveal.
// @Summary Reveal unmasked configurations
// @Description Retrieve unmasked values of API keys by re-authenticating with the user's password.
// @Tags configurations
// @Security Bearer
// @Accept json
// @Produce json
// @Param body body object true "User password"
// @Success 200 {object} map[string]interface{}
// @Router /configs/reveal [post]
func (h *ConfigsHandler) RevealConfigs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get username from context
	username, ok := r.Context().Value("username").(string)
	if !ok || username == "" {
		http.Error(w, "Unauthorized: user identity not found", http.StatusUnauthorized)
		return
	}

	// Validate password against the users table for the CURRENT user
	var user struct {
		PasswordHash string `gorm:"column:password_hash"`
	}
	if err := h.db.Table("users").Select("password_hash").Where("username = ?", username).First(&user).Error; err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "User not found",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Incorrect password",
		})
		return
	}

	configs, err := h.configsRepo.GetAllRevealed()
	if err != nil {
		http.Error(w, "Failed to get configs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"configs": configs,
	})
}

// UpdateConfig updates a single config value.
// UpdateConfig handles PUT /api/configs.
// @Summary Update API configuration
// @Description Update a specific configuration key (e.g. meta_api_token). Required 'admin' role.
// @Tags configurations
// @Security Bearer
// @Accept json
// @Produce json
// @Param body body object true "Config key-value"
// @Success 200 {object} map[string]interface{}
// @Router /configs [put]
func (h *ConfigsHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if body.Key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	if err := h.configsRepo.Set(body.Key, body.Value); err != nil {
		http.Error(w, "Failed to update config", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Config updated",
	})
}
