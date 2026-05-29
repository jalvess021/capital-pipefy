package logger

import "go.uber.org/zap"

func MessagingInfo(log *zap.Logger, message string, fields ...zap.Field) {
    log.Info(message, append(fields, zap.String("type", "messaging"))...)
}

func MessagingWarn(log *zap.Logger, message string, fields ...zap.Field) {
    log.Warn(message, append(fields, zap.String("type", "messaging"))...)
}

func MessagingError(log *zap.Logger, message string, err error, fields ...zap.Field) {
    log.Error(message, append(fields, zap.String("type", "messaging"), zap.Error(err))...)
}