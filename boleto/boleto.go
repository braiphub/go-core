package boleto

import "github.com/braiphub/go-core/boleto/internal/types"

// Re-export types from internal package to maintain public API.
type (
	Boleto      = types.Boleto
	Beneficiary = types.Beneficiary
	Payer       = types.Payer
	PIXInfo     = types.PIXInfo
)
