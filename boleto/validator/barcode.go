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
