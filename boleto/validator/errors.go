package validator

import "errors"

// Required field errors
var (
	ErrBarcodeRequired       = errors.New("barcode is required")
	ErrDigitableLineRequired = errors.New("digitable line is required")
	ErrBankCodeRequired      = errors.New("bank code is required")
	ErrBeneficiaryRequired   = errors.New("beneficiary name is required")
	ErrPayerRequired         = errors.New("payer name is required")
	ErrValueRequired         = errors.New("value is required")
	ErrDueDateRequired       = errors.New("due date is required")
)

// Format errors
var (
	ErrInvalidBarcode         = errors.New("invalid barcode (must be 44 digits)")
	ErrInvalidDigitableLine   = errors.New("invalid digitable line (must be 47 digits)")
	ErrInvalidBarcodeChecksum = errors.New("invalid barcode checksum")
	ErrInvalidCPF             = errors.New("invalid CPF")
	ErrInvalidCNPJ            = errors.New("invalid CNPJ")
	ErrInvalidBankCode        = errors.New("invalid bank code (must be 3 digits)")
	ErrInvalidPIXEMV          = errors.New("invalid PIX EMV string")
)
