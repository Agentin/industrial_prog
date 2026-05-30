package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/student/tech-ip-sem2/services/tasks/internal/grpcclient"
	taskshttp "github.com/student/tech-ip-sem2/services/tasks/internal/http"
	"github.com/student/tech-ip-sem2/services/tasks/internal/http/handlers"
	"github.com/student/tech-ip-sem2/services/tasks/internal/repository"
	"github.com/student/tech-ip-sem2/shared/cache"
	"github.com/student/tech-ip-sem2/shared/logger"
	"github.com/student/tech-ip-sem2/shared/rabbit"
)

func main() {
	log, err := logger.New("tasks", zap.InfoLevel)
	if err != nil {
		panic(err)
	}
	defer log.Sync()
	handlers.SetLogger(log)

	tasksPort := os.Getenv("TASKS_PORT")
	if tasksPort == "" {
		tasksPort = "8082"
	}
	authGrpcAddr := os.Getenv("AUTH_GRPC_ADDR")
	if authGrpcAddr == "" {
		authGrpcAddr = "localhost:50051"
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgres"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "tasksdb"
	}
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	repo, err := repository.NewPostgresTaskRepo(connStr)
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}
	defer repo.Close()

	var redisClient *cache.RedisClient
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisClient, err = cache.NewRedisClient(cache.Config{
		Addrs:    []string{redisAddr},
		Password: os.Getenv("REDIS_PASSWORD"),
		PoolSize: 10,
	}, log)
	if err != nil {
		log.Warn("failed to connect to redis, caching disabled", zap.Error(err))
		redisClient = nil
	}

	var finalRepo repository.TaskRepository = repo
	if redisClient != nil {
		finalRepo = repository.NewCachingTaskRepository(repo, redisClient, log)
		log.Info("redis caching enabled")
	} else {
		log.Warn("redis caching disabled")
	}

	authClient, err := grpcclient.NewAuthClient(authGrpcAddr)
	if err != nil {
		log.Fatal("failed to create auth gRPC client", zap.Error(err))
	}
	defer authClient.Close()
	authClient.SetLogger(log)

	var eventPublisher *rabbit.Publisher
	rabbitURL := os.Getenv("RABBIT_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}
	eventQueue := os.Getenv("QUEUE_NAME")
	if eventQueue == "" {
		eventQueue = "task_events"
	}
	eventPublisher, err = rabbit.NewPublisher(rabbitURL, eventQueue)
	if err != nil {
		log.Warn("failed to connect to RabbitMQ for events, events disabled", zap.Error(err))
		eventPublisher = nil
	} else {
		log.Info("RabbitMQ event publisher connected", zap.String("queue", eventQueue))
	}
	defer func() {
		if eventPublisher != nil {
			eventPublisher.Close()
		}
	}()

	var jobPublisher *rabbit.Publisher
	jobQueue := os.Getenv("JOB_QUEUE_NAME")
	if jobQueue == "" {
		jobQueue = "task_jobs"
	}
	jobPublisher, err = rabbit.NewPublisher(rabbitURL, jobQueue)
	if err != nil {
		log.Warn("failed to connect to RabbitMQ for jobs, jobs disabled", zap.Error(err))
		jobPublisher = nil
	} else {
		log.Info("RabbitMQ job publisher connected", zap.String("queue", jobQueue))
	}
	defer func() {
		if jobPublisher != nil {
			jobPublisher.Close()
		}
	}()

	router := taskshttp.NewRouter(finalRepo, authClient, log, eventPublisher, jobPublisher)

	server := &http.Server{
		Addr:         ":" + tasksPort,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info("Tasks service started", zap.String("port", tasksPort), zap.String("auth_grpc_addr", authGrpcAddr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", zap.Error(err))
		}
	}()

	<-done
	log.Info("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("shutdown error", zap.Error(err))
	}
}
