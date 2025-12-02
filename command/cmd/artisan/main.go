package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/braiphub/go-core/command"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()

	// Create kernel with generator commands
	kernel := command.NewKernel()
	command.RegisterGeneratorCommands(kernel, "internal/commands")

	// Run CLI
	command.Main(ctx, kernel,
		command.WithAppName("artisan"),
		command.WithVersion("1.0.0"),
	)
}
