package gohorizon

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/braiphub/go-core/log"
	"github.com/redis/go-redis/v9"
)

// SupervisorStatus represents the supervisor's current state
type SupervisorStatus string

const (
	SupervisorStatusRunning SupervisorStatus = "running"
	SupervisorStatusPaused  SupervisorStatus = "paused"
	SupervisorStatusStopped SupervisorStatus = "stopped"
)

// BalanceMode defines how workers are distributed
type BalanceMode string

const (
	BalanceModeSimple BalanceMode = "simple" // Static number of workers
	BalanceModeAuto   BalanceMode = "auto"   // Auto-scale based on load
	BalanceModeNull   BalanceMode = "null"   // One worker per queue
)

// SupervisorConfig defines supervisor behavior
type SupervisorConfig struct {
	Name         string        `json:"name"`
	Queues       []string      `json:"queues"`
	Balance      BalanceMode   `json:"balance"`
	MinProcesses int           `json:"min_processes"`
	MaxProcesses int           `json:"max_processes"`
	MaxTime      time.Duration `json:"max_time"`
	MaxJobs      int           `json:"max_jobs"`
	Tries        int           `json:"tries"`
	Timeout      time.Duration `json:"timeout"`
	Sleep        time.Duration `json:"sleep"`
}

// DefaultSupervisorConfig returns sensible defaults
func DefaultSupervisorConfig(name string) SupervisorConfig {
	return SupervisorConfig{
		Name:         name,
		Queues:       []string{"default"},
		Balance:      BalanceModeSimple,
		MinProcesses: 1,
		MaxProcesses: 10,
		MaxTime:      0,
		MaxJobs:      0,
		Tries:        3,
		Timeout:      60 * time.Second,
		Sleep:        3 * time.Second,
	}
}

// Supervisor manages a pool of workers for a specific queue configuration
type Supervisor struct {
	name        string
	config      SupervisorConfig
	queue       *Queue
	failedStore *FailedJobStore
	registry    *JobRegistry
	redis       *redis.Client
	keys        *keyBuilder
	prefix      string
	logger      log.LoggerI
	metrics     *MetricsCollector
	workers     []*Worker
	status      SupervisorStatus
	stopCh      chan struct{}
	mu          sync.RWMutex
	wg          sync.WaitGroup
}

// NewSupervisor creates a new supervisor
func NewSupervisor(
	config SupervisorConfig,
	queue *Queue,
	failedStore *FailedJobStore,
	registry *JobRegistry,
	redisClient *redis.Client,
	prefix string,
	logger log.LoggerI,
	metrics *MetricsCollector,
) *Supervisor {
	return &Supervisor{
		name:        config.Name,
		config:      config,
		queue:       queue,
		failedStore: failedStore,
		registry:    registry,
		redis:       redisClient,
		keys:        newKeyBuilder(prefix),
		prefix:      prefix,
		logger:      logger,
		metrics:     metrics,
		workers:     make([]*Worker, 0),
		status:      SupervisorStatusStopped,
		stopCh:      make(chan struct{}),
	}
}

// Name returns the supervisor name
func (s *Supervisor) Name() string {
	return s.name
}

// Config returns the supervisor config
func (s *Supervisor) Config() SupervisorConfig {
	return s.config
}

// Start begins the supervisor and its workers
func (s *Supervisor) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.status == SupervisorStatusRunning {
		s.mu.Unlock()
		return ErrAlreadyStarted
	}

	s.status = SupervisorStatusRunning
	s.stopCh = make(chan struct{})
	s.mu.Unlock()

	// Register supervisor in Redis
	s.registerSupervisor(ctx)

	// Start initial workers
	initialWorkers := s.config.MinProcesses
	if s.config.Balance == BalanceModeNull {
		initialWorkers = len(s.config.Queues)
	}

	for i := 0; i < initialWorkers; i++ {
		s.spawnWorker(ctx)
	}

	// Start balancer if auto mode
	if s.config.Balance == BalanceModeAuto {
		go s.runBalancer(ctx)
	}

	// Wait for stop signal
	select {
	case <-ctx.Done():
		return s.Stop(context.Background())
	case <-s.stopCh:
		return nil
	}
}

// Stop gracefully shuts down the supervisor
func (s *Supervisor) Stop(ctx context.Context) error {
	s.mu.Lock()
	if s.status == SupervisorStatusStopped {
		s.mu.Unlock()
		return ErrNotStarted
	}

	s.status = SupervisorStatusStopped
	close(s.stopCh)
	workers := make([]*Worker, len(s.workers))
	copy(workers, s.workers)
	s.mu.Unlock()

	// Stop all workers
	for _, w := range workers {
		w.Stop(ctx)
	}

	// Wait for workers to finish
	s.wg.Wait()

	// Unregister supervisor
	s.unregisterSupervisor(ctx)

	return nil
}

// Pause pauses all workers
func (s *Supervisor) Pause() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.status != SupervisorStatusRunning {
		return ErrNotStarted
	}

	s.status = SupervisorStatusPaused

	for _, w := range s.workers {
		w.Pause()
	}

	return nil
}

// Continue resumes all workers
func (s *Supervisor) Continue() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.status != SupervisorStatusPaused {
		return nil
	}

	s.status = SupervisorStatusRunning

	for _, w := range s.workers {
		w.Resume()
	}

	return nil
}

// Scale adjusts the number of workers
func (s *Supervisor) Scale(ctx context.Context, count int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Enforce limits
	if count < s.config.MinProcesses {
		count = s.config.MinProcesses
	}
	if count > s.config.MaxProcesses {
		count = s.config.MaxProcesses
	}

	current := len(s.workers)

	if count > current {
		// Spawn more workers
		for i := 0; i < count-current; i++ {
			s.spawnWorkerLocked(ctx)
		}
	} else if count < current {
		// Stop excess workers
		for i := 0; i < current-count; i++ {
			if len(s.workers) > 0 {
				worker := s.workers[len(s.workers)-1]
				s.workers = s.workers[:len(s.workers)-1]
				go worker.Stop(ctx)
			}
		}
	}

	return nil
}

// Status returns supervisor status
func (s *Supervisor) Status() SupervisorStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

// Workers returns all workers
func (s *Supervisor) Workers() []*Worker {
	s.mu.RLock()
	defer s.mu.RUnlock()
	workers := make([]*Worker, len(s.workers))
	copy(workers, s.workers)
	return workers
}

// WorkerCount returns the number of workers
func (s *Supervisor) WorkerCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.workers)
}

func (s *Supervisor) spawnWorker(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.spawnWorkerLocked(ctx)
}

func (s *Supervisor) spawnWorkerLocked(ctx context.Context) {
	worker := NewWorker(
		s.queue,
		s.failedStore,
		s.registry,
		s.redis,
		s.prefix,
		s.logger,
		s.metrics,
		WithWorkerQueues(s.config.Queues...),
		WithWorkerSleep(s.config.Sleep),
		WithWorkerMaxJobs(s.config.MaxJobs),
		WithWorkerMaxTime(s.config.MaxTime),
	)

	worker.SetSupervisorID(s.name)
	s.workers = append(s.workers, worker)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer s.removeWorker(worker)

		if err := worker.Start(ctx); err != nil && s.logger != nil {
			s.logger.Error(ctx, "worker stopped with error", err)
		}

		// Respawn if supervisor is still running and we're under min processes
		s.mu.RLock()
		shouldRespawn := s.status == SupervisorStatusRunning && len(s.workers) < s.config.MinProcesses
		s.mu.RUnlock()

		if shouldRespawn {
			s.spawnWorker(ctx)
		}
	}()
}

func (s *Supervisor) removeWorker(worker *Worker) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, w := range s.workers {
		if w.ID() == worker.ID() {
			s.workers = append(s.workers[:i], s.workers[i+1:]...)
			break
		}
	}
}

func (s *Supervisor) runBalancer(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.balance(ctx)
		}
	}
}

func (s *Supervisor) balance(ctx context.Context) {
	s.mu.RLock()
	if s.status != SupervisorStatusRunning {
		s.mu.RUnlock()
		return
	}
	currentWorkers := len(s.workers)
	s.mu.RUnlock()

	// Calculate total queue size
	var totalPending int64
	for _, queueName := range s.config.Queues {
		size, err := s.queue.Size(ctx, queueName)
		if err == nil {
			totalPending += size
		}
	}

	// Simple auto-scaling logic
	var targetWorkers int
	switch {
	case totalPending == 0:
		targetWorkers = s.config.MinProcesses
	case totalPending < 100:
		targetWorkers = s.config.MinProcesses + 1
	case totalPending < 500:
		targetWorkers = (s.config.MinProcesses + s.config.MaxProcesses) / 2
	case totalPending < 1000:
		targetWorkers = s.config.MaxProcesses - 1
	default:
		targetWorkers = s.config.MaxProcesses
	}

	// Only scale if difference is significant
	if abs(targetWorkers-currentWorkers) >= 1 {
		s.Scale(ctx, targetWorkers)
	}
}

func (s *Supervisor) registerSupervisor(ctx context.Context) {
	data := map[string]interface{}{
		"name":          s.name,
		"status":        string(s.status),
		"queues":        s.config.Queues,
		"balance":       string(s.config.Balance),
		"min_processes": s.config.MinProcesses,
		"max_processes": s.config.MaxProcesses,
		"started_at":    time.Now().Unix(),
	}

	dataJSON, _ := json.Marshal(data)

	pipe := s.redis.Pipeline()
	pipe.SAdd(ctx, s.keys.supervisors(), s.name)
	pipe.Set(ctx, s.keys.supervisor(s.name), dataJSON, 0)
	pipe.Exec(ctx)
}

func (s *Supervisor) unregisterSupervisor(ctx context.Context) {
	pipe := s.redis.Pipeline()
	pipe.SRem(ctx, s.keys.supervisors(), s.name)
	pipe.Del(ctx, s.keys.supervisor(s.name))
	pipe.Del(ctx, s.keys.supervisorWorkers(s.name))
	pipe.Exec(ctx)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
