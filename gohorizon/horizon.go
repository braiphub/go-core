package gohorizon

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/braiphub/go-core/log"
	"github.com/redis/go-redis/v9"
)

// Horizon is the main coordinator for queue monitoring and workers
type Horizon struct {
	config      Config
	logger      log.LoggerI
	redis       *redis.Client
	queue       *Queue
	failedStore *FailedJobStore
	registry    *JobRegistry
	supervisors map[string]*Supervisor
	metrics     *MetricsCollector
	httpServer  *HTTPServer
	started     bool
	stopCh      chan struct{}
	mu          sync.RWMutex
	wg          sync.WaitGroup
}

// New creates a new Horizon instance with the given options
func New(opts ...Option) (*Horizon, error) {
	h := &Horizon{
		config:      DefaultConfig(),
		registry:    NewJobRegistry(),
		supervisors: make(map[string]*Supervisor),
		stopCh:      make(chan struct{}),
	}

	for _, opt := range opts {
		opt(h)
	}

	if err := h.validate(); err != nil {
		return nil, err
	}

	// Initialize Redis client if not provided
	if h.redis == nil {
		h.redis = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", h.config.Redis.Host, h.config.Redis.Port),
			Password: h.config.Redis.Password,
			DB:       h.config.Redis.DB,
		})
	}

	// Initialize queue
	h.queue = NewQueue(h.redis, h.config.Prefix)

	// Initialize failed job store
	h.failedStore = NewFailedJobStore(h.redis, h.config.Prefix, h.queue)

	// Initialize metrics collector
	h.metrics = NewMetricsCollector(h.redis, h.config.Prefix, h.queue, h.failedStore)

	// Initialize supervisors
	for name, config := range h.config.Supervisors {
		h.supervisors[name] = NewSupervisor(
			config,
			h.queue,
			h.failedStore,
			h.registry,
			h.redis,
			h.config.Prefix,
			h.logger,
			h.metrics,
		)
	}

	// Initialize HTTP server
	if h.config.HTTP.Enabled {
		h.httpServer = NewHTTPServer(h, h.config.HTTP)
	}

	return h, nil
}

func (h *Horizon) validate() error {
	if h.config.Prefix == "" {
		h.config.Prefix = "horizon"
	}

	if h.config.Redis.Host == "" {
		h.config.Redis.Host = "localhost"
	}

	if h.config.Redis.Port == 0 {
		h.config.Redis.Port = 6379
	}

	return nil
}

// Start begins all supervisors and the HTTP server
func (h *Horizon) Start(ctx context.Context) error {
	h.mu.Lock()
	if h.started {
		h.mu.Unlock()
		return ErrAlreadyStarted
	}
	h.started = true
	h.stopCh = make(chan struct{})
	h.mu.Unlock()

	// Test Redis connection
	if err := h.redis.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis connection failed: %w", err)
	}

	if h.logger != nil {
		h.logger.Info(ctx, "starting horizon")
	}

	// Start metrics snapshot routine
	if h.config.Metrics.Enabled {
		h.wg.Add(1)
		go h.runMetricsCollector(ctx)
	}

	// Start HTTP server
	if h.httpServer != nil {
		h.wg.Add(1)
		go func() {
			defer h.wg.Done()
			if err := h.httpServer.Start(ctx); err != nil && h.logger != nil {
				h.logger.Error(ctx, "http server error", err)
			}
		}()
	}

	// Start supervisors
	for name, supervisor := range h.supervisors {
		h.wg.Add(1)
		go func(name string, sup *Supervisor) {
			defer h.wg.Done()
			if h.logger != nil {
				h.logger.Info(ctx, fmt.Sprintf("starting supervisor: %s", name))
			}
			if err := sup.Start(ctx); err != nil && h.logger != nil {
				h.logger.Error(ctx, fmt.Sprintf("supervisor %s error", name), err)
			}
		}(name, supervisor)
	}

	// Wait for stop signal
	select {
	case <-ctx.Done():
		return h.Stop(context.Background())
	case <-h.stopCh:
		return nil
	}
}

// Stop gracefully shuts down all workers and servers
func (h *Horizon) Stop(ctx context.Context) error {
	h.mu.Lock()
	if !h.started {
		h.mu.Unlock()
		return ErrNotStarted
	}
	h.started = false
	h.mu.Unlock()

	if h.logger != nil {
		h.logger.Info(ctx, "stopping horizon")
	}

	// Signal stop
	close(h.stopCh)

	// Stop HTTP server
	if h.httpServer != nil {
		h.httpServer.Stop(ctx)
	}

	// Stop supervisors
	for _, supervisor := range h.supervisors {
		supervisor.Stop(ctx)
	}

	// Wait for all goroutines
	h.wg.Wait()

	return nil
}

// RegisterJob adds a job type to the registry
func (h *Horizon) RegisterJob(factory func() Job) {
	h.registry.Register(factory)
}

// Dispatch queues a job for processing
func (h *Horizon) Dispatch(ctx context.Context, job Job, opts ...DispatchOption) error {
	options := &dispatchOptions{
		queue: "default",
	}

	// Check if job specifies its own queue
	if jq, ok := job.(JobWithQueue); ok {
		options.queue = jq.Queue()
	}

	for _, opt := range opts {
		opt(options)
	}

	payload, err := NewPayload(job, options.queue)
	if err != nil {
		return fmt.Errorf("failed to create payload: %w", err)
	}

	// Add additional tags
	if len(options.tags) > 0 {
		payload.Tags = append(payload.Tags, options.tags...)
	}

	if options.delay > 0 {
		return h.queue.Later(ctx, options.queue, payload, options.delay)
	}

	return h.queue.Push(ctx, options.queue, payload)
}

// Queue returns the queue instance
func (h *Horizon) Queue() *Queue {
	return h.queue
}

// Metrics returns the metrics collector
func (h *Horizon) Metrics() *MetricsCollector {
	return h.metrics
}

// FailedJobs returns the failed job store
func (h *Horizon) FailedJobs() *FailedJobStore {
	return h.failedStore
}

// Registry returns the job registry
func (h *Horizon) Registry() *JobRegistry {
	return h.registry
}

// Supervisors returns all supervisors
func (h *Horizon) Supervisors() map[string]*Supervisor {
	h.mu.RLock()
	defer h.mu.RUnlock()
	result := make(map[string]*Supervisor)
	for k, v := range h.supervisors {
		result[k] = v
	}
	return result
}

// GetSupervisor returns a supervisor by name
func (h *Horizon) GetSupervisor(name string) (*Supervisor, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	sup, ok := h.supervisors[name]
	if !ok {
		return nil, ErrSupervisorNotFound
	}
	return sup, nil
}

// PauseSupervisor pauses a supervisor
func (h *Horizon) PauseSupervisor(name string) error {
	sup, err := h.GetSupervisor(name)
	if err != nil {
		return err
	}
	return sup.Pause()
}

// ContinueSupervisor resumes a supervisor
func (h *Horizon) ContinueSupervisor(name string) error {
	sup, err := h.GetSupervisor(name)
	if err != nil {
		return err
	}
	return sup.Continue()
}

// ScaleSupervisor scales a supervisor
func (h *Horizon) ScaleSupervisor(ctx context.Context, name string, workers int) error {
	sup, err := h.GetSupervisor(name)
	if err != nil {
		return err
	}
	return sup.Scale(ctx, workers)
}

func (h *Horizon) runMetricsCollector(ctx context.Context) {
	defer h.wg.Done()

	ticker := time.NewTicker(h.config.Metrics.SnapshotInterval)
	defer ticker.Stop()

	trimTicker := time.NewTicker(time.Hour)
	defer trimTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-h.stopCh:
			return
		case <-ticker.C:
			if err := h.metrics.TakeSnapshot(ctx); err != nil && h.logger != nil {
				h.logger.Error(ctx, "failed to take metrics snapshot", err)
			}
		case <-trimTicker.C:
			if err := h.metrics.TrimSnapshots(ctx, h.config.Metrics.RetentionPeriod); err != nil && h.logger != nil {
				h.logger.Error(ctx, "failed to trim snapshots", err)
			}
		}
	}
}
