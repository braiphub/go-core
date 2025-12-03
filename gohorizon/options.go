package gohorizon

import (
	"time"

	"github.com/braiphub/go-core/log"
	"github.com/redis/go-redis/v9"
)

// Option is a functional option for configuring Horizon
type Option func(*Horizon)

// WithConfig sets the full configuration
func WithConfig(config Config) Option {
	return func(h *Horizon) {
		h.config = config
	}
}

// WithRedis sets the Redis client
func WithRedis(client *redis.Client) Option {
	return func(h *Horizon) {
		h.redis = client
	}
}

// WithLogger sets the logger
func WithLogger(logger log.LoggerI) Option {
	return func(h *Horizon) {
		h.logger = logger
	}
}

// WithSupervisor adds a supervisor configuration
func WithSupervisor(name string, config SupervisorConfig) Option {
	return func(h *Horizon) {
		if h.config.Supervisors == nil {
			h.config.Supervisors = make(map[string]SupervisorConfig)
		}
		config.Name = name
		h.config.Supervisors[name] = config
	}
}

// WithHTTP enables and configures the HTTP server
func WithHTTP(config HTTPConfig) Option {
	return func(h *Horizon) {
		h.config.HTTP = config
	}
}

// WithMetrics configures metrics collection
func WithMetrics(config MetricsConfig) Option {
	return func(h *Horizon) {
		h.config.Metrics = config
	}
}

// WithPrefix sets the Redis key prefix
func WithPrefix(prefix string) Option {
	return func(h *Horizon) {
		h.config.Prefix = prefix
	}
}

// DispatchOption configures job dispatch
type DispatchOption func(*dispatchOptions)

type dispatchOptions struct {
	queue string
	delay time.Duration
	tags  []string
}

// ToQueue sets the queue for the job
func ToQueue(queue string) DispatchOption {
	return func(o *dispatchOptions) {
		o.queue = queue
	}
}

// WithDelay delays the job execution
func WithDelay(delay time.Duration) DispatchOption {
	return func(o *dispatchOptions) {
		o.delay = delay
	}
}

// WithTags adds tags to the job
func WithTags(tags ...string) DispatchOption {
	return func(o *dispatchOptions) {
		o.tags = append(o.tags, tags...)
	}
}
