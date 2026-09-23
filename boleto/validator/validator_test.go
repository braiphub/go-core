package validator

import (
	"context"
	"testing"
	"time"

	"github.com/braiphub/go-core/boleto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validBoleto() *boleto.Boleto {
	return &boleto.Boleto{
		Barcode:        "00191234500000100000000000000000000000000000",
		DigitiableLine: "00190000090123456789500000000009100000000010000",
		OurNumber:      "12345678",
		DocumentNumber: "DM-001",
		BankCode:       "001",
		BankName:       "Banco do Brasil",
		Agency:         "1234",
		Account:        "56789",
		Beneficiary: boleto.Beneficiary{
			Name:     "Test Company",
			Document: "11.222.333/0001-81",
		},
		Payer: boleto.Payer{
			Name:     "John Doe",
			Document: "529.982.247-25",
		},
		Value:   100.00,
		DueDate: time.Now().AddDate(0, 0, 30),
	}
}

func TestValidator_Validate(t *testing.T) {
	ctx := context.Background()
	v := New()

	t.Run("valid boleto passes", func(t *testing.T) {
		b := validBoleto()
		err := v.Validate(ctx, b)
		assert.NoError(t, err)
	})

	t.Run("nil boleto fails", func(t *testing.T) {
		err := v.Validate(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("missing barcode fails", func(t *testing.T) {
		b := validBoleto()
		b.Barcode = ""
		err := v.Validate(ctx, b)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "barcode")
	})

	t.Run("invalid beneficiary CPF fails", func(t *testing.T) {
		b := validBoleto()
		b.Beneficiary.Document = "123.456.789-00"
		err := v.Validate(ctx, b)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "CPF")
	})

	t.Run("invalid payer CNPJ fails", func(t *testing.T) {
		b := validBoleto()
		b.Payer.Document = "11.222.333/0001-00"
		err := v.Validate(ctx, b)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "CNPJ")
	})

	t.Run("missing value fails", func(t *testing.T) {
		b := validBoleto()
		b.Value = 0
		err := v.Validate(ctx, b)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "value")
	})

	t.Run("invalid bank code fails", func(t *testing.T) {
		b := validBoleto()
		b.BankCode = "12"
		err := v.Validate(ctx, b)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "bank_code")
	})
}
