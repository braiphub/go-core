package main

import (
	"context"
	"log"
	"time"

	"github.com/braiphub/go-core/boleto"
	"github.com/braiphub/go-core/boleto/assets"
	"github.com/braiphub/go-core/boleto/renderer"
)

func main() {
	// Create generator with custom options
	gen, err := boleto.NewGenerator(
		// Customize renderer
		boleto.WithRenderer(renderer.NewFebraban(
			renderer.WithPrimaryColor("#003366"),
		)),
		// Bank logos (embedded)
		boleto.WithBankLogo("001", assets.LogoBancoDoBrasil),
		boleto.WithBankLogo("329", assets.LogoHPag),
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
		BankCode:     "329",
		BankName:     "HPag",
		Agency:       "1234",
		AgencyDigit:  "5",
		Account:      "56789",
		AccountDigit: "0",

		// Beneficiary
		Beneficiary: boleto.Beneficiary{
			Name:       "Minha Empresa LTDA",
			Document:   "11.222.333/0001-81",
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
