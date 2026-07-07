package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	app   *zap.Logger
	audit *zap.Logger
}

func New() (*Logger, error) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "ts"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	encoder := zapcore.NewJSONEncoder(encoderConfig)
	level := zap.NewAtomicLevelAt(zap.InfoLevel)

	appCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
	auditCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)

	return &Logger{
		app:   zap.New(appCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)).Named("app"),
		audit: zap.New(auditCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)).Named("audit"),
	}, nil
}

func (l *Logger) App() *zap.Logger {
	return l.app
}

func (l *Logger) Audit() *zap.Logger {
	return l.audit
}

func (l *Logger) Sync() {
	if l == nil {
		return
	}

	_ = l.app.Sync()
	_ = l.audit.Sync()
}
