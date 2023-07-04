package rabbitmq

import (
	"testing"
	"time"

	"github.com/braiphub/go-core/cache"
	"github.com/braiphub/go-core/log"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRabbitMQ_SetIdempotencyChecker(t *testing.T) {
	type fields struct {
		dsn           string
		logger        log.Logger
		loggerAdapter *LoggerAdapter
	}
	type args struct {
		cacheKey     string
		messageLabel string
		duration     time.Duration
		cache        cache.Cacherer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "error: key is empty",
			wantErr: true,
		},
		{
			name: "error: label is empty",
			args: args{
				cacheKey: "message-id",
			},
			wantErr: true,
		},
		{
			name: "error: expire time is empty",
			args: args{
				cacheKey:     "message-id",
				messageLabel: "label",
			},
			wantErr: true,
		},
		{
			name: "error: cache component is empty",
			args: args{
				cacheKey:     "message-id",
				messageLabel: "label",
				duration:     time.Second,
			},
			wantErr: true,
		},
		{
			name: "success",
			args: args{
				cacheKey:     "message-id",
				messageLabel: "label",
				duration:     time.Second,
				cache:        cache.NewMockCacherer(gomock.NewController(t)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rmq := &RabbitMQ{
				dsn:           tt.fields.dsn,
				logger:        tt.fields.logger,
				loggerAdapter: tt.fields.loggerAdapter,
			}
			if err := rmq.SetIdempotencyChecker(tt.args.cacheKey, tt.args.messageLabel, tt.args.duration, tt.args.cache); (err != nil) != tt.wantErr {
				t.Errorf("RabbitMQ.SetIdempotencyChecker() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRabbitMQ_SetIdempotencyChecker_SuccessCase(t *testing.T) {
	cache := cache.NewMockCacherer(gomock.NewController(t))
	expect := &RabbitMQ{
		messageIDLabel: "message-id",
		idempotencyChecker: &idempotencyChecker{
			cacheKey:     "cache.key",
			messageLabel: "message-id",
			duration:     time.Minute,
			cache:        cache,
		},
	}

	rmq := &RabbitMQ{}
	err := rmq.SetIdempotencyChecker("cache.key", "message-id", time.Minute, cache)
	assert.NoError(t, err)
	assert.Equal(t, expect, rmq)
}
