package health

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/rabbitmq/amqp091-go"
)

type rabbitmqChecker struct {
	dsn string
}

func newRabbitChecker(dsn string) (*rabbitmqChecker, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, errors.Wrap(ErrInvalidParam, "dsn")
	}

	return &rabbitmqChecker{
		dsn: dsn,
	}, nil
}

func (c *rabbitmqChecker) check(context.Context) error {
	conn, err := amqp091.Dial(c.dsn)
	if err != nil {
		return errors.Wrap(err, "rabbitmq: open conn")
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		return errors.Wrap(err, "rabbitmq: open channel")
	}
	defer channel.Close()

	return nil
}
