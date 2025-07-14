package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestJWT(t *testing.T) {
	secret := "test-secret"
	userID := uuid.New()
	expiresIn := time.Hour

	// Тест создания токена
	token, err := MakeJWT(userID, secret, expiresIn)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Тест валидации токена
	parsedID, err := ValidateJWT(token, secret)
	assert.NoError(t, err)
	assert.Equal(t, userID, parsedID)

	// Тест с неправильным секретом
	_, err = ValidateJWT(token, "wrong-secret")
	assert.Error(t, err)

	// Тест с истекшим токеном
	expiredToken, err := MakeJWT(userID, secret, -time.Hour)
	assert.NoError(t, err)
	_, err = ValidateJWT(expiredToken, secret)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}