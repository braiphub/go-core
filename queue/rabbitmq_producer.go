package queue

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
)

func (r *RabbitMQConnection) Produce(ctx context.Context, routingKey string, msg any) error {
	var body []byte
	headers := make(map[string]any)

	switch msg.(type) {
	case Message:
		body = msg.(Message).Body
		headers = msg.(Message).Headers

	case []byte:
		body = msg.([]byte)

	default:
		buf, err := json.Marshal(msg)
		if err != nil {
			return errors.Wrap(err, "marshal msg")
		}

		body = buf
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, publishTimeout)
	defer cancel()

	err := r.channel.PublishWithContext(
		timeoutCtx,
		r.config.Exchange, // exchange
		routingKey,        // routing key
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{ //nolint:exhaustruct
			ContentType: "text/plain",
			Body:        body,
			Headers:     headers,
		})
	if err != nil {
		return errors.Wrap(err, "publish")
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
