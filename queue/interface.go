package queue

import "context"

//go:generate mockgen -destination=interface_mock.go -package=queue . QueueI
type QueueI interface {
	Produce(ctx context.Context, eventName string, object any) error
	Consume(ctx context.Context, queue string, fn func(ctx context.Context, data []byte) error)
	ConsumeStream(ctx context.Context, eventName string, fn func(ctx context.Context, data []byte) error)
	ConsumeWithHeaders(ctx context.Context, queue string, fn func(ctx context.Context, data []byte, headers map[string]interface{}) error)
}
