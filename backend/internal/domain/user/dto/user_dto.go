package dto

// LoginRequest represents credentials for authentication.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents the output of authentication.
type LoginResponse struct {
	Token       string `json:"token,omitempty"`
	Requires2FA bool   `json:"requires_2fa"`
}

// VerifyOTPRequest represents the OTP submitted for 2FA verification.
type VerifyOTPRequest struct {
	Username string `json:"username"`
	OTP      string `json:"otp"`
}

// CreateUserRequest represents the payload to create a new user.
type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}
