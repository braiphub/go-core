package barcode

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateITF(t *testing.T) {
	t.Run("generates valid PNG", func(t *testing.T) {
		code := "00191234500000100000000000000000000000000000"
		img, err := GenerateITF(code, 500, 50)

		require.NoError(t, err)
		require.NotEmpty(t, img)

		// Check PNG magic bytes
		pngHeader := []byte{0x89, 0x50, 0x4E, 0x47}
		assert.True(t, bytes.HasPrefix(img, pngHeader), "should be valid PNG")
	})

	t.Run("rejects odd length", func(t *testing.T) {
		code := "12345" // 5 digits - odd
		_, err := GenerateITF(code, 400, 50)

		assert.Equal(t, ErrITFOddLength, err)
	})

	t.Run("rejects non-digits", func(t *testing.T) {
		code := "1234567890ABCD"
		_, err := GenerateITF(code, 400, 50)

		assert.Equal(t, ErrInvalidITFCode, err)
	})

	t.Run("rejects empty code", func(t *testing.T) {
		_, err := GenerateITF("", 400, 50)

		assert.Error(t, err)
	})
}
