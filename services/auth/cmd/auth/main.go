package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	authv1 "github.com/student/tech-ip-sem2/pkg/api/auth/v1"
	authgrpc "github.com/student/tech-ip-sem2/services/auth/internal/grpc"
	authhttp "github.com/student/tech-ip-sem2/services/auth/internal/http"
	"github.com/student/tech-ip-sem2/services/auth/internal/service"
	"github.com/student/tech-ip-sem2/shared/logger"
	"github.com/student/tech-ip-sem2/shared/middleware"
)

func main() {
	// Инициализация логгера
	log, err := logger.New("auth", zap.InfoLevel)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	httpPort := os.Getenv("AUTH_PORT")
	if httpPort == "" {
		httpPort = "8081"
	}
	grpcPort := os.Getenv("AUTH_GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	svc := service.NewAuthService()

	// ----- gRPC сервер -----
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.GrpcUnaryServerInterceptor(log)),
	)
	authv1.RegisterAuthServiceServer(grpcServer, authgrpc.NewAuthServer(svc))

	go func() {
		log.Info("gRPC server started", zap.String("port", grpcPort))
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("gRPC server error", zap.Error(err))
		}
	}()

	// ----- HTTP сервер -----
	router := authhttp.NewRouter(svc, log)
	httpServer := &http.Server{
		Addr:         ":" + httpPort,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Info("HTTP server started", zap.String("port", httpPort))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("HTTP server error", zap.Error(err))
		}
	}()

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-done
	log.Info("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	grpcServer.GracefulStop()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error("HTTP shutdown error", zap.Error(err))
	}
}
