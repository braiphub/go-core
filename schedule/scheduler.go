package schedule

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/braiphub/go-core/log"
	"github.com/dromara/carbon/v2"
	"github.com/go-co-op/gocron/v2"
	"github.com/pkg/errors"
)

// Scheduler manages scheduled jobs
type Scheduler struct {
	scheduler gocron.Scheduler
	logger    log.LoggerI
	location  *time.Location
	jobs      []Job
}

// SchedulerOption is a functional option for configuring the Scheduler
type SchedulerOption func(*Scheduler)

// WithLogger sets the logger for the scheduler
func WithLogger(logger log.LoggerI) SchedulerOption {
	return func(s *Scheduler) {
		s.logger = logger
	}
}

// WithLocation sets the timezone for the scheduler
func WithLocation(loc *time.Location) SchedulerOption {
	return func(s *Scheduler) {
		s.location = loc
	}
}

// WithJobs adds jobs to the scheduler during initialization
func WithJobs(jobs ...Job) SchedulerOption {
	return func(s *Scheduler) {
		s.jobs = append(s.jobs, jobs...)
	}
}

// New creates a new Scheduler with the given options
func New(opts ...SchedulerOption) (*Scheduler, error) {
	s := &Scheduler{
		location: time.UTC,
		jobs:     make([]Job, 0),
	}

	for _, opt := range opts {
		opt(s)
	}

	scheduler, err := gocron.NewScheduler(gocron.WithLocation(s.location))
	if err != nil {
		return nil, errors.Wrap(err, "create scheduler")
	}

	s.scheduler = scheduler
	return s, nil
}

// AddJob adds a job to the scheduler
func (s *Scheduler) AddJob(job Job) error {
	s.jobs = append(s.jobs, job)
	return nil
}

// AddJobs adds multiple jobs to the scheduler
func (s *Scheduler) AddJobs(jobs ...Job) error {
	s.jobs = append(s.jobs, jobs...)
	return nil
}

// Start starts the scheduler and blocks until context is cancelled
func (s *Scheduler) Start(ctx context.Context) error {
	ctx, terminate := context.WithCancel(ctx)

	// Capture OS terminate signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		terminate()
	}()

	// Register all jobs
	if err := s.registerJobs(ctx); err != nil {
		return err
	}

	// Start the scheduler
	s.scheduler.Start()
	defer s.scheduler.Shutdown()

	s.logInfo(ctx, "Schedule service: started...")
	<-ctx.Done()
	s.logInfo(ctx, "Schedule service: terminated")

	return nil
}

// registerJobs registers all jobs with the gocron scheduler
func (s *Scheduler) registerJobs(ctx context.Context) error {
	for _, job := range s.jobs {
		if err := s.registerJob(ctx, job); err != nil {
			return err
		}
	}
	return nil
}

// registerJob registers a single job with the gocron scheduler
func (s *Scheduler) registerJob(ctx context.Context, job Job) error {
	options := []gocron.JobOption{
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	}

	if job.StartImmediately {
		options = append(options, gocron.JobOption(gocron.WithStartImmediately()))
	}

	if job.StartDelay != nil {
		startAt := time.Now().Add(*job.StartDelay)
		options = append(options, gocron.JobOption(gocron.WithStartDateTime(startAt)))
	}

	jobDef, err := s.buildJobDefinition(job)
	if err != nil {
		return err
	}

	// Create the task
	capturedJob := job
	capturedCtx := ctx
	_, err = s.scheduler.NewJob(
		jobDef,
		gocron.NewTask(
			func(ctx context.Context) {
				if err := capturedJob.Func(ctx); err != nil {
					s.logError(ctx, "schedule job error", err, capturedJob.Name)
				}
			},
			capturedCtx,
		),
		options...,
	)
	if err != nil {
		return errors.Wrap(err, "schedule new job")
	}

	return nil
}

// buildJobDefinition builds the gocron job definition based on job configuration
func (s *Scheduler) buildJobDefinition(job Job) (gocron.JobDefinition, error) {
	// Cron expression takes priority
	if job.CronExpr != "" {
		return gocron.CronJob(job.CronExpr, false), nil
	}

	// Interval
	if job.Interval != nil {
		return gocron.DurationJob(*job.Interval), nil
	}

	// Daily at
	if job.DailyAt != nil {
		loc := s.location
		if job.Timezone != nil {
			loc = job.Timezone
		}

		t := carbon.
			CreateFromTime(job.DailyAt.Hour, job.DailyAt.Minute, job.DailyAt.Second, loc.String()).
			StdTime()

		return gocron.DailyJob(
			1,
			gocron.NewAtTimes(
				gocron.NewAtTime(
					uint(t.UTC().Hour()),
					uint(t.UTC().Minute()),
					uint(t.UTC().Second()),
				),
			),
		), nil
	}

	// Weekly at
	if job.WeeklyAt != nil {
		loc := s.location
		if job.Timezone != nil {
			loc = job.Timezone
		}

		t := carbon.
			CreateFromTime(job.WeeklyAt.Hour, job.WeeklyAt.Minute, job.WeeklyAt.Second, loc.String()).
			StdTime()

		return gocron.WeeklyJob(
			1,
			gocron.NewWeekdays(job.WeeklyAt.Weekday),
			gocron.NewAtTimes(
				gocron.NewAtTime(
					uint(t.UTC().Hour()),
					uint(t.UTC().Minute()),
					uint(t.UTC().Second()),
				),
			),
		), nil
	}

	return nil, errors.New("job must have an interval, dailyAt, weeklyAt, or cron expression")
}

func (s *Scheduler) logInfo(ctx context.Context, msg string, fields ...log.Field) {
	if s.logger == nil {
		return
	}
	// Convert fields to any slice for LoggerI interface
	anyFields := make([]any, 0, len(fields))
	for _, f := range fields {
		anyFields = append(anyFields, f)
	}
	s.logger.WithContext(ctx).Info(msg, anyFields...)
}

func (s *Scheduler) logError(ctx context.Context, msg string, err error, jobName string) {
	if s.logger == nil {
		return
	}
	if jobName != "" {
		s.logger.WithContext(ctx).Error(msg, err, log.Any("job", jobName))
	} else {
		s.logger.WithContext(ctx).Error(msg, err)
	}
}

// Run is a convenience function to create and start a scheduler
// This maintains backward compatibility with cron.RunCronServer
func Run(ctx context.Context, logger log.LoggerI, jobs ...Job) error {
	scheduler, err := New(WithLogger(logger))
	if err != nil {
		return err
	}

	if err := scheduler.AddJobs(jobs...); err != nil {
		return err
	}

	return scheduler.Start(ctx)
}
