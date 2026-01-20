package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

type Config struct {
	Level       string   // "debug", "info", "warn", "error"
	Environment string   // "development", "production"
	OutputPaths []string // ["stdout", "logs/app.log"]
}

func Init(cfg Config) error {
	var config zap.Config

	if cfg.Environment == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return err
	}
	config.Level = zap.NewAtomicLevelAt(level)

	if len(cfg.OutputPaths) > 0 {
		config.OutputPaths = cfg.OutputPaths
	}

	l, err := config.Build(
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return err
	}

	log = l
	return nil
}

func Get() *zap.Logger {
	if log == nil {
		// Fallback a development logger
		log, _ = zap.NewDevelopment()
	}
	return log
}

func Debug(msg string, fields ...zap.Field) {
	Get().Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	Get().Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Get().Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Get().Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Get().Fatal(msg, fields...)
}

func Sync() {
	if log != nil {
		_ = log.Sync()
	}
}
