package ddotl

import (
	"context"
	"errors"

	"github.com/braiphub/go-core/log"
	"github.com/braiphub/go-core/tracer"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	ddotel "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/opentelemetry"
)

//go:generate mockgen -destination=mocks/ddotl.go -package=mocks . TraceProviderI

type TraceProviderI interface {
	Shutdown() error
}

type DataDogOTL struct {
	echoContextKey string
	logger         log.LoggerI
	tracer         trace.Tracer
	tracerProvider TraceProviderI
}

var (
	ErrMissingContext = errors.New("main context is missing")
	ErrMissingLogger  = errors.New("logger dependency is missing")
)

func New(ctx context.Context, serviceName, version, echoContextKey string, logger log.LoggerI) (*DataDogOTL, error) {
	if ctx == nil {
		return nil, ErrMissingContext
	}
	if logger == nil {
		return nil, ErrMissingLogger
	}
	tracer := otel.GetTracerProvider().Tracer(
		serviceName,
		trace.WithInstrumentationVersion(version),
		trace.WithSchemaURL(semconv.SchemaURL),
	)

	tracerProvider := ddotel.NewTracerProvider()
	otel.SetTracerProvider(tracerProvider)

	return &DataDogOTL{
		logger:         logger,
		tracer:         tracer,
		tracerProvider: tracerProvider,
		echoContextKey: echoContextKey,
	}, nil
}

func (otl *DataDogOTL) Close(_ context.Context) {
	if err := otl.tracerProvider.Shutdown(); err != nil {
		otl.logger.Error("closing trace provider", err)
	}
}

func (otl *DataDogOTL) NewSpan(ctx context.Context, name string) (context.Context, tracer.Span) {
	ctx, otlSpan := otl.tracer.Start(ctx, name)
	span := &Span{
		otlSpan: otlSpan,
	}

	return ctx, span
}
