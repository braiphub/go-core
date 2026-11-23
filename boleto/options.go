package boleto

import (
	"image/color"

	"github.com/braiphub/go-core/boleto/barcode"
	"github.com/braiphub/go-core/boleto/renderer"
	"github.com/braiphub/go-core/boleto/validator"
)

// Option configures the Generator.
type Option func(*Generator)

// WithValidator sets a custom validator.
func WithValidator(v validator.ValidatorI) Option {
	return func(g *Generator) {
		g.validator = v
	}
}

// WithBarcodeGenerator sets a custom barcode generator.
func WithBarcodeGenerator(b barcode.GeneratorI) Option {
	return func(g *Generator) {
		g.barcode = b
	}
}

// WithRenderer sets a custom renderer.
func WithRenderer(r renderer.RendererI) Option {
	return func(g *Generator) {
		g.renderer = r
	}
}

// WithBankLogo registers a bank logo by bank code.
func WithBankLogo(bankCode string, logo []byte) Option {
	return func(g *Generator) {
		g.bankLogos[bankCode] = logo
	}
}

// WithBankLogoBackground fills the logo area of a bank with a solid color, drawn behind the
// logo registered with WithBankLogo. Without it the area stays white, so a transparent logo
// keeps its own colors. Banco do Brasil uses yellow: color.RGBA{R: 255, G: 204, B: 0, A: 255}.
func WithBankLogoBackground(bankCode string, background color.RGBA) Option {
	return func(g *Generator) {
		g.logoBgs[bankCode] = background
	}
}

// WithoutValidation disables validation (use with caution).
func WithoutValidation() Option {
	return func(g *Generator) {
		g.skipValidation = true
	}
}
