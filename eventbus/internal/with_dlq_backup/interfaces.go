package with_dlq_backup

import "context"

//go:generate mockgen -destination=interfaces_mock.go -package=eventbus . EventBusInterface,PubSubInterface

type EventRegister[T any] struct{}

type EventBusInterface interface {
	// Register - Handlers func should have signature: (context.Context, EventInterface) error
	Register(event EventInterface, handler interface{}) error
	RegisterList(eventList map[EventInterface][]interface{}) error
	StartListen(ctx context.Context) error
	Publish(events ...EventInterface) error
}

type EventInterface interface {
	EventType() string
}

type HandlerFunc func(ctx context.Context, event EventInterface) error

type SubscriberCallbackFunc func(ctx context.Context, eventName string, data []byte) error

type PubSubInterface interface {
	Configure(config Config) error
	ListenToEvents(ctx context.Context, callback SubscriberCallbackFunc) error
	Publish(eventName string, data []byte) error
}
