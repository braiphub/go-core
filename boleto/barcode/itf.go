package barcode

import (
	"bytes"
	"image/png"
	"unicode"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/twooffive"
)

// GenerateITF generates an ITF-25 (Interleaved 2 of 5) barcode as PNG bytes.
// This is the standard barcode format for Brazilian boletos.
// Width and height are in pixels.
func GenerateITF(code string, width, height int) ([]byte, error) {
	if code == "" {
		return nil, ErrInvalidITFCode
	}

	// Validate all digits
	for _, r := range code {
		if !unicode.IsDigit(r) {
			return nil, ErrInvalidITFCode
		}
	}

	// ITF requires even number of digits
	if len(code)%2 != 0 {
		return nil, ErrITFOddLength
	}

	// Generate barcode
	bc, err := twooffive.Encode(code, true) // true for interleaved
	if err != nil {
		return nil, err
	}

	// Scale to requested dimensions
	scaled, err := barcode.Scale(bc, width, height)
	if err != nil {
		return nil, err
	}

	// Encode as PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, scaled); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
