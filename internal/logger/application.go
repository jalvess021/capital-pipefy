package logger

import "go.uber.org/zap"

func ApplicationInfo(log *zap.Logger, message string, fields ...zap.Field) {
    log.Info(message, append(fields, zap.String("type", "application"))...)
}

func ApplicationWarn(log *zap.Logger, message string, fields ...zap.Field) {
    log.Warn(message, append(fields, zap.String("type", "application"))...)
}

func ApplicationError(log *zap.Logger, message string, err error, fields ...zap.Field) {
    log.Error(message, append(fields, zap.String("type", "application"), zap.Error(err))...)
}