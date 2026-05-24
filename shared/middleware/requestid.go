package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type contextKey string

const RequestIDKey contextKey = "requestID"

// RequestIDMiddleware читает или генерирует X-Request-ID и сохраняет в контекст.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		// Добавляем в ответ, чтобы клиент видел тот же ID
		w.Header().Set("X-Request-ID", requestID)
		// Сохраняем в контекст
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID извлекает requestID из контекста.
func GetRequestID(ctx context.Context) string {
	if val, ok := ctx.Value(RequestIDKey).(string); ok {
		return val
	}
	return ""
}

// generateRequestID создаёт случайный 16-байтовый hex-идентификатор.
func generateRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// fallback: используем фиксированную строку, если rand сломался
		return "fallback-request-id"
	}
	return hex.EncodeToString(b)
}