package sentrymiddleware

import (
	"github.com/getsentry/sentry-go"
	"time"
)

// AppPanicHandler - defer sentrymiddleware.AppPanicHandler()
func AppPanicHandler() {
	err := recover()
	if err == nil {
		return
	}

	sentry.CurrentHub().Recover(err)
	sentry.Flush(time.Second * 5) //nolint:mnd

	panic(err)
}
