package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Logs the HTTP request and response details like status code, duration, path, remote address & user agent
func LoggingMiddleware(logger *slog.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a response wrapper to capture the status code
			wrapper := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			start := time.Now()

			// Call the next handler
			next.ServeHTTP(wrapper, r)

			queryValues := r.URL.Query()
			for key, values := range queryValues {
				if key == "name" || key == "address" {
					for i := range values {
						values[i] = "[REDACTED]"
					}
				}
			}
			query := queryValues.Encode()

			// Log response details
			duration := time.Since(start)
			logger.Info("HTTP Request",
				"method", r.Method,
				"path", r.URL.Path,
				"query", query,
				"status", wrapper.statusCode,
				"duration_ms", duration.Milliseconds(),
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
		})
	}
}

// responseWriter is a wrapper for http.ResponseWriter that captures the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
