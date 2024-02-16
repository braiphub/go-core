package log

import "context"

//go:generate mockgen -destination=log_mock.go -package=log . LoggerI

type LoggerI interface {
	Trace(msg string, fields ...any)
	Debug(msg string, fields ...any)
	Info(msg string, fields ...any)
	Warn(msg string, fields ...any)
	Error(msg string, err error, fields ...any)
	Fatal(msg string, fields ...any)
	Write(p []byte) (n int, err error)
	WithContext(ctx context.Context) LoggerI
	WithFields(fields ...any) LoggerI
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
