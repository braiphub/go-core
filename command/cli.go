package command

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// CLI provides command-line interface functionality using Cobra
type CLI struct {
	kernel  *Kernel
	rootCmd *cobra.Command
	appName string
	version string
}

// CLIOption is a functional option for configuring the CLI
type CLIOption func(*CLI)

// WithAppName sets the application name for the CLI
func WithAppName(name string) CLIOption {
	return func(c *CLI) {
		c.appName = name
	}
}

// WithVersion sets the version for the CLI
func WithVersion(version string) CLIOption {
	return func(c *CLI) {
		c.version = version
	}
}

// NewCLI creates a new CLI instance for the kernel
func NewCLI(kernel *Kernel, opts ...CLIOption) *CLI {
	c := &CLI{
		kernel:  kernel,
		appName: "app",
		version: "1.0.0",
	}

	for _, opt := range opts {
		opt(c)
	}

	c.buildRootCommand()
	c.buildCommands()

	return c
}

// buildRootCommand creates the root cobra command
func (c *CLI) buildRootCommand() {
	c.rootCmd = &cobra.Command{
		Use:     c.appName,
		Short:   fmt.Sprintf("%s CLI", c.appName),
		Version: c.version,
	}

	// Add list command
	c.rootCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all available commands",
		Run: func(cmd *cobra.Command, args []string) {
			c.printCommandList()
		},
	})

	// Add schedule:run command for running scheduler
	c.rootCmd.AddCommand(&cobra.Command{
		Use:   "schedule:run",
		Short: "Run the scheduler",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.kernel.StartScheduler(cmd.Context())
		},
	})
}

// buildCommands converts all registered commands to cobra commands
func (c *CLI) buildCommands() {
	for _, cmd := range c.kernel.Commands() {
		cobraCmd := c.commandToCobra(cmd)
		c.rootCmd.AddCommand(cobraCmd)
	}
}

// commandToCobra converts a Command to a cobra.Command
func (c *CLI) commandToCobra(cmd Command) *cobra.Command {
	cobraCmd := &cobra.Command{
		Use:   cmd.Name(),
		Short: cmd.Description(),
		RunE: func(cobraCmd *cobra.Command, positionalArgs []string) error {
			args := c.parseCobraFlags(cobraCmd, cmd, positionalArgs)
			return c.kernel.Run(cobraCmd.Context(), cmd.Name(), args)
		},
	}

	// Add flags if command defines options
	if cmdWithOpts, ok := cmd.(CommandWithOptions); ok {
		for _, opt := range cmdWithOpts.DefineOptions() {
			c.addFlag(cobraCmd, opt)
		}
	}

	return cobraCmd
}

// addFlag adds a flag to a cobra command based on Option definition
func (c *CLI) addFlag(cobraCmd *cobra.Command, opt Option) {
	switch opt.Type {
	case StringOption:
		defaultVal := ""
		if opt.Default != nil {
			defaultVal = opt.Default.(string)
		}
		if opt.Shorthand != "" {
			cobraCmd.Flags().StringP(opt.Name, opt.Shorthand, defaultVal, opt.Description)
		} else {
			cobraCmd.Flags().String(opt.Name, defaultVal, opt.Description)
		}

	case BoolOption:
		defaultVal := false
		if opt.Default != nil {
			defaultVal = opt.Default.(bool)
		}
		if opt.Shorthand != "" {
			cobraCmd.Flags().BoolP(opt.Name, opt.Shorthand, defaultVal, opt.Description)
		} else {
			cobraCmd.Flags().Bool(opt.Name, defaultVal, opt.Description)
		}

	case IntOption:
		defaultVal := 0
		if opt.Default != nil {
			defaultVal = opt.Default.(int)
		}
		if opt.Shorthand != "" {
			cobraCmd.Flags().IntP(opt.Name, opt.Shorthand, defaultVal, opt.Description)
		} else {
			cobraCmd.Flags().Int(opt.Name, defaultVal, opt.Description)
		}

	case Int64Option:
		defaultVal := int64(0)
		if opt.Default != nil {
			defaultVal = opt.Default.(int64)
		}
		if opt.Shorthand != "" {
			cobraCmd.Flags().Int64P(opt.Name, opt.Shorthand, defaultVal, opt.Description)
		} else {
			cobraCmd.Flags().Int64(opt.Name, defaultVal, opt.Description)
		}

	case Float64Option:
		defaultVal := float64(0)
		if opt.Default != nil {
			defaultVal = opt.Default.(float64)
		}
		if opt.Shorthand != "" {
			cobraCmd.Flags().Float64P(opt.Name, opt.Shorthand, defaultVal, opt.Description)
		} else {
			cobraCmd.Flags().Float64(opt.Name, defaultVal, opt.Description)
		}

	case StringSliceOption:
		var defaultVal []string
		if opt.Default != nil {
			defaultVal = opt.Default.([]string)
		}
		if opt.Shorthand != "" {
			cobraCmd.Flags().StringSliceP(opt.Name, opt.Shorthand, defaultVal, opt.Description)
		} else {
			cobraCmd.Flags().StringSlice(opt.Name, defaultVal, opt.Description)
		}
	}

	if opt.Required {
		_ = cobraCmd.MarkFlagRequired(opt.Name)
	}
}

// parseCobraFlags extracts flag values from cobra command into Args
func (c *CLI) parseCobraFlags(cobraCmd *cobra.Command, cmd Command, positionalArgs []string) *Args {
	args := NewArgs()

	// Set positional arguments
	for i, arg := range positionalArgs {
		args.SetArgument(fmt.Sprintf("arg%d", i), arg)
	}

	// Get options if command defines them
	if cmdWithOpts, ok := cmd.(CommandWithOptions); ok {
		for _, opt := range cmdWithOpts.DefineOptions() {
			switch opt.Type {
			case StringOption:
				if val, err := cobraCmd.Flags().GetString(opt.Name); err == nil {
					args.SetOption(opt.Name, val)
				}
			case BoolOption:
				if val, err := cobraCmd.Flags().GetBool(opt.Name); err == nil {
					args.SetOption(opt.Name, val)
				}
			case IntOption:
				if val, err := cobraCmd.Flags().GetInt(opt.Name); err == nil {
					args.SetOption(opt.Name, val)
				}
			case Int64Option:
				if val, err := cobraCmd.Flags().GetInt64(opt.Name); err == nil {
					args.SetOption(opt.Name, val)
				}
			case Float64Option:
				if val, err := cobraCmd.Flags().GetFloat64(opt.Name); err == nil {
					args.SetOption(opt.Name, val)
				}
			case StringSliceOption:
				if val, err := cobraCmd.Flags().GetStringSlice(opt.Name); err == nil {
					args.SetOption(opt.Name, val)
				}
			}
		}
	}

	return args
}

// printCommandList prints all available commands grouped by namespace
func (c *CLI) printCommandList() {
	fmt.Printf("%s version %s\n\n", c.appName, c.version)
	fmt.Println("Available commands:")

	groups := c.kernel.Registry().GroupByNamespace()

	// Get sorted namespace keys
	namespaces := make([]string, 0, len(groups))
	for ns := range groups {
		namespaces = append(namespaces, ns)
	}
	sort.Strings(namespaces)

	for _, ns := range namespaces {
		cmds := groups[ns]

		if ns != "" {
			fmt.Printf("\n %s\n", ns)
		}

		// Sort commands within namespace
		sort.Slice(cmds, func(i, j int) bool {
			return cmds[i].Name() < cmds[j].Name()
		})

		for _, cmd := range cmds {
			name := cmd.Name()
			if ns != "" {
				// Remove namespace prefix for display
				name = strings.TrimPrefix(name, ns+":")
			}
			fmt.Printf("  %-20s %s\n", name, cmd.Description())
		}
	}
}

// Execute runs the CLI application
func (c *CLI) Execute(ctx context.Context) error {
	return c.rootCmd.ExecuteContext(ctx)
}

// ExecuteWithArgs runs the CLI with specific arguments (useful for testing)
func (c *CLI) ExecuteWithArgs(ctx context.Context, args []string) error {
	c.rootCmd.SetArgs(args)
	return c.rootCmd.ExecuteContext(ctx)
}

// RootCommand returns the root cobra command for advanced customization
func (c *CLI) RootCommand() *cobra.Command {
	return c.rootCmd
}

// RunCLI is a convenience function to create and run a CLI
func (k *Kernel) RunCLI(ctx context.Context, opts ...CLIOption) error {
	cli := NewCLI(k, opts...)
	return cli.Execute(ctx)
}

// RunCLIWithArgs runs CLI with specific arguments
func (k *Kernel) RunCLIWithArgs(ctx context.Context, args []string, opts ...CLIOption) error {
	cli := NewCLI(k, opts...)
	return cli.ExecuteWithArgs(ctx, args)
}

// Main is a convenience function to run the CLI as main entry point
func Main(ctx context.Context, kernel *Kernel, opts ...CLIOption) {
	if err := kernel.RunCLI(ctx, opts...); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
