package queue

import (
	"context"
)

//go:generate mockgen -destination=interface_mock.go -package=queue . QueueI
type QueueI interface {
	Produce(ctx context.Context, eventName string, msg Message) error
	Consume(ctx context.Context, queue string, fn func(ctx context.Context, msg Message) error)
	ConsumeStream(ctx context.Context, eventName string, fn func(ctx context.Context, msg Message) error)
}
