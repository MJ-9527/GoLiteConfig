package service

import (
	"context"

	"go.uber.org/zap"
)

type AuditLogger struct {
	logger *zap.Logger
}

func NewAuditLogger(logger *zap.Logger) *AuditLogger {
	return &AuditLogger{logger: logger}
}

func (a *AuditLogger) Log(ctx context.Context, action string, fields ...zap.Field) {
	if a == nil || a.logger == nil {
		return
	}

	payload := []zap.Field{zap.String("action", action)}
	payload = append(payload, fields...)
	a.logger.Info("audit_event", payload...)
}
