package command

import (
	"sort"
	"strings"
	"sync"
)

// Registry manages command registration and lookup
type Registry struct {
	mu       sync.RWMutex
	commands map[string]Command
}

// NewRegistry creates a new command registry
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]Command),
	}
}

// Register adds a command to the registry
func (r *Registry) Register(cmd Command) error {
	if cmd == nil {
		return ErrInvalidCommandName
	}

	name := strings.TrimSpace(cmd.Name())
	if name == "" {
		return ErrInvalidCommandName
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.commands[name]; exists {
		return ErrCommandAlreadyExists
	}

	r.commands[name] = cmd
	return nil
}

// RegisterMany adds multiple commands to the registry
func (r *Registry) RegisterMany(cmds ...Command) error {
	for _, cmd := range cmds {
		if err := r.Register(cmd); err != nil {
			return err
		}
	}
	return nil
}

// Get retrieves a command by name
func (r *Registry) Get(name string) (Command, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cmd, exists := r.commands[name]
	if !exists {
		return nil, ErrCommandNotFound
	}

	return cmd, nil
}

// Has checks if a command exists
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.commands[name]
	return exists
}

// All returns all registered commands
func (r *Registry) All() map[string]Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]Command, len(r.commands))
	for k, v := range r.commands {
		result[k] = v
	}
	return result
}

// Names returns all command names sorted alphabetically
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.commands))
	for name := range r.commands {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Count returns the number of registered commands
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.commands)
}

// Remove removes a command from the registry
func (r *Registry) Remove(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.commands, name)
}

// Clear removes all commands from the registry
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.commands = make(map[string]Command)
}

// FindByPrefix returns commands whose names start with the given prefix
func (r *Registry) FindByPrefix(prefix string) []Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []Command
	for name, cmd := range r.commands {
		if strings.HasPrefix(name, prefix) {
			result = append(result, cmd)
		}
	}
	return result
}

// GroupByNamespace groups commands by their namespace (part before ":")
func (r *Registry) GroupByNamespace() map[string][]Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	groups := make(map[string][]Command)
	for name, cmd := range r.commands {
		namespace := ""
		if idx := strings.Index(name, ":"); idx > 0 {
			namespace = name[:idx]
		}
		groups[namespace] = append(groups[namespace], cmd)
	}
	return groups
}
