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
