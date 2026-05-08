package logging

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
	sugar  *zap.SugaredLogger
)

// InitLogger 初始化日志
func InitLogger(env string) error {
	var config zap.Config

	if env == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var err error
	logger, err = config.Build()
	if err != nil {
		return err
	}

	sugar = logger.Sugar()
	return nil
}

// WithRequestID 添加请求 ID 到 context
func WithRequestID(ctx context.Context) context.Context {
	reqID := ctx.Value("request_id")
	if reqID == nil {
		reqID = uuid.New().String()
	}
	return context.WithValue(ctx, "request_id", reqID)
}

// GetRequestID 从 context 获取请求 ID
func GetRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value("request_id").(string); ok {
		return reqID
	}
	return ""
}

// LogWithFields 记录带字段的日志
func LogWithFields(ctx context.Context, msg string, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("request_id", GetRequestID(ctx)),
		zap.Time("timestamp", time.Now()),
	}, fields...)
	logger.Info(msg, allFields...)
}

// Info 记录信息日志
func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

// Error 记录错误日志
func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

// Warn 记录警告日志
func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

// Debug 记录调试日志
func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

// Sync 刷新日志缓冲
func Sync() error {
	if logger != nil {
		return logger.Sync()
	}
	return nil
}
