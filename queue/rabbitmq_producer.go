package queue

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
)

func (r *RabbitMQConnection) Produce(ctx context.Context, routingKey string, msg any) error {
	body, headers, err := r.buildMessageBodyAndHeaders(msg)
	if err != nil {
		return errors.Wrap(err, "build message")
	}

	publishFn := func() (any, error) {
		if r.conn == nil || r.conn.IsClosed() {
			return nil, errors.New("rabbitmq connection closed")
		}

		ch, err := r.conn.Channel()
		if err != nil {
			return nil, err
		}
		defer ch.Close()

		pubCtx, cancel := context.WithTimeout(ctx, publishTimeout)
		defer cancel()

		return nil, ch.PublishWithContext(
			pubCtx,
			r.config.Exchange,
			routingKey,
			false, // mandatory
			false, // immediate
			amqp.Publishing{ //nolint:exhaustruct
				ContentType:  "application/json",
				Body:         body,
				Headers:      headers,
				DeliveryMode: amqp.Persistent,
			},
		)
	}

	_, err = r.breaker.Execute(publishFn)
	if err != nil {
		return r.handlePublisherFallBack(ctx, routingKey, &Message{
			Headers: headers,
			Body:    body,
		})
	}

	return nil
}

func objectToPayload(object interface{}) ([]byte, error) {
	var payload []byte

	switch v := object.(type) {
	case string:
		payload = []byte(v)

	case []byte:
		payload = v

	case nil:
		return nil, ErrEmptyObject

	default:
		var err error
		payload, err = json.Marshal(object)
		if err != nil {
			return nil, errors.Wrap(err, "marshal payload")
		}
	}

	return payload, nil
}

func (r *RabbitMQConnection) handlePublisherFallBack(ctx context.Context, routingKey string, msg *Message) error {
	if r.databaseFallback == nil {
		return errors.New("database fallback not initialized and rabbitmq is not healthy")
	}

	headers, err := json.Marshal(msg.Headers)
	if err != nil {
		return errors.Wrap(err, "marshal headers")
	}

	model := &GormFallbackProducerModel{
		RoutingKey: routingKey,
		Headers:    headers,
		Body:       msg.Body,
	}

	if err := r.databaseFallback.database.WithContext(ctx).Debug().Create(model).Error; err != nil {
		return errors.Wrap(err, "create fallback producer model")
	}

	return nil
}

func (r *RabbitMQConnection) buildMessageBodyAndHeaders(
	msg any,
) (body []byte, headers map[string]any, err error) {
	switch m := msg.(type) {
	case Message:
		body = m.Body
		headers = m.Headers
	case *Message:
		body = m.Body
		headers = m.Headers
	case []byte:
		body = m
	default:
		buf, err := json.Marshal(msg)
		if err != nil {
			return nil, nil, errors.Wrap(err, "marshal msg")
		}
		body = buf
	}

	if headers == nil {
		headers = make(map[string]any)
	}

	return body, headers, nil
}
