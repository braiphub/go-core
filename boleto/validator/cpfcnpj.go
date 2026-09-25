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
