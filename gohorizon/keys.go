package gohorizon

import "fmt"

// Redis key builder for consistent key naming
type keyBuilder struct {
	prefix string
}

func newKeyBuilder(prefix string) *keyBuilder {
	return &keyBuilder{prefix: prefix}
}

// Queue keys
func (k *keyBuilder) queues() string {
	return fmt.Sprintf("%s:queues", k.prefix)
}

func (k *keyBuilder) queue(name string) string {
	return fmt.Sprintf("%s:queue:%s", k.prefix, name)
}

func (k *keyBuilder) queueDelayed(name string) string {
	return fmt.Sprintf("%s:queue:%s:delayed", k.prefix, name)
}

func (k *keyBuilder) queueReserved(name string) string {
	return fmt.Sprintf("%s:queue:%s:reserved", k.prefix, name)
}

func (k *keyBuilder) queueNotify(name string) string {
	return fmt.Sprintf("%s:queue:%s:notify", k.prefix, name)
}

// Job keys
func (k *keyBuilder) job(id string) string {
	return fmt.Sprintf("%s:job:%s", k.prefix, id)
}

// Failed jobs
func (k *keyBuilder) failedJobs() string {
	return fmt.Sprintf("%s:failed_jobs", k.prefix)
}

func (k *keyBuilder) failedJob(id string) string {
	return fmt.Sprintf("%s:failed_job:%s", k.prefix, id)
}

// Recent jobs
func (k *keyBuilder) recentJobs() string {
	return fmt.Sprintf("%s:recent_jobs", k.prefix)
}

// Metrics keys
func (k *keyBuilder) metricsSnapshots() string {
	return fmt.Sprintf("%s:metrics:snapshots", k.prefix)
}

func (k *keyBuilder) metricsQueue(queue string) string {
	return fmt.Sprintf("%s:metrics:queue:%s", k.prefix, queue)
}

func (k *keyBuilder) metricsJob(name string) string {
	return fmt.Sprintf("%s:metrics:job:%s", k.prefix, name)
}

func (k *keyBuilder) metricsThroughput() string {
	return fmt.Sprintf("%s:metrics:throughput", k.prefix)
}

func (k *keyBuilder) metricsJobsThroughput(queue string) string {
	return fmt.Sprintf("%s:metrics:queue:%s:throughput", k.prefix, queue)
}

func (k *keyBuilder) metricsFailedThroughput(queue string) string {
	return fmt.Sprintf("%s:metrics:queue:%s:failed_throughput", k.prefix, queue)
}

// Worker/Supervisor keys
func (k *keyBuilder) masters() string {
	return fmt.Sprintf("%s:masters", k.prefix)
}

func (k *keyBuilder) master(id string) string {
	return fmt.Sprintf("%s:master:%s", k.prefix, id)
}

func (k *keyBuilder) supervisors() string {
	return fmt.Sprintf("%s:supervisors", k.prefix)
}

func (k *keyBuilder) supervisor(name string) string {
	return fmt.Sprintf("%s:supervisor:%s", k.prefix, name)
}

func (k *keyBuilder) supervisorWorkers(name string) string {
	return fmt.Sprintf("%s:supervisor:%s:workers", k.prefix, name)
}

func (k *keyBuilder) worker(id string) string {
	return fmt.Sprintf("%s:worker:%s", k.prefix, id)
}

// Tags
func (k *keyBuilder) monitoredTags() string {
	return fmt.Sprintf("%s:monitored_tags", k.prefix)
}

func (k *keyBuilder) jobsByTag(tag string) string {
	return fmt.Sprintf("%s:tag:%s:jobs", k.prefix, tag)
}

// Locks
func (k *keyBuilder) lock(name string) string {
	return fmt.Sprintf("%s:lock:%s", k.prefix, name)
}
