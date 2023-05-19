package rabbitmq

type LoggerAdapter struct{}

func (LoggerAdapter) Fatalf(string, ...interface{}) {}
func (LoggerAdapter) Errorf(string, ...interface{}) {}
func (LoggerAdapter) Warnf(string, ...interface{})  {}
func (LoggerAdapter) Infof(string, ...interface{})  {}
func (LoggerAdapter) Debugf(string, ...interface{}) {}
func (LoggerAdapter) Tracef(string, ...interface{}) {}
