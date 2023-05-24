package log

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestAny(t *testing.T) {
	got := Any("key", "item data")
	assert.Equal(t, got, Field{Key: "key", Data: "item data"})
}

func TestError(t *testing.T) {
	err := errors.Wrap(errors.New("unknown error"), "doing anything")
	got := Error(err)
	assert.Equal(t, got, Field{Key: "error", Data: "doing anything: unknown error"})
}

func TestErrorWTrace(t *testing.T) {
	err := errors.Wrap(errors.New("unknown error"), "doing anything")
	got := ErrorWTrace(err)
	assert.Equal(t, got.Key, "error")
	assert.NotNil(t, got.Data)
}
