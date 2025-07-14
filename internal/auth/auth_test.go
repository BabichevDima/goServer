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

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name          string
		headers       map[string]string
		expectedToken string
		expectedError error
	}{
		{
			name:          "Valid token",
			headers:       map[string]string{"Authorization": "Bearer valid.token.123"},
			expectedToken: "valid.token.123",
			expectedError: nil,
		},
		{
			name:          "No auth header",
			headers:       map[string]string{},
			expectedToken: "",
			expectedError: ErrNoAuthHeader,
		},
		{
			name:          "Malformed header - no Bearer",
			headers:       map[string]string{"Authorization": "Token valid.token.123"},
			expectedToken: "",
			expectedError: ErrMalformedAuthHeader,
		},
		{
			name:          "Malformed header - empty token",
			headers:       map[string]string{"Authorization": "Bearer "},
			expectedToken: "",
			expectedError: ErrMalformedAuthHeader,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := http.Header{}
			for k, v := range tt.headers {
				headers.Set(k, v)
			}

			token, err := GetBearerToken(headers)

			assert.Equal(t, tt.expectedToken, token)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}