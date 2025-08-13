package with_dlq_backup

import (
	"context"
	"errors"
	"testing"

	"github.com/braiphub/go-core/log"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type testEventType struct {
	SampleValue string `json:"sample_value"`
}

func (testEventType) EventType() string {
	return "test_event"
}

func TestEventBus_Register(t *testing.T) {
	bus := &EventBus{
		registeredEvents: make(map[string][]EventRegisterConfig),
	}

	eventName := testEventType{}.EventType()
	func1 := func(ctx context.Context, event EventInterface) error { return nil }
	func2 := func(ctx context.Context, event EventInterface) error { return nil }
	func3 := func(ctx context.Context, event EventInterface) error { return nil }

	// error: input event should be be a pointer
	err := bus.Register(testEventType{}, func1)
	assert.Error(t, err)

	// success registering event and handlers
	err = bus.Register(&testEventType{}, func1)
	assert.NoError(t, err)
	err = bus.Register(&testEventType{}, func2)
	assert.NoError(t, err)
	assert.Len(t, bus.registeredEvents, 1)
	assert.Len(t, bus.registeredEvents[eventName], 2)

	// increment listenners
	err = bus.Register(&testEventType{}, func3)
	assert.NoError(t, err)
	assert.Len(t, bus.registeredEvents, 1)
	assert.Len(t, bus.registeredEvents[eventName], 3)
}

func TestEventBus_receivedEventFromPubSubHandler(t *testing.T) {
	handlerSuccess := func(ctx context.Context, event testEventType) error {
		assert.IsType(t, testEventType{}, event)
		return nil
	}
	handlerFail := func(ctx context.Context, event testEventType) error { return errors.New("unknown") }

	type fields struct {
		Config           Config
		pubSub           PubSubInterface
		registeredEvents map[string][]EventRegisterConfig
		logger           log.LoggerI
	}
	type args struct {
		ctx       context.Context
		eventName string
		data      []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "received an unregistered event + should dlq unregistered enabled = error",
			fields: fields{
				Config: Config{
					ShouldDLQUnregistered: true,
				},
				logger: func() log.LoggerI {
					loggerMock := log.NewMockLoggerI(gomock.NewController(t))
					loggerMock.EXPECT().Error("nacking unknown/unregistered event", gomock.Any(), log.Any("event_name", "unregistered-event-name")).Times(1)
					return loggerMock
				}(),
			},
			args: args{
				eventName: "unregistered-event-name",
				data:      []byte("{}"),
			},
			wantErr: true,
		},
		{
			name: "invalid payload should throw unmarshal error",
			fields: fields{
				registeredEvents: map[string][]EventRegisterConfig{
					testEventType{}.EventType(): {
						{
							eventModel: &testEventType{},
						},
					},
				},
				logger: func() log.LoggerI {
					loggerMock := log.NewMockLoggerI(gomock.NewController(t))
					loggerMock.EXPECT().Error("error unmarshaling message", gomock.Any(), log.Any("event_name", testEventType{}.EventType())).Times(1)
					return loggerMock
				}(),
			},
			args: args{
				eventName: testEventType{}.EventType(),
				data:      []byte(""),
			},
			wantErr: true,
		},
		{
			name: "handler 2 of 3 throw error",
			fields: fields{
				registeredEvents: map[string][]EventRegisterConfig{
					testEventType{}.EventType(): {
						{eventModel: &testEventType{}, handler: handlerSuccess},
						{eventModel: &testEventType{}, handler: handlerFail},
						{eventModel: &testEventType{}, handler: handlerSuccess},
					},
				},
				logger: func() log.LoggerI {
					loggerMock := log.NewMockLoggerI(gomock.NewController(t))
					loggerMock.EXPECT().Error("error on handler 2 of 3", gomock.Any(), log.Any("event_name", testEventType{}.EventType())).Times(1)
					return loggerMock
				}(),
			},
			args: args{
				eventName: testEventType{}.EventType(),
				data:      []byte("{}"),
			},
			wantErr: true,
		},
		{
			name: "success case",
			fields: fields{
				registeredEvents: map[string][]EventRegisterConfig{
					testEventType{}.EventType(): {
						{eventModel: &testEventType{}, handler: handlerSuccess},
						{eventModel: &testEventType{}, handler: handlerSuccess},
					},
				},
				logger: func() log.LoggerI {
					loggerMock := log.NewMockLoggerI(gomock.NewController(t))
					loggerMock.EXPECT().Debug("event processed", log.Any("event_name", testEventType{}.EventType())).Times(1)
					return loggerMock
				}(),
			},
			args: args{
				eventName: testEventType{}.EventType(),
				data:      []byte("{}"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bus := &EventBus{
				Config:           tt.fields.Config,
				pubSub:           tt.fields.pubSub,
				registeredEvents: tt.fields.registeredEvents,
				logger:           tt.fields.logger,
			}
			if err := bus.receivedEventFromPubSubHandler(tt.args.ctx, tt.args.eventName, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("EventBus.receivedEventFromPubSubHandler() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
