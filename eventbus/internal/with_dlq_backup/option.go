package with_dlq_backup

import "github.com/braiphub/go-core/log"

type Option func(*EventBus)

func WithDefaultConfig(e *EventBus) {
	e.Config.ShouldDLQUnregistered = true
}

func WithLogger(logger log.LoggerI) func(bus *EventBus) {
	return func(bus *EventBus) {
		bus.logger = logger
	}
}
