package barcode

import "context"

// Default dimensions for FEBRABAN boletos
const (
	DefaultBarcodeWidth  = 500 // ~103mm at 96 DPI (min 404 for ITF-25)
	DefaultBarcodeHeight = 50  // ~13mm at 96 DPI
	DefaultQRCodeSize    = 95  // ~25mm at 96 DPI
)

// GeneratorI generates barcode and QR code images.
type GeneratorI interface {
	// GenerateITF generates an ITF-25 barcode from a 44-digit code.
	GenerateITF(ctx context.Context, code string) ([]byte, error)

	// GenerateQRCode generates a QR code from a PIX EMV string.
	GenerateQRCode(ctx context.Context, emv string) ([]byte, error)
}

// Generator implements GeneratorI with default FEBRABAN dimensions.
type Generator struct {
	barcodeWidth  int
	barcodeHeight int
	qrCodeSize    int
}

// New creates a new Generator with default dimensions.
func New() *Generator {
	return &Generator{
		barcodeWidth:  DefaultBarcodeWidth,
		barcodeHeight: DefaultBarcodeHeight,
		qrCodeSize:    DefaultQRCodeSize,
	}
}

// GenerateITF generates an ITF-25 barcode as PNG bytes.
func (g *Generator) GenerateITF(ctx context.Context, code string) ([]byte, error) {
	return GenerateITF(code, g.barcodeWidth, g.barcodeHeight)
}

// GenerateQRCode generates a QR code as PNG bytes.
func (g *Generator) GenerateQRCode(ctx context.Context, emv string) ([]byte, error) {
	return GenerateQRCode(emv, g.qrCodeSize)
}
