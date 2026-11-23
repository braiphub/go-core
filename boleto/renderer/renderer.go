package renderer

import (
	"context"

	"github.com/braiphub/go-core/boleto/internal/types"
)

// RendererI renders a boleto to PDF bytes.
type RendererI interface {
	Render(ctx context.Context, data *RenderData) ([]byte, error)
}

// RenderData contains all data needed to render a boleto PDF.
type RenderData struct {
	Boleto       *types.Boleto
	BarcodeImage []byte // ITF-25 PNG bytes
	QRCodeImage  []byte // QR Code PNG bytes (nil if no PIX)
	BankLogo     []byte // Bank logo PNG bytes
}
