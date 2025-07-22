package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/braiphub/go-core/log"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sony/gobreaker/v2"
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
	databaseFallback  *GormFallback
	breaker           *gobreaker.CircuitBreaker[any]
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

	rabbitMQ.breaker = gobreaker.NewCircuitBreaker[any](
		gobreaker.Settings{
			Name:        "rabbitmq-connection",
			MaxRequests: 5,
			ReadyToTrip: func(c gobreaker.Counts) bool { return c.ConsecutiveFailures >= maxReconnectAttempts },
			OnStateChange: func(name string, from, to gobreaker.State) {
				if to == gobreaker.StateOpen {
					openErr := errors.Errorf("RabbitMQ circuit breaker changed from %s to %s", from.String(), to.String())
					rabbitMQ.logger.Error(openErr.Error(), openErr)

					return
				}

				rabbitMQ.logger.Warn(fmt.Sprintf("RabbitMQ circuit breaker changed from %s to %s", from.String(), to.String()))
			},
			Timeout: 1 * time.Second,
		},
	)

	return rabbitMQ
}

func (r *RabbitMQConnection) validate() {
	if r.logger == nil {
		panic("rabbit-mq: missing logger")
	}
}

func (r *RabbitMQConnection) Connect(ctx context.Context) error {
	var lastErr error

	for attempt := 0; attempt < maxReconnectAttempts; attempt++ {
		_, err := r.breaker.Execute(func() (any, error) {
			return nil, r.tryConnect(ctx)
		})

		if err != nil {
			lastErr = err
			msg := fmt.Sprintf("Error connecting to RabbitMQ: %v, (attempt: %d of %d)", err, attempt+1, maxReconnectAttempts)
			r.logger.WithContext(ctx).Error(msg, nil)
			time.Sleep(reconnectDelay)

			continue
		}

		r.logger.WithContext(ctx).Info("RabbitMQ connection established")
		go r.manageConnection(ctx)

		return nil
	}

	r.logger.
		WithContext(ctx).
		Warn("RabbitMQ unavailable; entering degraded mode", log.Error(lastErr))

	go r.manageConnection(ctx)

	return nil
}

func (r *RabbitMQConnection) Setup(
	ctx context.Context,
	exchange RabbitMQExchangeConfig,
	queues []RabbitMQQueueConfig,
) error {
	if r.conn == nil || r.conn.IsClosed() {
		r.logger.WithContext(ctx).Warn("RabbitMQ offline – skipping setup")
		return nil
	}

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

func (r *RabbitMQConnection) manageConnection(ctx context.Context) {
	for {
		if r.conn == nil {
			_, err := r.breaker.Execute(func() (any, error) {
				return nil, r.tryConnect(ctx)
			})
			if err != nil {
				r.logger.WithContext(ctx).Error("erro ao conectar, retry em", err)
				select {
				case <-time.After(reconnectDelay):
					continue
				case <-r.terminateCh:
					return
				}
			}
		}

		closeCh := r.conn.NotifyClose(make(chan *amqp.Error, 1))
		select {
		case amqpErr := <-closeCh:
			r.logger.WithContext(ctx).Error("RabbitMQ connection closed", amqpErr)
			_ = r.conn.Close()
			r.conn, r.channel = nil, nil

			continue

		case <-r.terminateCh:
			if r.conn != nil {
				_ = r.conn.Close()
			}
			return
		}
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

func (r *RabbitMQConnection) IsClosed() bool {
	return r.breaker.State() == gobreaker.StateClosed
}

func (r *RabbitMQConnection) tryConnect(ctx context.Context) error {
	conn, err := amqp.Dial(r.config.Dsn)
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return err
	}

	r.conn, r.channel = conn, ch
	r.logger.WithContext(ctx).Info("rabbit-mq connection established")
	return nil
}
