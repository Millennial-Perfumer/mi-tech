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
	Token       string `json:"token,omitempty"`
	Requires2FA bool   `json:"requires_2fa"`
}

type VerifyOTPRequest struct {
	Username string `json:"username"`
	OTP      string `json:"otp"`
}

// Login handles POST /api/auth/login.
// @Summary Login to the application
// @Description Authenticate a user and return a JWT token. If 2FA is enabled, it returns requires_2fa=true.
// @Tags auth
// @Accept json
// @Produce json
// @Param body body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 401 {string} string "unauthorized"
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	token, requires2FA, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{
		Token:       token,
		Requires2FA: requires2FA,
	})
}

// VerifyOTP handles POST /api/auth/verify-otp.
// @Summary Verify 2FA OTP
// @Description Complete the login process by verifying the 6-digit OTP.
// @Tags auth
// @Accept json
// @Produce json
// @Param body body VerifyOTPRequest true "OTP Verification"
// @Success 200 {object} LoginResponse
// @Failure 401 {string} string "unauthorized"
// @Router /auth/verify-otp [post]
func (h *AuthHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req VerifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	token, err := h.authService.VerifyOTP(req.Username, req.OTP)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{Token: token})
}

// VerifyAuth handles GET /api/auth/verify for Nginx auth_request.
// @Summary Validate JWT Session
// @Description Internal endpoint for Nginx to validate the Bearer token and retrieve user metadata.
// @Tags auth
// @Security Bearer
// @Success 200 {string} string "OK"
// @Failure 401 {string} string "unauthorized"
// @Router /auth/verify [get]
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
