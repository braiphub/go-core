package sentrymiddleware

import "github.com/getsentry/sentry-go"

func EventBusErrorHandler(err error) {
	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetTag("module", "eventbus")
		sentry.CaptureException(err)
	})
}
