package trace

import "context"

type TracerInterface interface {
	StartSpan(ctx context.Context, name string, attrs ...attribute) (context.Context, SpanInterface)
	StartSpanWithKind(ctx context.Context, kind spanKind, name string, attrs ...attribute) (context.Context, SpanInterface)
}

type SpanInterface interface {
	Status(s spanStatus, msg string)
	Close()
}

// span-kind
type spanKind struct{ k string }

var (
	KindUnset    = spanKind{"unset"}
	KindInternal = spanKind{"internal"}
	KindServer   = spanKind{"server"}
	KindClient   = spanKind{"client"}
	KindProducer = spanKind{"producer"}
	KindConsumer = spanKind{"consumer"}
)

// //////////////////
// span-status
type spanStatus struct{ s string }

var (
	StatusUnset = spanStatus{"unset"}
	StatusOK    = spanStatus{"ok"}
	StatusError = spanStatus{"error"}
)

// //////////////////
// attribute
type attribute struct {
	Key   string
	Value any
}

func Attr(key string, val any) attribute {
	return attribute{
		Key:   key,
		Value: val,
	}
}
