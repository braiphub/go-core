package queue

import (
	"context"
	"encoding/json"

	"github.com/braiphub/go-core/log"
	"github.com/rabbitmq/amqp091-go"
)

type JSONMessage map[string]interface{}

type ConsumeOptions struct {
	PrefetchCount   *int
	Priority        *int
	ConsumerTimeout *int
}

func (r *RabbitMQConnection) Consume(
	ctx context.Context,
	queue string,
	processMsgFn func(ctx context.Context, msg Message) error,
	opts ...func(*ConsumeOptions),
) {
	var (
		forever chan struct{}
		options ConsumeOptions
	)

	for _, opt := range opts {
		opt(&options)
	}

	go func() {
		if r.deferPanicHandler != nil {
			defer r.deferPanicHandler(queue)
		}

		for msg := range r.channelConsumer(ctx, queue, options) {
			func() {
				message := Message{
					Headers: msg.Headers,
					Body:    msg.Body,
				}

				// call message handler
				err := processMsgFn(ctx, message)

				// error: unacknownledge
				if err != nil {
					r.logger.WithContext(ctx).Error(
						"process message error",
						err,
						log.Any("queue", queue),
						getProcessMessageErrorField(msg),
					)

					r.callErrorHandler(queue, msg, err)

					if err := msg.Nack(false, false); err != nil {
						r.logger.WithContext(ctx).Error("nack message", err)
					}

					return
				}

				// ok: acknowledge
				if err := msg.Ack(false); err != nil {
					r.logger.WithContext(ctx).Error("ack message", err)
				}
			}()
		}
	}()

	r.logger.WithContext(ctx).Info("[*] Listening for messages in queue: " + queue)
	<-forever
}

func (r *RabbitMQConnection) callErrorHandler(queue string, msg amqp091.Delivery, err error) {
	if r.errorHandler == nil {
		return
	}

	r.errorHandler(queue, msg.Body, nil, err)
}

func (r *RabbitMQConnection) ConsumeStream(
	ctx context.Context,
	eventName string,
	processMsgFn func(ctx context.Context, msg Message) error,
) {
	routingKey := eventName

	go func() {
		for msg := range r.channelStreamConsumer(ctx, routingKey) {
			message := Message{
				Headers: msg.Headers,
				Body:    msg.Body,
			}

			if err := processMsgFn(ctx, message); err != nil {
				r.logger.WithContext(ctx).Error(
					"process message error",
					err,
					log.Any("message_data", string(msg.Body)),
				)
				if err := msg.Nack(false, false); err != nil {
					r.logger.WithContext(ctx).Error("nack message", err)
				}

				continue
			}

			if err := msg.Ack(false); err != nil {
				r.logger.WithContext(ctx).Error("ack message", err)
			}
		}
	}()

	<-ctx.Done()
}

func getProcessMessageErrorField(msg amqp091.Delivery) log.Field {
	var jsonMessage JSONMessage

	err := json.Unmarshal(msg.Body, &jsonMessage)
	if err != nil {
		return log.Any("message_data", string(msg.Body))
	}

	return log.Any("message_data", jsonMessage)
}
