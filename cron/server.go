package cron

import (
	"context"
	"github.com/braiphub/go-core/log"
	"github.com/dromara/carbon/v2"
	"github.com/go-co-op/gocron/v2"
	"github.com/pkg/errors"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func RunCronServer(
	ctx context.Context,
	logger log.LoggerI,
	jobs ...JobConfig,
) error {
	ctx, terminate := context.WithCancel(ctx)

	// capture O.S. terminate signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		terminate()
	}()

	// create a scheduler
	scheduler, err := gocron.NewScheduler(gocron.WithLocation(time.UTC))
	if err != nil {
		return err
	}

	// register jobs
	if err := enqueueCronJobs(ctx, scheduler, jobs, logger); err != nil {
		return err
	}

	// start the scheduler
	scheduler.Start()
	defer scheduler.Shutdown()

	// when you're done, shut it down
	logInfo(ctx, logger, "Cron service: started...")
	<-ctx.Done()
	logInfo(ctx, logger, "Cron service: terminated")

	return nil
}

func enqueueCronJobs(ctx context.Context, s gocron.Scheduler, configs []JobConfig, logger log.LoggerI) error {
	for _, cfg := range configs {
		options := []gocron.JobOption{
			gocron.WithSingletonMode(gocron.LimitModeReschedule),
		}

		if cfg.StartImmediately {
			options = append(options, gocron.JobOption(gocron.WithStartImmediately()))
		}

		if cfg.StartDelay != nil {
			startAt := time.Now().Add(*cfg.StartDelay)
			options = append(options, gocron.JobOption(gocron.WithStartDateTime(startAt)))
		}

		var job gocron.JobDefinition

		if cfg.Interval != nil {
			job = gocron.DurationJob(*cfg.Interval)
		}

		if cfg.DailyAt != nil {
			t := carbon.
				SetTimezone("America/Sao_Paulo").
				CreateFromTime(cfg.DailyAt.Hour, cfg.DailyAt.Minute, cfg.DailyAt.Second).
				StdTime()

			job = gocron.DailyJob(
				1,
				gocron.NewAtTimes(
					gocron.NewAtTime(
						uint(t.UTC().Hour()),
						uint(t.UTC().Minute()),
						uint(t.UTC().Second()),
					),
				),
			)
		}

		if job == nil {
			panic("you must specify job interval")
		}

		_, err := s.NewJob(
			job,
			gocron.NewTask(
				func(ctx context.Context) {
					if err := cfg.Func(ctx); err != nil {
						logError(ctx, logger, "cron job error", err)
					}
				},
				ctx,
			),
			options...,
		)
		if err != nil {
			return errors.Wrap(err, "schedule new job")
		}
	}

	return nil
}

func logInfo(ctx context.Context, logger log.LoggerI, msg string, params ...any) {
	if logger == nil {
		return
	}

	logger.WithContext(ctx).Info(msg, params...)
}

func logError(ctx context.Context, logger log.LoggerI, msg string, err error) {
	if logger == nil {
		return
	}

	logger.WithContext(ctx).Error(msg, err)
}
