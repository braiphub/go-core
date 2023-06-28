package zaplogger

import (
	"strings"

	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

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
		color.Black(s)
	}

	return len(p), nil
}

func newZapLoggerDev(skip int) (*zap.Logger, error) {
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel
	})

	// output should also go to standard out.
	syncer := zapcore.AddSync(&debugColorSyncer{})
	consoleDebugging := zapcore.Lock(syncer)
	consoleErrors := zapcore.Lock(syncer)
	consoleEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

	// Join the outputs, encoders, and level-handling functions into
	// zapcore.Cores, then tee the four cores together.
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleErrors, highPriority),
		zapcore.NewCore(consoleEncoder, consoleDebugging, lowPriority),
	)

	// From a zapcore.Core, it's easy to construct a Logger.
	logger := zap.New(core, zap.AddCallerSkip(skip), zap.AddStacktrace(zap.DPanicLevel))

	return logger, nil
}
