package log

import (
	"context"
	"os"
	reflect "reflect"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLoggerAdapter struct {
	zap *zap.Logger
}

func NewZap(env string, callerSkip int) (*ZapLoggerAdapter, error) {
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.TimeKey = "timestamp"
	encCfg.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	encCfg.CallerKey = "caller"
	encCfg.EncodeCaller = zapcore.ShortCallerEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encCfg),
		zapWriteSyncer(env),
		zapLevel(env),
	)

	return &ZapLoggerAdapter{
		zap: zap.New(core, zap.AddCallerSkip(callerSkip), zap.AddCaller()),
	}, nil
}

func zapWriteSyncer(env string) zapcore.WriteSyncer {
	if env == "local" {
		return zapcore.AddSync(&debugColorSyncer{})
	}

	return zapcore.AddSync(os.Stdout)
}

func zapLevel(env string) zapcore.LevelEnabler {
	if env == "local" {
		return zap.DebugLevel
	}

	return zap.InfoLevel
}

func (l *ZapLoggerAdapter) Trace(msg string, fields ...any) {
	l.zap.Debug(msg, l.zapFields(fields)...)
}

func (l *ZapLoggerAdapter) Debug(msg string, fields ...any) {
	l.zap.Debug(msg, l.zapFields(fields)...)
}

func (l *ZapLoggerAdapter) Info(msg string, fields ...any) {
	l.zap.Info(msg, l.zapFields(fields)...)
}

func (l *ZapLoggerAdapter) Warn(msg string, fields ...any) {
	l.zap.Warn(msg, l.zapFields(fields)...)
}

func (l *ZapLoggerAdapter) Error(msg string, err error, fields ...any) {
	if err != nil {
		fields = append(fields, Error(err))
	}

	l.zap.Error(msg, l.zapFields(fields)...)
}

func (l *ZapLoggerAdapter) Fatal(msg string, fields ...any) {
	l.zap.Fatal(msg, l.zapFields(fields)...)
}

func (l *ZapLoggerAdapter) Write(p []byte) (n int, err error) {
	l.Info(string(p))

	return len(p), nil
}

func (l *ZapLoggerAdapter) WithContext(ctx context.Context) LoggerI {

	// append traceable fields
	//if v := ctx.Value(logger.label.RequestID); v != nil {
	//	zapFields = append(zapFields, zap.Any(logger.label.RequestID, v))
	//}
	//if v := ctx.Value(logger.label.MessageID); v != nil {
	//	zapFields = append(zapFields, zap.Any(logger.label.MessageID, v))
	//}
	//if v := ctx.Value(logger.label.TraceID); v != nil {
	//	zapFields = append(zapFields, zap.Any(logger.label.LoggerTraceID, v))
	//}
	//if v := ctx.Value(logger.label.SpanID); v != nil {
	//	zapFields = append(zapFields, zap.Any(logger.label.LoggerSpanID, v))
	//}

	// trace-id
	//if logger.tracer != nil {
	//	if span := logger.tracer.SpanFromContext(ctx); span != nil {
	//		if span.TraceID() != "" {
	//			zapFields = append(zapFields, zap.Any("trace_id", span.TraceID()))
	//			zapFields = append(zapFields, zap.Any("dd.trace_id", span.TraceID()))
	//		}
	//		if span.SpanID() != "" {
	//			zapFields = append(zapFields, zap.Any("span_id", span.SpanID()))
	//			zapFields = append(zapFields, zap.Any("dd.span_id", span.SpanID()))
	//		}
	//
	//		// set span error
	//		for _, v := range fields {
	//			if v.Key == "error" {
	//				span.SetStatus(tracer.Error, v.Data.(string))
	//			}
	//		}
	//	}
	//}

	return l
}

func (l *ZapLoggerAdapter) WithFields(fields ...any) LoggerI {
	return &ZapLoggerAdapter{
		zap: l.zap.With(l.zapFields(fields)...),
	}
}

func (l *ZapLoggerAdapter) zapFields(fields []any) []zapcore.Field {
	zapFields := make([]zapcore.Field, len(fields))

	for i, v := range fields {
		switch v := v.(type) {
		case Field:
			zapFields[i] = zap.Any(v.Key, v.Data)

		default:
			typeName := reflect.TypeOf(v).String()
			zapFields[i] = zap.Any(typeName, v)
		}
	}

	return zapFields
}
