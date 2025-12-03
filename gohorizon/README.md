# GoHorizon

Laravel Horizon-like queue monitoring and job processing for Go applications. Provides Redis-based job queues with real-time monitoring, worker supervision, auto-scaling, and HTTP API for dashboards.

## Installation

```bash
go get github.com/braiphub/go-core/gohorizon
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/braiphub/go-core/gohorizon"
    "github.com/braiphub/go-core/log"
)

// Define a job
type SendEmailJob struct {
    To      string `json:"to"`
    Subject string `json:"subject"`
    Body    string `json:"body"`
}

func (j *SendEmailJob) Name() string {
    return "send-email"
}

func (j *SendEmailJob) Handle(ctx context.Context) error {
    fmt.Printf("Sending email to %s: %s\n", j.To, j.Subject)
    return nil
}

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Handle shutdown
    go func() {
        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
        <-sigCh
        cancel()
    }()

    // Create Horizon
    logger := log.NewZap()

    horizon, err := gohorizon.New(
        gohorizon.WithLogger(logger),
        gohorizon.WithPrefix("myapp"),
        gohorizon.WithSupervisor("default", gohorizon.SupervisorConfig{
            Queues:       []string{"default", "emails"},
            MinProcesses: 2,
            MaxProcesses: 10,
            Balance:      gohorizon.BalanceModeAuto,
        }),
        gohorizon.WithHTTP(gohorizon.HTTPConfig{
            Enabled:  true,
            Addr:     ":8080",
            BasePath: "/horizon",
        }),
    )
    if err != nil {
        panic(err)
    }

    // Register job types
    horizon.RegisterJob(func() gohorizon.Job { return &SendEmailJob{} })

    // Start Horizon (blocks until context cancelled)
    if err := horizon.Start(ctx); err != nil {
        logger.Error("horizon error", err)
    }
}
```

## Dispatching Jobs

### Basic Dispatch

```go
err := horizon.Dispatch(ctx, &SendEmailJob{
    To:      "user@example.com",
    Subject: "Welcome!",
    Body:    "Welcome to our platform.",
})
```

### Delayed Dispatch

```go
err := horizon.Dispatch(ctx, &SendEmailJob{...},
    gohorizon.WithDelay(5 * time.Minute),
)
```

### Dispatch to Specific Queue

```go
err := horizon.Dispatch(ctx, &SendEmailJob{...},
    gohorizon.ToQueue("high-priority"),
)
```

### Dispatch with Tags

```go
err := horizon.Dispatch(ctx, &SendEmailJob{...},
    gohorizon.WithTags("marketing", "campaign-2024"),
)
```

## Job Interface

### Basic Job

```go
type Job interface {
    Name() string
    Handle(ctx context.Context) error
}
```

### Job with Tags

```go
type JobWithTags interface {
    Job
    Tags() []string
}

// Example
func (j *SendEmailJob) Tags() []string {
    return []string{"email", j.To}
}
```

### Job with Custom Retry

```go
type JobWithRetry interface {
    Job
    MaxRetries() int
    RetryDelay() time.Duration
}

// Example
func (j *SendEmailJob) MaxRetries() int { return 5 }
func (j *SendEmailJob) RetryDelay() time.Duration { return 30 * time.Second }
```

### Job with Custom Timeout

```go
type JobWithTimeout interface {
    Job
    Timeout() time.Duration
}

// Example
func (j *SendEmailJob) Timeout() time.Duration { return 2 * time.Minute }
```

### Job with Custom Queue

```go
type JobWithQueue interface {
    Job
    Queue() string
}

// Example
func (j *SendEmailJob) Queue() string { return "emails" }
```

## Supervisor Configuration

```go
gohorizon.SupervisorConfig{
    Name:         "default",           // Supervisor name
    Queues:       []string{"default"}, // Queues to process
    Balance:      gohorizon.BalanceModeAuto, // simple, auto, null
    MinProcesses: 1,                   // Minimum workers
    MaxProcesses: 10,                  // Maximum workers
    MaxTime:      0,                   // Max worker lifetime (0 = unlimited)
    MaxJobs:      0,                   // Max jobs per worker (0 = unlimited)
    Tries:        3,                   // Default max retries
    Timeout:      60 * time.Second,    // Default job timeout
    Sleep:        3 * time.Second,     // Sleep when no jobs
}
```

### Balance Modes

| Mode | Description |
|------|-------------|
| `BalanceModeSimple` | Static number of workers (MinProcesses) |
| `BalanceModeAuto` | Auto-scale based on queue load |
| `BalanceModeNull` | One worker per queue |

## HTTP API

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/horizon/api/stats` | Overall statistics |
| GET | `/horizon/api/queues` | All queues and metrics |
| GET | `/horizon/api/workload` | Current workload by queue |
| GET | `/horizon/api/supervisors` | All supervisors and workers |
| GET | `/horizon/api/jobs/recent` | Recently processed jobs |
| GET | `/horizon/api/jobs/failed` | Failed jobs |
| POST | `/horizon/api/jobs/retry` | Retry a failed job |
| POST | `/horizon/api/jobs/retry-all` | Retry all failed jobs |
| POST | `/horizon/api/jobs/flush` | Delete all failed jobs |
| GET | `/horizon/api/metrics/snapshots` | Historical metrics snapshots |

### Authentication

```go
gohorizon.WithHTTP(gohorizon.HTTPConfig{
    Enabled:  true,
    Addr:     ":8080",
    BasePath: "/horizon",
    Auth: gohorizon.AuthConfig{
        Enabled:  true,
        Type:     "basic", // or "token"
        Username: "admin",
        Password: "secret",
    },
}),
```

### Example API Response

```json
GET /horizon/api/stats

{
    "status": "running",
    "jobs_per_minute": 150.5,
    "total_processed": 45230,
    "total_failed": 12,
    "total_pending": 89,
    "total_workers": 5,
    "queues": [
        {
            "queue": "default",
            "total_processed": 30000,
            "total_failed": 8,
            "pending_jobs": 45,
            "reserved_jobs": 5,
            "delayed_jobs": 10,
            "jobs_per_minute": 100.0,
            "fail_rate": 0.02
        }
    ],
    "updated_at": "2024-01-15T10:30:00Z"
}
```

## Programmatic Control

### Pause/Resume Supervisors

```go
// Pause a supervisor
horizon.PauseSupervisor("default")

// Resume a supervisor
horizon.ContinueSupervisor("default")
```

### Scale Workers

```go
// Scale to 5 workers
horizon.ScaleSupervisor(ctx, "default", 5)
```

### Access Components

```go
// Get metrics
stats, _ := horizon.Metrics().GetStats(ctx)

// Get failed jobs
failed, _ := horizon.FailedJobs().All(ctx, 50)

// Retry a failed job
horizon.FailedJobs().Retry(ctx, "job-id")

// Get queue sizes
size, _ := horizon.Queue().Size(ctx, "default")
```

## Configuration

### Full Configuration Example

```go
horizon, err := gohorizon.New(
    // Redis connection
    gohorizon.WithConfig(gohorizon.Config{
        Redis: gohorizon.RedisConfig{
            Host:     "localhost",
            Port:     6379,
            Password: "",
            DB:       0,
        },
        Prefix: "horizon",
        Metrics: gohorizon.MetricsConfig{
            Enabled:          true,
            SnapshotInterval: time.Minute,
            RetentionPeriod:  7 * 24 * time.Hour,
        },
    }),

    // Or use existing Redis client
    gohorizon.WithRedis(redisClient),

    // Logger
    gohorizon.WithLogger(logger),

    // Supervisors
    gohorizon.WithSupervisor("default", gohorizon.SupervisorConfig{
        Queues:       []string{"default"},
        MinProcesses: 2,
        MaxProcesses: 10,
        Balance:      gohorizon.BalanceModeAuto,
    }),
    gohorizon.WithSupervisor("emails", gohorizon.SupervisorConfig{
        Queues:       []string{"emails", "notifications"},
        MinProcesses: 1,
        MaxProcesses: 5,
        Balance:      gohorizon.BalanceModeSimple,
    }),

    // HTTP API
    gohorizon.WithHTTP(gohorizon.HTTPConfig{
        Enabled:  true,
        Addr:     ":8080",
        BasePath: "/horizon",
    }),
)
```

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/braiphub/go-core/gohorizon"
    "github.com/braiphub/go-core/log"
)

// Jobs

type SendEmailJob struct {
    To      string `json:"to"`
    Subject string `json:"subject"`
}

func (j *SendEmailJob) Name() string { return "send-email" }
func (j *SendEmailJob) Queue() string { return "emails" }
func (j *SendEmailJob) Tags() []string { return []string{"email", j.To} }
func (j *SendEmailJob) MaxRetries() int { return 3 }
func (j *SendEmailJob) RetryDelay() time.Duration { return 30 * time.Second }

func (j *SendEmailJob) Handle(ctx context.Context) error {
    fmt.Printf("Sending email to %s\n", j.To)
    return nil
}

type ProcessOrderJob struct {
    OrderID string `json:"order_id"`
}

func (j *ProcessOrderJob) Name() string { return "process-order" }
func (j *ProcessOrderJob) Timeout() time.Duration { return 5 * time.Minute }

func (j *ProcessOrderJob) Handle(ctx context.Context) error {
    fmt.Printf("Processing order %s\n", j.OrderID)
    return nil
}

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go func() {
        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
        <-sigCh
        cancel()
    }()

    logger := log.NewZap()

    horizon, _ := gohorizon.New(
        gohorizon.WithLogger(logger),
        gohorizon.WithSupervisor("default", gohorizon.SupervisorConfig{
            Queues:       []string{"default"},
            MinProcesses: 2,
            MaxProcesses: 10,
            Balance:      gohorizon.BalanceModeAuto,
        }),
        gohorizon.WithSupervisor("emails", gohorizon.SupervisorConfig{
            Queues:       []string{"emails"},
            MinProcesses: 1,
            MaxProcesses: 3,
            Balance:      gohorizon.BalanceModeSimple,
        }),
        gohorizon.WithHTTP(gohorizon.HTTPConfig{
            Enabled:  true,
            Addr:     ":8080",
            BasePath: "/horizon",
        }),
    )

    // Register jobs
    horizon.RegisterJob(func() gohorizon.Job { return &SendEmailJob{} })
    horizon.RegisterJob(func() gohorizon.Job { return &ProcessOrderJob{} })

    // Dispatch some jobs
    go func() {
        time.Sleep(2 * time.Second)

        horizon.Dispatch(ctx, &SendEmailJob{To: "user@example.com", Subject: "Welcome!"})
        horizon.Dispatch(ctx, &ProcessOrderJob{OrderID: "ORD-001"})
        horizon.Dispatch(ctx, &SendEmailJob{To: "admin@example.com", Subject: "Alert"},
            gohorizon.WithDelay(1 * time.Minute),
        )
    }()

    // Start (blocks)
    horizon.Start(ctx)
}
```

## Comparison with Laravel Horizon

| Feature | Laravel Horizon | GoHorizon |
|---------|----------------|-----------|
| Queue Backend | Redis | Redis |
| Dashboard | Vue.js SPA | HTTP API (JSON) |
| Supervisors | ✅ | ✅ |
| Auto-scaling | ✅ | ✅ |
| Job Metrics | ✅ | ✅ |
| Failed Jobs | ✅ | ✅ |
| Job Tags | ✅ | ✅ |
| Job Retries | ✅ | ✅ |
| Notifications | ✅ | 🚧 (planned) |
| Batches | ✅ | 🚧 (planned) |

## License

MIT
