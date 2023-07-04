package rabbitmq

import (
	"context"
	"database/sql"
	"time"

	"github.com/braiphub/go-core/cache"
	"github.com/pkg/errors"
)

type idempotencyChecker struct {
	cacheKey     string
	messageLabel string
	duration     time.Duration
	cache        cache.Cacherer
}

var ErrMessageAlreadyProcessed = errors.New("message is already processed")

func (checker *idempotencyChecker) SetProcessed(ctx context.Context, messageID string) error {
	key := checker.cacheKey + "." + messageID

	if err := checker.cache.Set(ctx, key, true, checker.duration); err != nil {
		return errors.Wrap(err, "set key processed")
	}

	return nil
}

func (checker *idempotencyChecker) CanProcess(ctx context.Context, messageID string) error {
	key := checker.cacheKey + "." + messageID

	_, err := checker.cache.Get(ctx, key)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "checking if key is processed")
	}

	return ErrMessageAlreadyProcessed
}
