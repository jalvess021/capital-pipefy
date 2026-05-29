package logger

import "go.uber.org/zap"

func InfraInfo(log *zap.Logger, message string, fields ...zap.Field) {
	log.Info(message, append(fields, zap.String("type", "infrastructure"))...)
}

func InfraWarn(log *zap.Logger, message string, fields ...zap.Field) {
	log.Warn(message, append(fields, zap.String("type", "infrastructure"))...)
}

func InfraError(log *zap.Logger, message string, err error, fields ...zap.Field) {
	log.Error(message, append(fields, zap.String("type", "infrastructure"), zap.Error(err))...)
}
