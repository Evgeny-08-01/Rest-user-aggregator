// Package service - бизнес-логика авторизации
package service

import (
	"context"
	"errors"

	"Rest-user-agregator/internal/authentication"
	"Rest-user-agregator/internal/models"
	"Rest-user-agregator/internal/repository"
	"Rest-user-agregator/pkg/logger"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ============================================================
// AuthService — сервис авторизации
// ============================================================
// Содержит бизнес-логику для регистрации и входа пользователей.
// Работает через репозиторий UserRepository.
// ============================================================
type AuthService struct {
	repo repository.UserRepository
}

// NewAuthService — конструктор сервиса авторизации
func NewAuthService(repo repository.UserRepository) *AuthService {
	return &AuthService{repo: repo}
}

// ============================================================
// 1. РЕГИСТРАЦИЯ НОВОГО ПОЛЬЗОВАТЕЛЯ
// ============================================================
// Register — создаёт нового пользователя в системе.
//
// ПАРАМЕТРЫ:
//   - ctx: контекст запроса.
//     Отменяется при:
//   - Ctrl+C (SIGINT)
//   - выключении ОС (SIGTERM)
//   - закрытии браузера
//   - истечении таймаута
//     Используется для передачи данных (user_id) и отмены операций.
//   - email: email пользователя (обязательно)
//   - password: пароль (обязательно, минимум 6 символов)
//   - role: роль пользователя (по умолчанию "user")
//
// ВОЗВРАЩАЕТ:
//   - error: ошибка, если валидация не пройдена или пользователь уже существует
//
// ЛОГИКА:
//  1. Валидация email и password
//  2. Проверка, что пользователь не существует
//  3. Хеширование пароля (bcrypt)
//  4. Сохранение пользователя в БД
//
// ============================================================
func (s *AuthService) Register(ctx context.Context, email, password, role string) error {
	// 1. Валидация входных данных
	if email == "" {
		logger.Warn("AuthService.Register: email is required")
		return errors.New("email is required")
	}
	if password == "" {
		logger.Warn("AuthService.Register: password is required")
		return errors.New("password is required")
	}
	if len(password) < 4 {
		logger.Warn("AuthService.Register: password too short: %d chars", len(password))
		return errors.New("password must be at least 4 characters.")
	}

	// 2. Проверяем, существует ли пользователь с таким email
	existing, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		logger.Error("AuthService.Register: failed to check existing user: %v", err)
		return err
	}
	if existing != nil {
		logger.Warn("AuthService.Register: user already exists: %s", email)
		return errors.New("user already exists")
	}

	// 3. Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("AuthService.Register: failed to hash password: %v", err)
		return err
	}

	// 4. Если роль не указана — ставим "user" по умолчанию
	if role == "" {
		role = "user"
	}
	// Запрет на создание админа через API
	if role == "admin" {
		logger.Warn("AuthService.Register: attempt to create admin via API: %s", email)
		return errors.New("admin role cannot be created via API")
	}
	// 5. Создаём пользователя
	user := models.User{
		ID:       uuid.New().String(),
		Email:    email,
		Password: string(hashedPassword),
		Role:     role,
	}

	// 6. Сохраняем в БД
	if err := s.repo.CreateUser(ctx, user); err != nil {
		logger.Error("AuthService.Register: failed to create user: %v", err)
		return err
	}

	logger.Info("AuthService.Register: user created successfully: %s (role: %s)", email, role)
	return nil
}

// ============================================================
// 2. ВХОД ПОЛЬЗОВАТЕЛЯ (ЛОГИН)
// ============================================================
// Login — проверяет credentials и генерирует JWT-токен.
//
// ПАРАМЕТРЫ:
//   - ctx: контекст для управления временем жизни запроса
//   - email: email пользователя (обязательно)
//   - password: пароль (обязательно)
//
// ВОЗВРАЩАЕТ:
//   - string: JWT-токен
//   - string: роль пользователя
//   - error: ошибка, если credentials неверны
//
// ЛОГИКА:
//  1. Валидация email и password
//  2. Поиск пользователя в БД по email
//  3. Проверка пароля (bcrypt)
//  4. Генерация JWT-токена
//  5. Возврат токена и роли
//
// ============================================================
func (s *AuthService) Login(ctx context.Context, email, password string) (string, string, error) {
	// 1. Валидация входных данных
	if email == "" {
		logger.Warn("AuthService.Login: email is required")
		return "", "", errors.New("email is required")
	}
	if password == "" {
		logger.Warn("AuthService.Login: password is required")
		return "", "", errors.New("password is required")
	}

	// 2. Ищем пользователя в БД
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		logger.Error("AuthService.Login: database error: %v", err)
		return "", "", err
	}
	if user == nil {
		logger.Warn("AuthService.Login: user not found: %s", email)
		return "", "", errors.New("invalid credentials")
	}

	// 3. Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		logger.Warn("AuthService.Login: invalid password for: %s", email)
		return "", "", errors.New("invalid credentials")
	}

	// 4. Генерируем JWT-токен
	token, err := authentication.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		logger.Error("AuthService.Login: failed to generate token: %v", err)
		return "", "", errors.New("failed to generate token")
	}

	logger.Info("AuthService.Login: user logged in successfully: %s (role: %s)", email, user.Role)
	return token, user.Role, nil
}
