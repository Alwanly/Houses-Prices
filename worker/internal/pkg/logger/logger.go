package logger

import (
    "fmt"

    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

// New creates a new logger based on level and format
func New(level, format string) (*zap.Logger, error) {
    var zapLevel zapcore.Level
    if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
        return nil, fmt.Errorf("invalid log level: %w", err)
    }

    var cfg zap.Config
    if format == "json" {
        cfg = zap.NewProductionConfig()
    } else {
        cfg = zap.NewDevelopmentConfig()
    }

    cfg.Level = zap.NewAtomicLevelAt(zapLevel)
    cfg.EncoderConfig.TimeKey = "timestamp"
    cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

    logger, err := cfg.Build()
    if err != nil {
        return nil, fmt.Errorf("building logger: %w", err)
    }

    return logger, nil
}
