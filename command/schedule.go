package command

import (
	"context"
	"time"

	"github.com/braiphub/go-core/cron"
)

// Schedule defines when and how a command should be executed
type Schedule struct {
	commandName      string
	args             *Args
	interval         *time.Duration
	dailyAt          *cron.DailyAt
	startImmediately bool
	startDelay       *time.Duration
}

// NewSchedule creates a new schedule for a command
func NewSchedule(commandName string) *Schedule {
	return &Schedule{
		commandName: commandName,
		args:        NewArgs(),
	}
}

// Command returns the command name
func (s *Schedule) Command() string {
	return s.commandName
}

// Args returns the configured arguments
func (s *Schedule) Arguments() *Args {
	return s.args
}

// Fluent API methods

// EverySecond schedules the command to run every second
func (s *Schedule) EverySecond() *Schedule {
	interval := time.Second
	s.interval = &interval
	return s
}

// EverySeconds schedules the command to run every N seconds
func (s *Schedule) EverySeconds(n int) *Schedule {
	interval := time.Duration(n) * time.Second
	s.interval = &interval
	return s
}

// EveryMinute schedules the command to run every minute
func (s *Schedule) EveryMinute() *Schedule {
	interval := time.Minute
	s.interval = &interval
	return s
}

// EveryMinutes schedules the command to run every N minutes
func (s *Schedule) EveryMinutes(n int) *Schedule {
	interval := time.Duration(n) * time.Minute
	s.interval = &interval
	return s
}

// EveryFiveMinutes schedules the command to run every 5 minutes
func (s *Schedule) EveryFiveMinutes() *Schedule {
	return s.EveryMinutes(5)
}

// EveryTenMinutes schedules the command to run every 10 minutes
func (s *Schedule) EveryTenMinutes() *Schedule {
	return s.EveryMinutes(10)
}

// EveryFifteenMinutes schedules the command to run every 15 minutes
func (s *Schedule) EveryFifteenMinutes() *Schedule {
	return s.EveryMinutes(15)
}

// EveryThirtyMinutes schedules the command to run every 30 minutes
func (s *Schedule) EveryThirtyMinutes() *Schedule {
	return s.EveryMinutes(30)
}

// Hourly schedules the command to run every hour
func (s *Schedule) Hourly() *Schedule {
	interval := time.Hour
	s.interval = &interval
	return s
}

// EveryHours schedules the command to run every N hours
func (s *Schedule) EveryHours(n int) *Schedule {
	interval := time.Duration(n) * time.Hour
	s.interval = &interval
	return s
}

// Daily schedules the command to run daily at midnight
func (s *Schedule) Daily() *Schedule {
	s.dailyAt = &cron.DailyAt{Hour: 0, Minute: 0, Second: 0}
	return s
}

// DailyAt schedules the command to run daily at the specified time
func (s *Schedule) DailyAt(hour, minute int) *Schedule {
	s.dailyAt = &cron.DailyAt{Hour: hour, Minute: minute, Second: 0}
	return s
}

// DailyAtTime schedules the command to run daily at the specified time with seconds
func (s *Schedule) DailyAtTime(hour, minute, second int) *Schedule {
	s.dailyAt = &cron.DailyAt{Hour: hour, Minute: minute, Second: second}
	return s
}

// WithInterval sets a custom interval
func (s *Schedule) WithInterval(d time.Duration) *Schedule {
	s.interval = &d
	return s
}

// Immediately runs the command immediately on scheduler start
func (s *Schedule) Immediately() *Schedule {
	s.startImmediately = true
	return s
}

// WithDelay delays the first execution by the specified duration
func (s *Schedule) WithDelay(d time.Duration) *Schedule {
	s.startDelay = &d
	return s
}

// WithArgs sets the arguments for the scheduled command
func (s *Schedule) WithArgs(args *Args) *Schedule {
	s.args = args
	return s
}

// WithOption adds an option to the scheduled command
func (s *Schedule) WithOption(name string, value any) *Schedule {
	s.args.SetOption(name, value)
	return s
}

// WithArgument adds an argument to the scheduled command
func (s *Schedule) WithArgument(name, value string) *Schedule {
	s.args.SetArgument(name, value)
	return s
}

// ToCronJob converts the schedule to a cron.JobConfig
func (s *Schedule) ToCronJob(handler func() error) (cron.JobConfig, error) {
	if s.interval == nil && s.dailyAt == nil {
		return cron.JobConfig{}, ErrScheduleNotConfigured
	}

	var params []cron.JobParam

	if s.interval != nil {
		params = append(params, cron.WithInterval(*s.interval))
	}

	if s.dailyAt != nil {
		params = append(params, cron.WithDailyAt(*s.dailyAt))
	}

	if s.startImmediately {
		params = append(params, cron.WithStartImmediately())
	}

	if s.startDelay != nil {
		params = append(params, cron.WithDelay(*s.startDelay))
	}

	return cron.NewJob(
		func(_ context.Context) error {
			return handler()
		},
		params...,
	), nil
}

// IsConfigured returns true if the schedule has timing configuration
func (s *Schedule) IsConfigured() bool {
	return s.interval != nil || s.dailyAt != nil
}
