package zaplogger

import (
	"testing"

	"github.com/braiphub/go-core/log"
	"github.com/braiphub/go-core/log/zaplogger/mocks"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestNew(t *testing.T) {
	logger, err := New(LoggerEnvDev, 7)
	assert.NoError(t, err)
	assert.NotNil(t, logger)

	logger, err = New(LoggerEnvProd, 7)
	assert.NoError(t, err)
	assert.NotNil(t, logger)

	logger, err = New("", 7)
	assert.Error(t, err)
	assert.Nil(t, logger)
}

func TestLogging(t *testing.T) {
	var errData = errors.New("unknown error")
	ctrl := gomock.NewController(t)
	zapMock := mocks.NewMockZapLoggerI(ctrl)
	zapMock.EXPECT().Debug("trace message", []zapcore.Field{{Key: "key", Type: zapcore.StringType, String: "val"}}).Times(1)
	zapMock.EXPECT().Debug("debug message", []zapcore.Field{{Key: "key", Type: zapcore.StringType, String: "val"}}).Times(1)
	zapMock.EXPECT().Info("info message", []zapcore.Field{{Key: "key", Type: zapcore.StringType, String: "val"}}).Times(1)
	zapMock.EXPECT().Warn("warn message", []zapcore.Field{{Key: "key", Type: zapcore.StringType, String: "val"}}).Times(1)
	zapMock.EXPECT().Error("error message", []zapcore.Field{{Key: "key", Type: zapcore.StringType, String: "val"}, {Key: "error", Type: zapcore.StringType, String: "unknown error"}}).Times(1)
	zapMock.EXPECT().Fatal("fatal message", []zapcore.Field{{Key: "key", Type: zapcore.StringType, String: "val"}}).Times(1)

	logger := &ZapLogger{
		zapMock,
	}
	logger.Trace("trace message", log.Any("key", "val"))
	logger.Debug("debug message", log.Any("key", "val"))
	logger.Info("info message", log.Any("key", "val"))
	logger.Warn("warn message", log.Any("key", "val"))
	logger.Error("error message", errData, log.Any("key", "val"))
	logger.Fatal("fatal message", log.Any("key", "val"))
}

func TestZapLogger_Write(t *testing.T) {
	ctrl := gomock.NewController(t)
	zapMock := mocks.NewMockZapLoggerI(ctrl)
	zapMock.EXPECT().Info("write test message").Times(1)

	logger := &ZapLogger{
		zapMock,
	}
	logger.Write([]byte("write test message"))
}
