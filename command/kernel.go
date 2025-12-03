package command

import (
	"context"
	"fmt"

	"github.com/braiphub/go-core/log"
	"github.com/braiphub/go-core/schedule"
)

// Kernel is the central coordinator for commands and scheduling.
// Similar to Laravel's Console Kernel.
type Kernel struct {
	registry  *Registry
	schedules []*Schedule
	logger    log.LoggerI
}

// NewKernel creates a new Kernel with the given options
func NewKernel(opts ...KernelOption) *Kernel {
	k := &Kernel{
		registry:  NewRegistry(),
		schedules: make([]*Schedule, 0),
	}

	for _, opt := range opts {
		opt(k)
	}

	return k
}

// Register adds a command to the kernel
func (k *Kernel) Register(cmd Command) error {
	return k.registry.Register(cmd)
}

// RegisterMany adds multiple commands to the kernel
func (k *Kernel) RegisterMany(cmds ...Command) error {
	return k.registry.RegisterMany(cmds...)
}

// Schedule adds a scheduled command
func (k *Kernel) Schedule(schedules ...*Schedule) *Kernel {
	k.schedules = append(k.schedules, schedules...)
	return k
}

// Commands returns all registered commands
func (k *Kernel) Commands() map[string]Command {
	return k.registry.All()
}

// CommandNames returns all command names
func (k *Kernel) CommandNames() []string {
	return k.registry.Names()
}

// GetCommand retrieves a command by name
func (k *Kernel) GetCommand(name string) (Command, error) {
	return k.registry.Get(name)
}

// HasCommand checks if a command exists
func (k *Kernel) HasCommand(name string) bool {
	return k.registry.Has(name)
}

// Run executes a command by name with the given arguments
func (k *Kernel) Run(ctx context.Context, name string, args *Args) error {
	cmd, err := k.registry.Get(name)
	if err != nil {
		return err
	}

	if args == nil {
		args = NewArgs()
	}

	return k.executeCommand(ctx, cmd, args)
}

// Call is an alias for Run
func (k *Kernel) Call(ctx context.Context, name string, args *Args) error {
	return k.Run(ctx, name, args)
}

// executeCommand handles the full command execution lifecycle
func (k *Kernel) executeCommand(ctx context.Context, cmd Command, args *Args) error {
	// Setup phase
	if setupCmd, ok := cmd.(CommandWithSetup); ok {
		if err := setupCmd.Setup(ctx); err != nil {
			return fmt.Errorf("command setup failed: %w", err)
		}
	}

	// Validation phase
	if validatableCmd, ok := cmd.(CommandWithValidation); ok {
		if err := validatableCmd.Validate(ctx, args); err != nil {
			return fmt.Errorf("command validation failed: %w", err)
		}
	}

	// Teardown deferred
	if teardownCmd, ok := cmd.(CommandWithTeardown); ok {
		defer func() {
			if err := teardownCmd.Teardown(ctx); err != nil && k.logger != nil {
				k.logger.WithContext(ctx).Error("command teardown failed", err)
			}
		}()
	}

	// Execute
	return cmd.Handle(ctx, args)
}

// StartScheduler starts the scheduler with all scheduled commands
func (k *Kernel) StartScheduler(ctx context.Context) error {
	if len(k.schedules) == 0 {
		if k.logger != nil {
			k.logger.WithContext(ctx).Warn("no scheduled commands configured")
		}
		return nil
	}

	jobs, err := k.buildScheduleJobs(ctx)
	if err != nil {
		return fmt.Errorf("failed to build schedule jobs: %w", err)
	}

	return schedule.Run(ctx, k.logger, jobs...)
}

// buildScheduleJobs converts scheduled commands to schedule.Job
func (k *Kernel) buildScheduleJobs(ctx context.Context) ([]schedule.Job, error) {
	jobs := make([]schedule.Job, 0, len(k.schedules))

	for _, sched := range k.schedules {
		cmd, err := k.registry.Get(sched.Command())
		if err != nil {
			return nil, fmt.Errorf("scheduled command '%s' not found: %w", sched.Command(), err)
		}

		if !sched.IsConfigured() {
			return nil, fmt.Errorf("schedule for command '%s' has no timing configuration", sched.Command())
		}

		// Capture for closure
		capturedCmd := cmd
		capturedSchedule := sched
		capturedCtx := ctx

		job, err := sched.ToScheduleJob(func() error {
			return k.executeCommand(capturedCtx, capturedCmd, capturedSchedule.Arguments())
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create schedule job for '%s': %w", sched.Command(), err)
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

// Registry returns the command registry
func (k *Kernel) Registry() *Registry {
	return k.registry
}

// Logger returns the kernel's logger
func (k *Kernel) Logger() log.LoggerI {
	return k.logger
}
