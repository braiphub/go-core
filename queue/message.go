package queue

import "encoding/json"

type Message struct {
	Headers map[string]any `json:"metadata"`
	Body    []byte         `json:"body"`
}

func NewMessage(data any, headers map[string]any) (*Message, error) {
	var buf []byte

	switch v := data.(type) {
	case []byte:
		buf = v

	default:
		var err error
		buf, err = json.Marshal(data)
		if err != nil {
			return nil, err
		}
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
