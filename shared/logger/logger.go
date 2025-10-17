package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger interface defines the logging methods
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	Sync() error
	GetZapLogger() *zap.Logger
}

// Config represents logger configuration
type Config struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"` // json or console
	OutputPath string `mapstructure:"output_path"`
}

// logger implements the Logger interface
type logger struct {
	zapLogger *zap.Logger
}

func NewLogger(config Config, serviceName string) (Logger, error) {
	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	var encoder zapcore.Encoder
	if config.Format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	var outputPaths []string
	if config.OutputPath != "" {
		outputPaths = append(outputPaths, config.OutputPath)
	} else {
		outputPaths = append(outputPaths, "stdout")
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)),
		level,
	)

	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	
	// Add service name to all logs
	zapLogger = zapLogger.With(zap.String("service", serviceName))

	return &logger{zapLogger: zapLogger}, nil
}

// NewConfig creates a default logger configuration
func NewConfig() Config {
	return Config{
		Level:      "info",
		Format:     "json",
		OutputPath: "",
	}
}

// Debug logs a debug message
func (l *logger) Debug(msg string, fields ...zap.Field) {
	l.zapLogger.Debug(msg, fields...)
}

// Info logs an info message
func (l *logger) Info(msg string, fields ...zap.Field) {
	l.zapLogger.Info(msg, fields...)
}

// Warn logs a warning message
func (l *logger) Warn(msg string, fields ...zap.Field) {
	l.zapLogger.Warn(msg, fields...)
}

// Error logs an error message
func (l *logger) Error(msg string, fields ...zap.Field) {
	l.zapLogger.Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func (l *logger) Fatal(msg string, fields ...zap.Field) {
	l.zapLogger.Fatal(msg, fields...)
}

// Sync flushes any buffered log entries
func (l *logger) Sync() error {
	return l.zapLogger.Sync()
}

// GetZapLogger returns the underlying zap logger
func (l *logger) GetZapLogger() *zap.Logger {
	return l.zapLogger
}
