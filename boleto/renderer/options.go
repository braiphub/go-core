package renderer

// Febraban is the FEBRABAN renderer implementation.
// This is a forward reference; the actual implementation is in Task 9.
type Febraban struct {
	primaryColor   string
	margin         float64
	showReceipt    bool
	showPIXSection bool
	pageWidth      float64
	pageHeight     float64
}

// FebOption configures the Febraban renderer.
type FebOption func(*Febraban)

// WithPrimaryColor sets the primary color (hex format, e.g., "#003366").
func WithPrimaryColor(hex string) FebOption {
	return func(f *Febraban) {
		f.primaryColor = hex
	}
}

// WithMargin sets the page margin in millimeters.
func WithMargin(mm float64) FebOption {
	return func(f *Febraban) {
		f.margin = mm
	}
}

// WithoutReceipt hides the payer receipt section.
func WithoutReceipt() FebOption {
	return func(f *Febraban) {
		f.showReceipt = false
	}
}

// WithoutPIXSection hides the PIX section even when QR code is available.
func WithoutPIXSection() FebOption {
	return func(f *Febraban) {
		f.showPIXSection = false
	}
}

// WithPageSize sets custom page dimensions in millimeters.
func WithPageSize(width, height float64) FebOption {
	return func(f *Febraban) {
		f.pageWidth = width
		f.pageHeight = height
	}
}
