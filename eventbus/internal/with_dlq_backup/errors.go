package with_dlq_backup

import "errors"

var (
	ErrEventModelShouldBePointer = errors.New("input event-model should be a pointer to EventInterface struct")
	ErrEventHandlersNotSpecified = errors.New("you must specify at least one event-handler")
	ErrUnregisteredEvent         = errors.New("unregistered event")
	ErrInvalidEventHandler       = errors.New("invalid event handler")
)
