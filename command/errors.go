package command

import "errors"

var (
	// ErrCommandNotFound is returned when a command is not registered
	ErrCommandNotFound = errors.New("command not found")

	// ErrCommandAlreadyExists is returned when trying to register a duplicate command
	ErrCommandAlreadyExists = errors.New("command already exists")

	// ErrInvalidCommandName is returned when command name is empty or invalid
	ErrInvalidCommandName = errors.New("invalid command name")

	// ErrValidationFailed is returned when command validation fails
	ErrValidationFailed = errors.New("command validation failed")

	// ErrRequiredOptionMissing is returned when a required option is not provided
	ErrRequiredOptionMissing = errors.New("required option missing")

	// ErrScheduleNotConfigured is returned when schedule has no timing configuration
	ErrScheduleNotConfigured = errors.New("schedule timing not configured")
)
