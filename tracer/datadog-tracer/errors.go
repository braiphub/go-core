package datadogtracer

import "errors"

var (
	ErrEvent                 = errors.New("datadog_error")
	ErrMissingContext        = errors.New("context is missing")
	ErrMissingEnvironment    = errors.New("env param is empty")
	ErrMissingServiceName    = errors.New("service name param is empty")
	ErrMissingServiceVersion = errors.New("service version param is empty")
	ErrMissingLogger         = errors.New("missing logger component")
)
