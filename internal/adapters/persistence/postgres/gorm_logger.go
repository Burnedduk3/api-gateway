// internal/adapters/persistence/gorm_logger.go
package persistence

import (
	"context"
	"errors"
	"time"

	"product-service/pkg/logger"

	gormLogger "gorm.io/gorm/logger"
)

// GormZapLogger adapts your zap logger to work with GORM
type GormZapLogger struct {
	logger                    logger.Logger
	logLevel                  gormLogger.LogLevel
	ignoreRecordNotFoundError bool
	slowThreshold             time.Duration
}

// NewGormZapLogger creates a new GORM logger using your zap logger
func NewGormZapLogger(zapLogger logger.Logger) gormLogger.Interface {
	return &GormZapLogger{
		logger:                    zapLogger.With("component", "gorm"),
		logLevel:                  gormLogger.Info,
		ignoreRecordNotFoundError: true,
		slowThreshold:             200 * time.Millisecond,
	}
}

// NewGormZapLoggerWithConfig creates a new GORM logger with custom configuration
func NewGormZapLoggerWithConfig(zapLogger logger.Logger, config GormLoggerConfig) gormLogger.Interface {
	return &GormZapLogger{
		logger:                    zapLogger.With("component", "gorm"),
		logLevel:                  config.LogLevel,
		ignoreRecordNotFoundError: config.IgnoreRecordNotFoundError,
		slowThreshold:             config.SlowThreshold,
	}
}

// GormLoggerConfig provides configuration for the GORM logger
type GormLoggerConfig struct {
	LogLevel                  gormLogger.LogLevel
	IgnoreRecordNotFoundError bool
	SlowThreshold             time.Duration
}

// LogMode implements gorm.io/gorm/logger.Interface
func (l *GormZapLogger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	newLogger := *l
	newLogger.logLevel = level
	return &newLogger
}

// Info implements gorm.io/gorm/logger.Interface
func (l *GormZapLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormLogger.Info {
		l.logger.Info(msg, data...)
	}
}

// Warn implements gorm.io/gorm/logger.Interface
func (l *GormZapLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormLogger.Warn {
		l.logger.Warn(msg, data...)
	}
}

// Error implements gorm.io/gorm/logger.Interface
func (l *GormZapLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= gormLogger.Error {
		l.logger.Error(msg, data...)
	}
}

// Trace implements gorm.io/gorm/logger.Interface
// This is where SQL queries are logged
func (l *GormZapLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.logLevel <= gormLogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	fields := []interface{}{
		"elapsed", elapsed,
		"rows", rows,
		"sql", sql,
	}

	switch {
	case err != nil && l.logLevel >= gormLogger.Error && (!errors.Is(err, gormLogger.ErrRecordNotFound) || !l.ignoreRecordNotFoundError):
		l.logger.Error("database query failed", append(fields, "error", err)...)
	case elapsed > l.slowThreshold && l.slowThreshold != 0 && l.logLevel >= gormLogger.Warn:
		l.logger.Warn("slow query detected", append(fields, "threshold", l.slowThreshold)...)
	case l.logLevel == gormLogger.Info:
		l.logger.Debug("database query executed", fields...)
	}
}

func StringToGormLogLevel(level string) gormLogger.LogLevel {
	switch level {
	case "silent":
		return gormLogger.Silent
	case "error":
		return gormLogger.Error
	case "warn":
		return gormLogger.Warn
	case "info":
		return gormLogger.Info
	default:
		return gormLogger.Warn
	}
}
