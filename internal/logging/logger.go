package logging

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	app        *zap.Logger
	audit      *zap.Logger
	csvWriters []*csvLogWriter
}

type csvLogWriter struct {
	mu     sync.Mutex
	file   *os.File
	writer *csv.Writer
}

func New() (*Logger, error) {
	if err := os.MkdirAll("logs", 0o755); err != nil {
		return nil, fmt.Errorf("create logs dir failed: %w", err)
	}

	appWriter, err := newCSVLogWriter(filepath.Join("logs", "runtime.csv"))
	if err != nil {
		return nil, err
	}

	auditWriter, err := newCSVLogWriter(filepath.Join("logs", "audit.csv"))
	if err != nil {
		_ = appWriter.Close()
		return nil, err
	}

	appCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(appWriter),
		zap.NewAtomicLevelAt(zap.InfoLevel),
	)
	auditCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(auditWriter),
		zap.NewAtomicLevelAt(zap.InfoLevel),
	)

	return &Logger{
		app:        zap.New(appCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)).Named("app"),
		audit:      zap.New(auditCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)).Named("audit"),
		csvWriters: []*csvLogWriter{appWriter, auditWriter},
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

	for _, writer := range l.csvWriters {
		_ = writer.Close()
	}
}

func newCSVLogWriter(path string) (*csvLogWriter, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file %s failed: %w", path, err)
	}

	writer := csv.NewWriter(file)
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("stat log file %s failed: %w", path, err)
	}

	if info.Size() == 0 {
		if err := writer.Write([]string{"timestamp", "level", "logger", "message", "caller", "trace_id", "fields_json"}); err != nil {
			_ = file.Close()
			return nil, fmt.Errorf("write log header failed: %w", err)
		}
		writer.Flush()
		if err := writer.Error(); err != nil {
			_ = file.Close()
			return nil, fmt.Errorf("flush log header failed: %w", err)
		}
	}

	return &csvLogWriter{
		file:   file,
		writer: writer,
	}, nil
}

func (w *csvLogWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	var entry map[string]any
	if err := json.Unmarshal(p, &entry); err != nil {
		return 0, err
	}

	timestamp := time.Now().Format(time.RFC3339)
	if ts, ok := entry["ts"].(string); ok && ts != "" {
		timestamp = ts
	}

	level, _ := entry["level"].(string)
	loggerName, _ := entry["logger"].(string)
	message, _ := entry["msg"].(string)
	caller, _ := entry["caller"].(string)

	traceID := ""
	if value, ok := entry["trace_id"].(string); ok {
		traceID = value
	}

	delete(entry, "ts")
	delete(entry, "level")
	delete(entry, "logger")
	delete(entry, "msg")
	delete(entry, "caller")
	delete(entry, "trace_id")

	fieldsJSON, err := json.Marshal(entry)
	if err != nil {
		return 0, err
	}

	record := []string{
		timestamp,
		level,
		loggerName,
		message,
		caller,
		traceID,
		string(fieldsJSON),
	}

	if err := w.writer.Write(record); err != nil {
		return 0, err
	}
	w.writer.Flush()
	if err := w.writer.Error(); err != nil {
		return 0, err
	}

	return len(p), nil
}

func (w *csvLogWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.writer.Flush()
	if err := w.writer.Error(); err != nil {
		return err
	}

	return w.file.Sync()
}

func (w *csvLogWriter) Close() error {
	_ = w.Sync()
	return w.file.Close()
}
