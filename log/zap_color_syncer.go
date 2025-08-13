package log

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

const (
	Reset = "\033[0m"
	Gray  = "\033[38;5;238m"
)

type debugColorSyncer struct{}

func (debugColorSyncer) Write(p []byte) (n int, err error) {
	s := string(p) + "\n"

	switch {
	case strings.Contains(s, `"level":"info"`) || strings.Contains(s, "\tinfo\t"):
		color.Cyan(s)

	case strings.Contains(s, `"level":"warn"`) || strings.Contains(s, "\twarn\t"):
		color.Yellow(s)

	case strings.Contains(s, `"level":"error"`) || strings.Contains(s, "\terror\t"):
		color.Red(s)

	case strings.Contains(s, `"level":"fatal"`) || strings.Contains(s, "\tfatal\t"):
		color.Magenta(s)

	default:
		fmt.Print(Gray + s + Reset)
	}

	return len(p), nil
}
