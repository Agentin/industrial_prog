package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New создаёт новый zap.Logger с JSON форматом и настройками по умолчанию.
// serviceName передаётся как поле "service" в каждом лог-событии.
func New(serviceName string, level zapcore.Level) (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.Level = zap.NewAtomicLevelAt(level)
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	// Добавляем глобальное поле service
	return logger.With(zap.String("service", serviceName)), nil
}
