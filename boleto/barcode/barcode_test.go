package barcode

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerator(t *testing.T) {
	ctx := context.Background()
	gen := New()

	t.Run("GenerateITF returns PNG", func(t *testing.T) {
		code := "00191234500000100000000000000000000000000000"
		img, err := gen.GenerateITF(ctx, code)

		require.NoError(t, err)
		pngHeader := []byte{0x89, 0x50, 0x4E, 0x47}
		assert.True(t, bytes.HasPrefix(img, pngHeader))
	})

	t.Run("GenerateQRCode returns PNG", func(t *testing.T) {
		emv := "00020126580014br.gov.bcb.pix0136test"
		img, err := gen.GenerateQRCode(ctx, emv)

		require.NoError(t, err)
		pngHeader := []byte{0x89, 0x50, 0x4E, 0x47}
		assert.True(t, bytes.HasPrefix(img, pngHeader))
	})
}

func TestGenerator_ImplementsInterface(t *testing.T) {
	var _ GeneratorI = (*Generator)(nil)
}
