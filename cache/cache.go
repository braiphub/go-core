package cache

import (
	"context"
	"time"
)

//go:generate mockgen -destination=cache_mock.go -package=cache . Cacherer

type Cacherer interface {
	TestConnection(ctx context.Context) error
	Set(ctx context.Context, key string, value interface{}, duration time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
	GetString(ctx context.Context, key string) (string, error)
	GetInt(ctx context.Context, key string) (int, error)
	GetUint(ctx context.Context, key string) (uint, error)
	Delete(ctx context.Context, key string) error
}
