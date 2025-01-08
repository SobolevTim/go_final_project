package server

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// generateJWTToken generates a JWT token valid for 24 hours
func generateJWTToken(pass string) (string, error) {
	// Хэшируем текущий пароль из переменной окружения
	hash := sha256.New()
	hash.Write([]byte(pass))
	expectedHash := hex.EncodeToString(hash.Sum(nil))

	// Создаем claims с хэшированным паролем
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(8 * time.Hour)),
		Issuer:    expectedHash, // сохраняем хэш пароля в поле Issuer
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWTSecretKey)
}
