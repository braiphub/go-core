package log

import "context"

//go:generate mockgen -destination=log_mock.go -package=log . LoggerI

type LoggerI interface {
	Trace(ctx context.Context, msg string, fields ...Field)
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, err error, fields ...Field)
	Fatal(ctx context.Context, msg string, fields ...Field)
	Write(p []byte) (n int, err error)
	With(ctx context.Context, fields ...Field) LoggerI
}

type Field struct {
	Key  string
	Data interface{}
}

func Any(key string, data interface{}) Field {
	return Field{
		Key:  key,
		Data: data,
	}
}

func Error(err error) Field {
	return Field{
		Key:  "error",
		Data: err.Error(),
	}
}

func ErrorWTrace(err error) Field {
	return Field{
		Key:  "error",
		Data: err,
	}
}
