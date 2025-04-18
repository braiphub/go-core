package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/braiphub/go-core/log"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
)

// assert meets contract
var _ QueueI = &RabbitMQConnection{}

type ErrorHandlerFunc func(queue string, msg []byte, headers map[string]any, err error)

type DeferPanicHandlerFunc func(queue string)

type RabbitMQQueueConfig struct {
	Name       string
	Exchange   string
	RoutingKey string
	Arguments  amqp.Table
}

type RabbitMQExchangeConfig struct {
	Name string
	Type string
}

type RabbitMQConnection struct {
	config            Config
	logger            log.LoggerI
	conn              *amqp.Connection
	channel           *amqp.Channel
	terminateCh       chan interface{}
	errorHandler      ErrorHandlerFunc
	deferPanicHandler DeferPanicHandlerFunc
}

type Config struct {
	Dsn         string
	ServiceName string
	Exchange    string
}

const (
	reconnectDelay       = 5 * time.Second
	maxReconnectAttempts = 5
	publishTimeout       = 5 * time.Second
)

var ErrEmptyObject = errors.New("input object is empty")

func NewRabbitMQConnection(config Config, opts ...func(*RabbitMQConnection)) *RabbitMQConnection {
	rabbitMQ := &RabbitMQConnection{
		config:      config,
		logger:      nil,
		conn:        nil,
		channel:     nil,
		terminateCh: make(chan interface{}),
	}

	for _, o := range opts {
		o(rabbitMQ)
	}

	rabbitMQ.validate()

	return rabbitMQ
}

func (r *RabbitMQConnection) validate() {
	if r.logger == nil {
		panic("rabbit-mq: missing logger")
	}
}

func (r *RabbitMQConnection) Connect(ctx context.Context) error {
	var err error

	tryCount := 0
	for {
		r.conn, err = amqp.Dial(r.config.Dsn)
		if err == nil {
			break
		}

		msg := fmt.Sprintf("Falha ao conectar com RabbitMQ: %s. Tentativa %d de %d", err, tryCount+1, maxReconnectAttempts)
		r.logger.WithContext(ctx).Error(msg, nil)
		if tryCount == maxReconnectAttempts {
			return errors.Wrap(err, "amqp dial")
		}

		time.Sleep(reconnectDelay)
		tryCount++
	}

	r.channel, err = r.conn.Channel()
	if err != nil {
		return errors.Wrap(err, "channel open")
	}
	r.logger.WithContext(ctx).Info("Conectado ao RabbitMQ")

	go r.handleReconnect(ctx)

	return nil
}

func (r *RabbitMQConnection) Setup(
	ctx context.Context,
	exchange RabbitMQExchangeConfig,
	queues []RabbitMQQueueConfig,
) error {
	channel, err := r.conn.Channel()
	if err != nil {
		return errors.Wrap(err, "channel open")
	}
	defer func(channel *amqp.Channel) {
		err := channel.Close()
		if err != nil {
			r.logger.Fatal("Erro ao fechar canal", log.Error(err))
		}
	}(channel)

	if err := r.DeclareExchange(ctx, exchange.Name, exchange.Type); err != nil {
		return errors.Wrap(err, "declare exchange")
	}

	for _, queue := range queues {
		if err := r.DeclareQueue(queue); err != nil {
			return errors.Wrap(err, "declare queue")
		}

		if err := r.BindQueue(ctx, queue.Name, queue.Exchange, queue.RoutingKey); err != nil {
			return errors.Wrap(err, "queue bind")
		}
	}

	return nil
}

func (r *RabbitMQConnection) DeclareExchange(ctx context.Context, exchangeName string, exchangeType string) error {
	channel, err := r.conn.Channel()
	if err != nil {
		return errors.Wrap(err, "channel open")
	}
	defer func(channel *amqp.Channel) {
		err := channel.Close()
		if err != nil {
			r.logger.WithContext(ctx).Fatal("Erro ao fechar canal", log.Error(err))
		}
	}(channel)

	err = channel.ExchangeDeclare(exchangeName, exchangeType, true, false, false, false, nil)
	if err != nil {
		return errors.Wrap(err, "exchange declare")
	}

	return nil
}

func (r *RabbitMQConnection) DeclareQueue(queue RabbitMQQueueConfig) error {
	channel, err := r.conn.Channel()
	if err != nil {
		return errors.Wrap(err, "channel open")
	}
	defer func(channel *amqp.Channel) {
		if err := channel.Close(); err != nil {
			return
		}
	}(channel)

	_, err = channel.QueueDeclare(
		queue.Name,
		true,
		false,
		false,
		false,
		queue.Arguments,
	)
	if err != nil {
		return errors.Wrap(err, "queue declare")
	}

	return nil
}

func (r *RabbitMQConnection) BindQueue(
	ctx context.Context,
	queueName string,
	exchangeName string,
	routingKey string,
) error {
	channel, err := r.conn.Channel()
	if err != nil {
		return errors.Wrap(err, "channel open")
	}
	defer func(channel *amqp.Channel) {
		if err := channel.Close(); err != nil {
			r.logger.WithContext(ctx).Warn("Erro ao fechar canal: %s", log.Error(err))
		}
	}(channel)

	err = channel.QueueBind(queueName, routingKey, exchangeName, false, nil)
	if err != nil {
		return errors.Wrap(err, "queue bind")
	}

	return nil
}

func (r *RabbitMQConnection) Close() {
	r.terminateCh <- true

	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			return
		}
	}
}

func (r *RabbitMQConnection) handleReconnect(ctx context.Context) {
	select {
	case amqpErr := <-r.conn.NotifyClose(make(chan *amqp.Error)):
		var err error

		// terminate current
		r.logger.WithContext(ctx).Error("closing connection: %s", amqpErr)
		r.conn.Close()

		// reconnect loop
		for {
			r.conn, err = amqp.Dial(r.config.Dsn)
			if err != nil {
				r.logger.WithContext(ctx).Error("connect error: %s", err)
				time.Sleep(reconnectDelay)

				continue
			} else {
				r.channel, err = r.conn.Channel()
				if err != nil {
					r.logger.WithContext(ctx).Error("channel open error: %s", err)
					time.Sleep(reconnectDelay)

					continue
				}

				r.logger.WithContext(ctx).Warn("rabbit-mq reconnected")
				go r.handleReconnect(ctx)

				break
			}
		}

	case <-r.terminateCh:
		r.logger.WithContext(ctx).Info("terminated")
	}
}

func (r *RabbitMQConnection) channelConsumer(ctx context.Context, queue string, options ConsumeOptions) <-chan amqp.Delivery {
	args := make(amqp.Table)

	if options.Priority != nil {
		args["x-priority"] = *options.Priority
	}

	outChannel := make(chan amqp.Delivery)

	go func(outChannel chan amqp.Delivery) {
		for {
			if r.conn == nil || r.conn.IsClosed() {
				time.Sleep(reconnectDelay)

				continue
			}

			channel, err := r.conn.Channel()
			if err != nil {
				r.logger.WithContext(ctx).Error("open channel error: %s\n", err)
				time.Sleep(reconnectDelay)

				continue
			}

			if options.PrefetchCount != nil {
				channel.Qos(*options.PrefetchCount, 0, false)
			}

			messageCh, err := channel.Consume(
				queue,                // queue
				r.config.ServiceName, // consumer
				false,                // auto-ack
				false,                // exclusive
				false,                // no-local
				false,                // no-wait
				args,                 // args
			)
			if err != nil {
				r.logger.WithContext(ctx).Error("channel consume error: %s\n", err)
				time.Sleep(reconnectDelay)

				continue
			}

			for msg := range messageCh {
				outChannel <- msg
			}
		}
	}(outChannel)

	return outChannel
}

func (r *RabbitMQConnection) channelStreamConsumer(ctx context.Context, routingKey string) <-chan amqp.Delivery {
	outChannel := make(chan amqp.Delivery)

	go func(outChannel chan amqp.Delivery) {
		channel, err := r.conn.Channel()
		if err != nil {
			r.logger.WithContext(ctx).Error("open channel error: %s\n", err)

			return
		}
		defer channel.Close()

		queueName := fmt.Sprintf("%s.%s.stream.%s", r.config.Exchange, routingKey, uuid.New().String())

		queue, err := channel.QueueDeclare(queueName, false, true, false, false, nil)
		if err != nil {
			r.logger.WithContext(ctx).Error("stream queue declare", err)

			return
		}

		if err := r.BindQueue(ctx, queue.Name, r.config.Exchange, routingKey); err != nil {
			r.logger.WithContext(ctx).Error("stream queue bind", err)

			return
		}

		messageCh, err := channel.Consume(
			queue.Name,           // queue
			r.config.ServiceName, // consumer name
			false,                // auto-ack
			false,                // exclusive
			false,                // no-local
			false,                // no-wait
			nil,                  // args
		)
		if err != nil {
			r.logger.WithContext(ctx).Error("channel consume error: %s\n", err)

			return
		}

		for {
			select {
			case msg := <-messageCh:
				outChannel <- msg

			case <-ctx.Done():
				return
			}
		}
	}(outChannel)

	return outChannel
}
