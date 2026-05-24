package middleware

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// GrpcUnaryServerInterceptor логирует входящие gRPC вызовы.
// Извлекает request-id из метаданных и добавляет в лог.
func GrpcUnaryServerInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Извлекаем request-id из metadata
		requestID := ""
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if vals := md.Get("x-request-id"); len(vals) > 0 {
				requestID = vals[0]
			}
		}

		// Выполняем обработчик
		resp, err := handler(ctx, req)

		duration := time.Since(start)

		// Логируем
		logger.Info("gRPC request",
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
			zap.String("request_id", requestID),
			zap.Error(err),
		)

		return resp, err
	}
}
