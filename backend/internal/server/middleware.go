package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"mi-tech/internal/service"
	"mi-tech/internal/telemetry"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/prometheus/client_golang/prometheus"
)

// normalizeOrigin trims spaces and trailing slashes from a URL string for consistent comparison.
func normalizeOrigin(url string) string {
	return strings.TrimRight(strings.TrimSpace(url), "/")
}

// CORSMiddleware adds CORS headers to all requests.
func CORSMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")
		rawOrigin := r.Header.Get("Origin")
		origin := normalizeOrigin(rawOrigin)
		isAllowed := false

		// Always set Vary: Origin to handle caching behind proxies/CDNs
		w.Header().Add("Vary", "Origin")

		isWildcardMatch := false
		if rawOrigin == "" {
			// If no origin, we can consider it a direct request or same-origin
			isAllowed = true
		} else if allowedOriginsEnv != "" {
			allowedOrigins := strings.Split(allowedOriginsEnv, ",")
			for _, allowed := range allowedOrigins {
				trimmed := normalizeOrigin(allowed)
				if trimmed == "*" {
					isAllowed = true
					isWildcardMatch = true
					// Continue checking in case there's an explicit match later
				} else if origin == trimmed {
					isAllowed = true
					isWildcardMatch = false
					break
				}
			}
		}

		if isAllowed {
			if rawOrigin != "" {
				if isWildcardMatch {
					// Security: If allowed by wildcard, do NOT allow credentials and use '*' for origin
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else {
					// Use the actual origin for the allow header to support multiple origins with credentials
					w.Header().Set("Access-Control-Allow-Origin", rawOrigin)
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-Requested-With, X-Amz-Date, X-Api-Key, X-Amz-Security-Token, X-Forwarded-For, X-Real-IP, Origin, Access-Control-Request-Method, Access-Control-Request-Headers")
				w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
			}
		} else {
			log.Printf("CORS REJECTED: Origin=[%s] (normalized=[%s]) Method=[%s] Path=[%s] not in ALLOWED_ORIGINS=[%s]", 
				rawOrigin, origin, r.Method, r.URL.Path, allowedOriginsEnv)
		}

		if r.Method == http.MethodOptions {
			if isAllowed {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusForbidden)
			}
			return
		}

		next(w, r)
	}
}

func AuthMiddleware(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			log.Printf("AuthMiddleware: %s %s", r.Method, path)
			if r.Method == "OPTIONS" ||
			   strings.HasPrefix(path, "/api/webhooks") || 
			   path == "/api/health" || 
			   path == "/api/auth/login" {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "authorization header missing", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]
			token, err := authService.ValidateToken(tokenString)
			if err != nil || !token.Valid {
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "invalid token claims", http.StatusUnauthorized)
				return
			}
			
			role, _ := claims["role"].(string)
			if role == "" {
				role = "read" // default fallback
			}

			username, _ := claims["username"].(string)

			// Add role and username to context
			log.Printf("AuthMiddleware: user=%s role=%s", username, role)
			ctx := context.WithValue(r.Context(), "userRole", role)
			ctx = context.WithValue(ctx, "username", username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole enforces that the user has one of the allowed roles.
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value("userRole").(string)
			if !ok {
				http.Error(w, "unauthorized: role not found in context", http.StatusUnauthorized)
				return
			}

			isAllowed := false
			for _, allowedRole := range allowedRoles {
				if role == allowedRole {
					isAllowed = true
					break
				}
			}

			if !isAllowed {
				log.Printf("RequireRole: forbidden. role=%s, allowed=%v", role, allowedRoles)
				http.Error(w, "forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			log.Printf("RequireRole: success. role=%s, allowed=%v", role, allowedRoles)
			next.ServeHTTP(w, r)
		})
	}
}

// MetricsMiddleware tracks HTTP requests with Prometheus.
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create a custom response writer to capture the status code
		rw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		
		next.ServeHTTP(rw, r)
		
		duration := time.Since(start).Seconds()
		path := r.URL.Path
		
		// Clean up path for cardinality (e.g. /api/orders/1 -> /api/orders/:id if possible)
		// For now simple path tracking
		
		// Better status as string
		statusStr := fmt.Sprintf("%d", rw.status)
		telemetry.HttpRequestsTotal.With(prometheus.Labels{
			"path":   path,
			"method": r.Method,
			"status": statusStr,
		}).Inc()

		telemetry.HttpRequestDuration.With(prometheus.Labels{
			"path":   path,
			"method": r.Method,
		}).Observe(duration)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
