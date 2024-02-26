package trace

import "context"

type TracerInterface interface {
	StartSpan(ctx context.Context, name string, attrs ...Attribute) (context.Context, SpanInterface)
	StartSpanWithKind(ctx context.Context, kind SpanKind, name string, attrs ...Attribute) (context.Context, SpanInterface)
}

type SpanInterface interface {
	Status(s SpanStatus, msg string)
	Close()
}

// span-kind
type SpanKind struct{ k string }

var (
	KindUnset    = SpanKind{"unset"}
	KindInternal = SpanKind{"internal"}
	KindServer   = SpanKind{"server"}
	KindClient   = SpanKind{"client"}
	KindProducer = SpanKind{"producer"}
	KindConsumer = SpanKind{"consumer"}
)

// //////////////////
// span-status
type SpanStatus struct{ s string }

var (
	StatusUnset = SpanStatus{"unset"}
	StatusOK    = SpanStatus{"ok"}
	StatusError = SpanStatus{"error"}
)

// //////////////////
// Attribute
type Attribute struct {
	Key   string
	Value any
}

func Attr(key string, val any) Attribute {
	return Attribute{
		Key:   key,
		Value: val,
	}
}
