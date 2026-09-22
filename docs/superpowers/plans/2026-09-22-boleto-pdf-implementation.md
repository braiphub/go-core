# Boleto PDF Generator - Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go library to generate FEBRABAN-standard bank slips (boletos) as PDF with optional PIX QR code support.

**Architecture:** Modular design with three subpackages (validator, barcode, renderer) orchestrated by a Generator. Each component has its own interface for testability and can be replaced via functional options.

**Tech Stack:** Go 1.24+, go-pdf/fpdf (PDF), boombuler/barcode (ITF-25/QR), pkg/errors, stretchr/testify

## Global Constraints

- Go version: 1.24+
- No CGO dependencies
- License: MIT
- Module path: `github.com/braiphub/go-core/boleto`
- Follow go-core patterns: interfaces, functional options, context-aware
- All public functions accept `context.Context` as first parameter
- Error wrapping with `github.com/pkg/errors`
- Tests with `github.com/stretchr/testify`

---

## File Structure

```
boleto/
├── go.mod                      # Module definition
├── go.sum                      # Dependencies lock
├── doc.go                      # Package documentation
├── boleto.go                   # Boleto, Beneficiary, Payer, PIXInfo structs
├── errors.go                   # Package-level errors
├── generator.go                # Generator orchestrator
├── options.go                  # Generator functional options
├── generator_test.go           # Integration tests
│
├── validator/
│   ├── errors.go               # Validation errors
│   ├── cpfcnpj.go              # CPF/CNPJ validation
│   ├── cpfcnpj_test.go         # CPF/CNPJ tests
│   ├── barcode.go              # Barcode checksum validation
│   ├── barcode_test.go         # Barcode tests
│   ├── validator.go            # ValidatorI interface + implementation
│   └── validator_test.go       # Validator integration tests
│
├── barcode/
│   ├── errors.go               # Barcode generation errors
│   ├── itf.go                  # ITF-25 generation
│   ├── itf_test.go             # ITF-25 tests
│   ├── qrcode.go               # QR Code generation
│   ├── qrcode_test.go          # QR Code tests
│   ├── barcode.go              # GeneratorI interface + implementation
│   └── barcode_test.go         # Barcode integration tests
│
├── renderer/
│   ├── renderer.go             # RendererI interface + RenderData struct
│   ├── options.go              # Febraban options
│   ├── febraban.go             # FEBRABAN layout implementation
│   ├── febraban_test.go        # Renderer tests
│   └── format.go               # Formatting helpers (currency, date, CPF/CNPJ)
│
├── example/
│   └── main.go                 # Runnable example
│
└── README.md                   # Documentation
```

---

### Task 1: Project Scaffold & Data Structs

**Files:**
- Create: `boleto/go.mod`
- Create: `boleto/doc.go`
- Create: `boleto/boleto.go`
- Create: `boleto/errors.go`

**Interfaces:**
- Consumes: nothing
- Produces: `Boleto`, `Beneficiary`, `Payer`, `PIXInfo` structs; `ErrNilBoleto`, `ErrGenerateFailed` errors

- [ ] **Step 1: Create go.mod**

```bash
mkdir -p boleto
cd boleto
```

Create `boleto/go.mod`:
```go
module github.com/braiphub/go-core/boleto

go 1.24

require (
	github.com/boombuler/barcode v1.0.1
	github.com/go-pdf/fpdf v0.9.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.8.4
)
```

- [ ] **Step 2: Create doc.go**

Create `boleto/doc.go`:
```go
// Package boleto provides PDF generation for Brazilian bank slips (boletos)
// following FEBRABAN standards, with optional PIX QR code support.
//
// Basic usage:
//
//	gen, err := boleto.NewGenerator(
//	    boleto.WithBankLogo("001", logoBB),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	b := &boleto.Boleto{
//	    Barcode:        "00190000090300009802889500012000174720000010000",
//	    DigitiableLine: "00190.00009 03000.098028 89500.012009 1 74720000010000",
//	    // ... other fields
//	}
//
//	pdf, err := gen.Generate(ctx, b)
package boleto
```

- [ ] **Step 3: Create boleto.go with data structs**

Create `boleto/boleto.go`:
```go
package boleto

import "time"

// Boleto represents a complete FEBRABAN bank slip with all required fields.
type Boleto struct {
	// Identification
	Barcode        string // 44-digit barcode
	DigitiableLine string // 47-digit digitable line
	OurNumber      string // Nosso número
	DocumentNumber string // Número do documento

	// Bank
	BankCode     string // 3-digit bank code
	BankName     string // Bank name
	Agency       string // Agency number
	AgencyDigit  string // Agency check digit
	Account      string // Account number
	AccountDigit string // Account check digit

	// Beneficiary (receiver)
	Beneficiary Beneficiary

	// Payer
	Payer Payer

	// Values and dates
	Value          float64   // Document value
	DueDate        time.Time // Due date
	DocumentDate   time.Time // Document date
	ProcessingDate time.Time // Processing date

	// PIX (optional)
	PIX *PIXInfo

	// Instructions (free text)
	Instructions []string

	// Description (what is being charged)
	Description []string
}

// Beneficiary represents the payment receiver.
type Beneficiary struct {
	Name       string
	Document   string // CPF or CNPJ
	Address    string
	City       string
	State      string
	PostalCode string
}

// Payer represents the person paying the boleto.
type Payer struct {
	Name       string
	Document   string // CPF or CNPJ
	Address    string
	City       string
	State      string
	PostalCode string
}

// PIXInfo contains PIX payment information.
type PIXInfo struct {
	EMV  string // EMV string for QR code generation
	TxID string // Transaction ID (optional, for display)
}
```

- [ ] **Step 4: Create errors.go**

Create `boleto/errors.go`:
```go
package boleto

import "errors"

var (
	// ErrNilBoleto is returned when a nil boleto is passed to Generate.
	ErrNilBoleto = errors.New("boleto cannot be nil")

	// ErrGenerateFailed is returned when PDF generation fails.
	ErrGenerateFailed = errors.New("failed to generate PDF")
)
```

- [ ] **Step 5: Run go mod tidy and verify**

```bash
cd boleto && go mod tidy
```

Expected: Dependencies downloaded, go.sum created.

- [ ] **Step 6: Commit**

```bash
git add boleto/
git commit -m "feat(boleto): add project scaffold and data structs"
```

---

### Task 2: Validator - CPF/CNPJ Validation

**Files:**
- Create: `boleto/validator/errors.go`
- Create: `boleto/validator/cpfcnpj.go`
- Create: `boleto/validator/cpfcnpj_test.go`

**Interfaces:**
- Consumes: nothing
- Produces: `ValidateCPF(cpf string) error`, `ValidateCNPJ(cnpj string) error`, `ValidateDocument(doc string) error`; `ErrInvalidCPF`, `ErrInvalidCNPJ` errors

- [ ] **Step 1: Create validator/errors.go**

Create `boleto/validator/errors.go`:
```go
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
```

- [ ] **Step 2: Write failing test for CPF validation**

Create `boleto/validator/cpfcnpj_test.go`:
```go
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
```

- [ ] **Step 3: Run test to verify it fails**

```bash
cd boleto && go test ./validator/... -v
```

Expected: FAIL - `ValidateCPF`, `ValidateCNPJ`, `ValidateDocument` undefined.

- [ ] **Step 4: Implement CPF/CNPJ validation**

Create `boleto/validator/cpfcnpj.go`:
```go
package validator

import (
	"regexp"
	"strconv"
)

var nonDigitRegex = regexp.MustCompile(`\D`)

// ValidateCPF validates a Brazilian CPF (11 digits).
func ValidateCPF(cpf string) error {
	// Remove non-digits
	cpf = nonDigitRegex.ReplaceAllString(cpf, "")

	if len(cpf) != 11 {
		return ErrInvalidCPF
	}

	// Check for all same digits
	if allSameDigits(cpf) {
		return ErrInvalidCPF
	}

	// Validate first check digit
	sum := 0
	for i := 0; i < 9; i++ {
		digit, err := strconv.Atoi(string(cpf[i]))
		if err != nil {
			return ErrInvalidCPF
		}
		sum += digit * (10 - i)
	}
	remainder := sum % 11
	firstCheck := 0
	if remainder >= 2 {
		firstCheck = 11 - remainder
	}

	if strconv.Itoa(firstCheck) != string(cpf[9]) {
		return ErrInvalidCPF
	}

	// Validate second check digit
	sum = 0
	for i := 0; i < 10; i++ {
		digit, err := strconv.Atoi(string(cpf[i]))
		if err != nil {
			return ErrInvalidCPF
		}
		sum += digit * (11 - i)
	}
	remainder = sum % 11
	secondCheck := 0
	if remainder >= 2 {
		secondCheck = 11 - remainder
	}

	if strconv.Itoa(secondCheck) != string(cpf[10]) {
		return ErrInvalidCPF
	}

	return nil
}

// ValidateCNPJ validates a Brazilian CNPJ (14 digits).
func ValidateCNPJ(cnpj string) error {
	// Remove non-digits
	cnpj = nonDigitRegex.ReplaceAllString(cnpj, "")

	if len(cnpj) != 14 {
		return ErrInvalidCNPJ
	}

	// Check for all same digits
	if allSameDigits(cnpj) {
		return ErrInvalidCNPJ
	}

	// First check digit weights
	weights1 := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	sum := 0
	for i := 0; i < 12; i++ {
		digit, err := strconv.Atoi(string(cnpj[i]))
		if err != nil {
			return ErrInvalidCNPJ
		}
		sum += digit * weights1[i]
	}
	remainder := sum % 11
	firstCheck := 0
	if remainder >= 2 {
		firstCheck = 11 - remainder
	}

	if strconv.Itoa(firstCheck) != string(cnpj[12]) {
		return ErrInvalidCNPJ
	}

	// Second check digit weights
	weights2 := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	sum = 0
	for i := 0; i < 13; i++ {
		digit, err := strconv.Atoi(string(cnpj[i]))
		if err != nil {
			return ErrInvalidCNPJ
		}
		sum += digit * weights2[i]
	}
	remainder = sum % 11
	secondCheck := 0
	if remainder >= 2 {
		secondCheck = 11 - remainder
	}

	if strconv.Itoa(secondCheck) != string(cnpj[13]) {
		return ErrInvalidCNPJ
	}

	return nil
}

// ValidateDocument validates either CPF (11 digits) or CNPJ (14 digits).
func ValidateDocument(doc string) error {
	cleaned := nonDigitRegex.ReplaceAllString(doc, "")

	switch len(cleaned) {
	case 11:
		return ValidateCPF(doc)
	case 14:
		return ValidateCNPJ(doc)
	default:
		if len(cleaned) < 12 {
			return ErrInvalidCPF
		}
		return ErrInvalidCNPJ
	}
}

func allSameDigits(s string) bool {
	if len(s) == 0 {
		return true
	}
	first := s[0]
	for i := 1; i < len(s); i++ {
		if s[i] != first {
			return false
		}
	}
	return true
}
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
cd boleto && go test ./validator/... -v
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add boleto/validator/
git commit -m "feat(boleto/validator): add CPF/CNPJ validation"
```

---

### Task 3: Validator - Barcode Validation

**Files:**
- Create: `boleto/validator/barcode.go`
- Create: `boleto/validator/barcode_test.go`

**Interfaces:**
- Consumes: nothing
- Produces: `ValidateBarcode(code string) error`, `ValidateDigitableLine(line string) error`; uses `ErrInvalidBarcode`, `ErrInvalidDigitableLine`, `ErrInvalidBarcodeChecksum`

- [ ] **Step 1: Write failing test for barcode validation**

Create `boleto/validator/barcode_test.go`:
```go
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
			code:    "00190.00009",
			wantErr: ErrInvalidDigitableLine,
		},
		{
			name:    "empty",
			code:    "",
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd boleto && go test ./validator/... -v -run "TestValidateBarcode|TestValidateDigitableLine"
```

Expected: FAIL - `ValidateBarcode`, `ValidateDigitableLine` undefined.

- [ ] **Step 3: Implement barcode validation**

Create `boleto/validator/barcode.go`:
```go
package validator

import (
	"strconv"
	"unicode"
)

// ValidateBarcode validates a 44-digit FEBRABAN barcode.
// The barcode structure is: BBBMC.UUUUXXXXXX
// Where: BBB=bank, M=currency, C=checksum, U=due date factor, X=value/data
func ValidateBarcode(code string) error {
	if len(code) != 44 {
		return ErrInvalidBarcode
	}

	// Check all digits
	for _, r := range code {
		if !unicode.IsDigit(r) {
			return ErrInvalidBarcode
		}
	}

	// Validate checksum (position 4, 0-indexed)
	checkDigit, err := strconv.Atoi(string(code[4]))
	if err != nil {
		return ErrInvalidBarcode
	}

	calculated := calculateBarcodeChecksum(code)
	if calculated != checkDigit {
		return ErrInvalidBarcodeChecksum
	}

	return nil
}

// calculateBarcodeChecksum calculates the modulo 11 check digit.
// The check digit is at position 4 (0-indexed).
// Calculation uses positions 0-3 + 5-43 (excluding position 4).
func calculateBarcodeChecksum(code string) int {
	// Build string without check digit position
	withoutCheck := code[:4] + code[5:]

	sum := 0
	weight := 2

	// Process from right to left
	for i := len(withoutCheck) - 1; i >= 0; i-- {
		digit, _ := strconv.Atoi(string(withoutCheck[i]))
		sum += digit * weight
		weight++
		if weight > 9 {
			weight = 2
		}
	}

	remainder := sum % 11
	result := 11 - remainder

	// Special cases for FEBRABAN
	if result == 0 || result == 10 || result == 11 {
		return 1
	}

	return result
}

// ValidateDigitableLine validates a 47-digit digitable line.
func ValidateDigitableLine(line string) error {
	// Remove formatting (dots, spaces)
	cleaned := ""
	for _, r := range line {
		if unicode.IsDigit(r) {
			cleaned += string(r)
		}
	}

	if len(cleaned) != 47 {
		return ErrInvalidDigitableLine
	}

	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd boleto && go test ./validator/... -v -run "TestValidateBarcode|TestValidateDigitableLine"
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add boleto/validator/barcode.go boleto/validator/barcode_test.go
git commit -m "feat(boleto/validator): add barcode and digitable line validation"
```

---

### Task 4: Validator - Main Interface

**Files:**
- Create: `boleto/validator/validator.go`
- Create: `boleto/validator/validator_test.go`

**Interfaces:**
- Consumes: `ValidateCPF`, `ValidateCNPJ`, `ValidateDocument`, `ValidateBarcode`, `ValidateDigitableLine` from previous tasks; `Boleto` struct from Task 1
- Produces: `ValidatorI` interface with `Validate(ctx context.Context, b *boleto.Boleto) error`; `New() *Validator`; `ValidationError`, `ValidationErrors` types

- [ ] **Step 1: Write failing test for validator**

Create `boleto/validator/validator_test.go`:
```go
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
		assert.Contains(t, err.Error(), "bank code")
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd boleto && go test ./validator/... -v -run TestValidator_Validate
```

Expected: FAIL - `New`, type errors.

- [ ] **Step 3: Implement validator**

Create `boleto/validator/validator.go`:
```go
package validator

import (
	"context"
	"fmt"
	"strings"

	"github.com/braiphub/go-core/boleto"
)

// ValidatorI validates boleto data.
type ValidatorI interface {
	Validate(ctx context.Context, b *boleto.Boleto) error
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
func (v *Validator) Validate(ctx context.Context, b *boleto.Boleto) error {
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
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd boleto && go test ./validator/... -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add boleto/validator/validator.go boleto/validator/validator_test.go
git commit -m "feat(boleto/validator): add main validator interface"
```

---

### Task 5: Barcode Generator - ITF-25

**Files:**
- Create: `boleto/barcode/errors.go`
- Create: `boleto/barcode/itf.go`
- Create: `boleto/barcode/itf_test.go`

**Interfaces:**
- Consumes: nothing
- Produces: `GenerateITF(code string, width, height int) ([]byte, error)`; `ErrInvalidITFCode`, `ErrITFOddLength` errors

- [ ] **Step 1: Create barcode/errors.go**

Create `boleto/barcode/errors.go`:
```go
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
```

- [ ] **Step 2: Write failing test for ITF-25 generation**

Create `boleto/barcode/itf_test.go`:
```go
package barcode

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateITF(t *testing.T) {
	t.Run("generates valid PNG", func(t *testing.T) {
		code := "00191234500000100000000000000000000000000000"
		img, err := GenerateITF(code, 400, 50)

		require.NoError(t, err)
		require.NotEmpty(t, img)

		// Check PNG magic bytes
		pngHeader := []byte{0x89, 0x50, 0x4E, 0x47}
		assert.True(t, bytes.HasPrefix(img, pngHeader), "should be valid PNG")
	})

	t.Run("rejects odd length", func(t *testing.T) {
		code := "12345" // 5 digits - odd
		_, err := GenerateITF(code, 400, 50)

		assert.Equal(t, ErrITFOddLength, err)
	})

	t.Run("rejects non-digits", func(t *testing.T) {
		code := "1234567890ABCD"
		_, err := GenerateITF(code, 400, 50)

		assert.Equal(t, ErrInvalidITFCode, err)
	})

	t.Run("rejects empty code", func(t *testing.T) {
		_, err := GenerateITF("", 400, 50)

		assert.Error(t, err)
	})
}
```

- [ ] **Step 3: Run test to verify it fails**

```bash
cd boleto && go test ./barcode/... -v -run TestGenerateITF
```

Expected: FAIL - `GenerateITF` undefined.

- [ ] **Step 4: Implement ITF-25 generation**

Create `boleto/barcode/itf.go`:
```go
package barcode

import (
	"bytes"
	"image/png"
	"unicode"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/twooffive"
)

// GenerateITF generates an ITF-25 (Interleaved 2 of 5) barcode as PNG bytes.
// This is the standard barcode format for Brazilian boletos.
// Width and height are in pixels.
func GenerateITF(code string, width, height int) ([]byte, error) {
	if code == "" {
		return nil, ErrInvalidITFCode
	}

	// Validate all digits
	for _, r := range code {
		if !unicode.IsDigit(r) {
			return nil, ErrInvalidITFCode
		}
	}

	// ITF requires even number of digits
	if len(code)%2 != 0 {
		return nil, ErrITFOddLength
	}

	// Generate barcode
	bc, err := twooffive.Encode(code, true) // true for interleaved
	if err != nil {
		return nil, err
	}

	// Scale to requested dimensions
	scaled, err := barcode.Scale(bc, width, height)
	if err != nil {
		return nil, err
	}

	// Encode as PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, scaled); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
cd boleto && go test ./barcode/... -v -run TestGenerateITF
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add boleto/barcode/
git commit -m "feat(boleto/barcode): add ITF-25 barcode generation"
```

---

### Task 6: Barcode Generator - QR Code

**Files:**
- Create: `boleto/barcode/qrcode.go`
- Create: `boleto/barcode/qrcode_test.go`

**Interfaces:**
- Consumes: `ErrEmptyEMV`, `ErrQRCodeGeneration` from Task 5
- Produces: `GenerateQRCode(emv string, size int) ([]byte, error)`

- [ ] **Step 1: Write failing test for QR Code generation**

Create `boleto/barcode/qrcode_test.go`:
```go
package barcode

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateQRCode(t *testing.T) {
	t.Run("generates valid PNG", func(t *testing.T) {
		emv := "00020126580014br.gov.bcb.pix0136a1b2c3d4-e5f6-7890-abcd-ef1234567890"
		img, err := GenerateQRCode(emv, 200)

		require.NoError(t, err)
		require.NotEmpty(t, img)

		// Check PNG magic bytes
		pngHeader := []byte{0x89, 0x50, 0x4E, 0x47}
		assert.True(t, bytes.HasPrefix(img, pngHeader), "should be valid PNG")
	})

	t.Run("rejects empty EMV", func(t *testing.T) {
		_, err := GenerateQRCode("", 200)

		assert.Equal(t, ErrEmptyEMV, err)
	})

	t.Run("handles long EMV strings", func(t *testing.T) {
		// Real PIX EMV can be quite long
		emv := "00020126580014br.gov.bcb.pix0136a1b2c3d4-e5f6-7890-abcd-ef12345678905204000053039865802BR5925EMPRESA TESTE LTDA6009SAO PAULO62070503***6304ABCD"
		img, err := GenerateQRCode(emv, 200)

		require.NoError(t, err)
		require.NotEmpty(t, img)
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd boleto && go test ./barcode/... -v -run TestGenerateQRCode
```

Expected: FAIL - `GenerateQRCode` undefined.

- [ ] **Step 3: Implement QR Code generation**

Create `boleto/barcode/qrcode.go`:
```go
package barcode

import (
	"bytes"
	"image/png"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
)

// GenerateQRCode generates a QR code from a PIX EMV string as PNG bytes.
// Size is both width and height in pixels (QR codes are square).
func GenerateQRCode(emv string, size int) ([]byte, error) {
	if emv == "" {
		return nil, ErrEmptyEMV
	}

	// Generate QR code with medium error correction
	qrCode, err := qr.Encode(emv, qr.M, qr.Auto)
	if err != nil {
		return nil, ErrQRCodeGeneration
	}

	// Scale to requested size
	scaled, err := barcode.Scale(qrCode, size, size)
	if err != nil {
		return nil, ErrQRCodeGeneration
	}

	// Encode as PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, scaled); err != nil {
		return nil, ErrQRCodeGeneration
	}

	return buf.Bytes(), nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd boleto && go test ./barcode/... -v -run TestGenerateQRCode
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add boleto/barcode/qrcode.go boleto/barcode/qrcode_test.go
git commit -m "feat(boleto/barcode): add QR Code generation for PIX"
```

---

### Task 7: Barcode Generator - Interface

**Files:**
- Create: `boleto/barcode/barcode.go`
- Create: `boleto/barcode/barcode_test.go`

**Interfaces:**
- Consumes: `GenerateITF`, `GenerateQRCode` from Tasks 5-6
- Produces: `GeneratorI` interface with `GenerateITF(ctx, code) ([]byte, error)`, `GenerateQRCode(ctx, emv) ([]byte, error)`; `New() *Generator`

- [ ] **Step 1: Write failing test for Generator interface**

Create `boleto/barcode/barcode_test.go`:
```go
package barcode

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerator(t *testing.T) {
	ctx := context.Background()
	gen := New()

	t.Run("GenerateITF returns PNG", func(t *testing.T) {
		code := "00191234500000100000000000000000000000000000"
		img, err := gen.GenerateITF(ctx, code)

		require.NoError(t, err)
		pngHeader := []byte{0x89, 0x50, 0x4E, 0x47}
		assert.True(t, bytes.HasPrefix(img, pngHeader))
	})

	t.Run("GenerateQRCode returns PNG", func(t *testing.T) {
		emv := "00020126580014br.gov.bcb.pix0136test"
		img, err := gen.GenerateQRCode(ctx, emv)

		require.NoError(t, err)
		pngHeader := []byte{0x89, 0x50, 0x4E, 0x47}
		assert.True(t, bytes.HasPrefix(img, pngHeader))
	})
}

func TestGenerator_ImplementsInterface(t *testing.T) {
	var _ GeneratorI = (*Generator)(nil)
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd boleto && go test ./barcode/... -v -run TestGenerator
```

Expected: FAIL - `GeneratorI`, `New` undefined.

- [ ] **Step 3: Implement Generator interface**

Create `boleto/barcode/barcode.go`:
```go
package barcode

import "context"

// Default dimensions for FEBRABAN boletos
const (
	DefaultBarcodeWidth  = 385 // ~103mm at 96 DPI
	DefaultBarcodeHeight = 50  // ~13mm at 96 DPI
	DefaultQRCodeSize    = 95  // ~25mm at 96 DPI
)

// GeneratorI generates barcode and QR code images.
type GeneratorI interface {
	// GenerateITF generates an ITF-25 barcode from a 44-digit code.
	GenerateITF(ctx context.Context, code string) ([]byte, error)

	// GenerateQRCode generates a QR code from a PIX EMV string.
	GenerateQRCode(ctx context.Context, emv string) ([]byte, error)
}

// Generator implements GeneratorI with default FEBRABAN dimensions.
type Generator struct {
	barcodeWidth  int
	barcodeHeight int
	qrCodeSize    int
}

// New creates a new Generator with default dimensions.
func New() *Generator {
	return &Generator{
		barcodeWidth:  DefaultBarcodeWidth,
		barcodeHeight: DefaultBarcodeHeight,
		qrCodeSize:    DefaultQRCodeSize,
	}
}

// GenerateITF generates an ITF-25 barcode as PNG bytes.
func (g *Generator) GenerateITF(ctx context.Context, code string) ([]byte, error) {
	return GenerateITF(code, g.barcodeWidth, g.barcodeHeight)
}

// GenerateQRCode generates a QR code as PNG bytes.
func (g *Generator) GenerateQRCode(ctx context.Context, emv string) ([]byte, error) {
	return GenerateQRCode(emv, g.qrCodeSize)
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd boleto && go test ./barcode/... -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add boleto/barcode/barcode.go boleto/barcode/barcode_test.go
git commit -m "feat(boleto/barcode): add Generator interface"
```

---

### Task 8: Renderer - Interface & Options

**Files:**
- Create: `boleto/renderer/renderer.go`
- Create: `boleto/renderer/options.go`
- Create: `boleto/renderer/format.go`

**Interfaces:**
- Consumes: `Boleto` struct from Task 1
- Produces: `RendererI` interface with `Render(ctx, data *RenderData) ([]byte, error)`; `RenderData` struct; `FebOption` type; option functions `WithPrimaryColor`, `WithMargin`, `WithoutReceipt`, `WithoutPIXSection`, `WithPageSize`

- [ ] **Step 1: Create renderer/renderer.go**

Create `boleto/renderer/renderer.go`:
```go
package renderer

import (
	"context"

	"github.com/braiphub/go-core/boleto"
)

// RendererI renders a boleto to PDF bytes.
type RendererI interface {
	Render(ctx context.Context, data *RenderData) ([]byte, error)
}

// RenderData contains all data needed to render a boleto PDF.
type RenderData struct {
	Boleto       *boleto.Boleto
	BarcodeImage []byte // ITF-25 PNG bytes
	QRCodeImage  []byte // QR Code PNG bytes (nil if no PIX)
	BankLogo     []byte // Bank logo PNG bytes
}
```

- [ ] **Step 2: Create renderer/options.go**

Create `boleto/renderer/options.go`:
```go
package renderer

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
```

- [ ] **Step 3: Create renderer/format.go**

Create `boleto/renderer/format.go`:
```go
package renderer

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var nonDigitRegex = regexp.MustCompile(`\D`)

// FormatCurrency formats a float as Brazilian currency (R$ 1.234,56).
func FormatCurrency(value float64) string {
	// Format with 2 decimal places
	str := fmt.Sprintf("%.2f", value)

	// Split integer and decimal parts
	parts := strings.Split(str, ".")
	intPart := parts[0]
	decPart := parts[1]

	// Add thousand separators
	var result strings.Builder
	for i, r := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			result.WriteRune('.')
		}
		result.WriteRune(r)
	}

	return "R$ " + result.String() + "," + decPart
}

// FormatDate formats a time.Time as DD/MM/YYYY.
func FormatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("02/01/2006")
}

// FormatCPF formats a CPF string as XXX.XXX.XXX-XX.
func FormatCPF(cpf string) string {
	cpf = nonDigitRegex.ReplaceAllString(cpf, "")
	if len(cpf) != 11 {
		return cpf
	}
	return fmt.Sprintf("%s.%s.%s-%s", cpf[0:3], cpf[3:6], cpf[6:9], cpf[9:11])
}

// FormatCNPJ formats a CNPJ string as XX.XXX.XXX/XXXX-XX.
func FormatCNPJ(cnpj string) string {
	cnpj = nonDigitRegex.ReplaceAllString(cnpj, "")
	if len(cnpj) != 14 {
		return cnpj
	}
	return fmt.Sprintf("%s.%s.%s/%s-%s", cnpj[0:2], cnpj[2:5], cnpj[5:8], cnpj[8:12], cnpj[12:14])
}

// FormatDocument formats either CPF or CNPJ.
func FormatDocument(doc string) string {
	cleaned := nonDigitRegex.ReplaceAllString(doc, "")
	switch len(cleaned) {
	case 11:
		return FormatCPF(doc)
	case 14:
		return FormatCNPJ(doc)
	default:
		return doc
	}
}

// FormatDigitableLine formats the 47-digit line with spaces.
func FormatDigitableLine(line string) string {
	cleaned := nonDigitRegex.ReplaceAllString(line, "")
	if len(cleaned) != 47 {
		return line
	}
	// Format: XXXXX.XXXXX XXXXX.XXXXXX XXXXX.XXXXXX X XXXXXXXXXXXXXX
	return fmt.Sprintf("%s.%s %s.%s %s.%s %s %s",
		cleaned[0:5], cleaned[5:10],
		cleaned[10:15], cleaned[15:21],
		cleaned[21:26], cleaned[26:32],
		cleaned[32:33],
		cleaned[33:47])
}
```

- [ ] **Step 4: Commit**

```bash
git add boleto/renderer/
git commit -m "feat(boleto/renderer): add interface, options and format helpers"
```

---

### Task 9: Renderer - FEBRABAN Implementation

**Files:**
- Create: `boleto/renderer/febraban.go`
- Create: `boleto/renderer/febraban_test.go`

**Interfaces:**
- Consumes: `RendererI`, `RenderData`, `FebOption` from Task 8; format helpers
- Produces: `NewFebraban(opts ...FebOption) *Febraban`; `Febraban` struct implementing `RendererI`

- [ ] **Step 1: Write failing test for Febraban renderer**

Create `boleto/renderer/febraban_test.go`:
```go
package renderer

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/braiphub/go-core/boleto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testBoleto() *boleto.Boleto {
	return &boleto.Boleto{
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
		Beneficiary: boleto.Beneficiary{
			Name:       "Test Company LTDA",
			Document:   "11.222.333/0001-81",
			Address:    "Rua das Flores, 100",
			City:       "São Paulo",
			State:      "SP",
			PostalCode: "01234-567",
		},
		Payer: boleto.Payer{
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
		b.PIX = &boleto.PIXInfo{
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd boleto && go test ./renderer/... -v -run TestFebraban
```

Expected: FAIL - `NewFebraban`, `Febraban` undefined.

- [ ] **Step 3: Implement Febraban renderer**

Create `boleto/renderer/febraban.go`:
```go
package renderer

import (
	"bytes"
	"context"
	"image"
	_ "image/png"
	"strconv"

	"github.com/go-pdf/fpdf"
)

// FEBRABAN dimension constants (in mm)
const (
	BarcodeWidth  = 103.0
	BarcodeHeight = 13.0
	QRCodeSize    = 25.0
	LogoMaxWidth  = 40.0
	LogoMaxHeight = 15.0
)

// Febraban implements RendererI with FEBRABAN-compliant layout.
type Febraban struct {
	primaryColor   string
	pageWidth      float64
	pageHeight     float64
	margin         float64
	showReceipt    bool
	showPIXSection bool
}

// NewFebraban creates a new Febraban renderer with default settings.
func NewFebraban(opts ...FebOption) *Febraban {
	f := &Febraban{
		primaryColor:   "#000000",
		pageWidth:      210, // A4
		pageHeight:     297, // A4
		margin:         10,
		showReceipt:    true,
		showPIXSection: true,
	}

	for _, opt := range opts {
		opt(f)
	}

	return f
}

// Render generates the boleto PDF.
func (f *Febraban) Render(ctx context.Context, data *RenderData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(f.margin, f.margin, f.margin)
	pdf.AddPage()

	y := f.margin

	// Receipt section (top)
	if f.showReceipt {
		y = f.renderReceipt(pdf, data, y)
		y = f.renderCutLine(pdf, y)
	}

	// Main boleto section
	y = f.renderHeader(pdf, data, y)
	y = f.renderBeneficiarySection(pdf, data, y)
	y = f.renderAmountSection(pdf, data, y)
	y = f.renderPayerSection(pdf, data, y)
	y = f.renderInstructions(pdf, data, y)

	// PIX section
	if f.showPIXSection && data.QRCodeImage != nil {
		y = f.renderPIXSection(pdf, data, y)
	}

	// Barcode and digitable line
	y = f.renderBarcode(pdf, data, y)
	f.renderDigitableLine(pdf, data, y)

	// Generate PDF bytes
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (f *Febraban) renderReceipt(pdf *fpdf.Fpdf, data *RenderData, y float64) float64 {
	pdf.SetFont("Arial", "B", 10)
	pdf.SetXY(f.margin, y)
	pdf.Cell(0, 5, "RECIBO DO PAGADOR")
	y += 7

	pdf.SetFont("Arial", "", 8)
	b := data.Boleto

	// Beneficiary info
	pdf.SetXY(f.margin, y)
	pdf.Cell(0, 4, "Beneficiário: "+b.Beneficiary.Name+" - "+FormatDocument(b.Beneficiary.Document))
	y += 5

	// Payer info
	pdf.SetXY(f.margin, y)
	pdf.Cell(0, 4, "Pagador: "+b.Payer.Name+" - "+FormatDocument(b.Payer.Document))
	y += 5

	// Value and due date
	pdf.SetXY(f.margin, y)
	pdf.Cell(90, 4, "Vencimento: "+FormatDate(b.DueDate))
	pdf.Cell(0, 4, "Valor: "+FormatCurrency(b.Value))
	y += 5

	// Nosso número
	pdf.SetXY(f.margin, y)
	pdf.Cell(0, 4, "Nosso Número: "+b.OurNumber)
	y += 10

	return y
}

func (f *Febraban) renderCutLine(pdf *fpdf.Fpdf, y float64) float64 {
	pdf.SetDrawColor(128, 128, 128)
	pdf.SetDashPattern([]float64{2, 2}, 0)
	pdf.Line(f.margin, y, f.pageWidth-f.margin, y)
	pdf.SetDashPattern([]float64{}, 0)
	y += 5
	return y
}

func (f *Febraban) renderHeader(pdf *fpdf.Fpdf, data *RenderData, y float64) float64 {
	b := data.Boleto

	// Bank logo
	if data.BankLogo != nil {
		f.addImageFromBytes(pdf, data.BankLogo, f.margin, y, LogoMaxWidth, LogoMaxHeight)
	}

	// Bank code and name
	pdf.SetFont("Arial", "B", 12)
	pdf.SetXY(f.margin+LogoMaxWidth+5, y)
	pdf.Cell(20, 8, b.BankCode)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 8, b.BankName)

	y += 12

	// Separator line
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetLineWidth(0.5)
	pdf.Line(f.margin, y, f.pageWidth-f.margin, y)
	y += 3

	return y
}

func (f *Febraban) renderBeneficiarySection(pdf *fpdf.Fpdf, data *RenderData, y float64) float64 {
	b := data.Boleto
	pdf.SetFont("Arial", "", 7)
	colWidth := (f.pageWidth - 2*f.margin) / 3

	// Row 1: Beneficiário | Agência/Código | Nosso Número
	f.renderField(pdf, f.margin, y, colWidth*2, "Beneficiário", b.Beneficiary.Name)
	f.renderField(pdf, f.margin+colWidth*2, y, colWidth, "Agência/Código Beneficiário",
		b.Agency+"-"+b.AgencyDigit+" / "+b.Account+"-"+b.AccountDigit)
	y += 10

	// Row 2: Document | Nosso Número | Vencimento
	f.renderField(pdf, f.margin, y, colWidth, "CPF/CNPJ", FormatDocument(b.Beneficiary.Document))
	f.renderField(pdf, f.margin+colWidth, y, colWidth, "Nosso Número", b.OurNumber)
	f.renderField(pdf, f.margin+colWidth*2, y, colWidth, "Vencimento", FormatDate(b.DueDate))
	y += 10

	return y
}

func (f *Febraban) renderAmountSection(pdf *fpdf.Fpdf, data *RenderData, y float64) float64 {
	b := data.Boleto
	colWidth := (f.pageWidth - 2*f.margin) / 4

	// Dates and amounts
	f.renderField(pdf, f.margin, y, colWidth, "Data Documento", FormatDate(b.DocumentDate))
	f.renderField(pdf, f.margin+colWidth, y, colWidth, "Nº Documento", b.DocumentNumber)
	f.renderField(pdf, f.margin+colWidth*2, y, colWidth, "Espécie", "R$")
	f.renderField(pdf, f.margin+colWidth*3, y, colWidth, "Valor Documento", FormatCurrency(b.Value))
	y += 10

	return y
}

func (f *Febraban) renderPayerSection(pdf *fpdf.Fpdf, data *RenderData, y float64) float64 {
	b := data.Boleto

	f.renderField(pdf, f.margin, y, f.pageWidth-2*f.margin, "Pagador",
		b.Payer.Name+" - "+FormatDocument(b.Payer.Document))
	y += 8

	pdf.SetFont("Arial", "", 7)
	pdf.SetXY(f.margin, y)
	address := b.Payer.Address + " - " + b.Payer.City + "/" + b.Payer.State + " - " + b.Payer.PostalCode
	pdf.Cell(0, 4, address)
	y += 8

	return y
}

func (f *Febraban) renderInstructions(pdf *fpdf.Fpdf, data *RenderData, y float64) float64 {
	b := data.Boleto

	pdf.SetFont("Arial", "B", 7)
	pdf.SetXY(f.margin, y)
	pdf.Cell(0, 4, "Instruções")
	y += 5

	pdf.SetFont("Arial", "", 7)
	for _, instruction := range b.Instructions {
		pdf.SetXY(f.margin, y)
		pdf.Cell(0, 4, "- "+instruction)
		y += 4
	}
	y += 5

	return y
}

func (f *Febraban) renderPIXSection(pdf *fpdf.Fpdf, data *RenderData, y float64) float64 {
	pdf.SetFont("Arial", "B", 8)
	pdf.SetXY(f.margin, y)
	pdf.Cell(0, 5, "Pague com PIX")
	y += 6

	if data.QRCodeImage != nil {
		f.addImageFromBytes(pdf, data.QRCodeImage, f.margin, y, QRCodeSize, QRCodeSize)
	}

	if data.Boleto.PIX != nil && data.Boleto.PIX.TxID != "" {
		pdf.SetFont("Arial", "", 7)
		pdf.SetXY(f.margin+QRCodeSize+5, y+10)
		pdf.Cell(0, 4, "TxID: "+data.Boleto.PIX.TxID)
	}

	y += QRCodeSize + 5
	return y
}

func (f *Febraban) renderBarcode(pdf *fpdf.Fpdf, data *RenderData, y float64) float64 {
	if data.BarcodeImage != nil {
		// Center barcode
		x := (f.pageWidth - BarcodeWidth) / 2
		f.addImageFromBytes(pdf, data.BarcodeImage, x, y, BarcodeWidth, BarcodeHeight)
	}
	y += BarcodeHeight + 3
	return y
}

func (f *Febraban) renderDigitableLine(pdf *fpdf.Fpdf, data *RenderData, y float64) {
	pdf.SetFont("Arial", "B", 10)
	pdf.SetXY(f.margin, y)
	pdf.Cell(0, 5, FormatDigitableLine(data.Boleto.DigitiableLine))
}

func (f *Febraban) renderField(pdf *fpdf.Fpdf, x, y, width float64, label, value string) {
	pdf.SetFont("Arial", "", 6)
	pdf.SetXY(x, y)
	pdf.Cell(width, 3, label)

	pdf.SetFont("Arial", "", 8)
	pdf.SetXY(x, y+3)
	pdf.Cell(width, 5, value)

	// Border
	pdf.Rect(x, y, width, 9, "D")
}

func (f *Febraban) addImageFromBytes(pdf *fpdf.Fpdf, imgBytes []byte, x, y, maxW, maxH float64) {
	// Decode image to get dimensions
	reader := bytes.NewReader(imgBytes)
	img, _, err := image.DecodeConfig(reader)
	if err != nil {
		return
	}

	// Calculate scaled dimensions maintaining aspect ratio
	w := float64(img.Width)
	h := float64(img.Height)
	ratio := w / h

	finalW := maxW
	finalH := maxW / ratio

	if finalH > maxH {
		finalH = maxH
		finalW = maxH * ratio
	}

	// Register and place image
	name := strconv.FormatInt(int64(x*1000+y*100), 36)
	pdf.RegisterImageOptionsReader(name, fpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(imgBytes))
	pdf.ImageOptions(name, x, y, finalW, finalH, false, fpdf.ImageOptions{}, 0, "")
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd boleto && go test ./renderer/... -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add boleto/renderer/febraban.go boleto/renderer/febraban_test.go
git commit -m "feat(boleto/renderer): add FEBRABAN layout implementation"
```

---

### Task 10: Generator - Orchestrator

**Files:**
- Create: `boleto/generator.go`
- Create: `boleto/options.go`
- Create: `boleto/generator_test.go`

**Interfaces:**
- Consumes: `validator.ValidatorI`, `validator.New()` from Task 4; `barcode.GeneratorI`, `barcode.New()` from Task 7; `renderer.RendererI`, `renderer.NewFebraban()`, `renderer.RenderData` from Tasks 8-9; `Boleto` struct from Task 1
- Produces: `Generator` struct; `NewGenerator(opts ...Option) (*Generator, error)`; `Generate(ctx, *Boleto) ([]byte, error)`; `GenerateFile(ctx, *Boleto, filepath) error`; `Option` type; option functions

- [ ] **Step 1: Create options.go**

Create `boleto/options.go`:
```go
package boleto

import (
	"github.com/braiphub/go-core/boleto/barcode"
	"github.com/braiphub/go-core/boleto/renderer"
	"github.com/braiphub/go-core/boleto/validator"
)

// Option configures the Generator.
type Option func(*Generator)

// WithValidator sets a custom validator.
func WithValidator(v validator.ValidatorI) Option {
	return func(g *Generator) {
		g.validator = v
	}
}

// WithBarcodeGenerator sets a custom barcode generator.
func WithBarcodeGenerator(b barcode.GeneratorI) Option {
	return func(g *Generator) {
		g.barcode = b
	}
}

// WithRenderer sets a custom renderer.
func WithRenderer(r renderer.RendererI) Option {
	return func(g *Generator) {
		g.renderer = r
	}
}

// WithBankLogo registers a bank logo by bank code.
func WithBankLogo(bankCode string, logo []byte) Option {
	return func(g *Generator) {
		g.bankLogos[bankCode] = logo
	}
}

// WithoutValidation disables validation (use with caution).
func WithoutValidation() Option {
	return func(g *Generator) {
		g.skipValidation = true
	}
}
```

- [ ] **Step 2: Write failing test for Generator**

Create `boleto/generator_test.go`:
```go
package boleto

import (
	"bytes"
	"context"
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
```

- [ ] **Step 3: Run test to verify it fails**

```bash
cd boleto && go test -v -run "TestGenerator"
```

Expected: FAIL - `NewGenerator` undefined.

- [ ] **Step 4: Implement Generator**

Create `boleto/generator.go`:
```go
package boleto

import (
	"context"
	"os"

	"github.com/braiphub/go-core/boleto/barcode"
	"github.com/braiphub/go-core/boleto/renderer"
	"github.com/braiphub/go-core/boleto/validator"
	"github.com/pkg/errors"
)

// Generator orchestrates boleto PDF generation.
type Generator struct {
	validator      validator.ValidatorI
	barcode        barcode.GeneratorI
	renderer       renderer.RendererI
	bankLogos      map[string][]byte
	skipValidation bool
}

// NewGenerator creates a new Generator with the given options.
func NewGenerator(opts ...Option) (*Generator, error) {
	g := &Generator{
		validator: validator.New(),
		barcode:   barcode.New(),
		renderer:  renderer.NewFebraban(),
		bankLogos: make(map[string][]byte),
	}

	for _, opt := range opts {
		opt(g)
	}

	return g, nil
}

// Generate creates a PDF from the boleto data and returns the bytes.
func (g *Generator) Generate(ctx context.Context, b *Boleto) ([]byte, error) {
	if b == nil {
		return nil, ErrNilBoleto
	}

	// Validate
	if !g.skipValidation {
		if err := g.validator.Validate(ctx, b); err != nil {
			return nil, errors.Wrap(err, "validation failed")
		}
	}

	// Generate barcode image
	barcodeImg, err := g.barcode.GenerateITF(ctx, b.Barcode)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate barcode")
	}

	// Generate QR code if PIX is present
	var qrCodeImg []byte
	if b.PIX != nil && b.PIX.EMV != "" {
		qrCodeImg, err = g.barcode.GenerateQRCode(ctx, b.PIX.EMV)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate QR code")
		}
	}

	// Get bank logo
	bankLogo := g.bankLogos[b.BankCode]

	// Render PDF
	data := &renderer.RenderData{
		Boleto:       b,
		BarcodeImage: barcodeImg,
		QRCodeImage:  qrCodeImg,
		BankLogo:     bankLogo,
	}

	pdf, err := g.renderer.Render(ctx, data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render PDF")
	}

	return pdf, nil
}

// GenerateFile creates a PDF and writes it to the specified file path.
func (g *Generator) GenerateFile(ctx context.Context, b *Boleto, filePath string) error {
	pdf, err := g.Generate(ctx, b)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filePath, pdf, 0644); err != nil {
		return errors.Wrap(err, "failed to write PDF file")
	}

	return nil
}
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
cd boleto && go test -v -run "TestGenerator"
```

Expected: PASS

- [ ] **Step 6: Run all tests**

```bash
cd boleto && go test ./... -v
```

Expected: All tests PASS

- [ ] **Step 7: Commit**

```bash
git add boleto/generator.go boleto/options.go boleto/generator_test.go
git commit -m "feat(boleto): add Generator orchestrator with full integration"
```

---

### Task 11: Example & Documentation

**Files:**
- Create: `boleto/example/main.go`
- Create: `boleto/README.md`

**Interfaces:**
- Consumes: All previous tasks
- Produces: Runnable example, documentation

- [ ] **Step 1: Create example/main.go**

Create `boleto/example/main.go`:
```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/braiphub/go-core/boleto"
	"github.com/braiphub/go-core/boleto/renderer"
)

func main() {
	// Create generator with custom options
	gen, err := boleto.NewGenerator(
		// Customize renderer
		boleto.WithRenderer(renderer.NewFebraban(
			renderer.WithPrimaryColor("#003366"),
		)),
		// Note: In production, load actual bank logos:
		// boleto.WithBankLogo("001", logoBB),
	)
	if err != nil {
		log.Fatalf("Failed to create generator: %v", err)
	}

	// Create boleto
	b := &boleto.Boleto{
		// Identification
		Barcode:        "00191234500000100000000000000000000000000000",
		DigitiableLine: "00190000090123456789500000000009100000000010000",
		OurNumber:      "12345678",
		DocumentNumber: "DM-001",

		// Bank
		BankCode:     "001",
		BankName:     "Banco do Brasil",
		Agency:       "1234",
		AgencyDigit:  "5",
		Account:      "56789",
		AccountDigit: "0",

		// Beneficiary
		Beneficiary: boleto.Beneficiary{
			Name:       "Minha Empresa LTDA",
			Document:   "12.345.678/0001-90",
			Address:    "Rua das Flores, 100",
			City:       "São Paulo",
			State:      "SP",
			PostalCode: "01234-567",
		},

		// Payer
		Payer: boleto.Payer{
			Name:       "João da Silva",
			Document:   "529.982.247-25",
			Address:    "Av. Brasil, 500",
			City:       "Rio de Janeiro",
			State:      "RJ",
			PostalCode: "20000-000",
		},

		// Values and dates
		Value:          100.00,
		DueDate:        time.Date(2024, 12, 15, 0, 0, 0, 0, time.Local),
		DocumentDate:   time.Now(),
		ProcessingDate: time.Now(),

		// PIX (optional)
		PIX: &boleto.PIXInfo{
			EMV:  "00020126580014br.gov.bcb.pix0136a1b2c3d4-e5f6-7890-abcd-ef1234567890",
			TxID: "ABC123",
		},

		// Instructions
		Instructions: []string{
			"Não receber após o vencimento",
			"Multa de 2% após o vencimento",
			"Juros de 1% ao mês",
		},

		// Description
		Description: []string{
			"Referente à mensalidade de Dezembro/2024",
		},
	}

	// Generate PDF
	ctx := context.Background()

	if err := gen.GenerateFile(ctx, b, "boleto.pdf"); err != nil {
		log.Fatalf("Failed to generate boleto: %v", err)
	}

	log.Println("Boleto generated: boleto.pdf")
}
```

- [ ] **Step 2: Create README.md**

Create `boleto/README.md`:
```markdown
# Boleto PDF Generator

Go library for generating Brazilian bank slips (boletos) in PDF format following FEBRABAN standards, with optional PIX QR code support.

## Installation

```bash
go get github.com/braiphub/go-core/boleto
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/braiphub/go-core/boleto"
)

func main() {
    // Create generator
    gen, _ := boleto.NewGenerator()

    // Create boleto
    b := &boleto.Boleto{
        Barcode:        "00191234500000100000000000000000000000000000",
        DigitiableLine: "00190000090123456789500000000009100000000010000",
        OurNumber:      "12345678",
        BankCode:       "001",
        BankName:       "Banco do Brasil",
        Agency:         "1234",
        Account:        "56789",
        Beneficiary: boleto.Beneficiary{
            Name:     "Company Name",
            Document: "12.345.678/0001-90",
        },
        Payer: boleto.Payer{
            Name:     "John Doe",
            Document: "529.982.247-25",
        },
        Value:   100.00,
        DueDate: time.Now().AddDate(0, 0, 30),
    }

    // Generate PDF
    ctx := context.Background()
    if err := gen.GenerateFile(ctx, b, "boleto.pdf"); err != nil {
        log.Fatal(err)
    }
}
```

## Features

- **FEBRABAN-compliant layout** - Standard bank slip format accepted by all Brazilian banks
- **PIX QR code support** - Optional PIX payment QR code
- **ITF-25 barcode generation** - Automatic barcode generation from 44-digit code
- **Full validation** - CPF/CNPJ, barcode checksum, required fields
- **Customizable** - Colors, margins, sections, bank logos
- **No CGO** - Pure Go, compiles anywhere

## Configuration

### Generator Options

```go
gen, _ := boleto.NewGenerator(
    // Add bank logos
    boleto.WithBankLogo("001", logoBBBytes),
    boleto.WithBankLogo("999", myBankLogoBytes),

    // Custom validator
    boleto.WithValidator(myValidator),

    // Custom renderer
    boleto.WithRenderer(renderer.NewFebraban(
        renderer.WithPrimaryColor("#003366"),
        renderer.WithMargin(15),
        renderer.WithoutReceipt(),
    )),
)
```

### Renderer Options

```go
r := renderer.NewFebraban(
    renderer.WithPrimaryColor("#003366"),  // Primary color (hex)
    renderer.WithMargin(15),                // Page margin in mm
    renderer.WithPageSize(210, 297),        // Page size in mm (default: A4)
    renderer.WithoutReceipt(),              // Hide payer receipt section
    renderer.WithoutPIXSection(),           // Hide PIX section
)
```

## PIX Support

To include a PIX QR code on the boleto:

```go
b := &boleto.Boleto{
    // ... other fields
    PIX: &boleto.PIXInfo{
        EMV:  "00020126580014br.gov.bcb.pix0136...", // PIX EMV string
        TxID: "ABC123",                              // Transaction ID (optional)
    },
}
```

## Validation

The library validates:

- **Barcode** - 44 digits with valid checksum
- **Digitable line** - 47 digits
- **CPF** - 11 digits with valid check digits
- **CNPJ** - 14 digits with valid check digits
- **Required fields** - Beneficiary, payer, value, due date, bank code

Validation errors are descriptive:

```go
if err := gen.Generate(ctx, b); err != nil {
    // err contains details: "beneficiary.document: invalid CNPJ (12.345.678/0001-00)"
}
```

## FEBRABAN Specifications

The generated PDF follows FEBRABAN (Brazilian Banking Federation) standards:

| Element | Dimension |
|---------|-----------|
| Barcode width | 103mm |
| Barcode height | 13mm |
| QR Code PIX | 25x25mm |
| Bank logo | 40x15mm (max) |
| Page size | A4 (210x297mm) |

## License

MIT
```

- [ ] **Step 3: Verify example compiles**

```bash
cd boleto/example && go build -o /dev/null
```

Expected: No errors

- [ ] **Step 4: Commit**

```bash
git add boleto/example/ boleto/README.md
git commit -m "docs(boleto): add example and README documentation"
```

---

### Task 12: Final Integration & go.work

**Files:**
- Modify: `go.work` (root)

**Interfaces:**
- Consumes: All previous tasks
- Produces: Integration with go-core workspace

- [ ] **Step 1: Add boleto to go.work**

Read current `go.work` and add boleto module:

```bash
cd /Users/matheuscarmo/Desktop/projects/go-core
```

Add `./boleto` to the `use` block in `go.work`.

- [ ] **Step 2: Run go work sync**

```bash
go work sync
```

Expected: No errors

- [ ] **Step 3: Run all tests from root**

```bash
go test ./boleto/... -v
```

Expected: All tests PASS

- [ ] **Step 4: Commit**

```bash
git add go.work
git commit -m "feat: add boleto package to workspace"
```

---

## Self-Review Checklist

1. **Spec coverage:**
   - [x] Data structs (Boleto, Beneficiary, Payer, PIXInfo) - Task 1
   - [x] Validator with CPF/CNPJ, barcode validation - Tasks 2-4
   - [x] ITF-25 barcode generation - Tasks 5, 7
   - [x] QR Code PIX generation - Tasks 6-7
   - [x] FEBRABAN renderer with options - Tasks 8-9
   - [x] Generator orchestrator - Task 10
   - [x] Example and documentation - Task 11
   - [x] Workspace integration - Task 12

2. **Placeholder scan:** No TBD, TODO, or incomplete sections.

3. **Type consistency:**
   - `Boleto` struct used consistently across all tasks
   - `ValidatorI`, `GeneratorI`, `RendererI` interfaces match implementations
   - Option functions match struct fields

---

**Plan complete and saved to `docs/superpowers/plans/2026-09-22-boleto-pdf-implementation.md`.**

Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

Which approach?
