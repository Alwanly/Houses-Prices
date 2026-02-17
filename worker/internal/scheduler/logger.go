package scheduler

import (
    "fmt"

    "go.uber.org/zap"
    "github.com/robfig/cron/v3"
)

// cronLogger adapts zap.Logger to cron.Logger interface
type cronLogger struct {
    logger *zap.Logger
}

func newCronLogger(logger *zap.Logger) cron.Logger {
    return &cronLogger{logger: logger}
}

func (l *cronLogger) Info(msg string, keysAndValues ...interface{}) {
    fields := make([]zap.Field, 0, len(keysAndValues)/2)
    for i := 0; i < len(keysAndValues); i += 2 {
        if i+1 < len(keysAndValues) {
            key := fmt.Sprintf("%v", keysAndValues[i])
            value := keysAndValues[i+1]
            fields = append(fields, zap.Any(key, value))
        }
    }
    l.logger.Info(msg, fields...)
}

func (l *cronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
    fields := make([]zap.Field, 0, len(keysAndValues)/2+1)
    fields = append(fields, zap.Error(err))
    
    for i := 0; i < len(keysAndValues); i += 2 {
        if i+1 < len(keysAndValues) {
            key := fmt.Sprintf("%v", keysAndValues[i])
            value := keysAndValues[i+1]
            fields = append(fields, zap.Any(key, value))
        }
    }
    l.logger.Error(msg, fields...)
}
