package log

//go:generate mockgen -destination=mocks/log_mock.go -package=mocks . LoggerI

type LoggerI interface {
	Trace(msg string, fields ...Field)
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, err error, fields ...Field)
	Fatal(msg string, fields ...Field)
	Write(p []byte) (n int, err error)
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
