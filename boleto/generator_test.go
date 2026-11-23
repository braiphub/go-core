package boleto

import (
	"bytes"
	"context"
	"image/color"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/braiphub/go-core/boleto/renderer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validBoleto() *Boleto {
	return &Boleto{
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
		Beneficiary: Beneficiary{
			Name:       "Test Company",
			Document:   "11.222.333/0001-81",
			Address:    "Rua Test, 100",
			City:       "São Paulo",
			State:      "SP",
			PostalCode: "01234-567",
		},
		Payer: Payer{
			Name:       "John Doe",
			Document:   "529.982.247-25",
			Address:    "Av. Test, 500",
			City:       "Rio de Janeiro",
			State:      "RJ",
			PostalCode: "20000-000",
		},
		Value:          100.00,
		DueDate:        time.Now().AddDate(0, 0, 30),
		DocumentDate:   time.Now(),
		ProcessingDate: time.Now(),
		Instructions: []string{
			"Não receber após vencimento",
		},
	}
}

// Minimal valid PNG (1x1 pixel)
var testLogo = []byte{
	0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
	0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xDE, 0x00, 0x00, 0x00,
	0x0C, 0x49, 0x44, 0x41, 0x54, 0x08, 0xD7, 0x63, 0xF8, 0xFF, 0xFF, 0x3F,
	0x00, 0x05, 0xFE, 0x02, 0xFE, 0xDC, 0xCC, 0x59, 0xE7, 0x00, 0x00, 0x00,
	0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
}

func TestGenerator_Generate(t *testing.T) {
	ctx := context.Background()

	t.Run("generates valid PDF", func(t *testing.T) {
		gen, err := NewGenerator(
			WithBankLogo("001", testLogo),
		)
		require.NoError(t, err)

		pdf, err := gen.Generate(ctx, validBoleto())

		require.NoError(t, err)
		require.NotEmpty(t, pdf)
		assert.True(t, bytes.HasPrefix(pdf, []byte("%PDF")), "should be valid PDF")
	})

	t.Run("generates PDF with PIX", func(t *testing.T) {
		gen, err := NewGenerator(
			WithBankLogo("001", testLogo),
		)
		require.NoError(t, err)

		b := validBoleto()
		b.PIX = &PIXInfo{
			EMV:  "00020126580014br.gov.bcb.pix0136test",
			TxID: "ABC123",
		}

		pdf, err := gen.Generate(ctx, b)

		require.NoError(t, err)
		assert.True(t, bytes.HasPrefix(pdf, []byte("%PDF")))
	})

	t.Run("fails on nil boleto", func(t *testing.T) {
		gen, _ := NewGenerator()

		_, err := gen.Generate(ctx, nil)

		assert.Equal(t, ErrNilBoleto, err)
	})

	t.Run("fails on invalid boleto", func(t *testing.T) {
		gen, _ := NewGenerator()
		b := validBoleto()
		b.Barcode = "" // Invalid

		_, err := gen.Generate(ctx, b)

		assert.Error(t, err)
	})

	t.Run("skips validation when configured", func(t *testing.T) {
		gen, _ := NewGenerator(
			WithoutValidation(),
			WithBankLogo("001", testLogo),
		)
		b := validBoleto()
		b.Barcode = "invalid" // Would fail validation

		// Should not error because validation is skipped
		// (will fail at barcode generation instead)
		_, err := gen.Generate(ctx, b)

		// Error is expected, but from barcode generation, not validation
		assert.Error(t, err)
	})

	t.Run("uses custom renderer", func(t *testing.T) {
		customRenderer := renderer.NewFebraban(
			renderer.WithPrimaryColor("#003366"),
			renderer.WithoutReceipt(),
		)

		gen, err := NewGenerator(
			WithRenderer(customRenderer),
			WithBankLogo("001", testLogo),
		)
		require.NoError(t, err)

		pdf, err := gen.Generate(ctx, validBoleto())

		require.NoError(t, err)
		assert.NotEmpty(t, pdf)
	})
}

func TestGenerator_GenerateFile(t *testing.T) {
	ctx := context.Background()

	t.Run("writes PDF to file", func(t *testing.T) {
		gen, _ := NewGenerator(
			WithBankLogo("001", testLogo),
		)

		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "boleto.pdf")

		err := gen.GenerateFile(ctx, validBoleto(), filePath)

		require.NoError(t, err)

		// Verify file exists and is valid PDF
		content, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.True(t, bytes.HasPrefix(content, []byte("%PDF")))
	})
}

type capturingRenderer struct {
	received *renderer.RenderData
}

func (r *capturingRenderer) Render(_ context.Context, data *renderer.RenderData) ([]byte, error) {
	r.received = data

	return []byte("%PDF-fake"), nil
}

func TestGenerator_Generate_BankLogoBackground(t *testing.T) {
	yellow := color.RGBA{R: 255, G: 204, B: 0, A: 255}

	t.Run("passes the background registered for the bank code", func(t *testing.T) {
		capture := &capturingRenderer{}
		gen, err := NewGenerator(
			WithRenderer(capture),
			WithBankLogo("001", testLogo),
			WithBankLogoBackground("001", yellow),
		)
		require.NoError(t, err)

		_, err = gen.Generate(context.Background(), validBoleto())

		require.NoError(t, err)
		require.NotNil(t, capture.received.BankLogoBackground)
		assert.Equal(t, yellow, *capture.received.BankLogoBackground)
	})

	t.Run("leaves background nil when only another bank registered one", func(t *testing.T) {
		capture := &capturingRenderer{}
		gen, err := NewGenerator(
			WithRenderer(capture),
			WithBankLogo("001", testLogo),
			WithBankLogo("783", testLogo),
			WithBankLogoBackground("001", yellow),
		)
		require.NoError(t, err)
		slip := validBoleto()
		slip.BankCode = "783"

		_, err = gen.Generate(context.Background(), slip)

		require.NoError(t, err)
		assert.Equal(t, testLogo, capture.received.BankLogo)
		assert.Nil(t, capture.received.BankLogoBackground)
	})
}
