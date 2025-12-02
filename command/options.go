package command

import "github.com/braiphub/go-core/log"

// KernelOption is a functional option for configuring the Kernel
type KernelOption func(*Kernel)

// WithLogger sets the logger for the kernel
func WithLogger(logger log.LoggerI) KernelOption {
	return func(k *Kernel) {
		k.logger = logger
	}
}

// WithCommands registers commands during kernel initialization
func WithCommands(cmds ...Command) KernelOption {
	return func(k *Kernel) {
		_ = k.RegisterMany(cmds...)
	}
}

// WithSchedules adds schedules during kernel initialization
func WithSchedules(schedules ...*Schedule) KernelOption {
	return func(k *Kernel) {
		k.Schedule(schedules...)
	}
}

// WithRegistry allows injecting a custom registry
func WithRegistry(registry *Registry) KernelOption {
	return func(k *Kernel) {
		k.registry = registry
	}
}
