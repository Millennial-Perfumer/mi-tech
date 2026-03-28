package handler

import (
	"encoding/json"
	"net/http"

	"mi-tech/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

// Login handles POST /api/auth/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	token, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{Token: token})
}

// VerifyAuth handles GET /api/auth/verify for Nginx auth_request.
func (h *AuthHandler) VerifyAuth(w http.ResponseWriter, r *http.Request) {
	username, _ := r.Context().Value("username").(string)
	role, _ := r.Context().Value("userRole").(string)

	if username == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Set headers for Nginx to capture and forward
	w.Header().Set("X-WEBAUTH-USER", username)
	w.Header().Set("X-WEBAUTH-ROLE", role)
	w.WriteHeader(http.StatusOK)
}
