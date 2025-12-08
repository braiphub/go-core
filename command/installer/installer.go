package installer

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Installer handles the CLI system setup
type Installer struct {
	projectRoot string
	moduleName  string
	hasConfigs  bool
	force       bool
}

// New creates a new installer instance
func New(force bool) (*Installer, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	return &Installer{
		projectRoot: cwd,
		force:       force,
	}, nil
}

// Run orchestrates the installation process
func (i *Installer) Run() error {
	fmt.Println("🚀 Starting CLI system installation...")
	fmt.Println()

	// Step 1: Detect module name
	if err := i.detectModuleName(); err != nil {
		return fmt.Errorf("❌ failed to detect module name: %w", err)
	}
	fmt.Printf("✅ Detected module: %s\n", i.moduleName)

	// Step 2: Detect project structure
	i.detectProjectStructure()
	if i.hasConfigs {
		fmt.Println("✅ Detected configs directory")
	}

	// Step 3: Create directories
	if err := i.createDirectories(); err != nil {
		return fmt.Errorf("❌ failed to create directories: %w", err)
	}

	// Step 4: Create CLI main file
	if err := i.createCLIMain(); err != nil {
		return fmt.Errorf("❌ failed to create CLI main: %w", err)
	}

	// Step 5: Create registry file
	if err := i.createRegistry(); err != nil {
		return fmt.Errorf("❌ failed to create registry: %w", err)
	}

	// Step 6: Update configs if exists
	if i.hasConfigs {
		if err := i.updateConfigs(); err != nil {
			return fmt.Errorf("⚠️  warning: failed to update configs: %w", err)
		}
	}

	// Step 7: Print success message
	i.printSuccess()

	return nil
}

// detectModuleName reads the go.mod file to get the module name
func (i *Installer) detectModuleName() error {
	goModPath := filepath.Join(i.projectRoot, "go.mod")

	file, err := os.Open(goModPath)
	if err != nil {
		return fmt.Errorf("go.mod not found - make sure you're in a Go project root")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			i.moduleName = strings.TrimSpace(strings.TrimPrefix(line, "module"))
			return nil
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading go.mod: %w", err)
	}

	return fmt.Errorf("module name not found in go.mod")
}

// detectProjectStructure checks if configs directory exists
func (i *Installer) detectProjectStructure() {
	configsPath := filepath.Join(i.projectRoot, "configs")
	if info, err := os.Stat(configsPath); err == nil && info.IsDir() {
		i.hasConfigs = true
	}
}

// createDirectories creates the necessary directories
func (i *Installer) createDirectories() error {
	dirs := []string{
		filepath.Join(i.projectRoot, "cmd", "cli"),
		filepath.Join(i.projectRoot, "internal", "commands"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		fmt.Printf("📁 Created directory: %s\n", dir)
	}

	return nil
}

// createCLIMain creates the cmd/cli/main.go file from template
func (i *Installer) createCLIMain() error {
	targetPath := filepath.Join(i.projectRoot, "cmd", "cli", "main.go")

	// Check if file exists and not force mode
	if _, err := os.Stat(targetPath); err == nil && !i.force {
		fmt.Printf("⚠️  Skipping: %s already exists (use --force to overwrite)\n", targetPath)
		return nil
	}

	tmpl, err := template.New("main").Parse(cliMainTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	data := struct {
		ModuleName string
		HasConfigs bool
	}{
		ModuleName: i.moduleName,
		HasConfigs: i.hasConfigs,
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	if err := os.WriteFile(targetPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("✅ Created: %s\n", targetPath)
	return nil
}

// createRegistry creates the internal/commands/registry.go file
func (i *Installer) createRegistry() error {
	targetPath := filepath.Join(i.projectRoot, "internal", "commands", "registry.go")

	// Check if file exists and not force mode
	if _, err := os.Stat(targetPath); err == nil && !i.force {
		fmt.Printf("⚠️  Skipping: %s already exists (use --force to overwrite)\n", targetPath)
		return nil
	}

	if err := os.WriteFile(targetPath, []byte(registryTemplate), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("✅ Created: %s\n", targetPath)
	return nil
}

// updateConfigs updates the config.yaml file to add appName
func (i *Installer) updateConfigs() error {
	configPath := filepath.Join(i.projectRoot, "configs", "config.yaml")

	// Check if config.yaml exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("⚠️  config.yaml not found, skipping config update\n")
		return nil
	}

	// Read existing config
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config.yaml: %w", err)
	}

	// Check if appName already exists
	if strings.Contains(string(content), "appName:") {
		fmt.Printf("⚠️  appName already exists in config.yaml, skipping\n")
		return nil
	}

	// Add appName to the beginning
	newContent := "appName: ${APP_NAME}\n" + string(content)

	if err := os.WriteFile(configPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write config.yaml: %w", err)
	}

	fmt.Printf("✅ Updated: %s (added appName)\n", configPath)
	return nil
}

// printSuccess prints the success message with next steps
func (i *Installer) printSuccess() {
	fmt.Println()
	fmt.Println("🎉 CLI system installation completed successfully!")
	fmt.Println()
	fmt.Println("📋 Next steps:")
	fmt.Println()
	fmt.Println("1. Add your commands to internal/commands/registry.go")
	fmt.Println("2. Build your CLI:")
	fmt.Println("   go build -o cli ./cmd/cli")
	fmt.Println()
	fmt.Println("3. Run your CLI:")
	fmt.Println("   ./cli")
	fmt.Println()
	fmt.Println("4. Use make:command to generate new commands:")
	fmt.Println("   ./cli make:command MyCommand")
	fmt.Println()
	fmt.Println("5. See all available commands:")
	fmt.Println("   ./cli --help")
	fmt.Println()
}
