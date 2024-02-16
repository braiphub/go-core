package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/braiphub/go-core/log"
)

type JSONMessage map[string]interface{}

func (r *RabbitMQConnection) Consume(
	ctx context.Context,
	queue string,
	processMsgFn func(ctx context.Context, data []byte) error,
) {
	var forever chan struct{}

	go func() {
		msgCount := 1
		for msg := range r.channelConsumer(ctx, queue) {
			if err := processMsgFn(ctx, msg.Body); err != nil {
				errorField := log.Any("message_data", string(msg.Body))

				// try parsing message as json, to enhance log visualization
				var jsonMessage JSONMessage
				if err := json.Unmarshal(msg.Body, &jsonMessage); err == nil {
					errorField = log.Any("message_data", jsonMessage)
				}

				r.logger.WithContext(ctx).Error("process message error", err, errorField)

				if err := msg.Nack(false, false); err != nil {
					r.logger.WithContext(ctx).Error("nack message", err)
				}

				continue
			}

			r.logger.WithContext(ctx).Debug(fmt.Sprintf("Message consumed successfully [%d]", msgCount))
			if err := msg.Ack(false); err != nil {
				r.logger.WithContext(ctx).Error("ack message", err)
			}
			msgCount++
		}
	}()

	r.logger.WithContext(ctx).Info("[*] Listening for messages in queue: " + queue)
	<-forever
}

func (r *RabbitMQConnection) ConsumeStream(
	ctx context.Context,
	eventName string,
	processMsgFn func(ctx context.Context, data []byte) error,
) {
	routingKey := eventName

	go func() {
		for msg := range r.channelStreamConsumer(ctx, routingKey) {
			if err := processMsgFn(ctx, msg.Body); err != nil {
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
