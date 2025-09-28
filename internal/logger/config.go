package logger

import (
	"os"
	"strings"
	"time"

	"go.uber.org/zap/zapcore"
)

const (
	DefaultServiceName = "mini-ecommerce"
	DefaultAppVersion  = "unknown"
	DefaultLogLevel    = zapcore.InfoLevel
	SessionTimeout     = 24 * time.Hour
)

type Config struct {
	ServiceName string
	AppVersion  string
	LogLevel    zapcore.Level
	Mode        string
}

func NewConfig() *Config {
	return &Config{
		ServiceName: getServiceName(),
		AppVersion:  getAppVersion(),
		LogLevel:    getLogLevelFromEnv(),
		Mode:        getEnvironmentMode(),
	}
}

func (c *Config) IsProduction() bool {
	mode := strings.ToLower(c.Mode)
	return mode == "release" || mode == "production"
}

func getServiceName() string {
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		return DefaultServiceName
	}
	return serviceName
}

func getAppVersion() string {
	version := os.Getenv("APP_VERSION")
	if version == "" {
		return DefaultAppVersion
	}
	return version
}

func getEnvironmentMode() string {
	return strings.ToLower(os.Getenv("GIN_MODE"))
}

func getLogLevelFromEnv() zapcore.Level {
	levelStr := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	switch levelStr {
	case "DEBUG":
		return zapcore.DebugLevel
	case "INFO":
		return zapcore.InfoLevel
	case "WARN", "WARNING":
		return zapcore.WarnLevel
	case "ERROR":
		return zapcore.ErrorLevel
	case "FATAL":
		return zapcore.FatalLevel
	default:
		return DefaultLogLevel
	}
}