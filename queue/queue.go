package queue

import (
	"context"
	"encoding/json"

	_ "github.com/golang/mock/mockgen/model"
)

//go:generate mockgen -destination=mocks/queue_mock.go -package=mocks . QueueI

type QueueI interface {
	// for rabbitmq topic must be an exchange name
	Publish(ctx context.Context, topic string, msg Message) error
	// for rabbitmq topic must be a queue name and retry must be an exchange
	Subscribe(ctx context.Context, topic, retry string, f func(Message) error)
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

	msg := &Message{
		Metadata: metadata,
		Body:     b,
	}

	return msg, nil
}

func (m *Message) Marshal() []byte {
	b, err := json.Marshal(m)
	if err != nil {
		return []byte{}
	}

	return b
}
