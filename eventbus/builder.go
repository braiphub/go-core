package eventbus

import "github.com/asaskevich/EventBus"

func buildDefaultInstance() *Bus {
	return New(Config{
		PanicOnSubscribeFailed: true,
	})
}

func New(cfg Config, opts ...OptionFn) *Bus {
	bus := &Bus{
		Bus:          EventBus.New(),
		config:       cfg,
		logger:       nil,
		errorHandler: nil,
	}

	for _, opt := range opts {
		opt(bus)
	}

	return bus
}

func SetConfig(cfg Config) {
	defaultInstance.config = cfg
}

func SetOptions(opts ...OptionFn) {
	for _, opt := range opts {
		opt(defaultInstance)
	}
}
