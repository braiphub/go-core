package ddotl

import (
	"context"
	"errors"
	"testing"

	"github.com/braiphub/go-core/log"
	"github.com/braiphub/go-core/tracer/dd-otl/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type mockedTrace struct{}

func (mockedTrace) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return context.Background(), otlSpanMocked{}
}

type otlSpanMocked struct {
	trace.Span
}

func (otlSpanMocked) AddEvent(name string, options ...trace.EventOption) {}
func (otlSpanMocked) End(options ...trace.SpanEndOption)                 {}
func (otlSpanMocked) SetStatus(code codes.Code, description string)      {}
func (otlSpanMocked) SetAttributes(kv ...attribute.KeyValue)             {}

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	logger := log.NewMockLogger(ctrl)

	type args struct {
		ctx            context.Context
		serviceName    string
		version        string
		echoContextKey string
		logger         log.Logger
	}
	tests := []struct {
		name    string
		args    args
		want    *DataDogOTL
		wantErr bool
	}{
		{
			name: "error: nil context",
			args: args{
				ctx:    nil,
				logger: logger,
			},
			wantErr: true,
		},
		{
			name: "error: missing logger",
			args: args{
				ctx:    context.Background(),
				logger: nil,
			},
			wantErr: true,
		},
		{
			name: "success case",
			args: args{
				ctx:    context.Background(),
				logger: logger,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.ctx, tt.args.serviceName, tt.args.version, tt.args.echoContextKey, tt.args.logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr == false {
				assert.NotNil(t, got)
			}
		})
	}
}

func TestDataDogOTL_NewSpan(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		otl := &DataDogOTL{
			tracer: mockedTrace{},
		}
		otl.NewSpan(context.Background(), "span name")
	})
}

func TestDataDogOTL_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	logger := log.NewMockLogger(ctrl)
	logger.EXPECT().Error("closing trace provider", gomock.Any())

	type fields struct {
		echoContextKey string
		tracer         trace.Tracer
		tracerProvider func() TraceProviderI
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "error: close",
			fields: fields{tracerProvider: func() TraceProviderI {
				ctrl := gomock.NewController(t)
				provider := mocks.NewMockTraceProviderI(ctrl)
				provider.EXPECT().Shutdown().Return(errors.New("unknown"))

				return provider
			}},
		},
		{
			name: "success",
			fields: fields{tracerProvider: func() TraceProviderI {
				ctrl := gomock.NewController(t)
				provider := mocks.NewMockTraceProviderI(ctrl)
				provider.EXPECT().Shutdown().Return(nil)

				return provider
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			otl := &DataDogOTL{
				echoContextKey: tt.fields.echoContextKey,
				tracer:         tt.fields.tracer,
				tracerProvider: tt.fields.tracerProvider(),
				logger:         logger,
			}
			otl.Close(context.Background())
		})
	}
}
