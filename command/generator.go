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

// RegisterGeneratorCommands adds make:* commands to the kernel
func RegisterGeneratorCommands(k *Kernel, outputDir string) {
	_ = k.Register(NewMakeCommandCommand(outputDir))
	_ = k.Register(NewMakeJobCommand(outputDir))
}
