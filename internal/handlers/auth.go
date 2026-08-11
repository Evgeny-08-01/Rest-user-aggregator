// Package handlers - HTTP обработчики для 1. РЕГИСТРАЦИЯ НОВОГО ПОЛЬЗОВАТЕЛЯ и 2. ВХОД ПОЛЬЗОВАТЕЛЯ (ЛОГИН)
package handlers

import (
	"net/http"

	"Rest-user-agregator/internal/models"
	"Rest-user-agregator/pkg/logger"
)

// ============================================================
// 1. РЕГИСТРАЦИЯ НОВОГО ПОЛЬЗОВАТЕЛЯ
// ============================================================
// RegistrationHandler — обработчик POST /api/register
// Принимает JSON: { "email": "...", "password": "...", "role": "user" }
// Возвращает: 201 Created { "message": "User registered successfully" }
// ============================================================
func (h *Handler) RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Парсим JSON из тела запроса в структуру User
	var req models.User
	if err := parseJSON(r, &req); err != nil {
		logger.Warn("RegistrationHandler: invalid JSON: %v", err)
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// 2. Валидация обязательных полей
	if req.Email == "" || req.Password == "" {
		logger.Warn("RegistrationHandler: missing email or password")
		writeJSONError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// 3. Если роль не указана — ставим "user" по умолчанию
	if req.Role == "" {
		req.Role = "user"
	}

	// 4. Вызов сервиса для создания пользователя
	err := h.AuthService.Register(r.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		// Если пользователь уже существует — возвращаем 409 Conflict
		if err.Error() == "user already exists" {
			logger.Warn("RegistrationHandler: user already exists: %s", req.Email)
			writeJSONError(w, http.StatusConflict, "User already exists")
			return
		}
		// Любая другая ошибка — 500 Internal Server Error
		logger.Error("RegistrationHandler: failed to register user: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Failed to register user")
		return
	}

	// 5. Успешный ответ
	logger.Info("RegistrationHandler: user registered successfully: %s", req.Email)
	writeJSON(w, http.StatusCreated, map[string]string{
		"message": "User registered successfully",
	})
}

// ============================================================
// 2. ВХОД ПОЛЬЗОВАТЕЛЯ (ЛОГИН)
// ============================================================
// LoginHandler — обработчик POST /api/login
// Принимает JSON: { "email": "...", "password": "..." }
// Возвращает: 200 OK { "token": "jwt_token", "email": "...", "role": "..." }
// ============================================================
func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Парсим JSON из тела запроса
	var req models.User
	if err := parseJSON(r, &req); err != nil {
		logger.Warn("LoginHandler: invalid JSON: %v", err)
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// 2. Валидация обязательных полей
	logger.Info("LoginHandler: raw request: %+v", req)  // ← добавить логирование
	if req.Email == "" || req.Password == "" {
		logger.Warn("LoginHandler: missing email or password")
		writeJSONError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// 3. Вызов сервиса для проверки credentials и генерации токена
	token, role, err := h.AuthService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		logger.Warn("LoginHandler: invalid credentials for %s", req.Email)
		writeJSONError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// 4. Успешный ответ с токеном и ролью
	logger.Info("LoginHandler: user logged in: %s (role: %s)", req.Email, role)
	writeJSON(w, http.StatusOK, map[string]string{
		"token": token,
		"email": req.Email,
		"role":  role,
	})
}