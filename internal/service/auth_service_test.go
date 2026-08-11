//go:build unit
// Сборка только для юнит-тестов (запуск: go test -tags=unit)
// Юнит-тесты используют моки и не требуют реальной БД

package service

import (
	"context"
	"errors"
	"testing"

	"Rest-user-agregator/internal/models"
	"github.com/stretchr/testify/assert"
)

// ============================================================
// 1. МОК-РЕПОЗИТОРИЙ ДЛЯ ТЕСТОВ
// ============================================================
// MockUserRepository — реализует интерфейс repository.UserRepository
// Используется для изолированного тестирования AuthService без подключения к БД
//
// ПОЛЯ:
//   - Users: мапа для хранения пользователей (ключ — email)
//   - ShouldBeError: если true — все методы возвращают ошибку
//   - ErrorMessage: текст ошибки, если ShouldBeError = true
//   - CreateUserCallCount: счётчик вызовов CreateUser
//   - GetUserByEmailCallCount: счётчик вызовов GetUserByEmail
//   - LastCreatedUser: последний созданный пользователь (для проверки данных)
//   - LastSearchedEmail: последний email, который искали (для проверки данных)
// ============================================================
type MockUserRepository struct {
	// Хранилище пользователей (ключ — email)
	Users map[string]models.User
	// Управление ошибками
	ShouldBeError bool   // если true — возвращаем ошибку
	ErrorMessage  string // текст ошибки
	// Счётчики вызовов (для проверки, что метод вызван)
	CreateUserCallCount     int
	GetUserByEmailCallCount int
	// Последние переданные данные (для проверки, что передано)
	LastCreatedUser   models.User
	LastSearchedEmail string
}

// NewMockUserRepository — конструктор мока
func NewMockUserRepository() *MockUserRepository {
	// Возвращаем указатель на новый экземпляр мока
	// make(map[string]models.User) — создаём пустую мапу для хранения пользователей
	return &MockUserRepository{
		Users: make(map[string]models.User),
	}
}

// ============================================================
// 2. РЕАЛИЗАЦИЯ ИНТЕРФЕЙСА repository.UserRepository
// ============================================================

// CreateUser — сохраняет пользователя в мок-хранилище
func (m *MockUserRepository) CreateUser(ctx context.Context, user models.User) error {
	// Увеличиваем счётчик вызовов (чтобы в тесте проверить, что метод был вызван)
	m.CreateUserCallCount++
	// Сохраняем переданные данные для проверки в тесте
	m.LastCreatedUser = user

	// Если включён режим ошибки — возвращаем ошибку вместо выполнения
	if m.ShouldBeError {
		if m.ErrorMessage == "" {
			// Если сообщение не задано — возвращаем стандартную ошибку
			return errors.New("database error")
		}
		// Возвращаем заданное сообщение ошибки
		return errors.New(m.ErrorMessage)
	}

	// Проверяем, существует ли уже пользователь с таким email
	if _, exists := m.Users[user.Email]; exists {
		// Возвращаем ошибку, как в реальном репозитории при нарушении уникальности
		return errors.New("user already exists")
	}

	// Сохраняем пользователя в мапу (ключ — email, значение — структура User)
	m.Users[user.Email] = user
	// Возвращаем nil — ошибок нет, операция успешна
	return nil
}

// GetUserByEmail — ищет пользователя по email в мок-хранилище
func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if ctx.Err() != nil {// Проверяем контекст на отмену операции
        return nil, ctx.Err()
    } 
	// Увеличиваем счётчик вызовов (чтобы в тесте проверить, что метод был вызван)
	m.GetUserByEmailCallCount++
	// Сохраняем переданный email для проверки в тесте
	m.LastSearchedEmail = email

	// Если включён режим ошибки — возвращаем ошибку вместо выполнения
	if m.ShouldBeError {
		if m.ErrorMessage == "" {
			return nil, errors.New("database error")
		}
		return nil, errors.New(m.ErrorMessage)
	}

	// Ищем пользователя в мапе по email
	user, exists := m.Users[email]
	if !exists {
		// Пользователь не найден — возвращаем nil, nil (как в реальном репозитории)
		return nil, nil
	}
	// Пользователь найден — возвращаем указатель на него
	return &user, nil
}

// ============================================================
// 3. ТЕСТЫ ДЛЯ РЕГИСТРАЦИИ (Register)
// ============================================================

// TestRegister_Success — проверяет успешную регистрацию нового пользователя
func TestRegister_Success(t *testing.T) {
	// Подготавливаем тестовые данные
	email := "test@mail.com"
	password := "123456"
	role := "user"

	// Создаём мок-репозиторий (пустой, пользователей нет)
	mock := NewMockUserRepository()
	// Создаём сервис с моком (внедрение зависимости)
	svc := NewAuthService(mock)

	// Вызываем метод Register с тестовыми данными
	err := svc.Register(context.Background(), email, password, role)

	// Проверяем, что ошибки нет
	assert.NoError(t, err)
	// Проверяем, что CreateUser был вызван ровно 1 раз
	assert.Equal(t, 1, mock.CreateUserCallCount)
	// Проверяем, что в метод передан правильный email
	assert.Equal(t, email, mock.LastCreatedUser.Email)

	// Проверяем, что пользователь реально сохранился в моке
	user, err := mock.GetUserByEmail(context.Background(), email)
	assert.NoError(t, err)        // ошибки при получении нет
	assert.NotNil(t, user)        // пользователь найден
	assert.Equal(t, email, user.Email) // email совпадает
}

// TestRegister_EmptyEmail — проверяет, что email обязателен
func TestRegister_EmptyEmail(t *testing.T) {
	// Подготавливаем тестовые данные с пустым email
	email := ""
	password := "123456"
	role := "user"

	mock := NewMockUserRepository()
	svc := NewAuthService(mock)

	// Вызываем Register с пустым email
	err := svc.Register(context.Background(), email, password, role)

	// Проверяем, что вернулась ошибка
	assert.Error(t, err)
	// Проверяем, что сообщение ошибки правильное
	assert.Equal(t, "email is required", err.Error())
	// Проверяем, что CreateUser НЕ был вызван (валидация сработала до вызова репозитория)
	assert.Equal(t, 0, mock.CreateUserCallCount)
}

// TestRegister_EmptyPassword — проверяет, что пароль обязателен
func TestRegister_EmptyPassword(t *testing.T) {
	email := "test@mail.com"
	password := "" // пустой пароль
	role := "user"

	mock := NewMockUserRepository()
	svc := NewAuthService(mock)

	err := svc.Register(context.Background(), email, password, role)

	assert.Error(t, err)
	assert.Equal(t, "password is required", err.Error())
	assert.Equal(t, 0, mock.CreateUserCallCount)
}

// TestRegister_PasswordTooShort — проверяет, что пароль должен быть минимум 6 символов
func TestRegister_PasswordTooShort(t *testing.T) {
	email := "test@mail.com"
	password := "123" // пароль короче 6 символов
	role := "user"

	mock := NewMockUserRepository()
	svc := NewAuthService(mock)

	err := svc.Register(context.Background(), email, password, role)

	assert.Error(t, err)
	assert.Equal(t, "password must be at least 6 characters", err.Error())
	assert.Equal(t, 0, mock.CreateUserCallCount)
}

// TestRegister_UserExists — проверяет, что нельзя зарегистрировать существующего пользователя
func TestRegister_UserExists(t *testing.T) {
	email := "test@mail.com"
	password := "123456"
	role := "user"

	mock := NewMockUserRepository()
	// Добавляем существующего пользователя в мок ДО вызова Register
	mock.Users[email] = models.User{Email: email, Role: role}
	svc := NewAuthService(mock)

	err := svc.Register(context.Background(), email, password, role)

	assert.Error(t, err)
	assert.Equal(t, "user already exists", err.Error())
	// CreateUser НЕ вызван, потому что сервис сначала проверяет существование
	assert.Equal(t, 0, mock.CreateUserCallCount)
}

// TestRegister_DefaultRole — проверяет, что если роль не указана, подставляется "user"
func TestRegister_DefaultRole(t *testing.T) {
	email := "test@mail.com"
	password := "123456"
	role := "" // роль не указана

	mock := NewMockUserRepository()
	svc := NewAuthService(mock)

	err := svc.Register(context.Background(), email, password, role)

	assert.NoError(t, err)
	assert.Equal(t, 1, mock.CreateUserCallCount)
	// Проверяем, что сохранился пользователь с ролью "user"
	assert.Equal(t, "user", mock.LastCreatedUser.Role)
}

// TestRegister_RepoError — проверяет, что сервис возвращает ошибку репозитория
func TestRegister_RepoError(t *testing.T) {
	email := "test@mail.com"
	password := "123456"
	role := "user"

	mock := NewMockUserRepository()
	// Включаем режим ошибки в моке
	mock.ShouldBeError = true
	mock.ErrorMessage = "connection failed"
	svc := NewAuthService(mock)

	err := svc.Register(context.Background(), email, password, role)

	assert.Error(t, err)
	assert.Equal(t, "connection failed", err.Error())
	// GetUserByEmailCallCount вызван, но вернул ошибку (счётчик увеличился)
	assert.Equal(t, 1, mock.GetUserByEmailCallCount)
}

// TestRegister_ContextCanceled — проверяет, что сервис обрабатывает отменённый контекст
func TestRegister_ContextCanceled(t *testing.T) {
	email := "test@mail.com"
	password := "123456"
	role := "user"

	// Создаём контекст и сразу отменяем его
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // отменяем контекст ДО вызова Register

	mock := NewMockUserRepository()
	svc := NewAuthService(mock)

	err := svc.Register(ctx, email, password, role)

	// Ожидаем ошибку, связанную с отменой контекста
	assert.Error(t, err)
	// GetUserByEmailCallCount НЕ вызван, потому что контекст отменён
	assert.Equal(t, 0, mock.GetUserByEmailCallCount)
}

// ============================================================
// 4. ТЕСТЫ ДЛЯ ВХОДА (Login)
// ============================================================

// TestLogin_Success — проверяет успешный вход
func TestLogin_Success(t *testing.T) {
	email := "test@mail.com"
	password := "123456"
// Хеш пароля "123456" (сгенерируй один раз)
    hashedPassword := "$2a$10$eDFCq/t577pPq9tdz5BcV.pZ3ozuFBxPYZP21lVLgqqfvn7y5vplu"
	mock := NewMockUserRepository()
	// Добавляем существующего пользователя в мок
	mock.Users[email] = models.User{
        Email:    email,
        Password: hashedPassword,  // ← добавить хеш
        Role:     "user",
    }

	svc := NewAuthService(mock)

	// Вызываем Login
	token, role, err := svc.Login(context.Background(), email, password)

	assert.NoError(t, err)             // ошибки нет
	assert.NotEmpty(t, token)          // токен сгенерирован
	assert.Equal(t, "user", role)      // роль совпадает
	assert.Equal(t, 1, mock.GetUserByEmailCallCount) // метод вызван 1 раз
	assert.Equal(t, email, mock.LastSearchedEmail)   // передан правильный email
}

// TestLogin_EmptyEmail — проверяет, что email обязателен при входе
func TestLogin_EmptyEmail(t *testing.T) {
	email := ""
	password := "123456"

	mock := NewMockUserRepository()
	svc := NewAuthService(mock)

	token, role, err := svc.Login(context.Background(), email, password)

	assert.Error(t, err)
	assert.Equal(t, "email is required", err.Error())
	assert.Empty(t, token)
	assert.Empty(t, role)
	// GetUserByEmail НЕ вызван (валидация сработала до вызова репозитория)
	assert.Equal(t, 0, mock.GetUserByEmailCallCount)
}

// TestLogin_EmptyPassword — проверяет, что пароль обязателен при входе
func TestLogin_EmptyPassword(t *testing.T) {
	email := "test@mail.com"
	password := ""

	mock := NewMockUserRepository()
	svc := NewAuthService(mock)

	token, role, err := svc.Login(context.Background(), email, password)

	assert.Error(t, err)
	assert.Equal(t, "password is required", err.Error())
	assert.Empty(t, token)
	assert.Empty(t, role)
	assert.Equal(t, 0, mock.GetUserByEmailCallCount)
}

// TestLogin_UserNotFound — проверяет вход с несуществующим email
func TestLogin_UserNotFound(t *testing.T) {
	email := "unknown@mail.com" // пользователь не существует
	password := "123456"

	mock := NewMockUserRepository() // пустой мок (пользователей нет)
	svc := NewAuthService(mock)

	token, role, err := svc.Login(context.Background(), email, password)

	assert.Error(t, err)
	assert.Equal(t, "invalid credentials", err.Error())
	assert.Empty(t, token)
	assert.Empty(t, role)
	assert.Equal(t, 1, mock.GetUserByEmailCallCount)
	assert.Equal(t, email, mock.LastSearchedEmail)
}

// TestLogin_WrongPassword — проверяет вход с неверным паролем
func TestLogin_WrongPassword(t *testing.T) {
	email := "test@mail.com"
	password := "wrongpassword" // неверный пароль

	mock := NewMockUserRepository()
	mock.Users[email] = models.User{Email: email, Role: "user"}
	svc := NewAuthService(mock)

	token, role, err := svc.Login(context.Background(), email, password)

	assert.Error(t, err)
	assert.Equal(t, "invalid credentials", err.Error())
	assert.Empty(t, token)
	assert.Empty(t, role)
	assert.Equal(t, 1, mock.GetUserByEmailCallCount)
}

// TestLogin_DatabaseError — проверяет ошибку БД при входе
func TestLogin_DatabaseError(t *testing.T) {
	email := "test@mail.com"
	password := "123456"

	mock := NewMockUserRepository()
	// Включаем режим ошибки в моке
	mock.ShouldBeError = true
	mock.ErrorMessage = "connection lost"
	svc := NewAuthService(mock)

	token, role, err := svc.Login(context.Background(), email, password)

	assert.Error(t, err)
	assert.Equal(t, "connection lost", err.Error())
	assert.Empty(t, token)
	assert.Empty(t, role)
	assert.Equal(t, 1, mock.GetUserByEmailCallCount)
}

// TestLogin_ContextCanceled — проверяет, что сервис обрабатывает отменённый контекст при входе
func TestLogin_ContextCanceled(t *testing.T) {
	email := "test@mail.com"
	password := "123456"

	// Создаём контекст и сразу отменяем его
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mock := NewMockUserRepository()
	svc := NewAuthService(mock)

	token, role, err := svc.Login(ctx, email, password)

	assert.Error(t, err)
	// GetUserByEmail НЕ вызван (контекст отменён)
	assert.Equal(t, 0, mock.GetUserByEmailCallCount)
	assert.Empty(t, token)
	assert.Empty(t, role)
}