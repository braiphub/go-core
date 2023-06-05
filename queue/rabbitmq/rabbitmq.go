package rabbitmq

import (
	"context"
	"fmt"
	"time"

	"github.com/braiphub/go-core/queue"
	"github.com/pkg/errors"

	gorabbitmq "github.com/wagslane/go-rabbitmq"
)

type RabbitMQ struct {
	dsn    string
	logger *LoggerAdapter
}

// for now, we only support classic queues with dead letter exchange
func New( /* logger log.Logger, */ user, pass, host string) (*RabbitMQ, error) {
	dsn := fmt.Sprintf("amqp://%s:%s@%s", user, pass, host) // amqp://user:pass@host
	logger := &LoggerAdapter{}

	// test connection
	conn, err := gorabbitmq.NewConn(
		dsn,
		gorabbitmq.WithConnectionOptionsLogger(logger),
	)
	if err != nil {
		return nil, errors.Wrap(err, "init rabbitmq connection")
	}
	defer conn.Close()

	return &RabbitMQ{
		dsn:    dsn,
		logger: logger,
	}, nil
}

// Publish: exchangeDsn: exchange-name@key
func (rmq *RabbitMQ) Publish(ctx context.Context, exchange string, msg queue.Message) error {
	conn, err := gorabbitmq.NewConn(
		rmq.dsn,
		gorabbitmq.WithConnectionOptionsLogger(rmq.logger),
	)
	if err != nil {
		return err
	}
	defer conn.Close()

	publisher, err := gorabbitmq.NewPublisher(
		conn,
		gorabbitmq.WithPublisherOptionsLogger(rmq.logger),
		gorabbitmq.WithPublisherOptionsExchangeName(exchange),
	)
	if err != nil {
		return err
	}
	defer publisher.Close()

	publisher.NotifyReturn(func(r gorabbitmq.Return) {
		// rmq.logger.Debug("message returned from server: " + string(r.Body))
	})

	publisher.NotifyPublish(func(c gorabbitmq.Confirmation) {
		// rmq.logger.Debug("message confirmed from server", log.Any("tag", c.DeliveryTag), log.Any("ack", c.Ack))
	})

	// prepare headers
	headers := gorabbitmq.Table{}
	for k, v := range msg.Metadata {
		headers[k] = v
	}

	// publish
	err = publisher.Publish(
		msg.Body,     // data
		[]string{""}, // routing keys
		gorabbitmq.WithPublishOptionsHeaders(headers),                // metadata
		gorabbitmq.WithPublishOptionsContentType("application/json"), // optionFuncs
		gorabbitmq.WithPublishOptionsExchange(exchange),              // optionFuncs
	)
	if err != nil {
		return err
	}

	return nil
}

func (rmq *RabbitMQ) Subscribe(ctx context.Context, topic, retry string, f func(queue.Message) error) {
	conn, err := gorabbitmq.NewConn(
		rmq.dsn,
		// gorabbitmq.WithConnectionOptionsLogging,
		gorabbitmq.WithConnectionOptionsLogger(rmq.logger),
	)
	if err != nil {
		// rmq.logger.Error("initializing rabbitmq connection", err, log.Any("dsn", rmq.dsn))

		return
	}

	consumer, err := gorabbitmq.NewConsumer(
		conn,
		func(d gorabbitmq.Delivery) gorabbitmq.Action {
			var msg queue.Message

			// message metadata + body assign
			msg.Metadata = map[string]string{}
			for k, v := range d.Headers {
				switch v := v.(type) {
				case string:
					msg.Metadata[k] = v
				}
			}
			msg.Body = d.Body

			// process message
			if err := f(msg); err != nil {
				// rmq.logger.Error("nacked message", err, log.Any("topic", topic), log.Error(err))

				return gorabbitmq.NackDiscard
			}
			// rmq.logger.Debug("consumed", log.Any("topic", topic))

			return gorabbitmq.Ack
		},
		topic,
		gorabbitmq.WithConsumerOptionsLogger(rmq.logger),
		gorabbitmq.WithConsumerOptionsExchangeName(topic),
		gorabbitmq.WithConsumerOptionsQueueArgs(gorabbitmq.Table{"x-dead-letter-exchange": retry}),
	)
	if err != nil {
		// rmq.logger.Error("initializing rabbitmq consumer", err, log.Any("queue", topic))

		return
	}
	defer consumer.Close()

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Millisecond):
		}
	}
}
