package barcode

import "errors"

var (
	// ErrInvalidITFCode is returned when the code contains non-digits.
	ErrInvalidITFCode = errors.New("invalid ITF code: must contain only digits")

	// ErrITFOddLength is returned when the code has odd length.
	ErrITFOddLength = errors.New("ITF-25 requires even number of digits")

	// ErrQRCodeGeneration is returned when QR code generation fails.
	ErrQRCodeGeneration = errors.New("failed to generate QR code")

	// ErrEmptyEMV is returned when EMV string is empty.
	ErrEmptyEMV = errors.New("EMV string cannot be empty")
)
