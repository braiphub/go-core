package gohorizon

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// QueueMetrics contains metrics for a specific queue
type QueueMetrics struct {
	Queue          string        `json:"queue"`
	TotalProcessed int64         `json:"total_processed"`
	TotalFailed    int64         `json:"total_failed"`
	PendingJobs    int64         `json:"pending_jobs"`
	ReservedJobs   int64         `json:"reserved_jobs"`
	DelayedJobs    int64         `json:"delayed_jobs"`
	AvgRuntime     time.Duration `json:"avg_runtime_ns"`
	MaxRuntime     time.Duration `json:"max_runtime_ns"`
	JobsPerMinute  float64       `json:"jobs_per_minute"`
	FailRate       float64       `json:"fail_rate"`
	WaitTime       time.Duration `json:"wait_time_ns"`
	Throughput     *Throughput   `json:"throughput"`
}

// Throughput tracks jobs processed over time
type Throughput struct {
	Minute int64 `json:"minute"`
	Hour   int64 `json:"hour"`
	Day    int64 `json:"day"`
}

// JobMetrics contains metrics for a specific job type
type JobMetrics struct {
	JobName    string        `json:"job_name"`
	TotalRuns  int64         `json:"total_runs"`
	TotalFailed int64        `json:"total_failed"`
	AvgRuntime time.Duration `json:"avg_runtime_ns"`
	MaxRuntime time.Duration `json:"max_runtime_ns"`
	MinRuntime time.Duration `json:"min_runtime_ns"`
	LastRunAt  time.Time     `json:"last_run_at"`
}

// WorkerMetrics contains worker status
type WorkerMetrics struct {
	SupervisorName string       `json:"supervisor"`
	WorkerID       string       `json:"worker_id"`
	Status         WorkerStatus `json:"status"`
	JobsProcessed  int64        `json:"jobs_processed"`
	CurrentJob     string       `json:"current_job,omitempty"`
	StartedAt      time.Time    `json:"started_at"`
}

// Snapshot represents a point-in-time metrics snapshot
type Snapshot struct {
	Timestamp     time.Time       `json:"timestamp"`
	TotalPending  int64           `json:"total_pending"`
	TotalFailed   int64           `json:"total_failed"`
	JobsPerMinute float64         `json:"jobs_per_minute"`
	Queues        []*QueueMetrics `json:"queues"`
}

// MetricsCollector gathers and stores queue metrics
type MetricsCollector struct {
	redis  *redis.Client
	keys   *keyBuilder
	queue  *Queue
	failed *FailedJobStore
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(
	redisClient *redis.Client,
	prefix string,
	queue *Queue,
	failed *FailedJobStore,
) *MetricsCollector {
	return &MetricsCollector{
		redis:  redisClient,
		keys:   newKeyBuilder(prefix),
		queue:  queue,
		failed: failed,
	}
}

// RecordJobProcessed records a successful job
func (m *MetricsCollector) RecordJobProcessed(ctx context.Context, queueName string, payload *Payload, runtime time.Duration) {
	now := time.Now()
	minute := now.Truncate(time.Minute).Unix()

	pipe := m.redis.Pipeline()

	// Increment queue counters
	pipe.HIncrBy(ctx, m.keys.metricsQueue(queueName), "total_processed", 1)

	// Track runtime (store max)
	pipe.HSet(ctx, m.keys.metricsQueue(queueName), "last_runtime_ns", runtime.Nanoseconds())

	// Track throughput by minute
	pipe.ZIncrBy(ctx, m.keys.metricsJobsThroughput(queueName), 1, strconv.FormatInt(minute, 10))

	// Track job type metrics
	pipe.HIncrBy(ctx, m.keys.metricsJob(payload.Name), "total_runs", 1)
	pipe.HSet(ctx, m.keys.metricsJob(payload.Name), "last_run_at", now.Unix())
	pipe.HSet(ctx, m.keys.metricsJob(payload.Name), "last_runtime_ns", runtime.Nanoseconds())

	// Expire old throughput data (keep last hour)
	pipe.ZRemRangeByScore(ctx, m.keys.metricsJobsThroughput(queueName), "-inf", strconv.FormatInt(minute-3600, 10))

	pipe.Exec(ctx)
}

// RecordJobFailed records a failed job
func (m *MetricsCollector) RecordJobFailed(ctx context.Context, queueName string, payload *Payload, err error) {
	now := time.Now()
	minute := now.Truncate(time.Minute).Unix()

	pipe := m.redis.Pipeline()

	// Increment failure counters
	pipe.HIncrBy(ctx, m.keys.metricsQueue(queueName), "total_failed", 1)

	// Track failed throughput by minute
	pipe.ZIncrBy(ctx, m.keys.metricsFailedThroughput(queueName), 1, strconv.FormatInt(minute, 10))

	// Track job type metrics
	pipe.HIncrBy(ctx, m.keys.metricsJob(payload.Name), "total_failed", 1)

	// Expire old throughput data
	pipe.ZRemRangeByScore(ctx, m.keys.metricsFailedThroughput(queueName), "-inf", strconv.FormatInt(minute-3600, 10))

	pipe.Exec(ctx)
}

// GetQueueMetrics returns metrics for a queue
func (m *MetricsCollector) GetQueueMetrics(ctx context.Context, queueName string) (*QueueMetrics, error) {
	// Get stored metrics
	data, err := m.redis.HGetAll(ctx, m.keys.metricsQueue(queueName)).Result()
	if err != nil {
		return nil, err
	}

	metrics := &QueueMetrics{
		Queue:      queueName,
		Throughput: &Throughput{},
	}

	// Parse stored metrics
	if v, ok := data["total_processed"]; ok {
		metrics.TotalProcessed, _ = strconv.ParseInt(v, 10, 64)
	}
	if v, ok := data["total_failed"]; ok {
		metrics.TotalFailed, _ = strconv.ParseInt(v, 10, 64)
	}
	if v, ok := data["last_runtime_ns"]; ok {
		ns, _ := strconv.ParseInt(v, 10, 64)
		metrics.AvgRuntime = time.Duration(ns)
	}

	// Get current queue sizes
	metrics.PendingJobs, _ = m.queue.Size(ctx, queueName)
	metrics.DelayedJobs, _ = m.queue.DelayedSize(ctx, queueName)
	metrics.ReservedJobs, _ = m.queue.ReservedSize(ctx, queueName)

	// Calculate throughput
	now := time.Now()
	minuteAgo := now.Add(-time.Minute).Truncate(time.Minute).Unix()
	hourAgo := now.Add(-time.Hour).Truncate(time.Minute).Unix()

	// Jobs in last minute
	minuteJobs, _ := m.redis.ZRangeByScore(ctx, m.keys.metricsJobsThroughput(queueName), &redis.ZRangeBy{
		Min: strconv.FormatInt(minuteAgo, 10),
		Max: "+inf",
	}).Result()

	var minuteTotal int64
	for _, score := range minuteJobs {
		val, _ := m.redis.ZScore(ctx, m.keys.metricsJobsThroughput(queueName), score).Result()
		minuteTotal += int64(val)
	}
	metrics.Throughput.Minute = minuteTotal
	metrics.JobsPerMinute = float64(minuteTotal)

	// Jobs in last hour
	hourJobs, _ := m.redis.ZRangeByScore(ctx, m.keys.metricsJobsThroughput(queueName), &redis.ZRangeBy{
		Min: strconv.FormatInt(hourAgo, 10),
		Max: "+inf",
	}).Result()

	var hourTotal int64
	for _, score := range hourJobs {
		val, _ := m.redis.ZScore(ctx, m.keys.metricsJobsThroughput(queueName), score).Result()
		hourTotal += int64(val)
	}
	metrics.Throughput.Hour = hourTotal

	// Calculate fail rate
	if metrics.TotalProcessed > 0 {
		metrics.FailRate = float64(metrics.TotalFailed) / float64(metrics.TotalProcessed+metrics.TotalFailed) * 100
	}

	return metrics, nil
}

// GetAllQueuesMetrics returns metrics for all queues
func (m *MetricsCollector) GetAllQueuesMetrics(ctx context.Context) ([]*QueueMetrics, error) {
	queues, err := m.queue.Queues(ctx)
	if err != nil {
		return nil, err
	}

	metrics := make([]*QueueMetrics, 0, len(queues))
	for _, queueName := range queues {
		qm, err := m.GetQueueMetrics(ctx, queueName)
		if err != nil {
			continue
		}
		metrics = append(metrics, qm)
	}

	return metrics, nil
}

// GetJobMetrics returns metrics for a job type
func (m *MetricsCollector) GetJobMetrics(ctx context.Context, jobName string) (*JobMetrics, error) {
	data, err := m.redis.HGetAll(ctx, m.keys.metricsJob(jobName)).Result()
	if err != nil {
		return nil, err
	}

	metrics := &JobMetrics{
		JobName: jobName,
	}

	if v, ok := data["total_runs"]; ok {
		metrics.TotalRuns, _ = strconv.ParseInt(v, 10, 64)
	}
	if v, ok := data["total_failed"]; ok {
		metrics.TotalFailed, _ = strconv.ParseInt(v, 10, 64)
	}
	if v, ok := data["last_runtime_ns"]; ok {
		ns, _ := strconv.ParseInt(v, 10, 64)
		metrics.AvgRuntime = time.Duration(ns)
	}
	if v, ok := data["last_run_at"]; ok {
		ts, _ := strconv.ParseInt(v, 10, 64)
		metrics.LastRunAt = time.Unix(ts, 0)
	}

	return metrics, nil
}

// TakeSnapshot creates a point-in-time snapshot
func (m *MetricsCollector) TakeSnapshot(ctx context.Context) error {
	queuesMetrics, err := m.GetAllQueuesMetrics(ctx)
	if err != nil {
		return err
	}

	var totalPending int64
	var totalFailed int64
	var totalJobsPerMinute float64

	for _, qm := range queuesMetrics {
		totalPending += qm.PendingJobs
		totalFailed += qm.TotalFailed
		totalJobsPerMinute += qm.JobsPerMinute
	}

	snapshot := &Snapshot{
		Timestamp:     time.Now(),
		TotalPending:  totalPending,
		TotalFailed:   totalFailed,
		JobsPerMinute: totalJobsPerMinute,
		Queues:        queuesMetrics,
	}

	data, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}

	// Store snapshot with timestamp as score
	return m.redis.ZAdd(ctx, m.keys.metricsSnapshots(), redis.Z{
		Score:  float64(snapshot.Timestamp.Unix()),
		Member: data,
	}).Err()
}

// GetSnapshots returns historical snapshots
func (m *MetricsCollector) GetSnapshots(ctx context.Context, from, to time.Time, limit int64) ([]*Snapshot, error) {
	results, err := m.redis.ZRangeByScore(ctx, m.keys.metricsSnapshots(), &redis.ZRangeBy{
		Min:   strconv.FormatInt(from.Unix(), 10),
		Max:   strconv.FormatInt(to.Unix(), 10),
		Count: limit,
	}).Result()
	if err != nil {
		return nil, err
	}

	snapshots := make([]*Snapshot, 0, len(results))
	for _, data := range results {
		var snapshot Snapshot
		if err := json.Unmarshal([]byte(data), &snapshot); err != nil {
			continue
		}
		snapshots = append(snapshots, &snapshot)
	}

	return snapshots, nil
}

// GetRecentJobs returns recently processed jobs
func (m *MetricsCollector) GetRecentJobs(ctx context.Context, limit int64) ([]*RecentJob, error) {
	results, err := m.redis.LRange(ctx, m.keys.recentJobs(), 0, limit-1).Result()
	if err != nil {
		return nil, err
	}

	jobs := make([]*RecentJob, 0, len(results))
	for _, data := range results {
		var job RecentJob
		if err := json.Unmarshal([]byte(data), &job); err != nil {
			continue
		}
		jobs = append(jobs, &job)
	}

	return jobs, nil
}

// TrimSnapshots removes old snapshots
func (m *MetricsCollector) TrimSnapshots(ctx context.Context, retention time.Duration) error {
	cutoff := time.Now().Add(-retention).Unix()
	return m.redis.ZRemRangeByScore(ctx, m.keys.metricsSnapshots(), "-inf", strconv.FormatInt(cutoff, 10)).Err()
}

// GetStats returns overall statistics
func (m *MetricsCollector) GetStats(ctx context.Context) (*StatsResponse, error) {
	queuesMetrics, err := m.GetAllQueuesMetrics(ctx)
	if err != nil {
		return nil, err
	}

	failedCount, _ := m.failed.Count(ctx)

	var totalProcessed, totalPending int64
	var totalJobsPerMinute float64

	for _, qm := range queuesMetrics {
		totalProcessed += qm.TotalProcessed
		totalPending += qm.PendingJobs + qm.DelayedJobs + qm.ReservedJobs
		totalJobsPerMinute += qm.JobsPerMinute
	}

	return &StatsResponse{
		Status:         "running",
		JobsPerMinute:  totalJobsPerMinute,
		TotalProcessed: totalProcessed,
		TotalFailed:    failedCount,
		TotalPending:   totalPending,
		Queues:         queuesMetrics,
		UpdatedAt:      time.Now(),
	}, nil
}
