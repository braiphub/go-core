package trace

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	otelAttribute "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type OpenTelemetry struct {
	serviceName   string
	tracer        trace.Tracer
	traceProvider *sdktrace.TracerProvider
	shutdownFunc  func(context.Context) error
}

var _ TracerInterface = &OpenTelemetry{}

var spanKindToOtelMap = map[spanKind]trace.SpanKind{
	KindUnset:    trace.SpanKindUnspecified,
	KindInternal: trace.SpanKindInternal,
	KindServer:   trace.SpanKindServer,
	KindClient:   trace.SpanKindClient,
	KindProducer: trace.SpanKindProducer,
	KindConsumer: trace.SpanKindConsumer,
}

func NewOpenTelemetry(
	ctx context.Context,
	serviceName string,
	opts ...func(*OpenTelemetry),
) *OpenTelemetry {
	t := &OpenTelemetry{
		serviceName: serviceName,
	}

	for _, o := range opts {
		o(t)
	}

	t.validate()

	t.startCollect(ctx)

	return t
}

func (ot *OpenTelemetry) validate() {
	switch {
	case ot.traceProvider == nil:
		panic("missing trace exporter")
	}
}

func (ot *OpenTelemetry) Close(ctx context.Context) {
	ot.shutdownFunc(ctx)
}

func (ot *OpenTelemetry) StartSpan(ctx context.Context, name string, attrs ...attribute) (context.Context, SpanInterface) {
	traceAttrs := ot.attributesToOtelAttribtes(attrs...)

	ctx, otelSpan := ot.tracer.Start(ctx, name, trace.WithAttributes(traceAttrs...))

	span := newOpenTelemetrySpan(otelSpan)

	return ctx, span
}

func (ot *OpenTelemetry) StartSpanWithKind(ctx context.Context, kind spanKind, name string, attrs ...attribute) (context.Context, SpanInterface) {
	otelKind := spanKindToOtelMap[kind]

	traceAttrs := ot.attributesToOtelAttribtes(attrs...)

	ctx, otelSpan := ot.tracer.Start(ctx, name, trace.WithAttributes(traceAttrs...), trace.WithSpanKind(otelKind))

	span := newOpenTelemetrySpan(otelSpan)

	return ctx, span
}

func (ot *OpenTelemetry) startCollect(ctx context.Context) {
	shutdownFunc, err := ot.setupOTelSDK(ctx)
	if err != nil {
		panic(err)
	}

	ot.shutdownFunc = shutdownFunc
}

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func (ot *OpenTelemetry) setupOTelSDK(
	ctx context.Context,
) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// Set up propagator.
	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	otel.SetTextMapPropagator(prop)

	shutdownFuncs = append(shutdownFuncs, ot.traceProvider.Shutdown)
	otel.SetTracerProvider(ot.traceProvider)

	ot.tracer = otel.Tracer("github.com/braiphub/go-core/tracer")

	return
}

func (ot *OpenTelemetry) attributesToOtelAttribtes(attrs ...attribute) []otelAttribute.KeyValue {
	traceAttrs := make([]otelAttribute.KeyValue, len(attrs))

	for _, attr := range attrs {
		switch v := attr.Value.(type) {
		case bool:
			traceAttrs = append(traceAttrs, otelAttribute.Bool(attr.Key, v))
		case int:
			traceAttrs = append(traceAttrs, otelAttribute.Int(attr.Key, v))
		case int64:
			traceAttrs = append(traceAttrs, otelAttribute.Int64(attr.Key, v))
		case string:
			traceAttrs = append(traceAttrs, otelAttribute.String(attr.Key, v))
		}
	}

	return traceAttrs
}
