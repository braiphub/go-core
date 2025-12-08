package installer

// cliMainTemplate is the template for cmd/cli/main.go
const cliMainTemplate = `package main

import (
	"context"
	"log"

	"github.com/braiphub/go-core/command"
	"{{.ModuleName}}/internal/commands"
{{- if .HasConfigs }}
	"{{.ModuleName}}/configs"
{{- end }}
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	ctx := context.Background()

{{- if .HasConfigs }}
	if err := configs.LoadConfig(); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
{{- end }}

	kernel := command.NewKernel(
		command.WithCommands(commands.RegisterCommands()...),
	)

	command.RegisterGeneratorCommands(kernel, "internal/commands")

{{- if .HasConfigs }}
	command.Main(ctx, kernel,
		command.WithAppName(configs.GetAppName()),
		command.WithVersion("1.0.0"),
	)
{{- else }}
	command.Main(ctx, kernel,
		command.WithAppName("app"),
		command.WithVersion("1.0.0"),
	)
{{- end }}
}
`

// registryTemplate is the template for internal/commands/registry.go
const registryTemplate = `package commands

import "github.com/braiphub/go-core/command"

// RegisterCommands returns all commands for the application
func RegisterCommands() []command.Command {
	return []command.Command{
		// Add your commands here
		// Example:
		// &HelloCommand{},
	}
}
`
