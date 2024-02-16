package auth

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

var (
	ErrEmptyParam       = errors.New("param is empty")
	ErrInvalidIVLength  = errors.New("iv must be equal to block-size")
	ErrInvalidKeyLength = errors.New("key length must be greater than or equal to 32")
)

type Encoder[T any] struct {
	jwtSecret         string
	tokenDuration     time.Duration
	encryptKey        string
	encryptSalt       string
	dataField         string
	issuer            string
	userProviderClass string
	jwtSigner         jwt.SigningMethod
}

func NewEncoder[T any](
	jwtSecret string,
	tokenDuration time.Duration,
	encryptKey string,
	encryptSalt string,
	dataField string,
	opts ...func(*Encoder[T]),
) Encoder[T] {
	encoder := Encoder[T]{
		jwtSecret:     jwtSecret,
		tokenDuration: tokenDuration,
		encryptKey:    encryptKey,
		encryptSalt:   encryptSalt,
		dataField:     dataField,
	}

	for _, o := range opts {
		o(&encoder)
	}

	return encoder
}

func (e *Encoder[T]) GenerateJWTToken(input T) (string, error) {
	encryptedData, err := e.encrypt(input)
	if err != nil {
		return "", err
	}

	// claims
	claims := jwt.MapClaims{
		"iat":       time.Now().Unix(),
		"nbf":       time.Now().Unix(),
		"exp":       time.Now().Add(e.tokenDuration).Unix(),
		e.dataField: encryptedData,
	}

	if e.issuer != "" {
		claims["iss"] = e.issuer
	}

	if e.userProviderClass != "" {
		claims["prv"] = md5hash(e.userProviderClass)
	}

	// token generate
	token := jwt.NewWithClaims(e.jwtSigner, claims)

	tokenString, err := token.SignedString([]byte(e.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (e *Encoder[T]) encrypt(input T) (string, error) {
	inputBuf, err := json.Marshal(input)
	if err != nil {
		return "", errors.Wrap(err, "marshal input")
	}

	encryptKey, err := hashSha512(e.encryptKey)
	if err != nil {
		return "", errors.Wrap(err, "sha512 encryptKey")
	}

	salt, err := hashSha512(e.encryptSalt)
	if err != nil {
		return "", errors.Wrap(err, "sha512 salt")
	}

	key, err := hashSha512(encryptKey + salt)
	if err != nil {
		return "", errors.Wrap(err, "sha512 key + salt")
	}

	pwd, err := hashSha512(randomString(32))
	if err != nil {
		return "", errors.Wrap(err, "sha512 pwd/iv")
	}
	iv := pwd[:aes.BlockSize]

	encrypted, err := aes256cbcEncB64(string(inputBuf), key, iv, aes.BlockSize)
	if err != nil {
		return "", errors.Wrap(err, "encode-aes-256")
	}

	encrypted = fmt.Sprintf("%s:%s:%s", iv, salt, encrypted)

	return encrypted, nil
}

func md5hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func randomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const (
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	)

	var src = rand.NewSource(time.Now().UnixNano())

	sb := strings.Builder{}
	sb.Grow(n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}

func hashSha512(str string) (string, error) {
	if str == "" {
		return "", errors.Wrap(ErrEmptyParam, "input str")
	}
	hasher := sha512.New()
	if _, err := hasher.Write([]byte(str)); err != nil {
		return "", errors.Wrap(err, "hasher write")
	}
	hash := hasher.Sum(nil)
	hashHexStr := hex.EncodeToString(hash)

	return hashHexStr, nil
}

func aes256cbcEncB64(plaintext string, key string, initialValue string, blockSize int) (string, error) {
	const keySize = 32

	if len(initialValue) != blockSize {
		return "", ErrInvalidIVLength
	}
	if len(key) < keySize {
		return "", ErrInvalidKeyLength
	}

	bKey := []byte(key[:keySize])
	bIV := []byte(initialValue)
	bPlaintext := pkcs5Padding([]byte(plaintext), blockSize)
	block, err := aes.NewCipher(bKey)
	if err != nil {
		return "", errors.Wrap(err, "init cipher")
	}
	ciphertext := make([]byte, len(bPlaintext))
	mode := cipher.NewCBCEncrypter(block, bIV)
	mode.CryptBlocks(ciphertext, bPlaintext)
	ciphertextEncoded := base64.StdEncoding.EncodeToString(ciphertext)

	return ciphertextEncoded, nil
}

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := (blockSize - len(ciphertext)%blockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)

	return append(ciphertext, padtext...)
}
