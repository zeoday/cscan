package logger

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/trace"
)

// Fields 日志字段
type Fields map[string]interface{}

// StructuredLogger 结构化日志器
type StructuredLogger struct {
	logx.Logger
	fields Fields
}

// WithContext 创建带追踪信息的日志器
func WithContext(ctx context.Context) *StructuredLogger {
	logger := &StructuredLogger{
		Logger: logx.WithContext(ctx),
		fields: make(Fields),
	}

	// 添加追踪信息
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		logger.fields["trace_id"] = span.SpanContext().TraceID().String()
		logger.fields["span_id"] = span.SpanContext().SpanID().String()
	}

	return logger
}

// New 创建新的结构化日志器
func New() *StructuredLogger {
	return &StructuredLogger{
		Logger: logx.WithContext(context.Background()),
		fields: make(Fields),
	}
}

// With 添加字段
func (l *StructuredLogger) With(fields Fields) *StructuredLogger {
	newFields := make(Fields, len(l.fields)+len(fields))
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}
	return &StructuredLogger{
		Logger: l.Logger,
		fields: newFields,
	}
}

// WithField 添加单个字段
func (l *StructuredLogger) WithField(key string, value interface{}) *StructuredLogger {
	return l.With(Fields{key: value})
}

// Info 信息日志
func (l *StructuredLogger) Info(msg string) {
	l.Logger.Infow(msg, l.toLogFields()...)
}

// Infof 格式化信息日志
func (l *StructuredLogger) Infof(format string, args ...interface{}) {
	l.Logger.Infof(format, args...)
}

// Error 错误日志
func (l *StructuredLogger) Error(msg string, err error) {
	fields := l.toLogFields()
	if err != nil {
		fields = append(fields, logx.Field("error", err.Error()))
	}
	l.Logger.Errorw(msg, fields...)
}

// Errorf 格式化错误日志
func (l *StructuredLogger) Errorf(format string, args ...interface{}) {
	l.Logger.Errorf(format, args...)
}

// Warn 警告日志
func (l *StructuredLogger) Warn(msg string) {
	l.Logger.Sloww(msg, l.toLogFields()...)
}

// Warnf 格式化警告日志
func (l *StructuredLogger) Warnf(format string, args ...interface{}) {
	l.Logger.Slowf(format, args...)
}

// Debug 调试日志
func (l *StructuredLogger) Debug(msg string) {
	l.Logger.Debugw(msg, l.toLogFields()...)
}

// Debugf 格式化调试日志
func (l *StructuredLogger) Debugf(format string, args ...interface{}) {
	l.Logger.Debugf(format, args...)
}

func (l *StructuredLogger) toLogFields() []logx.LogField {
	fields := make([]logx.LogField, 0, len(l.fields))
	for k, v := range l.fields {
		fields = append(fields, logx.Field(k, v))
	}
	return fields
}

// 便捷函数

// TaskLogger 任务日志器
func TaskLogger(ctx context.Context, taskId, taskType string) *StructuredLogger {
	return WithContext(ctx).With(Fields{
		"task_id":   taskId,
		"task_type": taskType,
	})
}

// ScanLogger 扫描日志器
func ScanLogger(ctx context.Context, scanner, target string) *StructuredLogger {
	return WithContext(ctx).With(Fields{
		"scanner": scanner,
		"target":  target,
	})
}

// WorkerLogger Worker日志器
func WorkerLogger(ctx context.Context, workerName string) *StructuredLogger {
	return WithContext(ctx).With(Fields{
		"worker": workerName,
	})
}

// APILogger API日志器
func APILogger(ctx context.Context, method, path string) *StructuredLogger {
	return WithContext(ctx).With(Fields{
		"method": method,
		"path":   path,
	})
}

// DBLogger 数据库日志器
func DBLogger(ctx context.Context, collection, operation string) *StructuredLogger {
	return WithContext(ctx).With(Fields{
		"collection": collection,
		"operation":  operation,
	})
}
