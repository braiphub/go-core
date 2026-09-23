package boleto

import "errors"

var (
	// ErrNilBoleto is returned when a nil boleto is passed to Generate.
	ErrNilBoleto = errors.New("boleto cannot be nil")

	// ErrGenerateFailed is returned when PDF generation fails.
	ErrGenerateFailed = errors.New("failed to generate PDF")
)
