package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestCORSMiddleware(t *testing.T) {
	// Setup allowed origins
	os.Setenv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173,http://localhost:8080,https://mi-tech.millennialperfumer.in,https://feedback-form.millennialperfumer.in")
	defer os.Unsetenv("ALLOWED_ORIGINS")

	tests := []struct {
		name           string
		origin         string
		expectedOrigin string
		expectedAllow  bool
	}{
		{
			name:           "Production Domain - Admin UI",
			origin:         "https://mi-tech.millennialperfumer.in",
			expectedOrigin: "https://mi-tech.millennialperfumer.in",
			expectedAllow:  true,
		},
		{
			name:           "Production Domain - Admin UI with Slash",
			origin:         "https://mi-tech.millennialperfumer.in/",
			expectedOrigin: "https://mi-tech.millennialperfumer.in/",
			expectedAllow:  true,
		},
		{
			name:           "Production Domain - Feedback Form",
			origin:         "https://feedback-form.millennialperfumer.in",
			expectedOrigin: "https://feedback-form.millennialperfumer.in",
			expectedAllow:  true,
		},
		{
			name:           "Localhost Dev",
			origin:         "http://localhost:5173",
			expectedOrigin: "http://localhost:5173",
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

func TestCORSMiddlewareWildcard(t *testing.T) {
	os.Setenv("ALLOWED_ORIGINS", "*")
	defer os.Unsetenv("ALLOWED_ORIGINS")

	handler := CORSMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/health", nil)
	req.Header.Set("Origin", "http://evil.com")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	// Currently, it reflects the origin and allows credentials - this is the vulnerability
	originHeader := res.Header().Get("Access-Control-Allow-Origin")
	credsHeader := res.Header().Get("Access-Control-Allow-Credentials")

	// Verify it meets our NEW desired state (Security Enhancement)
	if originHeader != "*" {
		t.Errorf("Expected Origin header '*', got %q", originHeader)
	}
	if credsHeader != "" {
		t.Errorf("Expected empty Access-Control-Allow-Credentials, got %q", credsHeader)
	}
}
