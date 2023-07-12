package rabbitmq

import (
	"context"
	"fmt"
	"time"

	"github.com/braiphub/go-core/cache"
	"github.com/braiphub/go-core/log"
	"github.com/braiphub/go-core/queue"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	gorabbitmq "github.com/wagslane/go-rabbitmq"
)

type checker interface {
	SetProcessed(ctx context.Context, messageID string) error
	CanProcess(ctx context.Context, messageID string) error
}

type RabbitMQ struct {
	dsn                string
	messageIDLabel     string
	logger             log.Logger
	loggerAdapter      *LoggerAdapter
	idempotencyChecker checker
}

var ErrInvalidParam = errors.New("parameter is invalid")

func New(logger log.Logger, user, pass, host, vhost string) (*RabbitMQ, error) {
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

func (rmq *RabbitMQ) SetIdempotencyChecker(cacheKey, messageLabel string, duration time.Duration, cache cache.Cacherer) error {
	switch {
	case cacheKey == "":
		return errors.Wrap(ErrInvalidParam, "cache-key is empty")
	case messageLabel == "":
		return errors.Wrap(ErrInvalidParam, "message-label is empty")
	case duration == 0:
		return errors.Wrap(ErrInvalidParam, "duration is empty")
	case cache == nil:
		return errors.Wrap(ErrInvalidParam, "cache is nil")
	}

	rmq.messageIDLabel = messageLabel
	rmq.idempotencyChecker = &idempotencyChecker{
		cacheKey:     cacheKey,
		messageLabel: messageLabel,
		duration:     duration,
		cache:        cache,
	}

	return nil
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

	// set message id for idempotency
	if rmq.idempotencyChecker != nil {
		headers[rmq.messageIDLabel] = uuid.New().String()
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
	reconnectInterval := 30 * time.Second

	conn, err := gorabbitmq.NewConn(
		rmq.dsn,
		gorabbitmq.WithConnectionOptionsLogger(rmq.loggerAdapter),
		gorabbitmq.WithConnectionOptionsReconnectInterval(reconnectInterval),
	)
	if err != nil {
		rmq.logger.Error("initializing rabbitmq connection", err, log.Any("dsn", rmq.dsn))

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
			messageID := msg.Metadata[rmq.messageIDLabel]
			idempotencyKey := topic + "." + messageID

			// check for idempotency
			if rmq.idempotencyChecker != nil {
				if messageID == "" {
					rmq.logger.Error("nacked message: message without message-id header", nil, log.Any("topic", topic))

					return gorabbitmq.NackDiscard
				}
				if err := rmq.idempotencyChecker.CanProcess(ctx, idempotencyKey); err != nil {
					rmq.logger.Error("nacked message: check can process", err, log.Any("message-id", idempotencyKey), log.Any("topic", topic))

					return gorabbitmq.NackDiscard
				}
			}

			// process message
			if err := f(ctx, msg); err != nil {
				rmq.logger.Error("nacked message", err, log.Any("message-id", idempotencyKey), log.Any("topic", topic))

				return gorabbitmq.NackDiscard
			}

			// set idempotency for message-id key
			if rmq.idempotencyChecker != nil {
				if err := rmq.idempotencyChecker.SetProcessed(ctx, idempotencyKey); err != nil {
					rmq.logger.Error("setting message as processed", err, log.Any("message-id", idempotencyKey), log.Any("topic", topic))
				}
			}

			return gorabbitmq.Ack
		},
		topic,
		gorabbitmq.WithConsumerOptionsLogger(rmq.loggerAdapter),
		gorabbitmq.WithConsumerOptionsQueueNoDeclare,
	)
	if err != nil {
		rmq.logger.Error("initializing rabbitmq consumer", err, log.Any("queue", topic))

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
