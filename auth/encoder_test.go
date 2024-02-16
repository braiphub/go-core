package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestEncodeAndDecode(t *testing.T) {
	type testStruct struct {
		Data string
	}
	input := testStruct{
		Data: "test",
	}

	encoder := NewEncoder[testStruct](
		"jwtSecret",
		time.Minute,
		"encryptKey",
		"encryptSalt",
		"dataField",
		WithEncoderOptionJWTSigningMethod[testStruct](jwt.SigningMethodHS256),
	)

	token, err := encoder.GenerateJWTToken(input)
	assert.NoError(t, err)

	decoder := NewDecoder[testStruct](
		"jwtSecret",
		"encryptKey",
		"dataField",
	)
	output, err := decoder.Data(token)
	assert.NoError(t, err)

	assert.Equal(t, &input, output)
}
