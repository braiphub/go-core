package gohorizon

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Queue handles Redis queue operations
type Queue struct {
	redis *redis.Client
	keys  *keyBuilder
}

// NewQueue creates a new queue instance
func NewQueue(client *redis.Client, prefix string) *Queue {
	return &Queue{
		redis: client,
		keys:  newKeyBuilder(prefix),
	}
}

// Push adds a job to the queue
func (q *Queue) Push(ctx context.Context, queueName string, payload *Payload) error {
	data, err := payload.Serialize()
	if err != nil {
		return err
	}

	pipe := q.redis.Pipeline()

	// Add queue to known queues set
	pipe.SAdd(ctx, q.keys.queues(), queueName)

	// Store job data
	pipe.Set(ctx, q.keys.job(payload.ID), data, 24*time.Hour)

	// Add to queue
	pipe.RPush(ctx, q.keys.queue(queueName), payload.ID)

	// Index by tags
	for _, tag := range payload.Tags {
		pipe.SAdd(ctx, q.keys.jobsByTag(tag), payload.ID)
	}

	_, err = pipe.Exec(ctx)
	return err
}

// Later schedules a job for delayed execution
func (q *Queue) Later(ctx context.Context, queueName string, payload *Payload, delay time.Duration) error {
	payload.AvailableAt = time.Now().Add(delay)

	data, err := payload.Serialize()
	if err != nil {
		return err
	}

	pipe := q.redis.Pipeline()

	// Add queue to known queues set
	pipe.SAdd(ctx, q.keys.queues(), queueName)

	// Store job data
	pipe.Set(ctx, q.keys.job(payload.ID), data, 24*time.Hour)

	// Add to delayed queue with score = availableAt timestamp
	pipe.ZAdd(ctx, q.keys.queueDelayed(queueName), redis.Z{
		Score:  float64(payload.AvailableAt.Unix()),
		Member: payload.ID,
	})

	// Index by tags
	for _, tag := range payload.Tags {
		pipe.SAdd(ctx, q.keys.jobsByTag(tag), payload.ID)
	}

	_, err = pipe.Exec(ctx)
	return err
}

// Pop retrieves the next job from the queue(s)
func (q *Queue) Pop(ctx context.Context, queues ...string) (*Payload, error) {
	// First, migrate delayed jobs that are now available
	for _, queueName := range queues {
		if err := q.migrateDelayedJobs(ctx, queueName); err != nil {
			return nil, err
		}
	}

	// Try to pop from each queue
	for _, queueName := range queues {
		result, err := q.redis.LPop(ctx, q.keys.queue(queueName)).Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			return nil, err
		}

		// Get job data
		data, err := q.redis.Get(ctx, q.keys.job(result)).Bytes()
		if err != nil {
			continue // Job expired or was deleted
		}

		payload, err := DeserializePayload(data)
		if err != nil {
			continue
		}

		// Mark as reserved
		now := time.Now()
		payload.ReservedAt = &now
		payload.Attempts++

		// Update job data
		updatedData, _ := payload.Serialize()
		q.redis.Set(ctx, q.keys.job(payload.ID), updatedData, 24*time.Hour)

		// Add to reserved set with timeout score
		q.redis.ZAdd(ctx, q.keys.queueReserved(queueName), redis.Z{
			Score:  float64(now.Add(payload.Timeout).Unix()),
			Member: payload.ID,
		})

		return payload, nil
	}

	return nil, ErrQueueEmpty
}

// migrateDelayedJobs moves delayed jobs that are ready to the main queue
func (q *Queue) migrateDelayedJobs(ctx context.Context, queueName string) error {
	now := float64(time.Now().Unix())

	// Get jobs that are ready (score <= now)
	jobs, err := q.redis.ZRangeByScore(ctx, q.keys.queueDelayed(queueName), &redis.ZRangeBy{
		Min: "-inf",
		Max: formatFloat(now),
	}).Result()
	if err != nil {
		return err
	}

	if len(jobs) == 0 {
		return nil
	}

	pipe := q.redis.Pipeline()
	for _, jobID := range jobs {
		pipe.ZRem(ctx, q.keys.queueDelayed(queueName), jobID)
		pipe.RPush(ctx, q.keys.queue(queueName), jobID)
	}
	_, err = pipe.Exec(ctx)
	return err
}

// Release returns a job to the queue for retry
func (q *Queue) Release(ctx context.Context, queueName string, payload *Payload, delay time.Duration) error {
	// Remove from reserved
	q.redis.ZRem(ctx, q.keys.queueReserved(queueName), payload.ID)

	if delay > 0 {
		return q.Later(ctx, queueName, payload, delay)
	}

	// Re-add to queue
	return q.redis.RPush(ctx, q.keys.queue(queueName), payload.ID).Err()
}

// Delete removes a job from the queue
func (q *Queue) Delete(ctx context.Context, queueName string, payload *Payload) error {
	pipe := q.redis.Pipeline()

	// Remove from all possible locations
	pipe.LRem(ctx, q.keys.queue(queueName), 0, payload.ID)
	pipe.ZRem(ctx, q.keys.queueDelayed(queueName), payload.ID)
	pipe.ZRem(ctx, q.keys.queueReserved(queueName), payload.ID)
	pipe.Del(ctx, q.keys.job(payload.ID))

	// Remove from tag indexes
	for _, tag := range payload.Tags {
		pipe.SRem(ctx, q.keys.jobsByTag(tag), payload.ID)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// Size returns the number of pending jobs in a queue
func (q *Queue) Size(ctx context.Context, queueName string) (int64, error) {
	return q.redis.LLen(ctx, q.keys.queue(queueName)).Result()
}

// DelayedSize returns the number of delayed jobs
func (q *Queue) DelayedSize(ctx context.Context, queueName string) (int64, error) {
	return q.redis.ZCard(ctx, q.keys.queueDelayed(queueName)).Result()
}

// ReservedSize returns the number of reserved jobs
func (q *Queue) ReservedSize(ctx context.Context, queueName string) (int64, error) {
	return q.redis.ZCard(ctx, q.keys.queueReserved(queueName)).Result()
}

// Clear removes all jobs from a queue
func (q *Queue) Clear(ctx context.Context, queueName string) error {
	pipe := q.redis.Pipeline()
	pipe.Del(ctx, q.keys.queue(queueName))
	pipe.Del(ctx, q.keys.queueDelayed(queueName))
	pipe.Del(ctx, q.keys.queueReserved(queueName))
	_, err := pipe.Exec(ctx)
	return err
}

// Queues returns all known queue names
func (q *Queue) Queues(ctx context.Context) ([]string, error) {
	return q.redis.SMembers(ctx, q.keys.queues()).Result()
}

// GetPendingJobs returns pending jobs for a queue
func (q *Queue) GetPendingJobs(ctx context.Context, queueName string, limit int64) ([]*Payload, error) {
	jobIDs, err := q.redis.LRange(ctx, q.keys.queue(queueName), 0, limit-1).Result()
	if err != nil {
		return nil, err
	}

	return q.getJobsByIDs(ctx, jobIDs)
}

// GetDelayedJobs returns delayed jobs for a queue
func (q *Queue) GetDelayedJobs(ctx context.Context, queueName string, limit int64) ([]*Payload, error) {
	jobIDs, err := q.redis.ZRange(ctx, q.keys.queueDelayed(queueName), 0, limit-1).Result()
	if err != nil {
		return nil, err
	}

	return q.getJobsByIDs(ctx, jobIDs)
}

func (q *Queue) getJobsByIDs(ctx context.Context, jobIDs []string) ([]*Payload, error) {
	if len(jobIDs) == 0 {
		return []*Payload{}, nil
	}

	pipe := q.redis.Pipeline()
	cmds := make([]*redis.StringCmd, len(jobIDs))
	for i, id := range jobIDs {
		cmds[i] = pipe.Get(ctx, q.keys.job(id))
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	payloads := make([]*Payload, 0, len(jobIDs))
	for _, cmd := range cmds {
		data, err := cmd.Bytes()
		if err != nil {
			continue
		}
		payload, err := DeserializePayload(data)
		if err != nil {
			continue
		}
		payloads = append(payloads, payload)
	}

	return payloads, nil
}

func formatFloat(f float64) string {
	return time.Unix(int64(f), 0).Format("20060102150405")
}
