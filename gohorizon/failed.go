package gohorizon

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// FailedJobStore manages failed jobs
type FailedJobStore struct {
	redis *redis.Client
	keys  *keyBuilder
	queue *Queue
}

// NewFailedJobStore creates a new failed job store
func NewFailedJobStore(client *redis.Client, prefix string, queue *Queue) *FailedJobStore {
	return &FailedJobStore{
		redis: client,
		keys:  newKeyBuilder(prefix),
		queue: queue,
	}
}

// Store saves a failed job
func (s *FailedJobStore) Store(ctx context.Context, payload *Payload, exception string) error {
	failedJob := &FailedJob{
		ID:        payload.ID,
		Queue:     payload.Queue,
		Payload:   payload,
		Exception: exception,
		FailedAt:  time.Now(),
	}

	data, err := json.Marshal(failedJob)
	if err != nil {
		return err
	}

	pipe := s.redis.Pipeline()

	// Store failed job data
	pipe.Set(ctx, s.keys.failedJob(payload.ID), data, 7*24*time.Hour)

	// Add to failed jobs list (sorted by time)
	pipe.ZAdd(ctx, s.keys.failedJobs(), redis.Z{
		Score:  float64(failedJob.FailedAt.Unix()),
		Member: payload.ID,
	})

	// Remove from reserved queue
	pipe.ZRem(ctx, s.keys.queueReserved(payload.Queue), payload.ID)

	// Delete original job data
	pipe.Del(ctx, s.keys.job(payload.ID))

	_, err = pipe.Exec(ctx)
	return err
}

// All retrieves all failed jobs
func (s *FailedJobStore) All(ctx context.Context, limit int64) ([]*FailedJob, error) {
	// Get job IDs sorted by most recent first
	jobIDs, err := s.redis.ZRevRange(ctx, s.keys.failedJobs(), 0, limit-1).Result()
	if err != nil {
		return nil, err
	}

	return s.getFailedJobsByIDs(ctx, jobIDs)
}

// Find retrieves a failed job by ID
func (s *FailedJobStore) Find(ctx context.Context, id string) (*FailedJob, error) {
	data, err := s.redis.Get(ctx, s.keys.failedJob(id)).Bytes()
	if err == redis.Nil {
		return nil, ErrFailedJobNotFound
	}
	if err != nil {
		return nil, err
	}

	var failedJob FailedJob
	if err := json.Unmarshal(data, &failedJob); err != nil {
		return nil, err
	}

	return &failedJob, nil
}

// Retry moves a failed job back to its queue
func (s *FailedJobStore) Retry(ctx context.Context, id string) error {
	failedJob, err := s.Find(ctx, id)
	if err != nil {
		return err
	}

	// Reset payload for retry
	failedJob.Payload.Attempts = 0
	failedJob.Payload.ReservedAt = nil

	// Push back to queue
	if err := s.queue.Push(ctx, failedJob.Queue, failedJob.Payload); err != nil {
		return err
	}

	// Remove from failed jobs
	return s.Forget(ctx, id)
}

// RetryAll retries all failed jobs
func (s *FailedJobStore) RetryAll(ctx context.Context) (int, error) {
	jobIDs, err := s.redis.ZRange(ctx, s.keys.failedJobs(), 0, -1).Result()
	if err != nil {
		return 0, err
	}

	count := 0
	for _, id := range jobIDs {
		if err := s.Retry(ctx, id); err == nil {
			count++
		}
	}

	return count, nil
}

// Forget removes a failed job without retrying
func (s *FailedJobStore) Forget(ctx context.Context, id string) error {
	pipe := s.redis.Pipeline()
	pipe.ZRem(ctx, s.keys.failedJobs(), id)
	pipe.Del(ctx, s.keys.failedJob(id))
	_, err := pipe.Exec(ctx)
	return err
}

// Flush removes all failed jobs
func (s *FailedJobStore) Flush(ctx context.Context) error {
	// Get all failed job IDs
	jobIDs, err := s.redis.ZRange(ctx, s.keys.failedJobs(), 0, -1).Result()
	if err != nil {
		return err
	}

	if len(jobIDs) == 0 {
		return nil
	}

	pipe := s.redis.Pipeline()

	// Delete all failed job data
	for _, id := range jobIDs {
		pipe.Del(ctx, s.keys.failedJob(id))
	}

	// Clear the failed jobs set
	pipe.Del(ctx, s.keys.failedJobs())

	_, err = pipe.Exec(ctx)
	return err
}

// Count returns the number of failed jobs
func (s *FailedJobStore) Count(ctx context.Context) (int64, error) {
	return s.redis.ZCard(ctx, s.keys.failedJobs()).Result()
}

// ByQueue returns failed jobs for a specific queue
func (s *FailedJobStore) ByQueue(ctx context.Context, queueName string, limit int64) ([]*FailedJob, error) {
	allJobs, err := s.All(ctx, -1)
	if err != nil {
		return nil, err
	}

	filtered := make([]*FailedJob, 0)
	for _, job := range allJobs {
		if job.Queue == queueName {
			filtered = append(filtered, job)
			if limit > 0 && int64(len(filtered)) >= limit {
				break
			}
		}
	}

	return filtered, nil
}

func (s *FailedJobStore) getFailedJobsByIDs(ctx context.Context, jobIDs []string) ([]*FailedJob, error) {
	if len(jobIDs) == 0 {
		return []*FailedJob{}, nil
	}

	pipe := s.redis.Pipeline()
	cmds := make([]*redis.StringCmd, len(jobIDs))
	for i, id := range jobIDs {
		cmds[i] = pipe.Get(ctx, s.keys.failedJob(id))
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	jobs := make([]*FailedJob, 0, len(jobIDs))
	for _, cmd := range cmds {
		data, err := cmd.Bytes()
		if err != nil {
			continue
		}
		var job FailedJob
		if err := json.Unmarshal(data, &job); err != nil {
			continue
		}
		jobs = append(jobs, &job)
	}

	return jobs, nil
}
