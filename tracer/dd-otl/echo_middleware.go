package ddotl

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/semconv/v1.13.0/httpconv"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// EchoMiddleware returns echo middleware which will trace incoming requests.
func (otl *DataDogOTL) EchoMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(otl.echoContextKey, otl.tracer)
			request := c.Request()
			savedCtx := request.Context()
			defer func() {
				request = request.WithContext(savedCtx)
				c.SetRequest(request)
			}()
			opts := []oteltrace.SpanStartOption{
				oteltrace.WithAttributes(httpconv.ServerRequest("", request)...),
				oteltrace.WithSpanKind(oteltrace.SpanKindServer),
			}
			if path := c.Path(); path != "" {
				rAttr := semconv.HTTPRoute(path)
				opts = append(opts, oteltrace.WithAttributes(rAttr))
			}
			spanName := c.Path()
			if spanName == "" {
				spanName = fmt.Sprintf("HTTP %s route not found", request.Method)
			}

			ctx, span := otl.tracer.Start(savedCtx, spanName, opts...)
			defer span.End()

			// pass the span through the request context
			c.SetRequest(request.WithContext(ctx))

			// serve the request to the next middleware
			err := next(c)
			if err != nil {
				span.SetAttributes(attribute.String("echo.error", err.Error()))
				// invokes the registered HTTP error handler
				c.Error(err)
			}

			status := c.Response().Status
			span.SetStatus(httpconv.ServerStatus(status))
			if status > 0 {
				span.SetAttributes(
					semconv.HTTPStatusCode(status),
				)
			}

			return err
		}
	}
}
