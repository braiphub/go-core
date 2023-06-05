package datadogtracer

import "gopkg.in/DataDog/dd-trace-go.v1/ddtrace"

type ddSpanAdapter struct {
	ddSpan ddtrace.Span
}

func (adapter *ddSpanAdapter) AddEvent(_ /*name*/ string) {}

func (adapter *ddSpanAdapter) Close() {
	adapter.ddSpan.Finish()
}
