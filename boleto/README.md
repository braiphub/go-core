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
