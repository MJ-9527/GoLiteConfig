package logging

import (
	"context"

	"go.uber.org/zap"
)

type contextKey string

const loggerContextKey contextKey = "logger"

func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey, logger)
}

func FromContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return zap.NewNop()
	}

	logger, ok := ctx.Value(loggerContextKey).(*zap.Logger)
	if !ok || logger == nil {
		return zap.NewNop()
	}

	return logger
}
