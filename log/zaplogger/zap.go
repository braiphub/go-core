package zaplogger

import (
	"github.com/braiphub/go-core/log"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	zap ZapLoggerI
}

type LoggerEnv string

const (
	LoggerEnvProd LoggerEnv = "production"
	LoggerEnvDev  LoggerEnv = "development"
)

var ErrInvalidEnv = errors.New("invalid logger environment, you must specify one")

func New(env LoggerEnv, skipCallers int) (*ZapLogger, error) {
	var logger *zap.Logger
	var err error

	skip := 1 + skipCallers

	switch env {
	case LoggerEnvDev:
		logger, err = zap.NewDevelopment(zap.AddCallerSkip(skip), zap.AddStacktrace(zap.DPanicLevel))

	case LoggerEnvProd:
		logger, err = zap.NewProduction(zap.AddCallerSkip(skip), zap.AddStacktrace(zap.DPanicLevel))

	default:
		err = ErrInvalidEnv
	}
	if err != nil {
		return nil, errors.Wrap(err, "init zap-logger instance")
	}

	return &ZapLogger{
		zap: logger,
	}, nil
}

func (logger *ZapLogger) Trace(msg string, fields ...log.Field) {
	logger.zap.Debug(msg, zapFields(fields...)...)
}

func (logger *ZapLogger) Debug(msg string, fields ...log.Field) {
	logger.zap.Debug(msg, zapFields(fields...)...)
}

func (logger *ZapLogger) Info(msg string, fields ...log.Field) {
	logger.zap.Info(msg, zapFields(fields...)...)
}

func (logger *ZapLogger) Warn(msg string, fields ...log.Field) {
	logger.zap.Warn(msg, zapFields(fields...)...)
}

func (logger *ZapLogger) Error(msg string, err error, fields ...log.Field) {
	fields = append(fields, log.Error(err))
	logger.zap.Error(msg, zapFields(fields...)...)
}

func (logger *ZapLogger) Fatal(msg string, fields ...log.Field) {
	logger.zap.Fatal(msg, zapFields(fields...)...)
}

func (logger *ZapLogger) Write(p []byte) (int, error) {
	logger.Info(string(p))

	return len(p), nil
}

func zapFields(fields ...log.Field) []zapcore.Field {
	zapFields := make([]zapcore.Field, len(fields))

	for i, v := range fields {
		zapFields[i] = zap.Any(v.Key, v.Data)
	}

	return zapFields
}
