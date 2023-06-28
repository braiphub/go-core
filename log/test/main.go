package main

import (
	"errors"

	"github.com/braiphub/go-core/log/zaplogger"
)

type Syncer struct {
	name string
}

func (s *Syncer) Write(p []byte) (n int, err error) {
	println("syncer [" + s.name + "] called: " + string(p))
	return len(p), nil
}

func main() {
	logger, _ := zaplogger.New(zaplogger.LoggerEnvDev, 0)
	logger.Debug("teste Debug")
	logger.Info("teste Info")
	logger.Warn("teste Warn")
	logger.Error("teste Error", errors.New("unknonw"))
	logger.Fatal("teste Fatal")
}
