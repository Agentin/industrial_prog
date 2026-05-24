package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/student/tech-ip-sem2/services/tasks/internal/metrics"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func MetricsMiddleware(route string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			metrics.HttpInFlightRequests.Inc()
			defer metrics.HttpInFlightRequests.Dec()

			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rw, r)

			duration := time.Since(start).Seconds()
			metrics.HttpRequestsTotal.WithLabelValues(r.Method, route, strconv.Itoa(rw.statusCode)).Inc()
			metrics.HttpRequestDuration.WithLabelValues(r.Method, route).Observe(duration)
		})
	}
}
