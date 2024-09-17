package redis

import (
	"context"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/braiphub/go-core/cache/redis/mocks"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestNewRedisAdapter(t *testing.T) {
	type args struct {
		host     string
		port     uint16
		password string
	}
	tests := []struct {
		name    string
		args    args
		want    *RedisAdapter
		wantErr bool
	}{
		{
			name:    "error: missing information: host addr",
			wantErr: true,
		},
		{
			name: "error: missing information: connection port",
			args: args{
				host: "localhost",
			},
			wantErr: true,
		},
		{
			name: "success",
			args: args{
				host:     "127.0.0.1",
				port:     1234,
				password: "pwd",
			},
			want: &RedisAdapter{
				host:     "127.0.0.1",
				port:     1234,
				password: "pwd",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRedisAdapter(tt.args.host, tt.args.port, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRedisAdapter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil && got != nil {
				tt.want.client = got.client
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConnection(t *testing.T) {
	type fields struct {
		host     string
		port     uint16
		password string
		client   func() ClientI
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "error: offline / any other",
			fields: fields{
				client: func() ClientI {
					err := net.OpError{Op: "dial", Net: "tcp", Err: errors.New("unknown error")}
					cmd := redis.NewIntCmd(context.Background())
					cmd.SetErr(err.Err)

					client := mocks.NewMockClientI(gomock.NewController(t))
					client.EXPECT().Exists(gomock.Any(), gomock.Any()).Return(cmd)

					return client
				},
			},
			wantErr: true,
		},
		{
			name: "success",
			fields: fields{
				client: func() ClientI {
					cmd := redis.NewIntCmd(context.Background())
					client := mocks.NewMockClientI(gomock.NewController(t))
					client.EXPECT().Exists(gomock.Any(), gomock.Any()).Return(cmd)

					return client
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &RedisAdapter{
				host:     tt.fields.host,
				port:     tt.fields.port,
				password: tt.fields.password,
				client:   tt.fields.client(),
			}
			if err := adapter.TestConnection(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("RedisAdapter.TestConnection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRedisAdapter_Set(t *testing.T) {
	type fields struct {
		host     string
		port     uint16
		password string
		client   func() ClientI
	}
	type args struct {
		ctx      context.Context
		key      string
		value    interface{}
		duration time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "error: empty key",
			fields: fields{
				client: func() ClientI { return nil },
			},
			wantErr: true,
		},
		{
			name: "error: set value",
			fields: fields{
				client: func() ClientI {
					cmd := redis.NewStatusCmd(context.Background())
					cmd.SetErr(errors.New("unknown"))

					client := mocks.NewMockClientI(gomock.NewController(t))
					client.EXPECT().Set(nil, "key", "val", gomock.Any()).Return(cmd)

					return client
				},
			},
			args: args{
				key:   "key",
				value: "val",
			},
			wantErr: true,
		},
		{
			name: "success",
			fields: fields{
				client: func() ClientI {
					cmd := redis.NewStatusCmd(context.Background())

					client := mocks.NewMockClientI(gomock.NewController(t))
					client.EXPECT().Set(nil, "key", "val", gomock.Any()).Return(cmd)

					return client
				},
			},
			args: args{
				key:   "key",
				value: "val",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &RedisAdapter{
				host:     tt.fields.host,
				port:     tt.fields.port,
				password: tt.fields.password,
				client:   tt.fields.client(),
			}
			if err := adapter.Set(tt.args.ctx, tt.args.key, tt.args.value, tt.args.duration); (err != nil) != tt.wantErr {
				t.Errorf("RedisAdapter.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRedisAdapter_get(t *testing.T) {
	var bytesOut []byte

	type fields struct {
		host     string
		port     uint16
		password string
		client   func() ClientI
	}
	type args struct {
		ctx    context.Context
		key    string
		output interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "error: empty key",
			fields: fields{
				client: func() ClientI { return nil },
			},
			wantErr: true,
		},
		{
			name: "error: get value: key not found",
			fields: fields{
				client: func() ClientI {
					cmd := redis.NewStringCmd(context.Background())
					cmd.SetErr(redis.Nil)

					client := mocks.NewMockClientI(gomock.NewController(t))
					client.EXPECT().Get(nil, "key").Return(cmd)

					return client
				},
			},
			args: args{
				key: "key",
			},
			wantErr: true,
		},
		{
			name: "error: get value: unknown err",
			fields: fields{
				client: func() ClientI {
					cmd := redis.NewStringCmd(context.Background())
					cmd.SetErr(errors.New("unknown"))

					client := mocks.NewMockClientI(gomock.NewController(t))
					client.EXPECT().Get(nil, "key").Return(cmd)

					return client
				},
			},
			args: args{
				key: "key",
			},
			wantErr: true,
		},
		{
			name: "error: scan to invalid interface",
			fields: fields{
				client: func() ClientI {
					cmd := redis.NewStringCmd(context.Background())

					client := mocks.NewMockClientI(gomock.NewController(t))
					client.EXPECT().Get(nil, "key").Return(cmd)

					return client
				},
			},
			args: args{
				key: "key",
			},
			wantErr: true,
		},
		{
			name: "success",
			fields: fields{
				client: func() ClientI {
					cmd := redis.NewStringCmd(context.Background())

					client := mocks.NewMockClientI(gomock.NewController(t))
					client.EXPECT().Get(nil, "key").Return(cmd)

					return client
				},
			},
			args: args{
				key:    "key",
				output: &bytesOut,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &RedisAdapter{
				host:     tt.fields.host,
				port:     tt.fields.port,
				password: tt.fields.password,
				client:   tt.fields.client(),
			}
			if err := adapter.get(tt.args.ctx, tt.args.key, tt.args.output); (err != nil) != tt.wantErr {
				t.Errorf("RedisAdapter.get() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRedisAdapter_Get(t *testing.T) {
	type fields struct {
		host     string
		port     uint16
		password string
		client   func() ClientI
	}
	type args struct {
		ctx context.Context
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "error: get value: empty key",
			fields: fields{
				client: func() ClientI { return nil },
			},
			wantErr: true,
		},
		{
			name: "success",
			fields: fields{
				client: func() ClientI {
					cmd := redis.NewStringCmd(context.Background())
					cmd.SetVal("value")

					client := mocks.NewMockClientI(gomock.NewController(t))
					client.EXPECT().Get(nil, "key").Return(cmd)

					return client
				},
			},
			args: args{
				key: "key",
			},
			want: []byte("value"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &RedisAdapter{
				host:     tt.fields.host,
				port:     tt.fields.port,
				password: tt.fields.password,
				client:   tt.fields.client(),
			}
			got, err := adapter.Get(tt.args.ctx, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("RedisAdapter.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RedisAdapter.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedisAdapter_GetString(t *testing.T) {
	type fields struct {
		host     string
		port     uint16
		password string
		client   func() ClientI
	}
	type args struct {
		ctx context.Context
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "error: get value: empty key",
			fields: fields{
				client: func() ClientI { return nil },
			},
			wantErr: true,
		},
		{
			name: "success",
			fields: fields{
				client: func() ClientI {
					cmd := redis.NewStringCmd(context.Background())
					cmd.SetVal("value")

					client := mocks.NewMockClientI(gomock.NewController(t))
					client.EXPECT().Get(nil, "key").Return(cmd)

					return client
				},
			},
			args: args{
				key: "key",
			},
			want: "value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &RedisAdapter{
				host:     tt.fields.host,
				port:     tt.fields.port,
				password: tt.fields.password,
				client:   tt.fields.client(),
			}
			got, err := adapter.GetString(tt.args.ctx, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("RedisAdapter.GetString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RedisAdapter.GetString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedisAdapter_GetInt(t *testing.T) {
	type fields struct {
		host     string
		port     uint16
		password string
		client   func() ClientI
	}
	type args struct {
		ctx context.Context
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "error: get value: empty key",
			fields: fields{
				client: func() ClientI { return nil },
			},
			wantErr: true,
		},
		{
			name: "success",
			fields: fields{
				client: func() ClientI {
					cmd := redis.NewStringCmd(context.Background())
					cmd.SetVal("-2")

					client := mocks.NewMockClientI(gomock.NewController(t))
					client.EXPECT().Get(nil, "key").Return(cmd)

					return client
				},
			},
			args: args{
				key: "key",
			},
			want: -2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &RedisAdapter{
				host:     tt.fields.host,
				port:     tt.fields.port,
				password: tt.fields.password,
				client:   tt.fields.client(),
			}
			got, err := adapter.GetInt(tt.args.ctx, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("RedisAdapter.GetInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RedisAdapter.GetInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedisAdapter_GetUint(t *testing.T) {
	type fields struct {
		host     string
		port     uint16
		password string
		client   func() ClientI
	}
	type args struct {
		ctx context.Context
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    uint
		wantErr bool
	}{
		{
			name: "error: get value: empty key",
			fields: fields{
				client: func() ClientI { return nil },
			},
			wantErr: true,
		},
		{
			name: "success",
			fields: fields{
				client: func() ClientI {
					cmd := redis.NewStringCmd(context.Background())
					cmd.SetVal("2")

					client := mocks.NewMockClientI(gomock.NewController(t))
					client.EXPECT().Get(nil, "key").Return(cmd)

					return client
				},
			},
			args: args{
				key: "key",
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := &RedisAdapter{
				host:     tt.fields.host,
				port:     tt.fields.port,
				password: tt.fields.password,
				client:   tt.fields.client(),
			}
			got, err := adapter.GetUint(tt.args.ctx, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("RedisAdapter.GetUint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RedisAdapter.GetUint() = %v, want %v", got, tt.want)
			}
		})
	}
}
