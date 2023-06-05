package tracer

import "context"

//go:generate mockgen -destination=mocks/tracer.go -package=mocks . Tracer,Span

type Tracer interface {
	NewSpan(ctx context.Context, name string) (context.Context, Span)
	Close(ctx context.Context)
}

type Span interface {
	AddEvent(name string)
	Close()
}
