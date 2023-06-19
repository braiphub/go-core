package rabbitmq

import (
	"context"
	"fmt"
	"time"

	"github.com/braiphub/go-core/log"
	"github.com/braiphub/go-core/queue"
	"github.com/pkg/errors"

	gorabbitmq "github.com/wagslane/go-rabbitmq"
)

type RabbitMQ struct {
	dsn           string
	logger        log.LoggerI
	loggerAdapter *LoggerAdapter
}

func New(logger log.LoggerI, user, pass, host, vhost string) (*RabbitMQ, error) {
	dsn := fmt.Sprintf("amqp://%s:%s@%s/%s", user, pass, host, vhost) // amqp://user:pass@host/vhost

	loggerAdapter := &LoggerAdapter{}

	// test connection
	conn, err := gorabbitmq.NewConn(
		dsn,
		gorabbitmq.WithConnectionOptionsLogger(loggerAdapter),
	)
	if err != nil {
		return nil, errors.Wrap(err, "init rabbitmq connection")
	}
	defer conn.Close()

	return &RabbitMQ{
		dsn:           dsn,
		logger:        logger,
		loggerAdapter: loggerAdapter,
	}, nil
}

// Publish: exchangeDsn: exchange-name@key
func (rmq *RabbitMQ) Publish(
	ctx context.Context,
	exchange string,
	routingKeys []string,
	msg queue.Message,
) error {
	if len(routingKeys) == 0 {
		routingKeys = []string{""}
	}

	conn, err := gorabbitmq.NewConn(
		rmq.dsn,
		gorabbitmq.WithConnectionOptionsLogger(rmq.loggerAdapter),
	)
	if err != nil {
		return err
	}
	defer conn.Close()

	publisher, err := gorabbitmq.NewPublisher(
		conn,
		gorabbitmq.WithPublisherOptionsLogger(rmq.loggerAdapter),
		gorabbitmq.WithPublisherOptionsExchangeName(exchange),
	)
	if err != nil {
		return err
	}
	defer publisher.Close()

	publisher.NotifyReturn(func(r gorabbitmq.Return) { /* do nothing */ })
	publisher.NotifyPublish(func(c gorabbitmq.Confirmation) { /* do nothing */ })

	// prepare headers
	headers := gorabbitmq.Table{}
	for k, v := range msg.Metadata {
		headers[k] = v
	}

	// publish
	err = publisher.Publish(
		msg.Body, // data
		routingKeys,
		gorabbitmq.WithPublishOptionsHeaders(headers),                // metadata
		gorabbitmq.WithPublishOptionsContentType("application/json"), // optionFuncs
		gorabbitmq.WithPublishOptionsExchange(exchange),              // optionFuncs
	)
	if err != nil {
		return err
	}

	return nil
}

func (rmq *RabbitMQ) Subscribe(ctx context.Context, topic, retry string, f func(context.Context, queue.Message) error) {
	conn, err := gorabbitmq.NewConn(
		rmq.dsn,
		gorabbitmq.WithConnectionOptionsLogger(rmq.loggerAdapter),
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
			if err := f(ctx, msg); err != nil {
				requestID := msg.Metadata["request-id"]
				rmq.logger.Error("nacked message", err, log.Any("request-id", requestID), log.Any("topic", topic))

				return gorabbitmq.NackDiscard
			}

			return gorabbitmq.Ack
		},
		topic,
		gorabbitmq.WithConsumerOptionsLogger(rmq.loggerAdapter),
		gorabbitmq.WithConsumerOptionsExchangeName(topic),
		gorabbitmq.WithConsumerOptionsQueueArgs(gorabbitmq.Table{"x-dead-letter-exchange": retry}),
		gorabbitmq.WithConsumerOptionsQueueNoDeclare,
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
