package trace

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"google.golang.org/grpc"
)

func WithGrpcOltpProvider(ctx context.Context, url string) func(*OpenTelemetry) {
	const connectTimeoutSecs = 10

	return func(ot *OpenTelemetry) {
		r, err := resource.Merge(
			resource.Default(),
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceName(ot.serviceName),
			),
		)
		if err != nil {
			panic(err)
		}

		traceExporter, err := otlptracegrpc.New(
			ctx,
			otlptracegrpc.WithEndpointURL(url),
			otlptracegrpc.WithDialOption(
				grpc.WithBlock(),
				grpc.WithTimeout(connectTimeoutSecs*time.Second),
			),
			// otlptracegrpc.WithInsecure(),
		)
		if err != nil {
			panic(err)
		}

		ot.traceProvider = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(traceExporter),
			sdktrace.WithResource(r),
			// sdktrace.WithSampler(sdktrace.AlwaysSample()), // don`t use in production
		)
	}
}

func WithHttpOltpProvider(ctx context.Context, url string) func(*OpenTelemetry) {
	return func(ot *OpenTelemetry) {
		r, err := resource.Merge(
			resource.Default(),
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceName(ot.serviceName),
			),
		)
		if err != nil {
			panic(err)
		}

		traceExporter, err := otlptracehttp.New(
			ctx,
			otlptracehttp.WithEndpointURL(url),
			// otlptracehttp.WithInsecure(),
		)
		if err != nil {
			panic(err)
		}

		ot.traceProvider = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(traceExporter),
			sdktrace.WithResource(r),
			// sdktrace.WithSampler(sdktrace.AlwaysSample()), // don`t use in production
		)
	}
}
