package ddotl

import (
	"go.opentelemetry.io/otel/trace"
)

type Span struct {
	otlSpan trace.Span
}

func (span *Span) AddEvent(name string) {
	span.otlSpan.AddEvent(name)
}

func (span *Span) Close() {
	span.otlSpan.End()
}
