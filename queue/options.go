package queue

import (
	"github.com/braiphub/go-core/log"
	"github.com/braiphub/go-core/trace"
)

func WithLogger(logger log.LoggerI) func(*RabbitMQConnection) {
	return func(rm *RabbitMQConnection) {
		rm.logger = logger
	}
}

func WithTracer(tracer trace.TracerInterface) func(*RabbitMQConnection) {
	return func(rm *RabbitMQConnection) {
		rm.tracer = tracer
	}
}

func WithErrorHandler(fn ErrorHandlerFunc) func(*RabbitMQConnection) {
	return func(rm *RabbitMQConnection) {
		rm.errorHandler = fn
	}
}

func WithPanicHandler(fn PanicHandlerFunc) func(*RabbitMQConnection) {
	return func(rm *RabbitMQConnection) {
		rm.panicHandler = fn
	}
}
