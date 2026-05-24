package grpc

import (
	"context"

	authv1 "github.com/student/tech-ip-sem2/pkg/api/auth/v1"
	"github.com/student/tech-ip-sem2/services/auth/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	authv1.UnimplementedAuthServiceServer
	authService *service.AuthService
}

func NewAuthServer(authService *service.AuthService) *AuthServer {
	return &AuthServer{authService: authService}
}

func (s *AuthServer) Verify(ctx context.Context, req *authv1.VerifyRequest) (*authv1.VerifyResponse, error) {
	valid, subject := s.authService.ValidateToken(req.Token)
	if !valid {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}
	return &authv1.VerifyResponse{
		Valid:   valid,
		Subject: subject,
	}, nil
}
