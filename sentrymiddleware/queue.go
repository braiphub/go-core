package sentrymiddleware

import (
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
)

func QueueErrorHandler(queue string, payload []byte, headers map[string]any, err error) {
	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetTag("module", "queue")
		scope.SetTag("queue", queue)
		scope.SetContext("message", map[string]interface{}{"queue": queue, "payload": string(payload)})
		sentry.CaptureException(err)
	})
}

func QueuePanicHandler(queue string) {
	err := recover()
	if err == nil {
		return
	}

	if _, ok := err.(string); ok {
		err = fmt.Sprintf("queue: %s. panic: %s", queue, err)
	}

	sentry.CurrentHub().Recover(err)
	sentry.Flush(time.Second * 5) //nolint:mnd
}
