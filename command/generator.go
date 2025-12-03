package command

import (
	"context"
	"fmt"
	"os"
	pathpkg "path/filepath"
	"strings"
	"text/template"
	"unicode"
)

// MakeCommandCommand generates new command files
type MakeCommandCommand struct {
	outputDir string
}

// NewMakeCommandCommand creates the make:command generator
func NewMakeCommandCommand(outputDir string) *MakeCommandCommand {
	return &MakeCommandCommand{outputDir: outputDir}
}

func (c *MakeCommandCommand) Name() string {
	return "make:command"
}

func (c *MakeCommandCommand) Description() string {
	return "Create a new command class"
}

func (c *MakeCommandCommand) DefineOptions() []Option {
	return []Option{
		{
			Name:        "name",
			Shorthand:   "n",
			Description: "The name of the command class (e.g., SendEmails)",
			Type:        StringOption,
			Required:    true,
		},
		{
			Name:        "signature",
			Shorthand:   "s",
			Description: "The command signature (e.g., email:send)",
			Type:        StringOption,
		},
		{
			Name:        "dir",
			Shorthand:   "d",
			Description: "Output directory for the command file",
			Type:        StringOption,
		},
	}
}

func (c *MakeCommandCommand) Handle(ctx context.Context, args *Args) error {
	name := args.GetString("name")
	if name == "" {
		// Try positional argument
		name = args.GetArgument("arg0")
	}

	if name == "" {
		return fmt.Errorf("command name is required")
	}

	// Determine signature
	signature := args.GetString("signature")
	if signature == "" {
		signature = toCommandSignature(name)
	}

	// Determine output directory
	outputDir := args.GetString("dir")
	if outputDir == "" {
		outputDir = c.outputDir
	}
	if outputDir == "" {
		outputDir = "internal/commands"
	}

	// Generate the command file
	return generateCommandFile(name, signature, outputDir)
}

// generateCommandFile creates the command file from template
func generateCommandFile(name, signature, outputDir string) error {
	// Ensure directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate filename
	filename := toSnakeCase(name) + ".go"
	fullPath := pathpkg.Join(outputDir, filename)

	// Check if file exists
	if _, err := os.Stat(fullPath); err == nil {
		return fmt.Errorf("command file already exists: %s", fullPath)
	}

	// Parse and execute template
	tmpl, err := template.New("command").Parse(commandTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Create file
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Get package name from directory
	packageName := pathpkg.Base(outputDir)

	// Execute template
	data := templateData{
		PackageName:   packageName,
		ClassName:     name,
		CommandName:   signature,
		Description:   fmt.Sprintf("%s command", name),
		StructName:    name + "Command",
		ConstructorFn: "New" + name + "Command",
	}

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to write template: %w", err)
	}

	fmt.Printf("✅ Command created successfully: %s\n", fullPath)
	fmt.Printf("   Signature: %s\n", signature)
	fmt.Printf("   Class: %s\n", data.StructName)

	return nil
}

type templateData struct {
	PackageName   string
	ClassName     string
	CommandName   string
	Description   string
	StructName    string
	ConstructorFn string
}

const commandTemplate = `package {{.PackageName}}

import (
	"context"

	"github.com/braiphub/go-core/command"
)

// {{.StructName}} handles the {{.CommandName}} command
type {{.StructName}} struct {
	// Add your dependencies here
}

// {{.ConstructorFn}} creates a new {{.StructName}}
func {{.ConstructorFn}}() *{{.StructName}} {
	return &{{.StructName}}{}
}

// Name returns the command signature
func (c *{{.StructName}}) Name() string {
	return "{{.CommandName}}"
}

// Description returns the command description
func (c *{{.StructName}}) Description() string {
	return "{{.Description}}"
}

// DefineOptions returns the command options/flags
func (c *{{.StructName}}) DefineOptions() []command.Option {
	return []command.Option{
		// Example:
		// {
		// 	Name:        "user",
		// 	Shorthand:   "u",
		// 	Description: "Target user ID",
		// 	Type:        command.StringOption,
		// },
	}
}

// Handle executes the command
func (c *{{.StructName}}) Handle(ctx context.Context, args *command.Args) error {
	// TODO: Implement your command logic here

	return nil
}
`

// toCommandSignature converts "SendEmails" to "send:emails" or "EnviarEmail" to "enviar:email"
func toCommandSignature(name string) string {
	var result strings.Builder
	var words []string
	var currentWord strings.Builder

	for i, r := range name {
		if unicode.IsUpper(r) && i > 0 {
			if currentWord.Len() > 0 {
				words = append(words, strings.ToLower(currentWord.String()))
				currentWord.Reset()
			}
		}
		currentWord.WriteRune(r)
	}

	if currentWord.Len() > 0 {
		words = append(words, strings.ToLower(currentWord.String()))
	}

	// First word becomes namespace, rest becomes command
	if len(words) == 1 {
		return words[0]
	}

	// Join with colon after first word, underscore for rest
	result.WriteString(words[0])
	result.WriteString(":")
	result.WriteString(strings.Join(words[1:], "-"))

	return result.String()
}

// toSnakeCase converts "SendEmails" to "send_emails"
func toSnakeCase(name string) string {
	var result strings.Builder

	for i, r := range name {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// MakeJobCommand generates new job files for gohorizon
type MakeJobCommand struct {
	outputDir string
}

// NewMakeJobCommand creates the make:job generator
func NewMakeJobCommand(outputDir string) *MakeJobCommand {
	return &MakeJobCommand{outputDir: outputDir}
}

func (c *MakeJobCommand) Name() string {
	return "make:job"
}

func (c *MakeJobCommand) Description() string {
	return "Create a new job class for gohorizon queue processing"
}

func (c *MakeJobCommand) DefineOptions() []Option {
	return []Option{
		{
			Name:        "name",
			Shorthand:   "n",
			Description: "The name of the job class (e.g., SendEmail)",
			Type:        StringOption,
			Required:    true,
		},
		{
			Name:        "queue",
			Shorthand:   "q",
			Description: "The default queue for the job (e.g., emails)",
			Type:        StringOption,
		},
		{
			Name:        "dir",
			Shorthand:   "d",
			Description: "Output directory for the job file",
			Type:        StringOption,
		},
	}
}

func (c *MakeJobCommand) Handle(ctx context.Context, args *Args) error {
	name := args.GetString("name")
	if name == "" {
		// Try positional argument
		name = args.GetArgument("arg0")
	}

	if name == "" {
		return fmt.Errorf("job name is required")
	}

	// Determine queue
	queue := args.GetString("queue")
	if queue == "" {
		queue = "default"
	}

	// Determine output directory
	outputDir := args.GetString("dir")
	if outputDir == "" {
		outputDir = c.outputDir
	}
	if outputDir == "" {
		outputDir = "internal/jobs"
	}

	// Generate the job file
	return generateJobFile(name, queue, outputDir)
}

// generateJobFile creates the job file from template
func generateJobFile(name, queue, outputDir string) error {
	// Ensure directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate filename
	filename := toSnakeCase(name) + "_job.go"
	fullPath := pathpkg.Join(outputDir, filename)

	// Check if file exists
	if _, err := os.Stat(fullPath); err == nil {
		return fmt.Errorf("job file already exists: %s", fullPath)
	}

	// Parse and execute template
	tmpl, err := template.New("job").Parse(jobTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Create file
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Get package name from directory
	packageName := pathpkg.Base(outputDir)

	// Generate job name in kebab-case
	jobName := toJobName(name)

	// Execute template
	data := jobTemplateData{
		PackageName:   packageName,
		ClassName:     name,
		JobName:       jobName,
		Queue:         queue,
		StructName:    name + "Job",
		ConstructorFn: "New" + name + "Job",
	}

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to write template: %w", err)
	}

	fmt.Printf("✅ Job created successfully: %s\n", fullPath)
	fmt.Printf("   Job name: %s\n", jobName)
	fmt.Printf("   Queue: %s\n", queue)
	fmt.Printf("   Class: %s\n", data.StructName)

	return nil
}

type jobTemplateData struct {
	PackageName   string
	ClassName     string
	JobName       string
	Queue         string
	StructName    string
	ConstructorFn string
}

// toJobName converts "SendEmail" to "send-email"
func toJobName(name string) string {
	var result strings.Builder
	var words []string
	var currentWord strings.Builder

	for i, r := range name {
		if unicode.IsUpper(r) && i > 0 {
			if currentWord.Len() > 0 {
				words = append(words, strings.ToLower(currentWord.String()))
				currentWord.Reset()
			}
		}
		currentWord.WriteRune(r)
	}

	if currentWord.Len() > 0 {
		words = append(words, strings.ToLower(currentWord.String()))
	}

	result.WriteString(strings.Join(words, "-"))
	return result.String()
}

const jobTemplate = `package {{.PackageName}}

import (
	"context"
	"time"

	"github.com/braiphub/go-core/gohorizon"
)

// {{.StructName}} handles the {{.JobName}} job
type {{.StructName}} struct {
	// Add your job data fields here (will be serialized to JSON)
	// Example:
	// UserID string ` + "`json:\"user_id\"`" + `
	// Email  string ` + "`json:\"email\"`" + `
}

// {{.ConstructorFn}} creates a new {{.StructName}}
func {{.ConstructorFn}}() *{{.StructName}} {
	return &{{.StructName}}{}
}

// Name returns the unique job type name
func (j *{{.StructName}}) Name() string {
	return "{{.JobName}}"
}

// Handle executes the job logic
func (j *{{.StructName}}) Handle(ctx context.Context) error {
	// TODO: Implement your job logic here

	return nil
}

// Queue returns the queue this job should be dispatched to
func (j *{{.StructName}}) Queue() string {
	return "{{.Queue}}"
}

// Tags returns searchable tags for this job
func (j *{{.StructName}}) Tags() []string {
	return []string{
		// Example: j.UserID, "{{.JobName}}",
	}
}

// MaxRetries returns the maximum number of retry attempts
func (j *{{.StructName}}) MaxRetries() int {
	return 3
}

// RetryDelay returns the delay between retry attempts
func (j *{{.StructName}}) RetryDelay() time.Duration {
	return 30 * time.Second
}

// Timeout returns the maximum execution time for this job
func (j *{{.StructName}}) Timeout() time.Duration {
	return 60 * time.Second
}

// Ensure {{.StructName}} implements all gohorizon job interfaces
var (
	_ gohorizon.Job            = (*{{.StructName}})(nil)
	_ gohorizon.JobWithQueue   = (*{{.StructName}})(nil)
	_ gohorizon.JobWithTags    = (*{{.StructName}})(nil)
	_ gohorizon.JobWithRetry   = (*{{.StructName}})(nil)
	_ gohorizon.JobWithTimeout = (*{{.StructName}})(nil)
)
`

// MakeHorizonCommand generates the complete gohorizon scaffolding
type MakeHorizonCommand struct{}

// NewMakeHorizonCommand creates the make:horizon generator
func NewMakeHorizonCommand() *MakeHorizonCommand {
	return &MakeHorizonCommand{}
}

func (c *MakeHorizonCommand) Name() string {
	return "make:horizon"
}

func (c *MakeHorizonCommand) Description() string {
	return "Scaffold complete gohorizon queue system (jobs, worker, config)"
}

func (c *MakeHorizonCommand) DefineOptions() []Option {
	return []Option{
		{
			Name:        "module",
			Shorthand:   "m",
			Description: "Go module name (auto-detected from go.mod if not specified)",
			Type:        StringOption,
		},
		{
			Name:        "jobs-dir",
			Description: "Directory for job files",
			Type:        StringOption,
			Default:     "internal/jobs",
		},
		{
			Name:        "worker-dir",
			Description: "Directory for horizon worker entry point",
			Type:        StringOption,
			Default:     "cmd/horizon",
		},
		{
			Name:        "force",
			Shorthand:   "f",
			Description: "Overwrite existing files",
			Type:        BoolOption,
		},
	}
}

func (c *MakeHorizonCommand) Handle(ctx context.Context, args *Args) error {
	moduleName := args.GetString("module")
	if moduleName == "" {
		// Try to detect from go.mod
		detected, err := detectModuleName()
		if err != nil {
			return fmt.Errorf("module name is required (use --module or ensure go.mod exists)")
		}
		moduleName = detected
	}

	jobsDir := args.GetString("jobs-dir")
	if jobsDir == "" {
		jobsDir = "internal/jobs"
	}

	workerDir := args.GetString("worker-dir")
	if workerDir == "" {
		workerDir = "cmd/horizon"
	}

	force := args.GetBool("force")

	fmt.Println("🚀 Scaffolding gohorizon queue system...")
	fmt.Println()

	// 1. Create jobs directory with example job
	if err := createExampleJob(jobsDir, force); err != nil {
		return err
	}

	// 2. Create jobs registry
	if err := createJobsRegistry(jobsDir, moduleName, force); err != nil {
		return err
	}

	// 3. Create horizon worker entry point
	if err := createHorizonWorker(workerDir, moduleName, jobsDir, force); err != nil {
		return err
	}

	// 4. Update Makefile if exists
	if err := updateMakefile(workerDir); err != nil {
		// Not fatal, just warn
		fmt.Printf("⚠️  Could not update Makefile: %v\n", err)
	}

	fmt.Println()
	fmt.Println("✅ Gohorizon scaffolding complete!")
	fmt.Println()
	fmt.Println("📁 Created files:")
	fmt.Printf("   %s/example_job.go      - Example job implementation\n", jobsDir)
	fmt.Printf("   %s/registry.go         - Job registry for horizon\n", jobsDir)
	fmt.Printf("   %s/main.go             - Horizon worker entry point\n", workerDir)
	fmt.Println()
	fmt.Println("🔧 Next steps:")
	fmt.Println("   1. Configure Redis connection in cmd/horizon/main.go")
	fmt.Println("   2. Create more jobs with: artisan make:job --name=YourJob")
	fmt.Println("   3. Register new jobs in internal/jobs/registry.go")
	fmt.Println("   4. Run worker with: go run ./cmd/horizon")
	fmt.Println()
	fmt.Println("📚 Documentation: https://github.com/braiphub/go-core/tree/master/gohorizon")

	return nil
}

// detectModuleName tries to read module name from go.mod
func detectModuleName() (string, error) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimPrefix(line, "module "), nil
		}
	}

	return "", fmt.Errorf("module not found in go.mod")
}

// createExampleJob creates an example job file
func createExampleJob(jobsDir string, force bool) error {
	if err := os.MkdirAll(jobsDir, 0755); err != nil {
		return fmt.Errorf("failed to create jobs directory: %w", err)
	}

	fullPath := pathpkg.Join(jobsDir, "example_job.go")

	if !force {
		if _, err := os.Stat(fullPath); err == nil {
			fmt.Printf("⏭️  Skipping %s (already exists, use --force to overwrite)\n", fullPath)
			return nil
		}
	}

	packageName := pathpkg.Base(jobsDir)

	tmpl, err := template.New("example_job").Parse(exampleJobTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, map[string]string{"PackageName": packageName}); err != nil {
		return fmt.Errorf("failed to write template: %w", err)
	}

	fmt.Printf("✅ Created %s\n", fullPath)
	return nil
}

const exampleJobTemplate = `package {{.PackageName}}

import (
	"context"
	"fmt"
	"time"

	"github.com/braiphub/go-core/gohorizon"
)

// ExampleJob is a sample job demonstrating gohorizon features
type ExampleJob struct {
	UserID  string ` + "`json:\"user_id\"`" + `
	Message string ` + "`json:\"message\"`" + `
}

// NewExampleJob creates a new ExampleJob
func NewExampleJob() *ExampleJob {
	return &ExampleJob{}
}

// Name returns the unique job type name
func (j *ExampleJob) Name() string {
	return "example"
}

// Handle executes the job logic
func (j *ExampleJob) Handle(ctx context.Context) error {
	fmt.Printf("Processing example job for user %s: %s\n", j.UserID, j.Message)

	// Simulate some work
	select {
	case <-time.After(100 * time.Millisecond):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Queue returns the queue this job should be dispatched to
func (j *ExampleJob) Queue() string {
	return "default"
}

// Tags returns searchable tags for this job
func (j *ExampleJob) Tags() []string {
	return []string{j.UserID, "example"}
}

// MaxRetries returns the maximum number of retry attempts
func (j *ExampleJob) MaxRetries() int {
	return 3
}

// RetryDelay returns the delay between retry attempts
func (j *ExampleJob) RetryDelay() time.Duration {
	return 30 * time.Second
}

// Timeout returns the maximum execution time for this job
func (j *ExampleJob) Timeout() time.Duration {
	return 60 * time.Second
}

// Ensure ExampleJob implements all gohorizon job interfaces
var (
	_ gohorizon.Job            = (*ExampleJob)(nil)
	_ gohorizon.JobWithQueue   = (*ExampleJob)(nil)
	_ gohorizon.JobWithTags    = (*ExampleJob)(nil)
	_ gohorizon.JobWithRetry   = (*ExampleJob)(nil)
	_ gohorizon.JobWithTimeout = (*ExampleJob)(nil)
)
`

// createJobsRegistry creates the jobs registry file
func createJobsRegistry(jobsDir, moduleName string, force bool) error {
	fullPath := pathpkg.Join(jobsDir, "registry.go")

	if !force {
		if _, err := os.Stat(fullPath); err == nil {
			fmt.Printf("⏭️  Skipping %s (already exists, use --force to overwrite)\n", fullPath)
			return nil
		}
	}

	packageName := pathpkg.Base(jobsDir)

	tmpl, err := template.New("registry").Parse(registryTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, map[string]string{"PackageName": packageName}); err != nil {
		return fmt.Errorf("failed to write template: %w", err)
	}

	fmt.Printf("✅ Created %s\n", fullPath)
	return nil
}

const registryTemplate = `package {{.PackageName}}

import "github.com/braiphub/go-core/gohorizon"

// RegisterJobs registers all job types with the horizon instance
func RegisterJobs(h *gohorizon.Horizon) {
	// Register all your jobs here
	// Each job type must be registered before it can be processed

	h.RegisterJob(func() gohorizon.Job { return NewExampleJob() })

	// Add more jobs as you create them:
	// h.RegisterJob(func() gohorizon.Job { return NewSendEmailJob() })
	// h.RegisterJob(func() gohorizon.Job { return NewProcessOrderJob() })
}
`

// createHorizonWorker creates the horizon worker entry point
func createHorizonWorker(workerDir, moduleName, jobsDir string, force bool) error {
	if err := os.MkdirAll(workerDir, 0755); err != nil {
		return fmt.Errorf("failed to create worker directory: %w", err)
	}

	fullPath := pathpkg.Join(workerDir, "main.go")

	if !force {
		if _, err := os.Stat(fullPath); err == nil {
			fmt.Printf("⏭️  Skipping %s (already exists, use --force to overwrite)\n", fullPath)
			return nil
		}
	}

	tmpl, err := template.New("worker").Parse(workerTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	data := map[string]string{
		"ModuleName": moduleName,
		"JobsImport": moduleName + "/" + jobsDir,
	}

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to write template: %w", err)
	}

	fmt.Printf("✅ Created %s\n", fullPath)
	return nil
}

const workerTemplate = `package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/braiphub/go-core/gohorizon"
	"github.com/braiphub/go-core/log"
	"{{.JobsImport}}"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	// Initialize logger
	logger := log.NewZap()

	// Create Horizon instance
	horizon, err := gohorizon.New(
		gohorizon.WithLogger(logger),
		gohorizon.WithPrefix("app"), // Change to your app name

		// Redis configuration (uses localhost:6379 by default)
		// Uncomment and configure for production:
		// gohorizon.WithConfig(gohorizon.Config{
		// 	Redis: gohorizon.RedisConfig{
		// 		Host:     os.Getenv("REDIS_HOST"),
		// 		Port:     6379,
		// 		Password: os.Getenv("REDIS_PASSWORD"),
		// 		DB:       0,
		// 	},
		// }),

		// Configure supervisors (worker pools)
		gohorizon.WithSupervisor("default", gohorizon.SupervisorConfig{
			Queues:       []string{"default"},
			MinProcesses: 2,
			MaxProcesses: 10,
			Balance:      gohorizon.BalanceModeAuto,
		}),

		// Add more supervisors for different queues:
		// gohorizon.WithSupervisor("emails", gohorizon.SupervisorConfig{
		// 	Queues:       []string{"emails", "notifications"},
		// 	MinProcesses: 1,
		// 	MaxProcesses: 5,
		// 	Balance:      gohorizon.BalanceModeSimple,
		// }),

		// Enable HTTP API for monitoring (optional)
		gohorizon.WithHTTP(gohorizon.HTTPConfig{
			Enabled:  true,
			Addr:     ":8080",
			BasePath: "/horizon",
			// Enable authentication for production:
			// Auth: gohorizon.AuthConfig{
			// 	Enabled:  true,
			// 	Type:     "basic",
			// 	Username: os.Getenv("HORIZON_USER"),
			// 	Password: os.Getenv("HORIZON_PASSWORD"),
			// },
		}),
	)
	if err != nil {
		logger.Error("failed to create horizon", err)
		os.Exit(1)
	}

	// Register all jobs
	jobs.RegisterJobs(horizon)

	// Start Horizon (blocks until context is cancelled)
	logger.Info("Starting Horizon worker...")
	if err := horizon.Start(ctx); err != nil {
		logger.Error("horizon error", err)
		os.Exit(1)
	}

	logger.Info("Horizon worker stopped")
}
`

// updateMakefile adds horizon commands to existing Makefile
func updateMakefile(workerDir string) error {
	makefilePath := "Makefile"

	// Check if Makefile exists
	data, err := os.ReadFile(makefilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No Makefile, skip silently
		}
		return err
	}

	content := string(data)

	// Check if horizon commands already exist
	if strings.Contains(content, "horizon:") || strings.Contains(content, "horizon-") {
		fmt.Println("⏭️  Makefile already contains horizon commands")
		return nil
	}

	// Add horizon commands
	horizonMakefileCommands := fmt.Sprintf(`
## Horizon Queue Commands
.PHONY: horizon horizon-build

horizon: ## Run horizon worker
	go run ./%s

horizon-build: ## Build horizon worker binary
	go build -o bin/horizon ./%s
`, workerDir, workerDir)

	// Append to Makefile
	file, err := os.OpenFile(makefilePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(horizonMakefileCommands); err != nil {
		return err
	}

	fmt.Println("✅ Updated Makefile with horizon commands")
	return nil
}

// RegisterGeneratorCommands adds make:* commands to the kernel
func RegisterGeneratorCommands(k *Kernel, outputDir string) {
	_ = k.Register(NewMakeCommandCommand(outputDir))
	_ = k.Register(NewMakeJobCommand(outputDir))
	_ = k.Register(NewMakeHorizonCommand())
}
