package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NoopLogger discards every log call. It is used by components (e.g. the
// dependency guard) that require a LoggerInterface but should not emit logs,
// and by unit tests that do not want log noise.
type NoopLogger struct{}

func (NoopLogger) Info(string, ...zap.Field)  {}
func (NoopLogger) Fatal(string, ...zap.Field) {}
func (NoopLogger) Debug(string, ...zap.Field) {}
func (NoopLogger) Error(string, ...zap.Field) {}
func (NoopLogger) Warn(string, ...zap.Field)  {}

func (NoopLogger) Check(zapcore.Level, string) *zapcore.CheckedEntry {
	return nil
}

func (NoopLogger) With(...zap.Field) LoggerInterface { return NoopLogger{} }
func (NoopLogger) Sync() error                       { return nil }
