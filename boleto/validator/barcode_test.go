package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateBarcode(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr error
	}{
		{
			name:    "valid barcode",
			code:    "00191234500000100000000000000000000000000000",
			wantErr: nil,
		},
		{
			name:    "too short",
			code:    "001912345",
			wantErr: ErrInvalidBarcode,
		},
		{
			name:    "too long",
			code:    "001912345000001000000000000000000000000000001",
			wantErr: ErrInvalidBarcode,
		},
		{
			name:    "empty",
			code:    "",
			wantErr: ErrInvalidBarcode,
		},
		{
			name:    "with letters",
			code:    "0019123450000010000000000000000000000000000A",
			wantErr: ErrInvalidBarcode,
		},
		{
			name:    "invalid checksum",
			code:    "00190234500000100000000000000000000000000000",
			wantErr: ErrInvalidBarcodeChecksum,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBarcode(tt.code)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestValidateDigitableLine(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantErr error
	}{
		{
			name:    "valid formatted",
			line:    "00190.00009 01234.567895 00000.000009 1 00000000010000",
			wantErr: nil,
		},
		{
			name:    "valid unformatted",
			line:    "00190000090123456789500000000009100000000010000",
			wantErr: nil,
		},
		{
			name:    "too short",
			line:    "00190.00009",
			wantErr: ErrInvalidDigitableLine,
		},
		{
			name:    "empty",
			line:    "",
			wantErr: ErrInvalidDigitableLine,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDigitableLine(tt.line)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
