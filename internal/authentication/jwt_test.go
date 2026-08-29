//go:build unit

// Сборка только для юнит-тестов (запуск: go test -tags=unit)
// Юнит-тесты используют моки и не требуют реальной БД

package authentication

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

// ============================================================
// 1. ТЕСТЫ ДЛЯ ГЕНЕРАЦИИ ТОКЕНА (GenerateToken)
// ============================================================

// TestGenerateToken_Success — проверяет успешную генерацию JWT-токена
func TestGenerateToken_Success(t *testing.T) {
	// 1. Устанавливаем тестовый секрет
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET") // очищаем после теста

	// 2. Генерируем токен
	token, err := GenerateToken("user-123", "test@mail.com", "user")

	// 3. Проверяем, что ошибки нет
	assert.NoError(t, err)
	// 4. Проверяем, что токен не пустой
	assert.NotEmpty(t, token)
}

// TestGenerateToken_MissingSecret — проверяет ошибку при отсутствии JWT_SECRET
func TestGenerateToken_MissingSecret(t *testing.T) {
	// 1. Удаляем секрет (если был)
	os.Unsetenv("JWT_SECRET")

	// 2. Пытаемся сгенерировать токен
	token, err := GenerateToken("user-123", "test@mail.com", "user")

	// 3. Проверяем, что вернулась ошибка
	assert.Error(t, err)
	// 4. Проверяем текст ошибки
	assert.Equal(t, "JWT_SECRET is not set in .env file", err.Error())
	// 5. Проверяем, что токен пустой
	assert.Empty(t, token)
}

// ============================================================
// 2. ТЕСТЫ ДЛЯ ВАЛИДАЦИИ ТОКЕНА (ValidateToken)
// ============================================================

// TestValidateToken_Success — проверяет успешную валидацию токена
func TestValidateToken_Success(t *testing.T) {
	// 1. Устанавливаем тестовый секрет
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	// 2. Генерируем токен
	tokenString, err := GenerateToken("user-123", "test@mail.com", "user")
	assert.NoError(t, err)

	// 3. Валидируем токен
	claims, err := ValidateToken(tokenString)

	// 4. Проверяем, что ошибки нет
	assert.NoError(t, err)
	// 5. Проверяем, что claims не пустые
	assert.NotNil(t, claims)
	// 6. Проверяем данные пользователя
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "test@mail.com", claims.Email)
	assert.Equal(t, "user", claims.Role)
}

// TestValidateToken_InvalidToken — проверяет ошибку при невалидном токене
func TestValidateToken_InvalidToken(t *testing.T) {
	// 1. Устанавливаем секрет
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	// 2. Передаём невалидный токен
	claims, err := ValidateToken("invalid-token")

	// 3. Проверяем, что вернулась ошибка
	assert.Error(t, err)
	// 4. Проверяем, что claims пустые
	assert.Nil(t, claims)
}

// TestValidateToken_ExpiredToken — проверяет ошибку при просроченном токене
func TestValidateToken_ExpiredToken(t *testing.T) {
	// 1. Устанавливаем секрет
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	// 2. Создаём токен с истекшим сроком
	claims := Claims{
		UserID: "user-123",
		Email:  "test@mail.com",
		Role:   "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), // просрочен на 1 час
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret-key"))

	// 3. Пытаемся валидировать
	_, err := ValidateToken(tokenString)

	// 4. Проверяем, что вернулась ошибка
	assert.Error(t, err)
}

// TestValidateToken_MissingSecret — проверяет ошибку при отсутствии секрета
func TestValidateToken_MissingSecret(t *testing.T) {
	// 1. Удаляем секрет
	os.Unsetenv("JWT_SECRET")

	// 2. Пытаемся валидировать любой токен
	claims, err := ValidateToken("some-token")

	// 3. Проверяем ошибку
	assert.Error(t, err)
	assert.Equal(t, "JWT_SECRET is not set", err.Error())
	assert.Nil(t, claims)
}

// ============================================================
// 3. ТЕСТЫ ДЛЯ ПОЛУЧЕНИЯ USER_ID ИЗ ТОКЕНА (GetUserIDFromToken)
// ============================================================

// TestGetUserIDFromToken_Success — проверяет успешное извлечение user_id
func TestGetUserIDFromToken_Success(t *testing.T) {
	// 1. Устанавливаем секрет
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	// 2. Генерируем токен
	tokenString, _ := GenerateToken("user-123", "test@mail.com", "user")

	// 3. Извлекаем user_id
	userID, err := GetUserIDFromToken(tokenString)

	// 4. Проверяем
	assert.NoError(t, err)
	assert.Equal(t, "user-123", userID)
}

// TestGetUserIDFromToken_InvalidToken — проверяет ошибку при невалидном токене
func TestGetUserIDFromToken_InvalidToken(t *testing.T) {
	// 1. Устанавливаем секрет
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	// 2. Передаём невалидный токен
	userID, err := GetUserIDFromToken("invalid-token")

	// 3. Проверяем ошибку
	assert.Error(t, err)
	assert.Empty(t, userID)
}

// TestGetUserIDFromToken_MissingSecret — проверяет ошибку при отсутствии секрета
func TestGetUserIDFromToken_MissingSecret(t *testing.T) {
	// 1. Удаляем секрет
	os.Unsetenv("JWT_SECRET")

	// 2. Пытаемся получить user_id
	userID, err := GetUserIDFromToken("some-token")

	// 3. Проверяем ошибку
	assert.Error(t, err)
	assert.Empty(t, userID)
}
