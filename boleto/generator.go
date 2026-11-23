package boleto

import (
	"context"
	"os"

	"github.com/braiphub/go-core/boleto/barcode"
	"github.com/braiphub/go-core/boleto/renderer"
	"github.com/braiphub/go-core/boleto/validator"
	"github.com/pkg/errors"
)

// Generator orchestrates boleto PDF generation.
type Generator struct {
	validator      validator.ValidatorI
	barcode        barcode.GeneratorI
	renderer       renderer.RendererI
	bankLogos      map[string][]byte
	skipValidation bool
}

// NewGenerator creates a new Generator with the given options.
func NewGenerator(opts ...Option) (*Generator, error) {
	g := &Generator{
		validator: validator.New(),
		barcode:   barcode.New(),
		renderer:  renderer.NewFebraban(),
		bankLogos: make(map[string][]byte),
	}

	for _, opt := range opts {
		opt(g)
	}

	return g, nil
}

// Generate creates a PDF from the boleto data and returns the bytes.
func (g *Generator) Generate(ctx context.Context, b *Boleto) ([]byte, error) {
	if b == nil {
		return nil, ErrNilBoleto
	}

	// Validate
	if !g.skipValidation {
		if err := g.validator.Validate(ctx, b); err != nil {
			return nil, errors.Wrap(err, "validation failed")
		}
	}

	// Generate barcode image
	barcodeImg, err := g.barcode.GenerateITF(ctx, b.Barcode)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate barcode")
	}

	// Generate QR code if PIX is present
	var qrCodeImg []byte
	if b.PIX != nil && b.PIX.EMV != "" {
		qrCodeImg, err = g.barcode.GenerateQRCode(ctx, b.PIX.EMV)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate QR code")
		}
	}

	// Get bank logo
	bankLogo := g.bankLogos[b.BankCode]

	// Render PDF
	data := &renderer.RenderData{
		Boleto:       b,
		BarcodeImage: barcodeImg,
		QRCodeImage:  qrCodeImg,
		BankLogo:     bankLogo,
	}

	pdf, err := g.renderer.Render(ctx, data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render PDF")
	}

	return pdf, nil
}

// GenerateFile creates a PDF and writes it to the specified file path.
func (g *Generator) GenerateFile(ctx context.Context, b *Boleto, filePath string) error {
	pdf, err := g.Generate(ctx, b)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filePath, pdf, 0644); err != nil {
		return errors.Wrap(err, "failed to write PDF file")
	}

	return nil
}
