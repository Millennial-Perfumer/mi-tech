package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSMiddleware_Wildcard(t *testing.T) {
	// Setup allowed origins with wildcard
	t.Setenv("ALLOWED_ORIGINS", "http://localhost:3000,*")

	tests := []struct {
		name           string
		origin         string
		expectedOrigin string
		expectedAllow  bool
		isWildcard     bool
	}{
		{
			name:           "Specific Allowed Origin",
			origin:         "http://localhost:3000",
			expectedOrigin: "http://localhost:3000",
			expectedAllow:  true,
			isWildcard:     false,
		},
		{
			name:           "Random Origin (Allowed via Wildcard)",
			origin:         "http://random.com",
			expectedOrigin: "*",
			expectedAllow:  true,
			isWildcard:     true,
		},
		{
			name:           "No origin",
			origin:         "",
			expectedOrigin: "",
			expectedAllow:  false,
			isWildcard:     false,
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
			if tt.expectedAllow {
				if tt.isWildcard {
					if credsHeader != "" {
						t.Errorf("Expected Access-Control-Allow-Credentials to be empty for wildcard")
					}
				} else {
					if credsHeader != "true" {
						t.Errorf("Expected Access-Control-Allow-Credentials to be true for specific origin")
					}
				}
			} else {
				if credsHeader != "" {
					t.Errorf("Expected Access-Control-Allow-Credentials to be empty, got %q", credsHeader)
				}
			}
		})
	}
}
