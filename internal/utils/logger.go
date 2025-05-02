package utils

import (
	"os"
	"sync"

	"github.com/riptano/iac_generator_cli/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.SugaredLogger
	once   sync.Once
)

// GetLogger returns a singleton logger instance
func GetLogger() *zap.SugaredLogger {
	once.Do(func() {
		// Determine log level from config
		level := zap.InfoLevel
		switch config.AppConfig.LogLevel {
		case "debug":
			level = zap.DebugLevel
		case "info":
			level = zap.InfoLevel
		case "warn":
			level = zap.WarnLevel
		case "error":
			level = zap.ErrorLevel
		}

		// Create encoder config
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

		// Create core
		core := zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			level,
		)

		// Create logger
		zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
		logger = zapLogger.Sugar()
	})

	return logger
}

// ShutdownLogger flushes any buffered log entries
func ShutdownLogger() {
	if logger != nil {
		_ = logger.Sync()
	}
}