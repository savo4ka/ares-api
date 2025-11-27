package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/savo4ka/ares-api/internal/metrics"
)

// CORSMiddleware добавляет CORS заголовки к ответам
func CORSMiddleware(allowedOrigins string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Если разрешены все origins (*) или origin в списке разрешённых
			if allowedOrigins == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if origin != "" {
				allowed := strings.Split(allowedOrigins, ",")
				for _, o := range allowed {
					if strings.TrimSpace(o) == origin {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						break
					}
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "3600")

			// Обработка preflight запросов
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// LoggingMiddleware логирует все HTTP запросы
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Простое логирование - можно расширить
		next.ServeHTTP(w, r)
	})
}

// responseWriter является wrapper для http.ResponseWriter для отслеживания status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// MetricsMiddleware отслеживает метрики HTTP запросов
func MetricsMiddleware(m *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Пропускаем метрики для /metrics эндпоинта
			if r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			rw := newResponseWriter(w)

			next.ServeHTTP(rw, r)

			duration := time.Since(start).Seconds()
			statusCode := strconv.Itoa(rw.statusCode)

			// Записываем метрики
			m.HTTPRequestsTotal.WithLabelValues(r.Method, r.URL.Path, statusCode).Inc()
			m.HTTPRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
		})
	}
}
