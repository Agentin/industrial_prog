package http

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/student/tech-ip-sem2/services/tasks/internal/grpcclient"
	"github.com/student/tech-ip-sem2/services/tasks/internal/http/handlers"
	authMiddleware "github.com/student/tech-ip-sem2/services/tasks/internal/http/handlers/middleware"
	"github.com/student/tech-ip-sem2/services/tasks/internal/http/middleware"
	"github.com/student/tech-ip-sem2/services/tasks/internal/repository"
	sharedMW "github.com/student/tech-ip-sem2/shared/middleware"
	"github.com/student/tech-ip-sem2/shared/rabbit"
)

func NewRouter(repo repository.TaskRepository, authClient *grpcclient.AuthClient, logger *zap.Logger, publisher *rabbit.Publisher) http.Handler {
	mux := http.NewServeMux()

	handler := sharedMW.RequestIDMiddleware(mux)
	handler = sharedMW.SecurityHeadersMiddleware(handler)
	handler = sharedMW.HTTPAccessLogMiddleware(logger)(handler)

	authMW := authMiddleware.AuthMiddleware(authClient)
	csrfMW := middleware.CSRFMiddleware

	createHandler := authMW(csrfMW(handlers.CreateTaskHandler(repo, publisher)))
	updateHandler := authMW(csrfMW(handlers.UpdateTaskHandler(repo)))
	deleteHandler := authMW(csrfMW(handlers.DeleteTaskHandler(repo)))
	getAllHandler := authMW(handlers.GetTasksHandler(repo))
	getOneHandler := authMW(handlers.GetTaskHandler(repo))
	searchHandler := authMW(handlers.SearchTasksHandler(repo))

	mux.HandleFunc("GET /health", handlers.HealthHandler)

	mux.Handle("POST /v1/tasks", middleware.MetricsMiddleware("/v1/tasks")(createHandler))
	mux.Handle("GET /v1/tasks", middleware.MetricsMiddleware("/v1/tasks")(getAllHandler))
	mux.Handle("GET /v1/tasks/{id}", middleware.MetricsMiddleware("/v1/tasks/:id")(getOneHandler))
	mux.Handle("PATCH /v1/tasks/{id}", middleware.MetricsMiddleware("/v1/tasks/:id")(updateHandler))
	mux.Handle("DELETE /v1/tasks/{id}", middleware.MetricsMiddleware("/v1/tasks/:id")(deleteHandler))
	mux.Handle("GET /v1/tasks/search", middleware.MetricsMiddleware("/v1/tasks/search")(searchHandler))
	mux.Handle("GET /metrics", promhttp.Handler())

	return handler
}
