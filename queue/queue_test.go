package queue

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMessage(t *testing.T) {
	type MyObj struct {
		Text string
	}
	obj := MyObj{Text: "test data"}
	b, err := json.Marshal(obj)
	assert.NoError(t, err)

	type args struct {
		obj      interface{}
		metadata map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    *Message
		wantErr bool
	}{
		{
			name: "error: marshal",
			args: args{
				obj: make(chan int),
			},
			wantErr: true,
		},
		{
			name: "success",
			args: args{
				obj:      MyObj{Text: "test data"},
				metadata: map[string]string{"input1": "value 1", "input2": "value 2"},
			},
			want: &Message{
				Body:     b,
				Metadata: map[string]string{"input1": "value 1", "input2": "value 2"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMessage(tt.args.obj, tt.args.metadata)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}
