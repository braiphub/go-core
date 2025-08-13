package eventbus

import "github.com/braiphub/go-core/log"

func WithLogger(logger log.LoggerI) OptionFn {
	return func(b *Bus) {
		b.logger = logger.WithFields(log.Any("module", "eventbus"))
	}
}

func WithErrorHandler(handler ErrorHandler) OptionFn {
	return func(b *Bus) {
		b.errorHandler = handler
	}
}
