package logger

import "go.uber.org/zap"

func WebhookInfo(log *zap.Logger, message string, fields ...zap.Field) {
    log.Info(message, append(fields, zap.String("type", "webhook"))...)
}

func WebhookWarn(log *zap.Logger, message string, fields ...zap.Field) {
    log.Warn(message, append(fields, zap.String("type", "webhook"))...)
}

func WebhookError(log *zap.Logger, message string, err error, fields ...zap.Field) {
    log.Error(message, append(fields, zap.String("type", "webhook"), zap.Error(err))...)
}