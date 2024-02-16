package log

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLoggerAdapter struct {
	zap *zap.Logger
}

func NewZap(env string, callerSkip int) (*ZapLoggerAdapter, error) {
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.TimeKey = "timestamp"
	encCfg.EncodeTime = zapcore.EpochMillisTimeEncoder
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

func (logger *ZapLoggerAdapter) Trace(ctx context.Context, msg string, fields ...Field) {
	logger.zap.Debug(msg, logger.zapFields(ctx, fields)...)
}

func (logger *ZapLoggerAdapter) Debug(ctx context.Context, msg string, fields ...Field) {
	logger.zap.Debug(msg, logger.zapFields(ctx, fields)...)
}

func (logger *ZapLoggerAdapter) Info(ctx context.Context, msg string, fields ...Field) {
	logger.zap.Info(msg, logger.zapFields(ctx, fields)...)
}

func (logger *ZapLoggerAdapter) Warn(ctx context.Context, msg string, fields ...Field) {
	logger.zap.Warn(msg, logger.zapFields(ctx, fields)...)
}

func (logger *ZapLoggerAdapter) Error(ctx context.Context, msg string, err error, fields ...Field) {
	if err != nil {
		fields = append(fields, Error(err))
	}

	logger.zap.Error(msg, logger.zapFields(ctx, fields)...)
}

func (logger *ZapLoggerAdapter) Fatal(ctx context.Context, msg string, fields ...Field) {
	logger.zap.Fatal(msg, logger.zapFields(ctx, fields)...)
}

func (logger *ZapLoggerAdapter) Write(p []byte) (n int, err error) {
	logger.Info(context.Background(), string(p))

	return len(p), nil
}

func (logger *ZapLoggerAdapter) With(ctx context.Context, fields ...Field) LoggerI {
	return &ZapLoggerAdapter{
		zap: logger.zap.With(logger.zapFields(ctx, fields)...),
	}
}

func (logger *ZapLoggerAdapter) zapFields(_ context.Context, fields []Field) []zapcore.Field {
	zapFields := make([]zapcore.Field, len(fields))

	for i, v := range fields {
		zapFields[i] = zap.Any(v.Key, v.Data)
	}

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

	return zapFields
}
