package logger

import "go.uber.org/zap"

func RequestInfo(log *zap.Logger, message string, fields ...zap.Field) {
    log.Info(message, append(fields, zap.String("type", "request"))...)
}

func RequestWarn(log *zap.Logger, message string, fields ...zap.Field) {
    log.Warn(message, append(fields, zap.String("type", "request"))...)
}

func RequestError(log *zap.Logger, message string, err error, fields ...zap.Field) {
    log.Error(message, append(fields, zap.String("type", "request"), zap.Error(err))...)
}