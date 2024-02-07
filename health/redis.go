package health

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type redisChecker struct {
	addr     string
	password string
	dbNumber int
}

func newRedisChecker(addr, password string, dbNumber int) (*redisChecker, error) {
	if strings.TrimSpace(addr) == "" {
		return nil, errors.Wrap(ErrInvalidParam, "addr")
	}

	return &redisChecker{
		addr:     addr,
		password: password,
		dbNumber: dbNumber,
	}, nil
}

func (c *redisChecker) check(ctx context.Context) error {
	rdb := redis.NewClient(&redis.Options{ //nolint:exhaustruct
		Addr:     c.addr,
		Password: c.password,
		DB:       c.dbNumber,
	})
	defer rdb.Close()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return errors.Wrap(err, "redis: ping")
	}

	return nil
}
