package queue

import (
	"context"
	"encoding/json"
	"time"

	"github.com/braiphub/go-core/cache"
	_ "github.com/golang/mock/mockgen/model"
)

//go:generate mockgen -destination=queue_mock.go -package=queue . Queuer

type Queuer interface {
	SetIdempotencyChecker(cacheKey, messageLabel string, duration time.Duration, cache cache.Cacherer) error
	// for rabbitmq topic must be an exchange name
	Publish(ctx context.Context, topic string, routingKeys []string, msg Message) error
	// for rabbitmq topic must be a queue name and retry must be an exchange
	Subscribe(ctx context.Context, topic, retry string, f func(context.Context, Message) error)
}

type Message struct {
	Metadata map[string]string `json:"metadata"`
	Body     []byte            `json:"body"`
}

func NewMessage(obj interface{}, metadata map[string]string) (*Message, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	if metadata == nil {
		metadata = map[string]string{}
	}

	msg := &Message{
		Metadata: metadata,
		Body:     b,
	}

	return msg, nil
}
