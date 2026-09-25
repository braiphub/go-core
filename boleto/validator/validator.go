package validator

import (
	"context"
	"fmt"
	"strings"

	"github.com/braiphub/go-core/boleto/internal/types"
)

// ValidatorI validates boleto data.
type ValidatorI interface {
	Validate(ctx context.Context, b *types.Boleto) error
}

// ValidationError represents a single validation error.
type ValidationError struct {
	Field   string
	Value   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s (%s)", e.Field, e.Message, e.Value)
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// Validator implements ValidatorI with full FEBRABAN validation.
type Validator struct{}

// New creates a new Validator.
func New() *Validator {
	return &Validator{}
}

// Validate validates all boleto fields.
func (v *Validator) Validate(ctx context.Context, b *types.Boleto) error {
	if b == nil {
		return &ValidationError{Field: "boleto", Message: "cannot be nil"}
	}

	var errs ValidationErrors

	// Required fields
	if b.Barcode == "" {
		errs = append(errs, ValidationError{Field: "barcode", Message: "is required"})
	} else if err := ValidateBarcode(b.Barcode); err != nil {
		errs = append(errs, ValidationError{Field: "barcode", Value: b.Barcode, Message: err.Error()})
	}

	if b.DigitiableLine == "" {
		errs = append(errs, ValidationError{Field: "digitable_line", Message: "is required"})
	} else if err := ValidateDigitableLine(b.DigitiableLine); err != nil {
		errs = append(errs, ValidationError{Field: "digitable_line", Value: b.DigitiableLine, Message: err.Error()})
	}

	if b.BankCode == "" {
		errs = append(errs, ValidationError{Field: "bank_code", Message: "is required"})
	} else if len(nonDigitRegex.ReplaceAllString(b.BankCode, "")) != 3 {
		errs = append(errs, ValidationError{Field: "bank_code", Value: b.BankCode, Message: "must be 3 digits"})
	}

	if b.Beneficiary.Name == "" {
		errs = append(errs, ValidationError{Field: "beneficiary.name", Message: "is required"})
	}

	if b.Beneficiary.Document != "" {
		if err := ValidateDocument(b.Beneficiary.Document); err != nil {
			errs = append(errs, ValidationError{Field: "beneficiary.document", Value: b.Beneficiary.Document, Message: err.Error()})
		}
	}

	if b.Payer.Name == "" {
		errs = append(errs, ValidationError{Field: "payer.name", Message: "is required"})
	}

	if b.Payer.Document != "" {
		if err := ValidateDocument(b.Payer.Document); err != nil {
			errs = append(errs, ValidationError{Field: "payer.document", Value: b.Payer.Document, Message: err.Error()})
		}
	}

	if b.Value <= 0 {
		errs = append(errs, ValidationError{Field: "value", Message: "must be greater than zero"})
	}

	if b.DueDate.IsZero() {
		errs = append(errs, ValidationError{Field: "due_date", Message: "is required"})
	}

	// PIX validation (optional)
	if b.PIX != nil && b.PIX.EMV == "" {
		errs = append(errs, ValidationError{Field: "pix.emv", Message: "is required when PIX is provided"})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
