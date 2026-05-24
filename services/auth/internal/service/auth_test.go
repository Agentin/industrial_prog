package service

import "testing"

func TestAuthService_CheckCredentials(t *testing.T) {
	svc := NewAuthService()

	tests := []struct {
		name     string
		username string
		password string
		want     bool
	}{
		{"valid student", "student", "student", true},
		{"wrong password", "student", "wrong", false},
		{"unknown user", "unknown", "any", false},
		{"empty credentials", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := svc.CheckCredentials(tt.username, tt.password); got != tt.want {
				t.Errorf("CheckCredentials() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	svc := NewAuthService()

	tests := []struct {
		name        string
		token       string
		wantValid   bool
		wantSubject string
	}{
		{"valid demo token", "demo-token", true, "student"},
		{"invalid token", "invalid", false, ""},
		{"empty token", "", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, subject := svc.ValidateToken(tt.token)
			if valid != tt.wantValid {
				t.Errorf("ValidateToken() valid = %v, want %v", valid, tt.wantValid)
			}
			if subject != tt.wantSubject {
				t.Errorf("ValidateToken() subject = %v, want %v", subject, tt.wantSubject)
			}
		})
	}
}
