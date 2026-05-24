package middleware

import "net/http"

// SecurityHeadersMiddleware устанавливает базовые заголовки безопасности.
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Защита от MIME-снифинга
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// Запрет загрузки страницы во фрейме (защита от clickjacking)
		w.Header().Set("X-Frame-Options", "DENY")
		// Простая CSP: разрешаем только ресурсы с того же источника
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		next.ServeHTTP(w, r)
	})
}
