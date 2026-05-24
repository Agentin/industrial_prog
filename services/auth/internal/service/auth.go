package service

// User представляет учётные данные пользователя (упрощённо).
type User struct {
	Username string
	Password string
}

// AuthService отвечает за логику аутентификации.
type AuthService struct {
	// В реальном проекте здесь было бы хранилище пользователей.
	validUsers map[string]string // username -> password
	validToken string
}

// NewAuthService создаёт новый экземпляр с тестовыми данными.
func NewAuthService() *AuthService {
	return &AuthService{
		validUsers: map[string]string{
			"student": "student",
		},
		validToken: "demo-token",
	}
}

// CheckCredentials проверяет логин и пароль.
func (s *AuthService) CheckCredentials(username, password string) bool {
	if pass, ok := s.validUsers[username]; ok {
		return pass == password
	}
	return false
}

// ValidateToken проверяет, что токен соответствует ожидаемому.
func (s *AuthService) ValidateToken(token string) (bool, string) {
	if token == s.validToken {
		return true, "student"
	}
	return false, ""
}