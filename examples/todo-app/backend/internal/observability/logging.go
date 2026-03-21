package observability

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func InitLogger(debug bool) error {
	config := zap.NewProductionConfig()

	if debug {
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	} else {
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return err
	}

	Logger = logger
	return nil
}

func Sync() error {
	if Logger != nil {
		return Logger.Sync()
	}
	return nil
}

// Helper functions for common logging patterns
func Info(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Info(msg, fields...)
	}
}

func Error(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Error(msg, fields...)
	}
}

func Debug(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Debug(msg, fields...)
	}
}

func Warn(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Warn(msg, fields...)
	}
}

func Fatal(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Fatal(msg, fields...)
	} else {
		// Fallback if logger not initialized
		os.Exit(1)
	}
}
