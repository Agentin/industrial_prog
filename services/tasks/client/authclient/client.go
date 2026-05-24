package authclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/student/tech-ip-sem2/shared/httpx"
)

type AuthClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAuthClient(baseURL string) *AuthClient {
	return &AuthClient{
		baseURL:    baseURL,
		httpClient: httpx.NewClient(httpx.DefaultTimeout),
	}
}

type verifyResponse struct {
	Valid   bool   `json:"valid"`
	Subject string `json:"subject"`
	Error   string `json:"error"`
}

// Verify проверяет токен через Auth service. Возвращает (true, nil) если токен валиден.
// В противном случае возвращает (false, nil) для 401, или (false, error) для других ошибок.
func (c *AuthClient) Verify(ctx context.Context, token, requestID string) (bool, error) {
	url := fmt.Sprintf("%s/v1/auth/verify", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if requestID != "" {
		req.Header.Set("X-Request-ID", requestID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var vResp verifyResponse
		if err := json.NewDecoder(resp.Body).Decode(&vResp); err != nil {
			return false, fmt.Errorf("decode response: %w", err)
		}
		return vResp.Valid, nil
	case http.StatusUnauthorized:
		// невалидный токен – это не ошибка клиента, а статус авторизации
		return false, nil
	default:
		return false, fmt.Errorf("auth service returned status %d", resp.StatusCode)
	}
}
