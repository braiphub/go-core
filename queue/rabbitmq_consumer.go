package queue

import (
	"context"
	"encoding/json"

	"github.com/braiphub/go-core/log"
	"github.com/braiphub/go-core/trace"
	"github.com/rabbitmq/amqp091-go"
)

type JSONMessage map[string]interface{}

func (r *RabbitMQConnection) Consume(
	ctx context.Context,
	queue string,
	processMsgFn func(ctx context.Context, data []byte) error,
) {
	var forever chan struct{}

	go func() {
		for msg := range r.channelConsumer(ctx, queue) {
			func() {
				// tracer span init
				spanCtx, cancel := context.WithCancel(ctx)
				defer cancel()

				_, processSpan := r.newSpan(spanCtx, trace.KindConsumer, "queue-consumer", trace.Attr("queue", queue))
				defer r.closeSpan(processSpan)

				// call message handler
				err := processMsgFn(ctx, msg.Body)

				// error: unacknownledge
				if err != nil {
					r.logger.WithContext(ctx).Error("process message error", err, getProcessMessageErrorField(msg))

					r.setSpanStatus(processSpan, trace.StatusError, "process message error")

					if err := msg.Nack(false, false); err != nil {
						r.logger.WithContext(ctx).Error("nack message", err)
					}

					return
				}

				// ok: acknownledge
				r.setSpanStatus(processSpan, trace.StatusOK, "")

				if err := msg.Ack(false); err != nil {
					r.logger.WithContext(ctx).Error("ack message", err)
				}
			}()
		}
	}()

	r.logger.WithContext(ctx).Info("[*] Listening for messages in queue: " + queue)
	<-forever
}

//nolint:ireturn
func (r *RabbitMQConnection) newSpan(
	ctx context.Context,
	kind trace.SpanKind,
	name string,
	attrs ...trace.Attribute,
) (context.Context, trace.SpanInterface) {
	if r.tracer == nil {
		return ctx, nil
	}

	return r.tracer.StartSpanWithKind(ctx, kind, name, attrs...)
}

func (r *RabbitMQConnection) setSpanStatus(span trace.SpanInterface, status trace.SpanStatus, msg string) {
	if span == nil {
		return
	}

	span.Status(status, msg)
}

func (r *RabbitMQConnection) closeSpan(span trace.SpanInterface) {
	if span == nil {
		return
	}

	span.Close()
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

func getProcessMessageErrorField(msg amqp091.Delivery) log.Field {
	var jsonMessage JSONMessage

	err := json.Unmarshal(msg.Body, &jsonMessage)
	if err != nil {
		return log.Any("message_data", string(msg.Body))
	}

	return log.Any("message_data", jsonMessage)
}
