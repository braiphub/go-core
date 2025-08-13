package cron

import (
	"context"
	"time"
)

type JobConfig struct {
	Func             func(ctx context.Context) error
	Interval         *time.Duration
	DailyAt          *DailyAt
	StartImmediately bool
	StartDelay       *time.Duration
}

type DailyAt struct {
	Hour   int
	Minute int
	Second int
}

type JobParam = func(*JobConfig)

func NewJob(fn func(context.Context) error, params ...JobParam) JobConfig {
	job := JobConfig{
		Func: fn,
	}

	for _, p := range params {
		p(&job)
	}

	return job
}

func WithInterval(interval time.Duration) func(*JobConfig) {
	return func(config *JobConfig) {
		config.Interval = &interval
	}
}

func WithDailyAt(dailyAt DailyAt) func(*JobConfig) {
	return func(config *JobConfig) {
		config.DailyAt = &dailyAt
	}
}

func WithDailyAtHour(hour int) func(*JobConfig) {
	return func(config *JobConfig) {
		config.DailyAt = &DailyAt{
			Hour:   hour,
			Minute: 0,
			Second: 0,
		}
	}
}

func WithStartImmediately() func(*JobConfig) {
	return func(config *JobConfig) {
		config.StartImmediately = true
	}
}

func WithDelay(delay time.Duration) func(*JobConfig) {
	return func(config *JobConfig) {
		config.StartDelay = &delay
	}
}
