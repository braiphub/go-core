package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/braiphub/go-core/log"
	"github.com/mohae/deepcopy"
	"github.com/pkg/errors"
)

type EventBus struct {
	Config           Config
	pubSub           PubSubInterface
	registeredEvents map[string][]EventRegisterConfig
	logger           log.LoggerI
}

func New(serviceName, daemonName string, opts ...Option) (*EventBus, error) {
	bus := &EventBus{
		Config: Config{
			ServiceName: serviceName,
			DaemonName:  daemonName,
		},
		registeredEvents: make(map[string][]EventRegisterConfig),
	}

	WithDefaultConfig(bus)

	for _, o := range opts {
		o(bus)
	}

	if err := bus.validateConfig(); err != nil {
		return nil, err
	}

	if err := bus.pubSub.Configure(bus.Config); err != nil {
		return nil, errors.Wrap(err, "configure pub/sub")
	}

	return bus, nil
}

// Register - Handlers func should have signature: (context.Context, EventInterface) error
func (bus *EventBus) Register(event EventInterface, handler interface{}) error {
	// validate
	if reflect.ValueOf(event).Kind() != reflect.Ptr {
		return errors.Wrap(ErrEventModelShouldBePointer, event.EventType())
	}

	if reflect.TypeOf(handler).Kind() != reflect.Func {
		return errors.Wrap(ErrInvalidEventHandler, event.EventType())
	}

	// TODO: must validate if func first param is a context and second is an EventInterface and if return is an error
	// 		 check if it has only 2 inputs, output is optional
	// TODO: must validate if second param is the same type as input event type but without pointer

	// init list if first use
	if _, ok := bus.registeredEvents[event.EventType()]; !ok {
		bus.registeredEvents[event.EventType()] = make([]EventRegisterConfig, 0)
	}

	// append event handler
	eventHandlerList := bus.registeredEvents[event.EventType()]

	eventHandlerList = append(eventHandlerList, EventRegisterConfig{
		eventModel: event,
		handler:    handler,
	})

	bus.registeredEvents[event.EventType()] = eventHandlerList

	return nil
}

// RegisterList - Handlers func should have signature: (context.Context, EventInterface) error
func (bus *EventBus) RegisterList(eventList map[EventInterface][]interface{}) error {
	for event, handlers := range eventList {
		for _, handler := range handlers {
			if err := bus.Register(event, handler); err != nil {
				return err
			}
		}
	}

	return nil
}

func (bus *EventBus) StartListen(ctx context.Context) error {
	if err := bus.pubSub.ListenToEvents(ctx, bus.receivedEventFromPubSubHandler); err != nil {
		return err
	}

	return nil
}

func (bus *EventBus) Publish(events ...EventInterface) error {
	for _, event := range events {
		data, err := json.Marshal(event)
		if err != nil {
			return err
		}

		if err := bus.pubSub.Publish(event.EventType(), data); err != nil {
			return errors.Wrap(err, "pub/sub publish event")
		}

		bus.logger.Debug("event published", log.Any("event_name", event.EventType()))
	}

	return nil
}

func (bus *EventBus) receivedEventFromPubSubHandler(
	ctx context.Context,
	eventName string,
	data []byte,
) error {
	eventHandlers, ok := bus.registeredEvents[eventName]
	if !ok && bus.Config.ShouldDLQUnregistered {
		bus.logger.Error("nacking unknown/unregistered event", ErrUnregisteredEvent, log.Any("event_name", eventName))

		return ErrUnregisteredEvent
	}

	for i, h := range eventHandlers {
		event := deepcopy.Copy(h.eventModel)
		if err := json.Unmarshal(data, &event); err != nil {
			bus.logger.Error("error unmarshaling message", err, log.Any("event_name", eventName))

			return errors.Wrap(err, "unmarshal")
		}

		if err := callFuncWithArgs(h.handler, ctx, event); err != nil {
			bus.logger.Error(fmt.Sprintf("error on handler %d of %d", i+1, len(eventHandlers)), err, log.Any("event_name", eventName))

			return errors.Wrap(err, "processing event")
		}
	}

	bus.logger.Debug("event processed", log.Any("event_name", eventName))

	return nil
}

func callFuncWithArgs(callback interface{}, ctx context.Context, eventPtr interface{}) error {
	passedArguments := make([]reflect.Value, 2)
	passedArguments[0] = reflect.ValueOf(ctx)
	passedArguments[1] = reflect.ValueOf(eventPtr).Elem()

	if ctx == nil {
		passedArguments[0] = reflect.New(reflect.ValueOf(callback).Type().In(0)).Elem()
	}

	result := reflect.ValueOf(callback).Call(passedArguments)
	if len(result) > 0 {
		if err, ok := result[len(result)-1].Interface().(error); ok && err != nil {
			return err
		}
	}

	return nil
}
