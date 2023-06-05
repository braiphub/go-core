package ddotl

import (
	"errors"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	ddotel "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/opentelemetry"
)

type mockedEchoContext struct {
	echo.Context
	path string
}

func (mockedEchoContext) Request() *http.Request {
	return &http.Request{}
}
func (m *mockedEchoContext) Path() string {
	path := m.path
	m.path = ""
	return path
}
func (mockedEchoContext) Response() *echo.Response {
	return &echo.Response{
		Status: http.StatusOK,
	}
}
func (mockedEchoContext) Error(error) {}

func TestDataDogOTL_EchoMiddleware(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		otl := &DataDogOTL{
			echoContextKey: "echo-key",
			tracer:         mockedTrace{},
			tracerProvider: &ddotel.TracerProvider{},
		}

		handlerF := func(c echo.Context) error {
			return errors.New("unknown")
		}
		e := &mockedEchoContext{
			Context: echo.New().AcquireContext(),
			path:    "/path",
		}
		middlewareFunc := otl.EchoMiddleware()(handlerF)
		middlewareFunc(e)
	})
}
