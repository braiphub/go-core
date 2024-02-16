package datadogtracer

import (
	"context"
	"testing"

	"github.com/braiphub/go-core/log"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	logger := log.NewMockLoggerI(ctrl)

	type args struct {
		ctx         context.Context
		env         string
		serviceName string
		version     string
		logger      func() log.LoggerI
	}
	tests := []struct {
		name        string
		args        args
		want        *DatadogTracer
		wantErr     bool
		wantErrType error
	}{
		{
			name: "error: invalid param: context",
			args: args{
				ctx:         nil,
				env:         "local",
				serviceName: "app-name",
				version:     "v0",
				logger: func() log.LoggerI {
					return logger
				},
			},
			wantErr:     true,
			wantErrType: ErrMissingContext,
		},
		{
			name: "error: invalid param: env",
			args: args{
				ctx:         context.Background(),
				env:         "",
				serviceName: "app-name",
				version:     "v0",
				logger: func() log.LoggerI {
					return logger
				},
			},
			wantErr:     true,
			wantErrType: ErrMissingEnvironment,
		},
		{
			name: "error: invalid param: serviceName",
			args: args{
				ctx:         context.Background(),
				env:         "local",
				serviceName: "",
				version:     "v0",
				logger: func() log.LoggerI {
					return logger
				},
			},
			wantErr:     true,
			wantErrType: ErrMissingServiceName,
		},
		{
			name: "error: invalid param: version",
			args: args{
				ctx:         context.Background(),
				env:         "local",
				serviceName: "app-name",
				version:     "",
				logger: func() log.LoggerI {
					return logger
				},
			},
			wantErr:     true,
			wantErrType: ErrMissingServiceVersion,
		},
		{
			name: "error: invalid param: logger",
			args: args{
				ctx:         context.Background(),
				env:         "local",
				serviceName: "app-name",
				version:     "v0",
				logger: func() log.LoggerI {
					return nil
				},
			},
			wantErr:     true,
			wantErrType: ErrMissingLogger,
		},
		{
			name: "success",
			args: args{
				ctx:         context.Background(),
				env:         "local",
				serviceName: "app-name",
				version:     "v0",
				logger: func() log.LoggerI {
					return logger
				},
			},
			wantErr: false,
			want: &DatadogTracer{
				logger:      logger,
				env:         "local",
				serviceName: "app-name",
				version:     "v0",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.ctx, tt.args.env, tt.args.serviceName, tt.args.version, tt.args.logger())
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.wantErrType, err)
			if got != nil {
				got.EchoMiddleware = nil
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
