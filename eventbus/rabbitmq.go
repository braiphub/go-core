package eventbus

import (
	"context"
	"fmt"

	"github.com/braiphub/go-core/log"
	"github.com/pkg/errors"
	"github.com/rabbitmq/amqp091-go"
	rabbitmq "github.com/wagslane/go-rabbitmq"
)

type rabbitMQPubSub struct {
	dsn           string
	serviceName   string
	daemonName    string
	exchange      string
	daemonQueue   string
	logger        log.LoggerI
	publisherConn *rabbitmq.Conn
	publisher     *rabbitmq.Publisher
}

func WithRabbitMQ(dsn string) func(bus *EventBus) {
	return func(bus *EventBus) {
		exchange := fmt.Sprintf("eventbus.%s", bus.Config.ServiceName)
		daemonQueue := fmt.Sprintf("eventbus.%s.%s", bus.Config.ServiceName, bus.Config.DaemonName)

		r := &rabbitMQPubSub{
			dsn:         dsn,
			serviceName: bus.Config.ServiceName,
			daemonName:  bus.Config.DaemonName,
			exchange:    exchange,
			daemonQueue: daemonQueue,
			logger:      bus.logger,
		}

		bus.pubSub = r
	}
}

func (r *rabbitMQPubSub) Configure(config Config) error {
	if err := r.declareExchange(); err != nil {
		return errors.Wrap(err, "declare exchange")
	}

	if err := r.declareDaemonQueue(); err != nil {
		return errors.Wrap(err, "declare daemon consumer queue")
	}

	return nil
}

func (r *rabbitMQPubSub) ListenToEvents(ctx context.Context, callback SubscriberCallbackFunc) error {
	conn, err := rabbitmq.NewConn(r.dsn, rabbitmq.WithConnectionOptionsLogger(nullLogger{}))
	if err != nil {
		return errors.Wrap(err, "connect")
	}
	defer conn.Close()

	consumer, err := rabbitmq.NewConsumer(
		conn,
		func(d rabbitmq.Delivery) rabbitmq.Action {
			eventName, ok := d.Headers["event_name"].(string)
			if !ok || eventName == "" {
				r.logger.Error(
					"eventbus: discarding message due to unknown event",
					nil,
					log.Any("body", string(d.Body)),
				)

				return rabbitmq.NackDiscard
			}

			if err := callback(ctx, eventName, d.Body); err != nil {
				return rabbitmq.NackDiscard
			}

			return rabbitmq.Ack
		},
		r.daemonQueue,
		rabbitmq.WithConsumerOptionsQueueNoDeclare,
		rabbitmq.WithConsumerOptionsLogger(nullLogger{}),
	)
	if err != nil {
		return errors.Wrap(err, "init consumer")
	}
	defer consumer.Close()

	<-ctx.Done()

	return nil
}

func (r *rabbitMQPubSub) Publish(eventName string, data []byte) error {
	if r.publisher == nil {
		conn, err := rabbitmq.NewConn(r.dsn, rabbitmq.WithConnectionOptionsLogger(nullLogger{}))
		if err != nil {
			return errors.Wrap(err, "open publisher conn")
		}

		publisher, err := rabbitmq.NewPublisher(
			conn,
			rabbitmq.WithPublisherOptionsExchangeName(r.exchange),
			rabbitmq.WithPublisherOptionsLogger(nullLogger{}),
		)
		if err != nil {
			return errors.Wrap(err, "init publisher")
		}

		r.publisherConn = conn
		r.publisher = publisher
	}

	headers := rabbitmq.Table{
		"event_name": eventName,
	}
	optionFuncs := []func(*rabbitmq.PublishOptions){
		rabbitmq.WithPublishOptionsExchange(r.exchange),
		rabbitmq.WithPublishOptionsHeaders(headers),
	}

	return r.publisher.Publish(
		data,
		[]string{"event." + eventName},
		optionFuncs...,
	)
}

func (r *rabbitMQPubSub) declareExchange() error {
	conn, err := amqp091.Dial(r.dsn)
	if err != nil {
		return errors.Wrap(err, "connecting to rabbitmq")
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		return errors.Wrap(err, "channel open")
	}
	defer channel.Close()

	err = channel.ExchangeDeclare(
		r.exchange,
		"topic",
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,   // args [amqp-table]
	)
	if err != nil {
		return errors.Wrap(err, "exchange declare")
	}

	return nil
}

func (r *rabbitMQPubSub) declareDaemonQueue() error {
	conn, err := amqp091.Dial(r.dsn)
	if err != nil {
		return errors.Wrap(err, "connecting to rabbitmq")
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		return errors.Wrap(err, "channel open")
	}
	defer channel.Close()

	deadQueueName := fmt.Sprintf("%s.dead", r.daemonQueue)

	normalQueueArgs := amqp091.Table{
		"x-dead-letter-exchange":    r.exchange,
		"x-dead-letter-routing-key": deadQueueName,
		"x-queue-type":              "quorum",
	}

	deadQueueArgs := amqp091.Table{}

	_, err = channel.QueueDeclare(
		deadQueueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		deadQueueArgs,
	)
	if err != nil {
		return errors.Wrap(err, "dead queue declare")
	}
	if err := r.bindQueue(channel, deadQueueName, deadQueueName); err != nil {
		return errors.Wrap(err, "dead queue bind")
	}

	_, err = channel.QueueDeclare(
		r.daemonQueue,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		normalQueueArgs,
	)
	if err != nil {
		return errors.Wrap(err, "daemon queue declare")
	}
	if err := r.bindQueue(channel, r.daemonQueue, "event.#"); err != nil {
		return errors.Wrap(err, "normal queue bind")
	}

	return nil
}

func (r *rabbitMQPubSub) bindQueue(
	channel *amqp091.Channel,
	queue string,
	routingKey string,
) error {
	err := channel.QueueBind(
		queue,
		routingKey,
		r.exchange,
		false, // no-wait
		nil,   // args [amqp-table]
	)
	if err != nil {
		return errors.Wrap(err, "queue bind")
	}

	return nil
}

type nullLogger struct{}

func (nullLogger) Fatalf(string, ...interface{}) {}
func (nullLogger) Errorf(string, ...interface{}) {}
func (nullLogger) Warnf(string, ...interface{})  {}
func (nullLogger) Infof(string, ...interface{})  {}
func (nullLogger) Debugf(string, ...interface{}) {}
func (nullLogger) Tracef(string, ...interface{}) {}
