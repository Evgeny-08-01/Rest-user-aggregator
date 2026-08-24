// Package database - реализация репозитория для работы с пользователями
package database

import (
	"context"
	"database/sql"
	"errors"

	"Rest-user-agregator/internal/models"
	"Rest-user-agregator/pkg/logger"
)

// ============================================================
// 1. СОЗДАНИЕ НОВОГО ПОЛЬЗОВАТЕЛЯ
// ============================================================
// CreateUser — сохраняет пользователя в БД.
// Принимает: context, models.User (с заполненными полями)
// Возвращает: error
// ============================================================
func (r *PostgresRepo) CreateUser(ctx context.Context, user models.User) error {
	query := `INSERT INTO users (id, email, password_hash, role) VALUES ($1, $2, $3, $4)`

	_, err := r.db.ExecContext(ctx, query, user.ID, user.Email, user.Password, user.Role)
	if err != nil {
		logger.Error("CreateUser: failed to create user %s: %v", user.Email, err)
		return err
	}

	logger.Debug("CreateUser: user created successfully: %s", user.Email)
	return nil
}

// ============================================================
// 2. ПОЛУЧЕНИЕ ПОЛЬЗОВАТЕЛЯ ПО EMAIL
// ============================================================
// GetUserByEmail — ищет пользователя по email.
// Принимает: context, email (string)
// Возвращает: *models.User, error
// Если пользователь не найден — возвращает nil, nil
// ============================================================
func (r *PostgresRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	
    logger.Debug("GetUserByEmail: searching for email: %s", email)
	query := `SELECT id, email, password_hash, role, created_at FROM users WHERE email = $1`

	row := r.db.QueryRowContext(ctx, query, email)

	var user models.User
	var createdAt string

	err := row.Scan(&user.ID, &user.Email, &user.Password, &user.Role, &createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Debug("GetUserByEmail: user not found: %s", email)
			return nil, nil // пользователь не найден
		}
		logger.Error("GetUserByEmail: failed to get user %s: %v", email, err)
		return nil, err
	}

	user.CreatedAt = createdAt
	logger.Debug("GetUserByEmail: user found: %s", email)
	return &user, nil
}