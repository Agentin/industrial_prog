package metrics

import (
	"net/http"
	"regexp"
	"strconv"
	"time"
)

// NormalizePath заменяет числовые идентификаторы в пути на :id
func NormalizePath(path string) string {
	// Пример: /v1/tasks/123 -> /v1/tasks/:id
	re := regexp.MustCompile(`/\d+`)
	return re.ReplaceAllString(path, "/:id")
}

// HTTPMetricsMiddleware – middleware для сбора метрик
func HTTPMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Увеличиваем in-flight при входе
		HttpInFlightRequests.Inc()
		defer HttpInFlightRequests.Dec()

		start := time.Now()

		// Оборачиваем ResponseWriter для получения статуса
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		// Вызов следующего обработчика
		next.ServeHTTP(rw, r)

		// Длительность
		duration := time.Since(start).Seconds()

		// Нормализуем путь для метрик (избегаем высокой кардинальности)
		route := NormalizePath(r.URL.Path)

		// Счётчик запросов
		HttpRequestsTotal.WithLabelValues(r.Method, route, strconv.Itoa(rw.status)).Inc()

		// Гистограмма длительности
		HttpRequestDuration.WithLabelValues(r.Method, route).Observe(duration)
	})
}

// responseWriter – обёртка для перехвата статуса
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
