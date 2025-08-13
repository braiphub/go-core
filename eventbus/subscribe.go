package eventbus

import (
	"fmt"
	"reflect"

	"github.com/braiphub/go-core/log"
	"github.com/pkg/errors"
)

func SubscribeAsync(topic string, eventHandler any, transactional bool) {
	defaultInstance.SubscribeAsync(topic, eventHandler, transactional)
}

func Subscribe(topic string, eventHandler any) {
	defaultInstance.Subscribe(topic, eventHandler)
}

func (b *Bus) SubscribeAsync(topic string, eventHandler any, transactional bool) {
	topicLogField := log.Any("topic", topic)

	handlerKind := reflect.TypeOf(eventHandler).Kind()

	if handlerKind != reflect.Func {
		if b.config.PanicOnSubscribeFailed {
			panic(fmt.Errorf("%s is not of type reflect.Func", handlerKind))
		}

		if b.logger != nil {
			b.logger.Error("topic async-subscribe", fmt.Errorf("%s is not of type reflect.Func", handlerKind))
		}

		return
	}

	err := b.Bus.SubscribeAsync(topic, b.decorate(eventHandler), transactional)
	if err == nil {
		return
	}

	if b.config.PanicOnSubscribeFailed {
		panic(err)
	}

	if b.logger != nil {
		b.logger.Error("topic async-subscribe", err, topicLogField)
	}
}

func (b *Bus) Subscribe(topic string, eventHandler any) {
	topicLogField := log.Any("topic", topic)

	handlerKind := reflect.TypeOf(eventHandler).Kind()

	if handlerKind != reflect.Func {
		if b.config.PanicOnSubscribeFailed {
			panic(fmt.Errorf("%s is not of type reflect.Func", handlerKind))
		}

		if b.logger != nil {
			b.logger.Error("topic async-subscribe", fmt.Errorf("%s is not of type reflect.Func", handlerKind))
		}

		return
	}

	err := b.Bus.Subscribe(topic, b.decorate(eventHandler))
	if err == nil {
		return
	}

	if b.config.PanicOnSubscribeFailed {
		panic(err)
	}

	if b.logger != nil {
		b.logger.Error("topic async-subscribe", err, topicLogField)
	}
}

func (b *Bus) decorate(f any) func(args ...any) {
	fnValue := reflect.ValueOf(f)
	fnType := fnValue.Type()

	return func(args ...any) {
		vArgs := make([]reflect.Value, len(args))

		for i := 0; i < len(args); i++ { //nolint:intrange
			if args[i] != nil {
				vArgs[i] = reflect.ValueOf(args[i])
			} else {
				// Use the zero value for the arg.
				vArgs[i] = reflect.Zero(fnType.In(i))
			}
		}

		rets := fnValue.Call(vArgs)

		if b.logger == nil {
			return
		}

		for _, ret := range rets {
			if ret.IsNil() {
				continue
			}

			err, ok := ret.Interface().(error)
			if !ok || err == nil {
				continue
			}

			if b.logger != nil {
				b.logger.Error("event handler returned an error", err)
			}

			if b.errorHandler != nil {
				b.errorHandler(errors.Wrap(err, "event handler returned an error"))
			}
		}
	}
}
