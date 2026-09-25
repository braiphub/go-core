// Package assets provides embedded bank logos for boleto generation.
package assets

import (
	_ "embed"
)

// LogoBancoDoBrasil contains the Banco do Brasil logo PNG bytes.
//
//go:embed bb-logo.png
var LogoBancoDoBrasil []byte

// LogoHPag contains the HPag logo PNG bytes.
//
//go:embed hpag-logo.png
var LogoHPag []byte
