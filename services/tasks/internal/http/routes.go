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
)

func NewRouter(repo repository.TaskRepository, authClient *grpcclient.AuthClient, logger *zap.Logger) http.Handler {
	mux := http.NewServeMux()

	// 1. Базовые middleware, оборачивающие весь маршрутизатор
	handler := sharedMW.RequestIDMiddleware(mux)
	handler = sharedMW.SecurityHeadersMiddleware(handler)
	handler = sharedMW.HTTPAccessLogMiddleware(logger)(handler)

	// 2. Middleware для аутентификации и CSRF (это функции-обёртки)
	authMW := authMiddleware.AuthMiddleware(authClient) // func(http.Handler) http.Handler
	csrfMW := middleware.CSRFMiddleware                 // func(http.Handler) http.Handler

	// 3. Создаём конечные обработчики, обёрнутые в нужные middleware
	// Для методов, изменяющих состояние: аутентификация + CSRF
	createHandler := authMW(csrfMW(http.HandlerFunc(handlers.CreateTaskHandler(repo))))
	updateHandler := authMW(csrfMW(http.HandlerFunc(handlers.UpdateTaskHandler(repo))))
	deleteHandler := authMW(csrfMW(http.HandlerFunc(handlers.DeleteTaskHandler(repo))))

	// Для GET-запросов: только аутентификация (без CSRF)
	getAllHandler := authMW(http.HandlerFunc(handlers.GetTasksHandler(repo)))
	getOneHandler := authMW(http.HandlerFunc(handlers.GetTaskHandler(repo)))
	searchHandler := authMW(http.HandlerFunc(handlers.SearchTasksHandler(repo)))

	// 4. Добавляем маршруты с метриками
	mux.Handle("POST /v1/tasks", middleware.MetricsMiddleware("/v1/tasks")(createHandler))
	mux.Handle("GET /v1/tasks", middleware.MetricsMiddleware("/v1/tasks")(getAllHandler))
	mux.Handle("GET /v1/tasks/{id}", middleware.MetricsMiddleware("/v1/tasks/:id")(getOneHandler))
	mux.Handle("PATCH /v1/tasks/{id}", middleware.MetricsMiddleware("/v1/tasks/:id")(updateHandler))
	mux.Handle("DELETE /v1/tasks/{id}", middleware.MetricsMiddleware("/v1/tasks/:id")(deleteHandler))
	mux.Handle("GET /v1/tasks/search", middleware.MetricsMiddleware("/v1/tasks/search")(searchHandler))

	// 5. Эндпоинт метрик (без аутентификации)
	mux.Handle("GET /metrics", promhttp.Handler())

	return handler
}
