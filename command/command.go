package command

import "context"

// Command is the core interface that all commands must implement.
type Command interface {
	Name() string

	Description() string

	Handle(ctx context.Context, args *Args) error
}

type CommandWithOptions interface {
	Command
	DefineOptions() []Option
}

type CommandWithValidation interface {
	Command
	Validate(ctx context.Context, args *Args) error
}

type CommandWithSetup interface {
	Command
	Setup(ctx context.Context) error
}

type CommandWithTeardown interface {
	Command
	Teardown(ctx context.Context) error
}
