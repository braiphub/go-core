# Schedule Package

Job scheduling package for Go applications. Provides a clean API for scheduling recurring tasks with support for intervals, daily schedules, weekly schedules, and cron expressions.

## Installation

```bash
go get github.com/braiphub/go-core/schedule
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"

    "github.com/braiphub/go-core/schedule"
)

func main() {
    ctx := context.Background()

    // Create and run scheduler
    err := schedule.Run(ctx, nil,
        schedule.NewJob(func(ctx context.Context) error {
            fmt.Println("Job executed!")
            return nil
        }, schedule.EveryMinute()),
    )

    if err != nil {
        panic(err)
    }
}
```

## Creating Jobs

### Basic Job with Interval

```go
job := schedule.NewJob(
    func(ctx context.Context) error {
        // Your job logic here
        return nil
    },
    schedule.WithInterval(5 * time.Minute),
)
```

### Job with Name (for logging)

```go
job := schedule.NewJob(
    myFunction,
    schedule.WithName("send-emails"),
    schedule.EveryHour(),
)
```

### Daily Job

```go
// Daily at midnight
job := schedule.NewJob(myFunc, schedule.Daily())

// Daily at 8:30 AM
job := schedule.NewJob(myFunc, schedule.DailyAtTime(8, 30))

// Daily at specific time with seconds
job := schedule.NewJob(myFunc, schedule.WithDailyAt(schedule.DailyAt{
    Hour:   14,
    Minute: 30,
    Second: 0,
}))
```

### Weekly Job

```go
job := schedule.NewJob(myFunc, schedule.WithWeeklyAt(schedule.WeeklyAt{
    Weekday: time.Monday,
    Hour:    9,
    Minute:  0,
    Second:  0,
}))
```

### Cron Expression

```go
// Every day at 2:30 AM
job := schedule.NewJob(myFunc, schedule.WithCron("30 2 * * *"))

// Every Monday at 9 AM
job := schedule.NewJob(myFunc, schedule.WithCron("0 9 * * 1"))
```

## Convenience Functions

### Intervals

```go
schedule.EverySecond()          // Every second
schedule.EverySeconds(30)       // Every 30 seconds
schedule.EveryMinute()          // Every minute
schedule.EveryMinutes(5)        // Every 5 minutes
schedule.EveryHour()            // Every hour
schedule.EveryHours(2)          // Every 2 hours
```

### Daily

```go
schedule.Daily()                // Daily at midnight
schedule.DailyAtTime(8, 30)     // Daily at 08:30
```

## Job Options

### Start Immediately

Run the job immediately when the scheduler starts:

```go
job := schedule.NewJob(myFunc,
    schedule.EveryHour(),
    schedule.WithStartImmediately(),
)
```

### Delay First Execution

Delay the first execution:

```go
job := schedule.NewJob(myFunc,
    schedule.EveryMinute(),
    schedule.WithDelay(10 * time.Second),
)
```

### Custom Timezone

```go
loc, _ := time.LoadLocation("America/Sao_Paulo")

job := schedule.NewJob(myFunc,
    schedule.DailyAtTime(8, 0),
    schedule.WithTimezone(loc),
)
```

## Using the Scheduler

### Method 1: Simple Run Function

```go
err := schedule.Run(ctx, logger, job1, job2, job3)
```

### Method 2: Scheduler Instance

```go
// Create scheduler with options
scheduler, err := schedule.New(
    schedule.WithLogger(logger),
    schedule.WithLocation(time.UTC),
)
if err != nil {
    panic(err)
}

// Add jobs
scheduler.AddJob(job1)
scheduler.AddJobs(job2, job3)

// Start (blocks until context is cancelled)
err = scheduler.Start(ctx)
```

### Method 3: Initialize with Jobs

```go
scheduler, err := schedule.New(
    schedule.WithLogger(logger),
    schedule.WithJobs(job1, job2, job3),
)
if err != nil {
    panic(err)
}

err = scheduler.Start(ctx)
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

    "github.com/braiphub/go-core/log"
    "github.com/braiphub/go-core/schedule"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Handle shutdown signals
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-sigs
        cancel()
    }()

    // Initialize logger
    logger := log.NewZap()

    // Define jobs
    sendEmailsJob := schedule.NewJob(
        func(ctx context.Context) error {
            fmt.Println("Sending emails...")
            return nil
        },
        schedule.WithName("send-emails"),
        schedule.DailyAtTime(8, 0),
    )

    backupJob := schedule.NewJob(
        func(ctx context.Context) error {
            fmt.Println("Running backup...")
            return nil
        },
        schedule.WithName("backup"),
        schedule.DailyAtTime(2, 0),
    )

    cleanupJob := schedule.NewJob(
        func(ctx context.Context) error {
            fmt.Println("Cleaning up...")
            return nil
        },
        schedule.WithName("cleanup"),
        schedule.EveryHour(),
        schedule.WithStartImmediately(),
    )

    healthCheckJob := schedule.NewJob(
        func(ctx context.Context) error {
            fmt.Println("Health check...")
            return nil
        },
        schedule.WithName("health-check"),
        schedule.EveryMinutes(5),
    )

    // Create and start scheduler
    scheduler, err := schedule.New(
        schedule.WithLogger(logger),
        schedule.WithJobs(
            sendEmailsJob,
            backupJob,
            cleanupJob,
            healthCheckJob,
        ),
    )
    if err != nil {
        logger.Fatal("Failed to create scheduler", err)
    }

    if err := scheduler.Start(ctx); err != nil {
        logger.Error("Scheduler error", err)
        os.Exit(1)
    }
}
```

## Migration from cron Package

If you're migrating from the `cron` package:

### Before (cron)

```go
import "github.com/braiphub/go-core/cron"

job := cron.NewJob(
    myFunc,
    cron.WithInterval(5 * time.Minute),
    cron.WithStartImmediately(),
)

err := cron.RunCronServer(ctx, logger, job)
```

### After (schedule)

```go
import "github.com/braiphub/go-core/schedule"

job := schedule.NewJob(
    myFunc,
    schedule.WithInterval(5 * time.Minute),
    schedule.WithStartImmediately(),
)

err := schedule.Run(ctx, logger, job)
```

### Mapping

| cron | schedule |
|------|----------|
| `cron.JobConfig` | `schedule.Job` |
| `cron.NewJob()` | `schedule.NewJob()` |
| `cron.WithInterval()` | `schedule.WithInterval()` |
| `cron.WithDailyAt()` | `schedule.WithDailyAt()` |
| `cron.WithDailyAtHour()` | `schedule.WithDailyAtHour()` |
| `cron.WithStartImmediately()` | `schedule.WithStartImmediately()` |
| `cron.WithDelay()` | `schedule.WithDelay()` |
| `cron.RunCronServer()` | `schedule.Run()` |

### New Features in schedule

- `schedule.WithName()` - Name jobs for better logging
- `schedule.WithWeeklyAt()` - Weekly scheduling
- `schedule.WithCron()` - Cron expression support
- `schedule.WithTimezone()` - Per-job timezone
- `schedule.New()` - Scheduler instance with more control
- Convenience functions: `EverySecond()`, `EveryMinute()`, `Daily()`, etc.

## Scheduler Options

```go
schedule.New(
    // Set logger for error reporting
    schedule.WithLogger(logger),

    // Set default timezone (default: UTC)
    schedule.WithLocation(time.UTC),

    // Add jobs during initialization
    schedule.WithJobs(job1, job2, job3),
)
```

## Job Configuration Reference

| Option | Description | Example |
|--------|-------------|---------|
| `WithName(name)` | Set job name for logging | `WithName("send-emails")` |
| `WithInterval(d)` | Run every duration | `WithInterval(5 * time.Minute)` |
| `WithDailyAt(da)` | Run daily at time | `WithDailyAt(DailyAt{Hour: 8})` |
| `WithDailyAtHour(h)` | Run daily at hour | `WithDailyAtHour(8)` |
| `WithDailyAtTime(h,m)` | Run daily at hour:minute | `WithDailyAtTime(8, 30)` |
| `WithWeeklyAt(wa)` | Run weekly at day/time | `WithWeeklyAt(WeeklyAt{...})` |
| `WithCron(expr)` | Use cron expression | `WithCron("0 8 * * *")` |
| `WithStartImmediately()` | Run on scheduler start | - |
| `WithDelay(d)` | Delay first execution | `WithDelay(10 * time.Second)` |
| `WithTimezone(loc)` | Set job timezone | `WithTimezone(loc)` |

## Convenience Functions Reference

| Function | Equivalent |
|----------|------------|
| `EverySecond()` | `WithInterval(time.Second)` |
| `EverySeconds(n)` | `WithInterval(n * time.Second)` |
| `EveryMinute()` | `WithInterval(time.Minute)` |
| `EveryMinutes(n)` | `WithInterval(n * time.Minute)` |
| `EveryHour()` | `WithInterval(time.Hour)` |
| `EveryHours(n)` | `WithInterval(n * time.Hour)` |
| `Daily()` | `WithDailyAt(DailyAt{0, 0, 0})` |
| `DailyAtTime(h, m)` | `WithDailyAt(DailyAt{h, m, 0})` |

## License

MIT
