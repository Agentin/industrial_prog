package middleware

import (
	"net/http"
)

// CSRFMiddleware защищает от CSRF атак через Double Submit Cookie.
// Проверяет для методов POST, PATCH, DELETE наличие заголовка X-CSRF-Token,
// который должен совпадать со значением cookie csrf_token.
func CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Защищаем только методы, изменяющие состояние
		if r.Method == http.MethodPost || r.Method == http.MethodPatch || r.Method == http.MethodDelete {
			// Получаем csrf_token из cookie
			cookie, err := r.Cookie("csrf_token")
			if err != nil {
				http.Error(w, `{"error":"missing csrf_token cookie"}`, http.StatusForbidden)
				return
			}
			csrfCookie := cookie.Value

			// Получаем заголовок X-CSRF-Token
			csrfHeader := r.Header.Get("X-CSRF-Token")
			if csrfHeader == "" {
				http.Error(w, `{"error":"missing X-CSRF-Token header"}`, http.StatusForbidden)
				return
			}

			// Сравниваем значения
			if csrfCookie != csrfHeader {
				http.Error(w, `{"error":"invalid CSRF token"}`, http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
