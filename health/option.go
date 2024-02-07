package health

import "github.com/pkg/errors"

func WithPostgres(dsn string) func(*Checker) error {
	return func(hc *Checker) error {
		pgChecker, err := newPostgresChecker(dsn)
		if err != nil {
			return errors.Wrap(err, "postgres checker init")
		}

		hc.checkers = append(hc.checkers, pgChecker)

		return nil
	}
}

func WithRedis(addr, password string, db int) func(*Checker) error {
	return func(hc *Checker) error {
		redisChecker, err := newRedisChecker(addr, password, db)
		if err != nil {
			return errors.Wrap(err, "redis checker init")
		}

		hc.checkers = append(hc.checkers, redisChecker)

		return nil
	}
}

func WithRabbitMQ(dsn string) func(*Checker) error {
	return func(hc *Checker) error {
		rabbitMQChecker, err := newRabbitChecker(dsn)
		if err != nil {
			return errors.Wrap(err, "rabbitmq checker init")
		}

		hc.checkers = append(hc.checkers, rabbitMQChecker)

		return nil
	}
}
