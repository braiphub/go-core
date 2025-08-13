package eventbus

//go:generate mockgen -source=publisher.go -destination=publisher_mock.go -package=eventbus . Publisher
type Publisher interface {
	Publish(topic string, args ...interface{})
}
