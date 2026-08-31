//go:build unit

// Сборка только для юнит-тестов (запуск: go test -tags=unit)
// Юнит-тесты не требуют реальной БД и используют моки

package authentication

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================
// ТЕСТЫ ДЛЯ MIDDLEWARE AuthMiddleware
// ============================================================

// TestAuthMiddleware_Success — проверяет успешный проход запроса через middleware
// с валидным JWT-токеном в заголовке Authorization.
// Ожидаемый результат: HTTP 200 OK
func TestAuthMiddleware_Success(t *testing.T) {
	// 1. Устанавливаем тестовый секрет для подписи JWT
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET") // очищаем после теста

	// 2. Генерируем валидный JWT-токен
	token, err := GenerateToken("user-123", "test@mail.com", "user")
	assert.NoError(t, err) // проверяем, что токен создался без ошибок

	// 3. Создаём тестовый хэндлер, который должен выполниться после middleware
	// Он просто возвращает статус 200 OK
	next := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	// 4. Оборачиваем хэндлер в middleware AuthMiddleware
	handler := AuthMiddleware(next)

	// 5. Создаём тестовый HTTP-запрос
	req := httptest.NewRequest("GET", "/", nil)
	// 6. Добавляем заголовок Authorization с токеном
	req.Header.Set("Authorization", "Bearer "+token)

	// 7. Создаём ResponseRecorder для записи ответа
	w := httptest.NewRecorder()

	// 8. Вызываем handler
	handler(w, req)

	// 9. Проверяем, что статус ответа 200 OK (успешный проход)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestAuthMiddleware_MissingHeader — проверяет, что middleware возвращает ошибку,
// если заголовок Authorization отсутствует.
// Ожидаемый результат: HTTP 401 Unauthorized
func TestAuthMiddleware_MissingHeader(t *testing.T) {
	// 1. Создаём тестовый хэндлер (не должен вызваться)
	next := func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	}

	// 2. Оборачиваем в middleware
	handler := AuthMiddleware(next)

	// 3. Создаём запрос БЕЗ заголовка Authorization
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// 4. Вызываем handler
	handler(w, req)

	// 5. Проверяем статус 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	// 6. Проверяем сообщение об ошибке
	assert.Contains(t, w.Body.String(), "Authorization header required")
}

// TestAuthMiddleware_InvalidFormat — проверяет, что middleware возвращает ошибку,
// если заголовок Authorization имеет неправильный формат (не начинается с "Bearer ").
// Ожидаемый результат: HTTP 401 Unauthorized
func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	// 1. Устанавливаем секрет (для генерации, но не используется)
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	// 2. Создаём тестовый хэндлер (не должен вызваться)
	next := func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	}

	// 3. Оборачиваем в middleware
	handler := AuthMiddleware(next)

	// 4. Создаём запрос с неправильным префиксом (не "Bearer")
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "InvalidPrefix some-token")
	w := httptest.NewRecorder()

	// 5. Вызываем handler
	handler(w, req)

	// 6. Проверяем статус 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	// 7. Проверяем сообщение об ошибке
	assert.Contains(t, w.Body.String(), "Invalid authorization format")
}

// TestAuthMiddleware_InvalidToken — проверяет, что middleware возвращает ошибку,
// если передан невалидный (битый) JWT-токен.
// Ожидаемый результат: HTTP 401 Unauthorized
func TestAuthMiddleware_InvalidToken(t *testing.T) {
	// 1. Устанавливаем секрет
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	// 2. Создаём тестовый хэндлер (не должен вызваться)
	next := func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	}

	// 3. Оборачиваем в middleware
	handler := AuthMiddleware(next)

	// 4. Создаём запрос с заведомо невалидным токеном
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	// 5. Вызываем handler
	handler(w, req)

	// 6. Проверяем статус 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	// 7. Проверяем сообщение об ошибке
	assert.Contains(t, w.Body.String(), "Invalid or expired token")
}

// TestAuthMiddleware_MissingSecret — проверяет, что middleware возвращает ошибку,
// если переменная окружения JWT_SECRET не установлена.
// Ожидаемый результат: HTTP 401 Unauthorized
func TestAuthMiddleware_MissingSecret(t *testing.T) {
	// 1. Удаляем секрет (если был установлен)
	os.Unsetenv("JWT_SECRET")

	// 2. Создаём тестовый хэндлер (не должен вызваться)
	next := func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	}

	// 3. Оборачиваем в middleware
	handler := AuthMiddleware(next)

	// 4. Создаём запрос с любым токеном (проверка секрета происходит до валидации)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer some-token")
	w := httptest.NewRecorder()

	// 5. Вызываем handler
	handler(w, req)

	// 6. Проверяем статус 401
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	// 7. Проверяем, что в ответе нет деталей ошибки (безопасность)
	assert.Contains(t, w.Body.String(), "Invalid or expired token")
}

// TestGetUserID — проверяет, что middleware сохраняет user_id в контексте
// и функция GetUserID может его извлечь.
// Ожидаемый результат: user_id из токена совпадает с извлечённым
func TestGetUserID(t *testing.T) {
	// 1. Устанавливаем секрет
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	// 2. Генерируем токен с известным user_id
	token, err := GenerateToken("user-123", "test@mail.com", "user")
	assert.NoError(t, err)

	// 3. Переменная для захвата user_id из контекста
	var capturedUserID string

	// 4. Создаём тестовый хэндлер, который извлекает user_id из контекста
	next := func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = GetUserID(r.Context()) // ← извлекаем user_id
		w.WriteHeader(http.StatusOK)
	}

	// 5. Оборачиваем в middleware
	handler := AuthMiddleware(next)

	// 6. Создаём запрос с токеном
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	// 7. Вызываем handler
	handler(w, req)

	// 8. Проверяем, что хэндлер выполнился (статус 200)
	assert.Equal(t, http.StatusOK, w.Code)
	// 9. Проверяем, что user_id извлечён правильно
	assert.Equal(t, "user-123", capturedUserID)
}

// TestGetUserID_Empty — проверяет возврат пустой строки, если user_id нет
func TestGetUserID_Empty(t *testing.T) {
	ctx := context.Background()
	userID := GetUserID(ctx)
	assert.Equal(t, "", userID)
}
