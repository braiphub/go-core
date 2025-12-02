package command

import (
	"strconv"
	"strings"
)

// OptionType defines the type of a command option
type OptionType int

const (
	StringOption OptionType = iota
	BoolOption
	IntOption
	Int64Option
	Float64Option
	StringSliceOption
)

// Option defines a command flag/option (similar to Laravel's signature)
type Option struct {
	Name        string
	Shorthand   string // e.g., "v" for --verbose / -v
	Description string
	Default     any
	Required    bool
	Type        OptionType
}

// Args holds parsed command arguments and options
type Args struct {
	Arguments map[string]string
	Options   map[string]any
}

// NewArgs creates a new Args instance
func NewArgs() *Args {
	return &Args{
		Arguments: make(map[string]string),
		Options:   make(map[string]any),
	}
}

// SetArgument sets a positional argument
func (a *Args) SetArgument(name, value string) *Args {
	a.Arguments[name] = value
	return a
}

// SetOption sets an option value
func (a *Args) SetOption(name string, value any) *Args {
	a.Options[name] = value
	return a
}

// GetArgument returns an argument value or default
func (a *Args) GetArgument(name string, defaultValue ...string) string {
	if val, ok := a.Arguments[name]; ok {
		return val
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

// GetString returns an option as string
func (a *Args) GetString(name string, defaultValue ...string) string {
	if val, ok := a.Options[name]; ok {
		switch v := val.(type) {
		case string:
			return v
		case *string:
			if v != nil {
				return *v
			}
		default:
			return ""
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

// GetBool returns an option as bool
func (a *Args) GetBool(name string, defaultValue ...bool) bool {
	if val, ok := a.Options[name]; ok {
		switch v := val.(type) {
		case bool:
			return v
		case *bool:
			if v != nil {
				return *v
			}
		case string:
			b, _ := strconv.ParseBool(v)
			return b
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return false
}

// GetInt returns an option as int
func (a *Args) GetInt(name string, defaultValue ...int) int {
	if val, ok := a.Options[name]; ok {
		switch v := val.(type) {
		case int:
			return v
		case *int:
			if v != nil {
				return *v
			}
		case int64:
			return int(v)
		case float64:
			return int(v)
		case string:
			i, _ := strconv.Atoi(v)
			return i
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

// GetInt64 returns an option as int64
func (a *Args) GetInt64(name string, defaultValue ...int64) int64 {
	if val, ok := a.Options[name]; ok {
		switch v := val.(type) {
		case int64:
			return v
		case *int64:
			if v != nil {
				return *v
			}
		case int:
			return int64(v)
		case float64:
			return int64(v)
		case string:
			i, _ := strconv.ParseInt(v, 10, 64)
			return i
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

// GetFloat64 returns an option as float64
func (a *Args) GetFloat64(name string, defaultValue ...float64) float64 {
	if val, ok := a.Options[name]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case *float64:
			if v != nil {
				return *v
			}
		case int:
			return float64(v)
		case int64:
			return float64(v)
		case string:
			f, _ := strconv.ParseFloat(v, 64)
			return f
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return 0
}

// GetStringSlice returns an option as []string
func (a *Args) GetStringSlice(name string, defaultValue ...[]string) []string {
	if val, ok := a.Options[name]; ok {
		switch v := val.(type) {
		case []string:
			return v
		case *[]string:
			if v != nil {
				return *v
			}
		case string:
			if v == "" {
				return []string{}
			}
			return strings.Split(v, ",")
		}
	}
	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return []string{}
}

// Has checks if an option exists and is truthy
func (a *Args) Has(name string) bool {
	val, ok := a.Options[name]
	if !ok {
		return false
	}

	switch v := val.(type) {
	case bool:
		return v
	case string:
		return v != ""
	case int:
		return v != 0
	default:
		return val != nil
	}
}

// HasArgument checks if an argument exists
func (a *Args) HasArgument(name string) bool {
	_, ok := a.Arguments[name]
	return ok
}
