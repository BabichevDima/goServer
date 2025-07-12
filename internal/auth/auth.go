package auth

import (
	"golang.org/x/crypto/bcrypt"
)

const (
	// Cost определяет сложность хеширования (4-31)
	// Рекомендуемое значение для production: 10-14
	Cost = 12
)

// HashPassword хеширует пароль с использованием bcrypt
func HashPassword(password string) (string, error) {
	// GenerateFromPassword возвращает bcrypt хеш пароля
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), Cost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// CheckPasswordHash сравнивает пароль с его хешем
func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}