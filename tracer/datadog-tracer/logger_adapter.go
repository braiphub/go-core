package datadogtracer

import (
	"strings"

	"github.com/braiphub/go-core/log"
)

type ddLoggerAdapter struct {
	logger log.Logger
}

func (adapter *ddLoggerAdapter) Log(msg string) {
	switch {
	case strings.Contains(msg, "INFO"):
		adapter.logger.Info("datadog_info", log.Any("msg", msg))

	case strings.Contains(msg, "WARN"):
		adapter.logger.Warn("datadog_warn", log.Any("msg", msg))

	case strings.Contains(msg, "ERROR"):
		adapter.logger.Error("datadog_error", ErrEvent, log.Any("msg", msg))
	}
}
