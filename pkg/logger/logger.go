package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new zap.Logger. If verbose is true, the log level is DEBUG.
// Otherwise, it defaults to INFO.
func New(verbose bool) (*zap.Logger, error) {
	config := zap.NewProductionConfig()

	// Ensure structured JSON output
	config.Encoding = "json"

	// Include system date/time and level
	config.EncoderConfig.TimeKey = "time"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	if verbose {
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else {
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	return config.Build()
}