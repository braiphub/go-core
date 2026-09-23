package barcode

import (
	"bytes"
	"image/png"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
)

// GenerateQRCode generates a QR code from a PIX EMV string as PNG bytes.
// Size is both width and height in pixels (QR codes are square).
func GenerateQRCode(emv string, size int) ([]byte, error) {
	if emv == "" {
		return nil, ErrEmptyEMV
	}

	// Generate QR code with medium error correction
	qrCode, err := qr.Encode(emv, qr.M, qr.Auto)
	if err != nil {
		return nil, ErrQRCodeGeneration
	}

	// Scale to requested size
	scaled, err := barcode.Scale(qrCode, size, size)
	if err != nil {
		return nil, ErrQRCodeGeneration
	}

	// Encode as PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, scaled); err != nil {
		return nil, ErrQRCodeGeneration
	}

	return buf.Bytes(), nil
}
