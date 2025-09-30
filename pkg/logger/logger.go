package logger

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Fatal(msg string, args ...interface{})

	With(fields ...interface{}) Logger
	Sync() error
}

type zapLogger struct {
	sugar *zap.SugaredLogger
	base  *zap.Logger
}

func New(env string) Logger {
	config := getZapConfig(env)

	base, err := config.Build(
		zap.AddCallerSkip(1), // Skip one level to show the actual caller
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		panic("Failed to initialize logging: " + err.Error())
	}

	return &zapLogger{
		sugar: base.Sugar(),
		base:  base,
	}
}

func getZapConfig(env string) zap.Config {
	switch strings.ToLower(env) {
	case "development", "dev":
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
		return config

	case "test":
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.EncodeCaller = zapcore.FullCallerEncoder
		return config

	case "production", "prod":
		config := zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.LevelKey = "level"
		config.EncoderConfig.CallerKey = "caller"
		config.EncoderConfig.MessageKey = "message"
		config.EncoderConfig.StacktraceKey = "stacktrace"

		// Add service information to all logs
		config.InitialFields = map[string]interface{}{
			"service": "api-gateway",
			"version": "1.0.0",
		}
		return config

	default:
		config := zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		return config
	}
}

func (l *zapLogger) Debug(msg string, args ...interface{}) {
	l.sugar.Debugw(msg, args...)
}

func (l *zapLogger) Info(msg string, args ...interface{}) {
	l.sugar.Infow(msg, args...)
}

func (l *zapLogger) Warn(msg string, args ...interface{}) {
	l.sugar.Warnw(msg, args...)
}

func (l *zapLogger) Error(msg string, args ...interface{}) {
	l.sugar.Errorw(msg, args...)
}

func (l *zapLogger) Fatal(msg string, args ...interface{}) {
	l.sugar.Fatalw(msg, args...)
}

func (l *zapLogger) With(fields ...interface{}) Logger {
	return &zapLogger{
		sugar: l.sugar.With(fields...),
		base:  l.base,
	}
}

func (l *zapLogger) Sync() error {
	return l.sugar.Sync()
}
