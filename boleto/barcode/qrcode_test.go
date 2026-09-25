package barcode

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateQRCode(t *testing.T) {
	t.Run("generates valid PNG", func(t *testing.T) {
		emv := "00020126580014br.gov.bcb.pix0136a1b2c3d4-e5f6-7890-abcd-ef1234567890"
		img, err := GenerateQRCode(emv, 200)

		require.NoError(t, err)
		require.NotEmpty(t, img)

		// Check PNG magic bytes
		pngHeader := []byte{0x89, 0x50, 0x4E, 0x47}
		assert.True(t, bytes.HasPrefix(img, pngHeader), "should be valid PNG")
	})

	t.Run("rejects empty EMV", func(t *testing.T) {
		_, err := GenerateQRCode("", 200)

		assert.Equal(t, ErrEmptyEMV, err)
	})

	t.Run("handles long EMV strings", func(t *testing.T) {
		// Real PIX EMV can be quite long
		emv := "00020126580014br.gov.bcb.pix0136a1b2c3d4-e5f6-7890-abcd-ef12345678905204000053039865802BR5925EMPRESA TESTE LTDA6009SAO PAULO62070503***6304ABCD"
		img, err := GenerateQRCode(emv, 200)

		require.NoError(t, err)
		require.NotEmpty(t, img)
	})
}
