package http

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/student/tech-ip-sem2/services/auth/internal/http/handlers"
	"github.com/student/tech-ip-sem2/services/auth/internal/service"
	"github.com/student/tech-ip-sem2/shared/middleware"
)

func NewRouter(svc *service.AuthService, logger *zap.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /v1/auth/login", handlers.LoginHandler(svc))
	mux.HandleFunc("GET /v1/auth/verify", handlers.VerifyHandler(svc))

	// Middleware: request-id и access log
	handler := middleware.RequestIDMiddleware(mux)
	handler = middleware.HTTPAccessLogMiddleware(logger)(handler)

	return handler
}
