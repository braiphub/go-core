package queue

import (
	"github.com/braiphub/go-core/log"
)

func WithLogger(logger log.LoggerI) func(*RabbitMQConnection) {
	return func(rm *RabbitMQConnection) {
		rm.logger = logger
	}
}

func WithErrorHandler(fn ErrorHandlerFunc) func(*RabbitMQConnection) {
	return func(rm *RabbitMQConnection) {
		rm.errorHandler = fn
	}
}

func WithDeferPanicHandler(fn DeferPanicHandlerFunc) func(*RabbitMQConnection) {
	return func(rm *RabbitMQConnection) {
		rm.deferPanicHandler = fn
	}
}
