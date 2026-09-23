package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateCPF(t *testing.T) {
	tests := []struct {
		name    string
		cpf     string
		wantErr error
	}{
		{"valid formatted", "529.982.247-25", nil},
		{"valid unformatted", "52998224725", nil},
		{"invalid checksum", "529.982.247-26", ErrInvalidCPF},
		{"all same digits", "111.111.111-11", ErrInvalidCPF},
		{"too short", "123", ErrInvalidCPF},
		{"too long", "123456789012", ErrInvalidCPF},
		{"empty", "", ErrInvalidCPF},
		{"with letters", "529.982.247-2A", ErrInvalidCPF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCPF(tt.cpf)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestValidateCNPJ(t *testing.T) {
	tests := []struct {
		name    string
		cnpj    string
		wantErr error
	}{
		{"valid formatted", "11.222.333/0001-81", nil},
		{"valid unformatted", "11222333000181", nil},
		{"invalid checksum", "11.222.333/0001-82", ErrInvalidCNPJ},
		{"all same digits", "11.111.111/1111-11", ErrInvalidCNPJ},
		{"too short", "123", ErrInvalidCNPJ},
		{"too long", "123456789012345", ErrInvalidCNPJ},
		{"empty", "", ErrInvalidCNPJ},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCNPJ(tt.cnpj)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestValidateDocument(t *testing.T) {
	tests := []struct {
		name    string
		doc     string
		wantErr error
	}{
		{"valid CPF", "529.982.247-25", nil},
		{"valid CNPJ", "11.222.333/0001-81", nil},
		{"invalid CPF", "123.456.789-00", ErrInvalidCPF},
		{"invalid CNPJ", "11.222.333/0001-00", ErrInvalidCNPJ},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDocument(tt.doc)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
