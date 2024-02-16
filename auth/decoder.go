package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

var (
	ErrMissingParam         = errors.New("parameter is missing")
	ErrInvalidToken         = errors.New("your token is invalid")
	ErrInvalidEncryptedData = errors.New("invalid encrypted data")
)

type Decoder[T any] struct {
	jwtSecret  string
	encryptKey string
	dataField  string
}

func NewDecoder[T any](
	jwtSecret string,
	encryptKey string,
	dataField string,
	opts ...func(*Decoder[T]),
) Decoder[T] {
	decoder := Decoder[T]{
		jwtSecret:  jwtSecret,
		encryptKey: encryptKey,
		dataField:  dataField,
	}

	for _, o := range opts {
		o(&decoder)
	}

	return decoder
}

func (d *Decoder[T]) Data(token string) (*T, error) {
	// // parse token
	tokenJwt, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.Wrap(ErrInvalidToken, fmt.Sprintf("unexpected signing method: %v", token.Header["alg"]))
		}
		hmacSampleSecret := []byte(d.jwtSecret)

		return hmacSampleSecret, nil
	})
	if err != nil {
		return nil, errors.Wrap(ErrInvalidToken, err.Error())
	}

	claims, ok := tokenJwt.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.Wrap(ErrInvalidToken, "claims not found")
	}
	if claims[d.dataField] == nil {
		return nil, errors.Wrap(ErrInvalidToken, "data-field not found")
	}

	decrypted, err := d.decrypt([]byte(claims[d.dataField].(string)))
	if err != nil {
		return nil, errors.Wrap(err, "data-field decrypt")
	}

	return decrypted, nil
}

func (d *Decoder[T]) decrypt(data []byte) (*T, error) {
	const expectParts = 3

	parts := strings.Split(string(data), ":")
	if len(parts) != expectParts {
		return nil, errors.Wrap(ErrInvalidEncryptedData, "expect splitted len of 3")
	}

	buf, err := decrypt(
		parts[2], // data
		d.encryptKey,
		parts[1], // salt
		parts[0], // iv
	)
	if err != nil {
		return nil, errors.Wrap(err, "decrypt")
	}

	dataStruct := new(T)
	if err := json.Unmarshal(buf, dataStruct); err != nil {
		return nil, errors.Wrap(err, "unmarshal into specified struct")
	}

	return dataStruct, nil
}

func decrypt(input, encryptKey, salt, initialValue string) ([]byte, error) {
	encryptKey, err := hashSha512(encryptKey)
	if err != nil {
		return nil, errors.Wrap(err, "sha512 encryptKey")
	}

	key, err := hashSha512(encryptKey + salt)
	if err != nil {
		return nil, errors.Wrap(err, "sha512 key + salt")
	}

	decrypted, err := aes256cbcDecB64(input, key, initialValue, aes.BlockSize)
	if err != nil {
		return nil, errors.Wrap(err, "aes256 decrypt")
	}

	return decrypted, nil
}

func aes256cbcDecB64(cipherText string, key string, initialValue string, blockSize int) ([]byte, error) {
	const keySize = 32

	if len(initialValue) != blockSize {
		return nil, ErrInvalidIVLength
	}
	if len(key) < keySize {
		return nil, ErrInvalidKeyLength
	}

	bKey := []byte(key[:keySize])
	bIV := []byte(initialValue)
	cipherTextDecoded, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return nil, errors.Wrap(err, "base64 decode")
	}

	block, err := aes.NewCipher(bKey)
	if err != nil {
		return nil, errors.Wrap(err, "init cipher")
	}

	mode := cipher.NewCBCDecrypter(block, bIV)
	mode.CryptBlocks(cipherTextDecoded, cipherTextDecoded)

	return pkcs5UnPadding(cipherTextDecoded), nil
}

func pkcs5UnPadding(src []byte) []byte {
	length := len(src)
	unpadding := int(src[length-1])

	return src[:(length - unpadding)]
}
