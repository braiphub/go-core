package redis

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type RedisAdapter struct {
	host     string
	port     uint16
	username string
	password string
	client   ClientI
}

//go:generate mockgen -destination=mocks/redis_mock.go -package=mocks . ClientI
type ClientI interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Exists(ctx context.Context, keys ...string) *redis.IntCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}

var (
	ErrMissingParam = errors.New("parameter is missing")
	ErrConnectTest  = errors.New("test connection failed")
	ErrEmptyKey     = errors.New("can't perform operation with an empty key")
)

func NewRedisAdapter(host string, port uint16, username, password string) (*RedisAdapter, error) {
	switch {
	case host == "":
		return nil, errors.Wrap(ErrMissingParam, "host")

	case port == 0:
		return nil, errors.Wrap(ErrMissingParam, "port")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Username: username,
		Password: password,
	})

	adapter := &RedisAdapter{
		host:     host,
		port:     port,
		username: username,
		password: password,
		client:   client,
	}

	return adapter, nil
}

func (adapter *RedisAdapter) TestConnection(ctx context.Context) error {
	if err := adapter.client.Exists(ctx, "test-connection").Err(); err != nil {
		return errors.Wrap(err, "test connection")
	}

	return nil
}

func (adapter *RedisAdapter) Set(ctx context.Context, key string, value interface{}, duration time.Duration) error {
	if key == "" {
		return ErrEmptyKey
	}
	if err := adapter.client.Set(ctx, key, value, duration).Err(); err != nil {
		return errors.Wrap(err, "set redis value")
	}

	return nil
}

func (adapter *RedisAdapter) Get(ctx context.Context, key string) ([]byte, error) {
	var result []byte

	if err := adapter.get(ctx, key, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (adapter *RedisAdapter) GetString(ctx context.Context, key string) (string, error) {
	var result string

	if err := adapter.get(ctx, key, &result); err != nil {
		return "", err
	}

	return result, nil
}

func (adapter *RedisAdapter) GetInt(ctx context.Context, key string) (int, error) {
	var result int

	if err := adapter.get(ctx, key, &result); err != nil {
		return 0, err
	}

	return result, nil
}

func (adapter *RedisAdapter) GetUint(ctx context.Context, key string) (uint, error) {
	var result uint

	if err := adapter.get(ctx, key, &result); err != nil {
		return 0, err
	}

	return result, nil
}

func (adapter *RedisAdapter) get(ctx context.Context, key string, output interface{}) error {
	if key == "" {
		return ErrEmptyKey
	}
	cmd := adapter.client.Get(ctx, key)
	if errors.Is(cmd.Err(), redis.Nil) {
		return sql.ErrNoRows
	}
	if cmd.Err() != nil {
		return errors.Wrap(cmd.Err(), "get redis value")
	}
	if err := cmd.Scan(output); err != nil {
		return errors.Wrap(err, "redis scan to interface")
	}

	return nil
}

func (adapter *RedisAdapter) Delete(ctx context.Context, key string) error {
	if key == "" {
		return ErrEmptyKey
	}

	cmd := adapter.client.Del(ctx, key)
	if errors.Is(cmd.Err(), redis.Nil) {
		return nil
	}
	if cmd.Err() != nil {
		return errors.Wrap(cmd.Err(), "delete redis value")
	}

	return nil
}
