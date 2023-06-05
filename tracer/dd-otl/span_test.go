package ddotl

import (
	"testing"
)

func TestSpan_AddEvent(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		span := &Span{
			otlSpan: otlSpanMocked{},
		}
		span.AddEvent("name")
	})
}

func TestSpan_Close(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		span := &Span{
			otlSpan: otlSpanMocked{},
		}
		span.Close()
	})
}
