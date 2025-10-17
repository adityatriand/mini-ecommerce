package database

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config represents database configuration
type Config struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// Connect establishes a database connection
func Connect(config Config, zapLogger *zap.Logger) (*gorm.DB, error) {
	zapLogger.Info("Connecting to database...")

	// Configure GORM logger
	gormLogger := logger.Default.LogMode(logger.Info)
	if zapLogger.Core().Enabled(zapcore.DebugLevel) {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(postgres.Open(config.URL), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		zapLogger.Error("Failed to connect to database", zap.Error(err))
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		zapLogger.Error("Failed to get underlying sql.DB", zap.Error(err))
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		zapLogger.Error("Failed to ping database", zap.Error(err))
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	zapLogger.Info("Database connection established successfully",
		zap.Int("max_open_conns", config.MaxOpenConns),
		zap.Int("max_idle_conns", config.MaxIdleConns),
		zap.Duration("conn_max_lifetime", config.ConnMaxLifetime),
	)

	return db, nil
}

// Migrate runs database migrations
func Migrate(db *gorm.DB, models []interface{}, zapLogger *zap.Logger) error {
	zapLogger.Info("Starting database migration...")

	if err := db.AutoMigrate(models...); err != nil {
		zapLogger.Error("Database migration failed", zap.Error(err))
		return fmt.Errorf("database migration failed: %w", err)
	}

	zapLogger.Info("Database migration completed successfully")
	return nil
}

// HealthCheck checks database health
func HealthCheck(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}
