package gohorizon

import "time"

// Config holds all Horizon configuration
type Config struct {
	// Redis connection
	Redis RedisConfig `json:"redis"`

	// Prefix for all Redis keys
	Prefix string `json:"prefix"`

	// Supervisors configuration
	Supervisors map[string]SupervisorConfig `json:"supervisors"`

	// Metrics configuration
	Metrics MetricsConfig `json:"metrics"`

	// HTTP server configuration
	HTTP HTTPConfig `json:"http"`
}

// RedisConfig for Redis connection
type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

// MetricsConfig for metrics collection
type MetricsConfig struct {
	Enabled          bool          `json:"enabled"`
	SnapshotInterval time.Duration `json:"snapshot_interval"`
	RetentionPeriod  time.Duration `json:"retention_period"`
}

// DefaultConfig returns sensible defaults
func DefaultConfig() Config {
	return Config{
		Prefix: "horizon",
		Redis: RedisConfig{
			Host: "localhost",
			Port: 6379,
			DB:   0,
		},
		Supervisors: make(map[string]SupervisorConfig),
		Metrics: MetricsConfig{
			Enabled:          true,
			SnapshotInterval: time.Minute,
			RetentionPeriod:  7 * 24 * time.Hour,
		},
		HTTP: DefaultHTTPConfig(),
	}
}
