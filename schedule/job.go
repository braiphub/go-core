package schedule

import (
	"context"
	"time"
)

// Job represents a scheduled job configuration
type Job struct {
	Name             string
	Func             func(ctx context.Context) error
	Interval         *time.Duration
	DailyAt          *DailyAt
	WeeklyAt         *WeeklyAt
	CronExpr         string
	StartImmediately bool
	StartDelay       *time.Duration
	Timezone         *time.Location
}

// DailyAt specifies a time of day for daily jobs
type DailyAt struct {
	Hour   int
	Minute int
	Second int
}

// WeeklyAt specifies a day and time for weekly jobs
type WeeklyAt struct {
	Weekday time.Weekday
	Hour    int
	Minute  int
	Second  int
}

// JobOption is a functional option for configuring a Job
type JobOption func(*Job)

// NewJob creates a new Job with the given function and options
func NewJob(fn func(context.Context) error, opts ...JobOption) Job {
	job := Job{
		Func: fn,
	}

	for _, opt := range opts {
		opt(&job)
	}

	return job
}

// WithName sets a name for the job (useful for logging)
func WithName(name string) JobOption {
	return func(j *Job) {
		j.Name = name
	}
}

// WithInterval sets a duration interval for the job
func WithInterval(interval time.Duration) JobOption {
	return func(j *Job) {
		j.Interval = &interval
	}
}

// WithDailyAt sets the job to run daily at the specified time
func WithDailyAt(dailyAt DailyAt) JobOption {
	return func(j *Job) {
		j.DailyAt = &dailyAt
	}
}

// WithDailyAtHour sets the job to run daily at the specified hour (minute and second = 0)
func WithDailyAtHour(hour int) JobOption {
	return func(j *Job) {
		j.DailyAt = &DailyAt{
			Hour:   hour,
			Minute: 0,
			Second: 0,
		}
	}
}

// WithDailyAtTime sets the job to run daily at the specified hour and minute
func WithDailyAtTime(hour, minute int) JobOption {
	return func(j *Job) {
		j.DailyAt = &DailyAt{
			Hour:   hour,
			Minute: minute,
			Second: 0,
		}
	}
}

// WithWeeklyAt sets the job to run weekly at the specified day and time
func WithWeeklyAt(weeklyAt WeeklyAt) JobOption {
	return func(j *Job) {
		j.WeeklyAt = &weeklyAt
	}
}

// WithCron sets a cron expression for the job
func WithCron(expr string) JobOption {
	return func(j *Job) {
		j.CronExpr = expr
	}
}

// WithStartImmediately makes the job run immediately when the scheduler starts
func WithStartImmediately() JobOption {
	return func(j *Job) {
		j.StartImmediately = true
	}
}

// WithDelay delays the first execution of the job
func WithDelay(delay time.Duration) JobOption {
	return func(j *Job) {
		j.StartDelay = &delay
	}
}

// WithTimezone sets the timezone for the job
func WithTimezone(loc *time.Location) JobOption {
	return func(j *Job) {
		j.Timezone = loc
	}
}

// Convenience functions for common intervals

// EverySecond creates an interval option for every second
func EverySecond() JobOption {
	return WithInterval(time.Second)
}

// EverySeconds creates an interval option for every N seconds
func EverySeconds(n int) JobOption {
	return WithInterval(time.Duration(n) * time.Second)
}

// EveryMinute creates an interval option for every minute
func EveryMinute() JobOption {
	return WithInterval(time.Minute)
}

// EveryMinutes creates an interval option for every N minutes
func EveryMinutes(n int) JobOption {
	return WithInterval(time.Duration(n) * time.Minute)
}

// EveryHour creates an interval option for every hour
func EveryHour() JobOption {
	return WithInterval(time.Hour)
}

// EveryHours creates an interval option for every N hours
func EveryHours(n int) JobOption {
	return WithInterval(time.Duration(n) * time.Hour)
}

// Daily creates an option for daily execution at midnight
func Daily() JobOption {
	return WithDailyAt(DailyAt{Hour: 0, Minute: 0, Second: 0})
}

// DailyAt creates an option for daily execution at the specified time
func DailyAtTime(hour, minute int) JobOption {
	return WithDailyAt(DailyAt{Hour: hour, Minute: minute, Second: 0})
}
