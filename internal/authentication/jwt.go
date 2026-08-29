// Package authentication - работа с JWT-токенами (генерация, валидация, middleware)
package authentication

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ============================================================
// 1. СТРУКТУРА КЛАЙМОВ (ДАННЫЕ, КОТОРЫЕ ХРАНЯТСЯ В ТОКЕНЕ)
// ============================================================
// Claims — это данные, которые мы вшиваем в JWT-токен.
// Они будут доступны серверу при каждом запросе.
//
// ПОЛЯ:
//   - UserID → идентификатор пользователя (из БД)
//   - Email  → email пользователя (для быстрого доступа)
//   - Role   → роль пользователя (admin/user) для проверки прав
//
// RegisteredClaims — стандартные поля JWT:
//   - ExpiresAt → когда токен истекает
//   - IssuedAt  → когда выдан
//
// ============================================================
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// ============================================================
// 2. ГЕНЕРАЦИЯ JWT-ТОКЕНА
// ============================================================
// GenerateToken — создаёт подписанный JWT-токен для пользователя.
//
// ПАРАМЕТРЫ:
//   - userID → ID пользователя (из БД)
//   - email  → email пользователя
//   - role   → роль пользователя (admin/user)
//
// ВОЗВРАЩАЕТ:
//   - string → готовый JWT-токен
//   - error  → ошибка, если не удалось создать
//
// АЛГОРИТМ:
//  1. Берём секрет из переменной окружения JWT_SECRET
//  2. Устанавливаем время жизни токена (24 часа)
//  3. Создаём claims с данными пользователя
//  4. Подписываем токен с использованием HS256
//
// ============================================================
func GenerateToken(userID, email, role string) (string, error) {
	// 1. Получаем секрет из переменной окружения
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("JWT_SECRET is not set in .env file")
	}

	// 2. Устанавливаем время жизни токена (10000 часа-для упрощения тестирования)
	expirationTime := time.Now().Add(10000 * time.Hour) // 10000- для упрощения тестирования

	// 3. Создаём claims (данные, которые будут в токене)
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime), // Когда токен протухнет
			IssuedAt:  jwt.NewNumericDate(time.Now()),     // Когда токен выдан
		},
	}

	// 4. Создаём токен с алгоритмом HS256
	//    HS256 — симметричное шифрование (один секрет для подписи и проверки)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 5. Подписываем токен секретом и возвращаем строку
	return token.SignedString([]byte(secret))
}

// ============================================================
// 3. ПРОВЕРКА И ВАЛИДАЦИЯ JWT-ТОКЕНА
// ============================================================
// ValidateToken — проверяет токен и возвращает данные (claims).
//
// ПАРАМЕТРЫ:
//   - tokenString → JWT-токен (строка из заголовка Authorization)
//
// ВОЗВРАЩАЕТ:
//   - *Claims → данные пользователя (если токен валиден)
//   - error   → ошибка, если токен недействителен
//
// АЛГОРИТМ:
//  1. Берём секрет из переменной окружения
//  2. Парсим токен, проверяем подпись
//  3. Проверяем срок действия (не истёк ли)
//  4. Возвращаем claims, если всё ок
//
// ============================================================
func ValidateToken(tokenString string) (*Claims, error) {
	// 1. Получаем секрет из переменной окружения
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, errors.New("JWT_SECRET is not set")
	}

	// 2. Парсим и проверяем токен
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Проверяем, что используется правильный алгоритм подписи
		// Если злоумышленник попытается использовать другой алгоритм — отклоним
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		// Возвращаем секрет для проверки подписи
		return []byte(secret), nil
	})

	// 3. Обрабатываем ошибки парсинга
	if err != nil {
		return nil, err
	}

	// 4. Проверяем, валидный ли токен (не просрочен, подпись верна)
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// 5. Возвращаем claims (данные пользователя)
	return claims, nil
}

// ============================================================
// 4. ИЗВЛЕЧЕНИЕ ID ПОЛЬЗОВАТЕЛЯ ИЗ ТОКЕНА (УДОБНАЯ ОБЁРТКА)
// ============================================================
// GetUserIDFromToken — извлекает UserID из валидного токена.
//
// ПАРАМЕТРЫ:
//   - tokenString → JWT-токен
//
// ВОЗВРАЩАЕТ:
//   - string → UserID
//   - error  → ошибка, если токен невалиден
//
// ============================================================
func GetUserIDFromToken(tokenString string) (string, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}
