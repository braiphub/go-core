package eventbus

import (
	"github.com/asaskevich/EventBus"
	"github.com/braiphub/go-core/log"
)

var defaultInstance = buildDefaultInstance() //nolint:gochecknoglobals

type Bus struct {
	EventBus.Bus
	config       Config
	logger       log.LoggerI
	errorHandler ErrorHandler
}

type Config struct {
	PanicOnSubscribeFailed bool
}

type OptionFn func(*Bus)

type ErrorHandler func(err error)

func DefaultInstance() *Bus {
	return defaultInstance
}
