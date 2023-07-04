package rabbitmq

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/braiphub/go-core/cache"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
)

func Test_idempotencyChecker_SetProcessed(t *testing.T) {
	type fields struct {
		cacheKey     string
		messageLabel string
		duration     time.Duration
		cache        func() cache.Cacherer
	}
	type args struct {
		ctx       context.Context
		messageID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "error: cache set key",
			fields: fields{
				cacheKey: "path",
				duration: time.Hour,
				cache: func() cache.Cacherer {
					cache := cache.NewMockCacherer(gomock.NewController(t))
					cache.EXPECT().Set(gomock.Any(), "path.msg-1234", true, time.Hour).Return(errors.New("unknonw"))

					return cache
				},
			},
			args: args{
				messageID: "msg-1234",
			},
			wantErr: true,
		},
		{
			name: "success",
			fields: fields{
				cacheKey: "path",
				duration: time.Hour,
				cache: func() cache.Cacherer {
					cache := cache.NewMockCacherer(gomock.NewController(t))
					cache.EXPECT().Set(gomock.Any(), "path.msg-1234", true, time.Hour)

					return cache
				},
			},
			args: args{
				messageID: "msg-1234",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := &idempotencyChecker{
				cacheKey:     tt.fields.cacheKey,
				messageLabel: tt.fields.messageLabel,
				duration:     tt.fields.duration,
				cache:        tt.fields.cache(),
			}
			if err := checker.SetProcessed(tt.args.ctx, tt.args.messageID); (err != nil) != tt.wantErr {
				t.Errorf("idempotencyChecker.SetProcessed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_idempotencyChecker_CanProcess(t *testing.T) {
	type fields struct {
		cacheKey     string
		messageLabel string
		duration     time.Duration
		cache        func() cache.Cacherer
	}
	type args struct {
		ctx       context.Context
		messageID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "error: cache error",
			fields: fields{
				cache: func() cache.Cacherer {
					cache := cache.NewMockCacherer(gomock.NewController(t))
					cache.EXPECT().Get(nil, ".").Return([]byte{}, errors.New("unknonw"))

					return cache
				},
			},
			wantErr: true,
		},
		{
			name: "error: already processed",
			fields: fields{
				cache: func() cache.Cacherer {
					cache := cache.NewMockCacherer(gomock.NewController(t))
					cache.EXPECT().Get(nil, ".").Return([]byte{}, nil)

					return cache
				},
			},
			wantErr: true,
		},
		{
			name: "success: can process",
			fields: fields{
				cache: func() cache.Cacherer {
					cache := cache.NewMockCacherer(gomock.NewController(t))
					cache.EXPECT().Get(nil, ".").Return([]byte{}, sql.ErrNoRows)

					return cache
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := &idempotencyChecker{
				cacheKey:     tt.fields.cacheKey,
				messageLabel: tt.fields.messageLabel,
				duration:     tt.fields.duration,
				cache:        tt.fields.cache(),
			}
			if err := checker.CanProcess(tt.args.ctx, tt.args.messageID); (err != nil) != tt.wantErr {
				t.Errorf("idempotencyChecker.CanProcess() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
