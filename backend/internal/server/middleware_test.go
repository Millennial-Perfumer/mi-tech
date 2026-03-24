package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestCORSMiddleware(t *testing.T) {
	// Setup allowed origins
	os.Setenv("ALLOWED_ORIGINS", "http://allowed.com, https://another.com")
	defer os.Unsetenv("ALLOWED_ORIGINS")

	tests := []struct {
		name           string
		origin         string
		expectedOrigin string
		expectedAllow  bool
	}{
		{
			name:           "Allowed origin 1",
			origin:         "http://allowed.com",
			expectedOrigin: "http://allowed.com",
			expectedAllow:  true,
		},
		{
			name:           "Allowed origin 2",
			origin:         "https://another.com",
			expectedOrigin: "https://another.com",
			expectedAllow:  true,
		},
		{
			name:           "Disallowed origin",
			origin:         "http://evil.com",
			expectedOrigin: "",
			expectedAllow:  false,
		},
		{
			name:           "No origin",
			origin:         "",
			expectedOrigin: "",
			expectedAllow:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := CORSMiddleware(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest("GET", "/api/health", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			res := httptest.NewRecorder()

			handler.ServeHTTP(res, req)

			originHeader := res.Header().Get("Access-Control-Allow-Origin")
			if originHeader != tt.expectedOrigin {
				t.Errorf("Expected Origin header %q, got %q", tt.expectedOrigin, originHeader)
			}

			credsHeader := res.Header().Get("Access-Control-Allow-Credentials")
			if tt.expectedAllow && credsHeader != "true" {
				t.Errorf("Expected Access-Control-Allow-Credentials to be true")
			} else if !tt.expectedAllow && credsHeader != "" {
				t.Errorf("Expected Access-Control-Allow-Credentials to be empty, got %q", credsHeader)
			}
		})
	}
}
