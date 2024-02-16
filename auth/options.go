package auth

import "github.com/golang-jwt/jwt/v5"

func WithEncoderOptionJWTSigningMethod[T any](method jwt.SigningMethod) func(encoder *Encoder[T]) {
	return func(encoder *Encoder[T]) {
		encoder.jwtSigner = method
	}
}
