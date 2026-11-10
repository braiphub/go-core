package renderer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatBankCode(t *testing.T) {
	cases := map[string]string{
		"001": "001-9",
		"237": "237-2",
		"341": "341-7",
		"033": "033-7",
		"104": "104-0",
		"12":  "12",
		"":    "",
	}

	for code, expected := range cases {
		assert.Equal(t, expected, FormatBankCode(code), code)
	}
}

func TestFormatAgencyAccount(t *testing.T) {
	assert.Equal(t, "1234-5 / 56789-0", FormatAgencyAccount("1234", "5", "56789", "0"))
	assert.Equal(t, "1234 / 56789-0", FormatAgencyAccount("1234", "", "56789", "0"))
	assert.Equal(t, "1234-5 / 56789", FormatAgencyAccount("1234", "5", "56789", ""))
	assert.Equal(t, "1234 / 56789", FormatAgencyAccount("1234", "", "56789", ""))
}
