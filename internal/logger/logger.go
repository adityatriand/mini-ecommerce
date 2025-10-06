package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	Sync() error
	WithContext(c any) ContextLogger
	GetZapLogger() *zap.Logger
}

type ZapLogger struct {
	logger *zap.Logger
	config *Config
}

func NewLogger(config *Config) (Logger, error) {
	logger, err := createLogger(config)
	if err != nil {
		return nil, err
	}

	return &ZapLogger{
		logger: logger,
		config: config,
	}, nil
}

func createLogger(config *Config) (*zap.Logger, error) {
	if config.IsProduction() {
		return createProductionLogger(config)
	}
	return createDevelopmentLogger(config)
}

func createProductionLogger(config *Config) (*zap.Logger, error) {
	zapConfig := zap.NewProductionConfig()
	zapConfig.Level = zap.NewAtomicLevelAt(config.LogLevel)
	zapConfig.EncoderConfig.TimeKey = "timestamp"
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapConfig.InitialFields = map[string]any{
		"service": config.ServiceName,
		"version": config.AppVersion,
	}
	return zapConfig.Build()
}

func createDevelopmentLogger(config *Config) (*zap.Logger, error) {
	zapConfig := zap.NewDevelopmentConfig()
	zapConfig.Level = zap.NewAtomicLevelAt(config.LogLevel)
	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return zapConfig.Build()
}

func (l *ZapLogger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

func (l *ZapLogger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

func (l *ZapLogger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

func (l *ZapLogger) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, fields...)
}

func (l *ZapLogger) Fatal(msg string, fields ...zap.Field) {
	l.logger.Fatal(msg, fields...)
}

func (l *ZapLogger) Sync() error {
	if l.logger != nil {
		return l.logger.Sync()
	}
	return nil
}

func (l *ZapLogger) WithContext(c any) ContextLogger {
	requestID, userID := extractContextValues(c)
	return NewContextLogger(l.logger, requestID, userID)
}

func (l *ZapLogger) GetZapLogger() *zap.Logger {
	return l.logger
}
