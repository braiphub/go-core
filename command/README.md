# Command Package

Laravel-like command system for Go applications. Provides a familiar interface for creating CLI commands and scheduling tasks.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Generating Commands](#generating-commands)
- [Creating Commands Manually](#creating-commands-manually)
- [Kernel Setup](#kernel-setup)
- [Execution Modes](#execution-modes)
- [Schedule DSL](#schedule-dsl)
- [Command Interfaces](#command-interfaces)
- [Option Types](#option-types)
- [Args Helper Methods](#args-helper-methods)
- [Kernel Options](#kernel-options)
- [CLI Options](#cli-options)
- [Complete Example](#complete-example)
- [Comparison with Laravel](#comparison-with-laravel)

## Installation

```bash
go get github.com/braiphub/go-core/command
```

## Quick Start

```go
package main

import (
    "context"
    "github.com/braiphub/go-core/command"
)

func main() {
    ctx := context.Background()

    // Create kernel and register commands
    kernel := command.NewKernel()
    kernel.Register(&MyCommand{})

    // Run as CLI
    command.Main(ctx, kernel)
}
```

## Generating Commands

Generate command files automatically using `make:command` (like Laravel Artisan):

### Using go run

```bash
# Basic usage
go run github.com/braiphub/go-core/command/cmd/artisan make:command --name=SendEmails

# With custom directory
go run github.com/braiphub/go-core/command/cmd/artisan make:command --name=EnviarEmail --dir=internal/commands

# With custom signature
go run github.com/braiphub/go-core/command/cmd/artisan make:command --name=ClearCache --signature=cache:flush
```

### Installing globally

```bash
go install github.com/braiphub/go-core/command/cmd/artisan@latest

# Then use directly
artisan make:command --name=SendEmails
artisan make:command --name=BackupDatabase --dir=internal/commands
```

### Generator flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--name` | `-n` | Class name (e.g., `SendEmails`) | Required |
| `--signature` | `-s` | Command signature (e.g., `email:send`) | Auto-generated |
| `--dir` | `-d` | Output directory | `internal/commands` |

### Auto-generated signatures

| Class Name | File | Signature |
|------------|------|-----------|
| `SendEmails` | `send_emails.go` | `send:emails` |
| `EnviarEmail` | `enviar_email.go` | `enviar:email` |
| `BackupDatabase` | `backup_database.go` | `backup:database` |
| `ClearCache` | `clear_cache.go` | `clear:cache` |

### Generated file example

```bash
artisan make:command --name=EnviarEmail
```

Output:

```
✅ Command created successfully: internal/commands/enviar_email.go
   Signature: enviar:email
   Class: EnviarEmailCommand
```

## Creating Commands Manually

### Basic Command

```go
package commands

import (
    "context"
    "github.com/braiphub/go-core/command"
)

type SendEmailsCommand struct {
    mailer MailerService
}

func NewSendEmailsCommand(mailer MailerService) *SendEmailsCommand {
    return &SendEmailsCommand{mailer: mailer}
}

func (c *SendEmailsCommand) Name() string {
    return "email:send"
}

func (c *SendEmailsCommand) Description() string {
    return "Send pending emails to users"
}

func (c *SendEmailsCommand) Handle(ctx context.Context, args *command.Args) error {
    return c.mailer.SendEmails(ctx)
}
```

### Command with Options/Flags

```go
func (c *SendEmailsCommand) DefineOptions() []command.Option {
    return []command.Option{
        {
            Name:        "user",
            Shorthand:   "u",
            Description: "Target user ID",
            Type:        command.StringOption,
            Required:    true,
        },
        {
            Name:        "queue",
            Shorthand:   "q",
            Description: "Queue instead of sending immediately",
            Type:        command.BoolOption,
            Default:     false,
        },
        {
            Name:        "limit",
            Shorthand:   "l",
            Description: "Maximum emails to send",
            Type:        command.IntOption,
            Default:     100,
        },
        {
            Name:        "tags",
            Shorthand:   "t",
            Description: "Email tags to filter",
            Type:        command.StringSliceOption,
        },
    }
}

func (c *SendEmailsCommand) Handle(ctx context.Context, args *command.Args) error {
    userID := args.GetString("user")
    queue := args.GetBool("queue")
    limit := args.GetInt("limit", 100)
    tags := args.GetStringSlice("tags")

    // Use the options...
    return nil
}
```

### Command with Validation

```go
func (c *SendEmailsCommand) Validate(ctx context.Context, args *command.Args) error {
    if args.GetString("user") == "" {
        return fmt.Errorf("user ID is required")
    }

    if args.GetInt("limit") < 1 {
        return fmt.Errorf("limit must be at least 1")
    }

    return nil
}
```

### Command with Setup and Teardown

```go
func (c *SendEmailsCommand) Setup(ctx context.Context) error {
    // Initialize resources before execution
    return c.mailer.Connect(ctx)
}

func (c *SendEmailsCommand) Teardown(ctx context.Context) error {
    // Cleanup after execution
    return c.mailer.Disconnect(ctx)
}
```

## Kernel Setup

### Basic Setup

```go
kernel := command.NewKernel()
kernel.Register(&MyCommand{})
```

### With Options

```go
kernel := command.NewKernel(
    command.WithLogger(logger),
    command.WithCommands(cmd1, cmd2, cmd3),
    command.WithSchedules(schedule1, schedule2),
)
```

### Registering Multiple Commands

```go
kernel.RegisterMany(
    commands.NewSendEmailsCommand(mailer),
    commands.NewBackupDatabaseCommand(db),
    commands.NewClearCacheCommand(cache),
)
```

### Adding Schedules

```go
kernel.Schedule(
    command.NewSchedule("email:send").DailyAt(8, 0),
    command.NewSchedule("backup:database").DailyAt(2, 0),
    command.NewSchedule("cache:clear").EveryThirtyMinutes(),
)
```

## Execution Modes

### 1. CLI Mode (Single Command)

Execute a command directly and exit:

```bash
# Run a command
./myapp email:send --user=123 --queue

# List all commands
./myapp list

# Get help for a command
./myapp email:send --help

# Show version
./myapp --version
```

### 2. Scheduler Mode (Background)

Run scheduled commands continuously (like Laravel's `schedule:run`):

```bash
# Via CLI
./myapp schedule:run

# Or via environment variable
MODE=scheduler ./myapp
```

### 3. Programmatic Execution

Call commands from your code:

```go
// Simple call
err := kernel.Run(ctx, "email:send", nil)

// With arguments
args := command.NewArgs().
    SetOption("user", "123").
    SetOption("queue", true).
    SetOption("limit", 50)

err := kernel.Run(ctx, "email:send", args)

// Alternative syntax
err := kernel.Call(ctx, "email:send", args)
```

### Complete main.go Example

```go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"

    "github.com/braiphub/go-core/command"
    "github.com/braiphub/go-core/log"
    "your-project/internal/commands"
)

func main() {
    // Context with graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-sigs
        cancel()
    }()

    // Initialize dependencies
    logger := log.NewZap()
    mailer := initMailer()
    db := initDatabase()
    cache := initCache()

    // Create and configure kernel
    kernel := command.NewKernel(
        command.WithLogger(logger),
    )

    // Register commands
    kernel.RegisterMany(
        commands.NewSendEmailsCommand(mailer),
        commands.NewBackupDatabaseCommand(db),
        commands.NewClearCacheCommand(cache),
    )

    // Define schedules (like Laravel's Kernel.php)
    kernel.Schedule(
        command.NewSchedule("email:send").
            WithOption("user", "all").
            DailyAt(8, 0),

        command.NewSchedule("backup:database").
            DailyAt(2, 0).
            WithOption("compress", true),

        command.NewSchedule("cache:clear").
            EveryThirtyMinutes(),
    )

    // Determine execution mode
    if len(os.Args) > 1 && os.Args[1] == "schedule:run" {
        // Scheduler mode (background)
        logger.Info("Starting scheduler...")
        if err := kernel.StartScheduler(ctx); err != nil {
            logger.Error("Scheduler error", err)
            os.Exit(1)
        }
    } else {
        // CLI mode (single command)
        command.Main(ctx, kernel,
            command.WithAppName("myapp"),
            command.WithVersion("1.0.0"),
        )
    }
}
```

## Schedule DSL

Laravel-like fluent syntax for scheduling:

### Time Intervals

```go
// Seconds
command.NewSchedule("cmd").EverySecond()
command.NewSchedule("cmd").EverySeconds(30)

// Minutes
command.NewSchedule("cmd").EveryMinute()
command.NewSchedule("cmd").EveryMinutes(5)
command.NewSchedule("cmd").EveryFiveMinutes()
command.NewSchedule("cmd").EveryTenMinutes()
command.NewSchedule("cmd").EveryFifteenMinutes()
command.NewSchedule("cmd").EveryThirtyMinutes()

// Hours
command.NewSchedule("cmd").Hourly()
command.NewSchedule("cmd").EveryHours(2)

// Daily
command.NewSchedule("cmd").Daily()                    // Midnight
command.NewSchedule("cmd").DailyAt(8, 30)            // 08:30
command.NewSchedule("cmd").DailyAtTime(8, 30, 15)    // 08:30:15

// Custom interval
command.NewSchedule("cmd").WithInterval(45 * time.Minute)
```

### Execution Modifiers

```go
// Run immediately when scheduler starts
command.NewSchedule("cmd").EveryMinutes(5).Immediately()

// Delay first execution
command.NewSchedule("cmd").Hourly().WithDelay(10 * time.Minute)
```

### With Arguments

```go
command.NewSchedule("email:send").
    WithOption("user", "all").
    WithOption("force", true).
    WithArgument("type", "newsletter").
    DailyAt(8, 0)
```

### Chaining Multiple Schedules

```go
kernel.Schedule(
    command.NewSchedule("reports:generate").
        DailyAt(6, 0).
        WithOption("format", "pdf"),

    command.NewSchedule("cleanup:temp").
        EveryHours(4).
        Immediately(),

    command.NewSchedule("sync:external").
        EveryMinutes(15).
        WithDelay(5 * time.Minute),
)
```

## Command Interfaces

### Required Interface

```go
type Command interface {
    Name() string                                    // Command identifier (e.g., "email:send")
    Description() string                             // Help text
    Handle(ctx context.Context, args *Args) error    // Execution logic
}
```

### Optional Interfaces

```go
// Define command flags/options
type CommandWithOptions interface {
    Command
    DefineOptions() []Option
}

// Pre-execution validation
type CommandWithValidation interface {
    Command
    Validate(ctx context.Context, args *Args) error
}

// Initialization before execution
type CommandWithSetup interface {
    Command
    Setup(ctx context.Context) error
}

// Cleanup after execution (always runs, even on error)
type CommandWithTeardown interface {
    Command
    Teardown(ctx context.Context) error
}
```

### Execution Lifecycle

```
1. Setup()     → Initialize resources (if CommandWithSetup)
2. Validate()  → Validate arguments (if CommandWithValidation)
3. Handle()    → Execute command logic
4. Teardown()  → Cleanup (if CommandWithTeardown, always runs)
```

## Option Types

```go
command.StringOption      // string    → args.GetString("name")
command.BoolOption        // bool      → args.GetBool("flag")
command.IntOption         // int       → args.GetInt("count")
command.Int64Option       // int64     → args.GetInt64("id")
command.Float64Option     // float64   → args.GetFloat64("rate")
command.StringSliceOption // []string  → args.GetStringSlice("tags")
```

### Option Definition

```go
command.Option{
    Name:        "user",           // Flag name (--user)
    Shorthand:   "u",              // Short flag (-u)
    Description: "Target user ID", // Help text
    Type:        command.StringOption,
    Required:    true,             // Mark as required
    Default:     "all",            // Default value
}
```

## Args Helper Methods

### Getting Option Values

```go
// With defaults
args.GetString("name", "default")
args.GetBool("flag", false)
args.GetInt("count", 0)
args.GetInt64("id", 0)
args.GetFloat64("rate", 0.0)
args.GetStringSlice("tags", []string{})

// Without defaults (returns zero value)
args.GetString("name")
args.GetBool("flag")
args.GetInt("count")
```

### Checking Options

```go
args.Has("option")           // Check if option exists and is truthy
args.HasArgument("name")     // Check if positional argument exists
```

### Getting Positional Arguments

```go
args.GetArgument("arg0", "default")  // First positional argument
args.GetArgument("arg1", "default")  // Second positional argument
```

### Setting Values (for programmatic calls)

```go
args := command.NewArgs().
    SetOption("user", "123").
    SetOption("queue", true).
    SetOption("tags", []string{"urgent", "vip"}).
    SetArgument("file", "/path/to/file")
```

## Kernel Options

```go
command.NewKernel(
    // Set logger for error reporting
    command.WithLogger(logger),

    // Register commands during initialization
    command.WithCommands(cmd1, cmd2, cmd3),

    // Add schedules during initialization
    command.WithSchedules(schedule1, schedule2),

    // Use custom registry (advanced)
    command.WithRegistry(customRegistry),
)
```

## CLI Options

```go
command.Main(ctx, kernel,
    command.WithAppName("myapp"),    // Application name in help
    command.WithVersion("1.0.0"),    // Version shown with --version
)
```

### Alternative CLI Usage

```go
// Create CLI instance for more control
cli := command.NewCLI(kernel,
    command.WithAppName("myapp"),
    command.WithVersion("1.0.0"),
)

// Execute
err := cli.Execute(ctx)

// Execute with specific args (useful for testing)
err := cli.ExecuteWithArgs(ctx, []string{"email:send", "--user=123"})

// Access root Cobra command for customization
rootCmd := cli.RootCommand()
rootCmd.AddCommand(customCobraCommand)
```

## Complete Example

### Project Structure

```
myproject/
├── cmd/
│   └── app/
│       └── main.go
├── internal/
│   └── commands/
│       ├── send_emails.go
│       ├── backup_database.go
│       └── clear_cache.go
└── go.mod
```

### internal/commands/send_emails.go

```go
package commands

import (
    "context"
    "fmt"

    "github.com/braiphub/go-core/command"
)

type SendEmailsCommand struct {
    mailer MailerService
}

func NewSendEmailsCommand(mailer MailerService) *SendEmailsCommand {
    return &SendEmailsCommand{mailer: mailer}
}

func (c *SendEmailsCommand) Name() string {
    return "email:send"
}

func (c *SendEmailsCommand) Description() string {
    return "Send pending emails to users"
}

func (c *SendEmailsCommand) DefineOptions() []command.Option {
    return []command.Option{
        {
            Name:        "user",
            Shorthand:   "u",
            Description: "Target user ID (use 'all' for everyone)",
            Type:        command.StringOption,
            Default:     "all",
        },
        {
            Name:        "queue",
            Shorthand:   "q",
            Description: "Queue instead of sending immediately",
            Type:        command.BoolOption,
            Default:     false,
        },
        {
            Name:        "dry-run",
            Description: "Simulate without actually sending",
            Type:        command.BoolOption,
            Default:     false,
        },
    }
}

func (c *SendEmailsCommand) Validate(ctx context.Context, args *command.Args) error {
    user := args.GetString("user")
    if user == "" {
        return fmt.Errorf("user cannot be empty")
    }
    return nil
}

func (c *SendEmailsCommand) Handle(ctx context.Context, args *command.Args) error {
    userID := args.GetString("user")
    queue := args.GetBool("queue")
    dryRun := args.GetBool("dry-run")

    if dryRun {
        fmt.Printf("DRY RUN: Would send emails to user: %s (queue: %v)\n", userID, queue)
        return nil
    }

    if queue {
        return c.mailer.QueueEmails(ctx, userID)
    }
    return c.mailer.SendEmails(ctx, userID)
}
```

### cmd/app/main.go

```go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"

    "github.com/braiphub/go-core/command"
    "github.com/braiphub/go-core/log"
    "myproject/internal/commands"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Graceful shutdown
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-sigs
        cancel()
    }()

    // Dependencies
    logger := log.NewZap()
    mailer := &MockMailer{}

    // Kernel
    kernel := command.NewKernel(command.WithLogger(logger))

    // Commands
    kernel.RegisterMany(
        commands.NewSendEmailsCommand(mailer),
    )

    // Schedules
    kernel.Schedule(
        command.NewSchedule("email:send").
            WithOption("user", "all").
            DailyAt(8, 0),
    )

    // Run
    command.Main(ctx, kernel,
        command.WithAppName("myapp"),
        command.WithVersion("1.0.0"),
    )
}
```

### Usage

```bash
# Build
go build -o myapp ./cmd/app

# List commands
./myapp list

# Run command
./myapp email:send --user=123 --queue

# Run with dry-run
./myapp email:send --user=all --dry-run

# Get help
./myapp email:send --help

# Run scheduler
./myapp schedule:run
```

## Comparison with Laravel

| Laravel | go-core/command |
|---------|-----------------|
| `php artisan make:command` | `artisan make:command --name=...` |
| `$this->signature = 'email:send'` | `func (c *Cmd) Name() string` |
| `$this->description` | `func (c *Cmd) Description() string` |
| `handle()` | `Handle(ctx, args) error` |
| `$this->argument('name')` | `args.GetArgument("name")` |
| `$this->option('queue')` | `args.GetBool("queue")` |
| `Kernel::$commands` | `kernel.RegisterMany(...)` |
| `$schedule->command('cmd')->daily()` | `NewSchedule("cmd").Daily()` |
| `$schedule->command('cmd')->dailyAt('8:00')` | `NewSchedule("cmd").DailyAt(8, 0)` |
| `$schedule->command('cmd')->everyMinute()` | `NewSchedule("cmd").EveryMinute()` |
| `php artisan schedule:run` | `kernel.StartScheduler(ctx)` |
| `php artisan email:send` | `./app email:send` |
| `Artisan::call('email:send')` | `kernel.Run(ctx, "email:send", args)` |

## License

MIT
