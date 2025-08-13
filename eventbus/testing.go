package eventbus

type Testing struct {
	*Bus
	calledEvents map[string]int
}

func NewTesting() *Testing {
	return &Testing{
		Bus:          buildDefaultInstance(),
		calledEvents: make(map[string]int),
	}
}

func (t *Testing) Publish(topic string, args ...interface{}) {
	t.calledEvents[topic]++
	t.Bus.Publish(topic, args...)
}

func (t *Testing) WasEventFired(topic string) bool {
	return t.calledEvents[topic] > 0
}

func (t *Testing) FireCount(topic string) int {
	return t.calledEvents[topic]
}
