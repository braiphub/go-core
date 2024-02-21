package eventbus

import "errors"

type Config struct {
	ServiceName           string
	DaemonName            string
	ShouldDLQUnregistered bool
}

type EventRegisterConfig struct {
	eventModel EventInterface
	handler    interface{}
}

func (e *EventBus) validateConfig() error {
	switch {
	case e.Config.ServiceName == "":
		return errors.New("invalid config: service name is empty")

	case e.Config.DaemonName == "":
		return errors.New("invalid config: daemon name is empty")

	case e.pubSub == nil:
		return errors.New("invalid config: pub/sub isn't setted up")

	case e.logger == nil:
		return errors.New("invalid config: logger isn't setted up")
	}

	return nil
}
