package gohorizon

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/braiphub/go-core/log"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// WorkerStatus represents the worker's current state
type WorkerStatus string

const (
	WorkerStatusIdle     WorkerStatus = "idle"
	WorkerStatusRunning  WorkerStatus = "running"
	WorkerStatusPaused   WorkerStatus = "paused"
	WorkerStatusStopping WorkerStatus = "stopping"
	WorkerStatusStopped  WorkerStatus = "stopped"
)

// Worker processes jobs from queues
type Worker struct {
	id            string
	supervisorID  string
	queue         *Queue
	failedStore   *FailedJobStore
	registry      *JobRegistry
	queues        []string
	logger        log.LoggerI
	metrics       *MetricsCollector
	redis         *redis.Client
	keys          *keyBuilder
	status        atomic.Value
	currentJob    atomic.Value
	jobsProcessed int64
	startedAt     time.Time
	sleep         time.Duration
	maxJobs       int
	maxTime       time.Duration
	stopCh        chan struct{}
	pauseCh       chan struct{}
	resumeCh      chan struct{}
	mu            sync.RWMutex
}

// WorkerOption configures a worker
type WorkerOption func(*Worker)

// WithWorkerQueues sets the queues to process
func WithWorkerQueues(queues ...string) WorkerOption {
	return func(w *Worker) {
		w.queues = queues
	}
}

// WithWorkerSleep sets the sleep duration when no jobs
func WithWorkerSleep(d time.Duration) WorkerOption {
	return func(w *Worker) {
		w.sleep = d
	}
}

// WithWorkerMaxJobs sets max jobs before worker restart
func WithWorkerMaxJobs(n int) WorkerOption {
	return func(w *Worker) {
		w.maxJobs = n
	}
}

// WithWorkerMaxTime sets max time before worker restart
func WithWorkerMaxTime(d time.Duration) WorkerOption {
	return func(w *Worker) {
		w.maxTime = d
	}
}

// NewWorker creates a new worker
func NewWorker(
	queue *Queue,
	failedStore *FailedJobStore,
	registry *JobRegistry,
	redisClient *redis.Client,
	prefix string,
	logger log.LoggerI,
	metrics *MetricsCollector,
	opts ...WorkerOption,
) *Worker {
	w := &Worker{
		id:          uuid.New().String(),
		queue:       queue,
		failedStore: failedStore,
		registry:    registry,
		redis:       redisClient,
		keys:        newKeyBuilder(prefix),
		logger:      logger,
		metrics:     metrics,
		queues:      []string{"default"},
		sleep:       3 * time.Second,
		maxJobs:     0, // unlimited
		maxTime:     0, // unlimited
		stopCh:      make(chan struct{}),
		pauseCh:     make(chan struct{}),
		resumeCh:    make(chan struct{}),
	}

	w.status.Store(WorkerStatusIdle)

	for _, opt := range opts {
		opt(w)
	}

	return w
}

// ID returns the worker ID
func (w *Worker) ID() string {
	return w.id
}

// SetSupervisorID sets the supervisor ID
func (w *Worker) SetSupervisorID(id string) {
	w.supervisorID = id
}

// Start begins processing jobs
func (w *Worker) Start(ctx context.Context) error {
	if w.Status() == WorkerStatusRunning {
		return ErrAlreadyStarted
	}

	w.status.Store(WorkerStatusRunning)
	w.startedAt = time.Now()

	// Register worker in Redis
	w.registerWorker(ctx)

	defer func() {
		w.status.Store(WorkerStatusStopped)
		w.unregisterWorker(ctx)
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-w.stopCh:
			return nil
		case <-w.pauseCh:
			w.status.Store(WorkerStatusPaused)
			w.updateWorkerStatus(ctx)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-w.stopCh:
				return nil
			case <-w.resumeCh:
				w.status.Store(WorkerStatusRunning)
				w.updateWorkerStatus(ctx)
			}
		default:
			// Check max jobs limit
			if w.maxJobs > 0 && atomic.LoadInt64(&w.jobsProcessed) >= int64(w.maxJobs) {
				return nil
			}

			// Check max time limit
			if w.maxTime > 0 && time.Since(w.startedAt) >= w.maxTime {
				return nil
			}

			// Process next job
			if err := w.processNextJob(ctx); err != nil {
				if err == ErrQueueEmpty {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(w.sleep):
						continue
					}
				}
				// Log error but continue
				if w.logger != nil {
					w.logger.WithContext(ctx).Error("worker error processing job", err)
				}
			}
		}
	}
}

// Stop gracefully stops the worker
func (w *Worker) Stop(ctx context.Context) error {
	if w.Status() == WorkerStatusStopped {
		return ErrNotStarted
	}

	w.status.Store(WorkerStatusStopping)
	close(w.stopCh)
	return nil
}

// Pause temporarily stops job processing
func (w *Worker) Pause() {
	select {
	case w.pauseCh <- struct{}{}:
	default:
	}
}

// Resume continues processing after pause
func (w *Worker) Resume() {
	select {
	case w.resumeCh <- struct{}{}:
	default:
	}
}

// Status returns the current worker status
func (w *Worker) Status() WorkerStatus {
	return w.status.Load().(WorkerStatus)
}

// CurrentJob returns the job currently being processed
func (w *Worker) CurrentJob() *Payload {
	if val := w.currentJob.Load(); val != nil {
		return val.(*Payload)
	}
	return nil
}

// JobsProcessed returns the number of jobs processed
func (w *Worker) JobsProcessed() int64 {
	return atomic.LoadInt64(&w.jobsProcessed)
}

func (w *Worker) processNextJob(ctx context.Context) error {
	payload, err := w.queue.Pop(ctx, w.queues...)
	if err != nil {
		return err
	}

	w.currentJob.Store(payload)
	defer w.currentJob.Store((*Payload)(nil))

	start := time.Now()

	// Create job context with timeout
	jobCtx, cancel := context.WithTimeout(ctx, payload.Timeout)
	defer cancel()

	// Execute job
	err = w.executeJob(jobCtx, payload)
	runtime := time.Since(start)

	if err != nil {
		return w.handleFailure(ctx, payload, err, runtime)
	}

	return w.handleSuccess(ctx, payload, runtime)
}

func (w *Worker) executeJob(ctx context.Context, payload *Payload) (err error) {
	// Recover from panics
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("job panicked: %v\n%s", r, debug.Stack())
		}
	}()

	// Hydrate job from registry
	job, err := w.registry.Hydrate(payload)
	if err != nil {
		return err
	}

	// Execute job
	return job.Handle(ctx)
}

func (w *Worker) handleSuccess(ctx context.Context, payload *Payload, runtime time.Duration) error {
	atomic.AddInt64(&w.jobsProcessed, 1)

	// Delete job from queue
	if err := w.queue.Delete(ctx, payload.Queue, payload); err != nil {
		if w.logger != nil {
			w.logger.WithContext(ctx).Error("failed to delete completed job", err)
		}
	}

	// Record metrics
	if w.metrics != nil {
		w.metrics.RecordJobProcessed(ctx, payload.Queue, payload, runtime)
	}

	// Store in recent jobs
	w.storeRecentJob(ctx, payload, StatusCompleted, runtime)

	return nil
}

func (w *Worker) handleFailure(ctx context.Context, payload *Payload, jobErr error, runtime time.Duration) error {
	atomic.AddInt64(&w.jobsProcessed, 1)

	// Check if we should retry
	if payload.Attempts < payload.MaxAttempts {
		// Release back to queue for retry
		if err := w.queue.Release(ctx, payload.Queue, payload, payload.RetryDelay); err != nil {
			if w.logger != nil {
				w.logger.WithContext(ctx).Error("failed to release job for retry", err)
			}
		}
		return nil
	}

	// Max retries exceeded, store in failed jobs
	if err := w.failedStore.Store(ctx, payload, jobErr.Error()); err != nil {
		if w.logger != nil {
			w.logger.WithContext(ctx).Error("failed to store failed job", err)
		}
	}

	// Record metrics
	if w.metrics != nil {
		w.metrics.RecordJobFailed(ctx, payload.Queue, payload, jobErr)
	}

	// Store in recent jobs
	w.storeRecentJob(ctx, payload, StatusFailed, runtime)

	return nil
}

func (w *Worker) storeRecentJob(ctx context.Context, payload *Payload, status Status, runtime time.Duration) {
	recent := &RecentJob{
		ID:          payload.ID,
		Name:        payload.Name,
		Queue:       payload.Queue,
		Status:      status,
		Attempts:    payload.Attempts,
		Runtime:     runtime,
		CompletedAt: time.Now(),
		Tags:        payload.Tags,
	}

	data, err := json.Marshal(recent)
	if err != nil {
		return
	}

	pipe := w.redis.Pipeline()

	// Add to recent jobs (limit to 1000)
	pipe.LPush(ctx, w.keys.recentJobs(), data)
	pipe.LTrim(ctx, w.keys.recentJobs(), 0, 999)

	pipe.Exec(ctx)
}

func (w *Worker) registerWorker(ctx context.Context) {
	data := map[string]interface{}{
		"id":           w.id,
		"supervisor":   w.supervisorID,
		"status":       string(w.Status()),
		"queues":       w.queues,
		"started_at":   w.startedAt.Unix(),
		"pid":          0, // Go doesn't have easy process ID access
		"memory":       0,
		"current_job":  "",
		"last_seen_at": time.Now().Unix(),
	}

	dataJSON, _ := json.Marshal(data)
	w.redis.Set(ctx, w.keys.worker(w.id), dataJSON, 5*time.Minute)

	if w.supervisorID != "" {
		w.redis.SAdd(ctx, w.keys.supervisorWorkers(w.supervisorID), w.id)
	}
}

func (w *Worker) updateWorkerStatus(ctx context.Context) {
	w.registerWorker(ctx)
}

func (w *Worker) unregisterWorker(ctx context.Context) {
	w.redis.Del(ctx, w.keys.worker(w.id))

	if w.supervisorID != "" {
		w.redis.SRem(ctx, w.keys.supervisorWorkers(w.supervisorID), w.id)
	}
}
