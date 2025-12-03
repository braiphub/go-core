package gohorizon

import "errors"

var (
	// ErrJobNotRegistered is returned when trying to process an unregistered job type
	ErrJobNotRegistered = errors.New("job type not registered")

	// ErrQueueEmpty is returned when the queue has no jobs
	ErrQueueEmpty = errors.New("queue is empty")

	// ErrJobNotFound is returned when a job cannot be found
	ErrJobNotFound = errors.New("job not found")

	// ErrFailedJobNotFound is returned when a failed job cannot be found
	ErrFailedJobNotFound = errors.New("failed job not found")

	// ErrSupervisorNotFound is returned when a supervisor cannot be found
	ErrSupervisorNotFound = errors.New("supervisor not found")

	// ErrWorkerNotFound is returned when a worker cannot be found
	ErrWorkerNotFound = errors.New("worker not found")

	// ErrAlreadyStarted is returned when trying to start an already running component
	ErrAlreadyStarted = errors.New("already started")

	// ErrNotStarted is returned when trying to stop a component that hasn't started
	ErrNotStarted = errors.New("not started")

	// ErrRedisNotConfigured is returned when Redis client is not set
	ErrRedisNotConfigured = errors.New("redis client not configured")

	// ErrInvalidConfig is returned when configuration is invalid
	ErrInvalidConfig = errors.New("invalid configuration")

	// ErrJobTimeout is returned when a job exceeds its timeout
	ErrJobTimeout = errors.New("job execution timeout")

	// ErrMaxRetriesExceeded is returned when a job has exceeded max retries
	ErrMaxRetriesExceeded = errors.New("max retries exceeded")
)
