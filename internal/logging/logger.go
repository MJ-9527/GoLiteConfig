package logging

import (
	"go.uber.org/zap"
)

type Logger struct {
	app   *zap.Logger
	audit *zap.Logger
}

func New() (*Logger, error) {
	base, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	return &Logger{
		app:   base.Named("app"),
		audit: base.Named("audit"),
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
