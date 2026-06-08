package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSMiddleware_Security(t *testing.T) {
	tests := []struct {
		name               string
		allowedOrigins     string
		requestOrigin      string
		expectedAllow      string
		expectedAllowCreds string
	}{
		{
			name:               "Explicit origin match - Allow credentials",
			allowedOrigins:     "https://mi-tech.millennialperfumer.in",
			requestOrigin:      "https://mi-tech.millennialperfumer.in",
			expectedAllow:      "https://mi-tech.millennialperfumer.in",
			expectedAllowCreds: "true",
		},
		{
			name:               "Wildcard origin - No credentials",
			allowedOrigins:     "*",
			requestOrigin:      "https://evil.com",
			expectedAllow:      "*",
			expectedAllowCreds: "", // Crucial security fix
		},
		{
			name:               "Multiple origins - Match first",
			allowedOrigins:     "https://a.com,https://b.com",
			requestOrigin:      "https://a.com",
			expectedAllow:      "https://a.com",
			expectedAllowCreds: "true",
		},
		{
			name:               "Disallowed origin",
			allowedOrigins:     "https://trusted.com",
			requestOrigin:      "https://evil.com",
			expectedAllow:      "",
			expectedAllowCreds: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("ALLOWED_ORIGINS", tt.allowedOrigins)

			handler := CORSMiddleware(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest("GET", "/api/health", nil)
			req.Header.Set("Origin", tt.requestOrigin)
			res := httptest.NewRecorder()

			handler.ServeHTTP(res, req)

			allowOrigin := res.Header().Get("Access-Control-Allow-Origin")
			if allowOrigin != tt.expectedAllow {
				t.Errorf("expected Allow-Origin %q, got %q", tt.expectedAllow, allowOrigin)
			}

			allowCreds := res.Header().Get("Access-Control-Allow-Credentials")
			if allowCreds != tt.expectedAllowCreds {
				t.Errorf("expected Allow-Credentials %q, got %q", tt.expectedAllowCreds, allowCreds)
			}
		})
	}
}
