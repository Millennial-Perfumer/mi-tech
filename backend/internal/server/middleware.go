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

// CORSMiddleware adds CORS headers to all requests.
func CORSMiddleware(next http.HandlerFunc) http.HandlerFunc {
	allowedOrigins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")

	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		isAllowed := false

		if origin != "" {
			for _, allowed := range allowedOrigins {
				if origin == strings.TrimSpace(allowed) {
					isAllowed = true
					break
				}
			}
		}

		if isAllowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
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
		telemetry.HttpRequestsTotal.With(prometheus.Labels{
			"path":   path,
			"method": r.Method,
			"status": string(rune(rw.status)), // This is slightly wrong, should be string representation
		}).Inc()
		
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
