package grpcclient

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	authv1 "github.com/student/tech-ip-sem2/pkg/api/auth/v1"
)

type AuthClient struct {
	client authv1.AuthServiceClient
	conn   *grpc.ClientConn
	logger *zap.Logger
}

func NewAuthClient(addr string) (*AuthClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	// logger будет установлен позже через SetLogger
	return &AuthClient{
		client: authv1.NewAuthServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *AuthClient) SetLogger(logger *zap.Logger) {
	c.logger = logger
}

func (c *AuthClient) Close() error {
	return c.conn.Close()
}

func (c *AuthClient) Verify(ctx context.Context, token, requestID string) (bool, string, error) {
	// Добавляем request-id в metadata
	ctx = metadata.AppendToOutgoingContext(ctx, "x-request-id", requestID)

	// Устанавливаем таймаут (можно сделать настраиваемым)
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	resp, err := c.client.Verify(ctx, &authv1.VerifyRequest{Token: token})
	if err != nil {
		if c.logger != nil {
			c.logger.Error("gRPC verify failed",
				zap.Error(err),
				zap.String("request_id", requestID),
				zap.String("token_preview", token[:min(len(token), 4)]+"..."),
			)
		}
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.Unauthenticated:
				return false, "", nil
			default:
				return false, "", err
			}
		}
		return false, "", err
	}
	return resp.Valid, resp.Subject, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
