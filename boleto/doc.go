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
