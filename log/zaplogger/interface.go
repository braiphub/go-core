package zaplogger

import "go.uber.org/zap"

//go:generate mockgen -destination=mocks/zaplogger_mock.go -package=mocks . ZapLoggerI

type ZapLoggerI interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
}
