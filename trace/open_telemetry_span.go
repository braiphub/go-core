package trace

import (
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// span.
var _ SpanInterface = &openTelemetrySpan{}

type openTelemetrySpan struct {
	span trace.Span
}

func newOpenTelemetrySpan(span trace.Span) *openTelemetrySpan {
	return &openTelemetrySpan{span: span}
}

func (s *openTelemetrySpan) Close() {
	s.span.End()
}

func (s *openTelemetrySpan) Status(status SpanStatus, msg string) {
	switch status {
	case StatusOK:
		s.span.SetStatus(codes.Ok, msg)
	case StatusError:
		s.span.SetStatus(codes.Error, msg)
	case StatusUnset:
		s.span.SetStatus(codes.Unset, msg)
	}
}
