package httpx

import (
	"net/http"
	"time"
)

// NewClient создаёт HTTP-клиент с заданным таймаутом.
func NewClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
	}
}

// DefaultTimeout стандартный таймаут для межсервисных вызовов.
const DefaultTimeout = 3 * time.Second