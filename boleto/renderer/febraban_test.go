package renderer

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/braiphub/go-core/boleto/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testBoleto() *types.Boleto {
	return &types.Boleto{
		Barcode:        "00191234500000100000000000000000000000000000",
		DigitiableLine: "00190000090123456789500000000009100000000010000",
		OurNumber:      "12345678",
		DocumentNumber: "DM-001",
		BankCode:       "001",
		BankName:       "Banco do Brasil",
		Agency:         "1234",
		AgencyDigit:    "5",
		Account:        "56789",
		AccountDigit:   "0",
		Beneficiary: types.Beneficiary{
			Name:       "Test Company LTDA",
			Document:   "11.222.333/0001-81",
			Address:    "Rua das Flores, 100",
			City:       "São Paulo",
			State:      "SP",
			PostalCode: "01234-567",
		},
		Payer: types.Payer{
			Name:       "John Doe",
			Document:   "529.982.247-25",
			Address:    "Av. Brasil, 500",
			City:       "Rio de Janeiro",
			State:      "RJ",
			PostalCode: "20000-000",
		},
		Value:          100.00,
		DueDate:        time.Date(2024, 12, 15, 0, 0, 0, 0, time.Local),
		DocumentDate:   time.Date(2024, 11, 15, 0, 0, 0, 0, time.Local),
		ProcessingDate: time.Date(2024, 11, 15, 0, 0, 0, 0, time.Local),
		Instructions: []string{
			"Não receber após o vencimento",
			"Multa de 2% após o vencimento",
		},
		Description: []string{
			"Mensalidade Dezembro/2024",
		},
	}
}

// Minimal valid PNG (1x1 pixel)
var testPNG = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
	0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xDE, 0x00, 0x00, 0x00,
	0x0C, 0x49, 0x44, 0x41, 0x54, 0x08, 0xD7, 0x63, 0xF8, 0xFF, 0xFF, 0x3F,
	0x00, 0x05, 0xFE, 0x02, 0xFE, 0xDC, 0xCC, 0x59, 0xE7, 0x00, 0x00, 0x00,
	0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
}

func TestFebraban_Render(t *testing.T) {
	ctx := context.Background()

	t.Run("generates valid PDF", func(t *testing.T) {
		renderer := NewFebraban()
		data := &RenderData{
			Boleto:       testBoleto(),
			BarcodeImage: testPNG,
			BankLogo:     testPNG,
		}

		pdf, err := renderer.Render(ctx, data)

		require.NoError(t, err)
		require.NotEmpty(t, pdf)

		// Check PDF magic bytes
		assert.True(t, bytes.HasPrefix(pdf, []byte("%PDF")), "should be valid PDF")
	})

	t.Run("generates PDF with PIX", func(t *testing.T) {
		renderer := NewFebraban()
		b := testBoleto()
		b.PIX = &types.PIXInfo{
			EMV:  "00020126580014br.gov.bcb.pix",
			TxID: "ABC123",
		}
		data := &RenderData{
			Boleto:       b,
			BarcodeImage: testPNG,
			QRCodeImage:  testPNG,
			BankLogo:     testPNG,
		}

		pdf, err := renderer.Render(ctx, data)

		require.NoError(t, err)
		assert.True(t, bytes.HasPrefix(pdf, []byte("%PDF")))
	})

	t.Run("applies options", func(t *testing.T) {
		renderer := NewFebraban(
			WithPrimaryColor("#003366"),
			WithMargin(15),
			WithoutReceipt(),
		)
		data := &RenderData{
			Boleto:       testBoleto(),
			BarcodeImage: testPNG,
			BankLogo:     testPNG,
		}

		pdf, err := renderer.Render(ctx, data)

		require.NoError(t, err)
		assert.NotEmpty(t, pdf)
	})
}

func TestFebraban_ImplementsInterface(t *testing.T) {
	var _ RendererI = (*Febraban)(nil)
}
