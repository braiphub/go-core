package datadogtracer

import (
	"context"
	"strings"

	"github.com/braiphub/go-core/log"
	"github.com/braiphub/go-core/tracer"
	"github.com/labstack/echo/v4"
	ddtraceecho "gopkg.in/DataDog/dd-trace-go.v1/contrib/labstack/echo.v4"
	ddTracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type DatadogTracer struct {
	env            string
	serviceName    string
	version        string
	logger         log.LoggerI
	EchoMiddleware echo.MiddlewareFunc
}

type TracerI interface {
	StartSpanFromContext(
		ctx context.Context,
		operationName string,
		opts ...ddTracer.StartSpanOption,
	) (ddTracer.Span, context.Context)
}

func New(ctx context.Context, env, serviceName, version string, logger log.LoggerI) (*DatadogTracer, error) {
	switch {
	case ctx == nil:
		return nil, ErrMissingContext

	case strings.TrimSpace(env) == "":
		return nil, ErrMissingEnvironment

	case strings.TrimSpace(serviceName) == "":
		return nil, ErrMissingServiceName

	case strings.TrimSpace(version) == "":
		return nil, ErrMissingServiceVersion

	case logger == nil:
		return nil, ErrMissingLogger
	}

	loggerAdapter := &ddLoggerAdapter{logger}
	ddTracer.Start(
		ddTracer.WithEnv(env),
		ddTracer.WithService(serviceName),
		ddTracer.WithServiceVersion(version),
		ddTracer.WithLogger(loggerAdapter),
		ddTracer.WithLogStartup(false),
	)

	return &DatadogTracer{
		env:            env,
		serviceName:    serviceName,
		version:        version,
		logger:         logger,
		EchoMiddleware: ddtraceecho.Middleware(ddtraceecho.NoDebugStack()),
	}, nil
}

func (DatadogTracer) NewSpan(ctx context.Context, name string) (context.Context, tracer.Span) {
	ddSpan, ctx := ddTracer.StartSpanFromContext(ctx, name)

	return ctx, &ddSpanAdapter{
		ddSpan: ddSpan,
	}
}

func (DatadogTracer) Close(_ context.Context) {
	ddTracer.Stop()
}
