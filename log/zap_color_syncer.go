package log

import "strings"
import "github.com/fatih/color"

type debugColorSyncer struct{}

func (debugColorSyncer) Write(p []byte) (n int, err error) {
	s := string(p) + "\n"

	switch {
	case strings.Contains(s, `"level":"info"`):
		color.Cyan(s)

	case strings.Contains(s, `"level":"warn"`):
		color.Yellow(s)

	case strings.Contains(s, `"level":"error"`):
		color.Red(s)

	case strings.Contains(s, `"level":"fatal"`):
		color.Magenta(s)

	default:
		color.White(s)
	}

	return len(p), nil
}
