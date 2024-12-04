package queue

import (
	"context"
	"encoding/json"
)

//go:generate mockgen -destination=interface_mock.go -package=queue . QueueI
type QueueI interface {
	Produce(ctx context.Context, eventName string, msg Message) error
	Consume(ctx context.Context, queue string, fn func(ctx context.Context, msg Message) error)
	ConsumeStream(ctx context.Context, eventName string, fn func(ctx context.Context, msg Message) error)
}

type Message struct {
	Headers map[string]any `json:"metadata"`
	Body    []byte         `json:"body"`
}

func NewMessage(obj any, headers map[string]any) (*Message, error) {
	buf, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	if headers == nil {
		headers = make(map[string]any)
	}

	msg := &Message{
		Headers: headers,
		Body:    buf,
	}

	return msg, nil
}
