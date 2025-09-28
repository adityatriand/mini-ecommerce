package logger

import "go.uber.org/zap"

type ContextLogger interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
}

type contextLogger struct {
	logger    *zap.Logger
	requestID string
	userID    string
}

func (cl *contextLogger) baseFields() []zap.Field {
	return []zap.Field{
		zap.String(RequestIDKey, cl.requestID),
		zap.String(UserIDKey, cl.userID),
	}
}

func (cl *contextLogger) Info(msg string, fields ...zap.Field) {
	allFields := append(cl.baseFields(), fields...)
	cl.logger.Info(msg, allFields...)
}

func (cl *contextLogger) Error(msg string, fields ...zap.Field) {
	allFields := append(cl.baseFields(), fields...)
	cl.logger.Error(msg, allFields...)
}

func (cl *contextLogger) Debug(msg string, fields ...zap.Field) {
	allFields := append(cl.baseFields(), fields...)
	cl.logger.Debug(msg, allFields...)
}

func (cl *contextLogger) Warn(msg string, fields ...zap.Field) {
	allFields := append(cl.baseFields(), fields...)
	cl.logger.Warn(msg, allFields...)
}

func (cl *contextLogger) Fatal(msg string, fields ...zap.Field) {
	allFields := append(cl.baseFields(), fields...)
	cl.logger.Fatal(msg, allFields...)
}

func NewContextLogger(logger *zap.Logger, requestID, userID string) ContextLogger {
	return &contextLogger{
		logger:    logger,
		requestID: requestID,
		userID:    userID,
	}
}