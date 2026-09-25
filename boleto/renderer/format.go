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
