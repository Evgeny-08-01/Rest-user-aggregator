// Package database-инициализация базы данных
package database

import (
	"database/sql"
	"errors"

	"Rest-user-agregator/internal/repository"
	"Rest-user-agregator/pkg/logger"
	_ "github.com/lib/pq"
)

// db -пакетная переменная (уровня пакета) с соединением с БД. Доступна только внутри пакета database (приватная).
var db *sql.DB

// Init открывает соединение с БД и проверяет его работоспособность
func Init(databasePath string) error {
	var err error
	db, err = sql.Open("postgres", databasePath)
	if err != nil {
		logger.Error("Failed to open database connection: %v", err)
		return err
	}
	//  Проверяем подключение
	err = db.Ping()
	if err != nil {
			logger.Error("Failed to ping database: %v", err)
		return err
	}
	logger.Info("Database connected successfully")
	return nil
}

// Close закрывает соединение с базой данных
func Close() error {
    if db == nil {
        return nil
    }
    err := db.Close()
    if err != nil {
        logger.Error("Failed to close database connection: %v", err)
        return err
    }
    logger.Debug("Database connection closed")
    return nil
}

// PostgresRepo структура - реализует интерфейс SubscriptionRepository....и это структура, 
// которая содержит в себе подключение к базе данных *sql.DB и предоставляет методы для выполнения
//  всех SQL-запросов из пакета interface 
//
type PostgresRepo struct {
    db *sql.DB
}
// ЯВНАЯ ПРОВЕРКА: гарантирует, что PostgresRepo реализует интерфейс repository.SubscriptionRepository
var _ repository.SubscriptionRepository = (*PostgresRepo)(nil)

// NewPostgresRepo - конструктор(обертка)- возвращает указатель на структуру PostgresRepo, 
// которая содержит пул соединений и все методы для работы с БД 
func NewPostgresRepo() *PostgresRepo {
	 if db == nil {
        logger.Error("Database not initialized. Call Init() first")
        return nil
    }
    return &PostgresRepo{db: db}
}

// GetDB - для тестов возвращает соединение с БД
func GetDB() *sql.DB {
    return db
}
// CreateTestTable создаёт таблицу для тестов (если её нет)
func CreateTestTable() error {
    if db == nil {
        err := errors.New("database not initialized")
        logger.Error("CreateTestTable: %v", err)
        return err
    }
    _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS subscriptions (
            id SERIAL PRIMARY KEY,
            service_name VARCHAR(255) NOT NULL,
            price INTEGER NOT NULL,
            user_id UUID NOT NULL,
            start_date DATE NOT NULL,
            end_date DATE
        )
    `)
    if err != nil {
        logger.Error("CreateTestTable: failed to create table: %v", err)
        return err
    }
    logger.Debug("CreateTestTable: table created or already exists")
    return nil
}

// CleanTestTable очищает таблицу перед тестами
func CleanTestTable() error {
    if db == nil {
        err := errors.New("database not initialized")
        logger.Error("CleanTestTable: %v", err)
        return err
    }
    _, err := db.Exec("TRUNCATE subscriptions RESTART IDENTITY")
    if err != nil {
        logger.Error("CleanTestTable: failed to truncate table: %v", err)
        return err
    }
    logger.Debug("CleanTestTable: table truncated successfully")
    return nil
}
// DeleteSubscriptionsByUserID удаляет все подписки пользователя (для тестов)
func DeleteSubscriptionsByUserID(userID string) error {
    if db == nil {
        err := errors.New("database not initialized")
        logger.Error("DeleteSubscriptionsByUserID: %v", err)
        return err
    }
    _, err := db.Exec("DELETE FROM subscriptions WHERE user_id = $1", userID)
    if err != nil {
        logger.Error("DeleteSubscriptionsByUserID: failed to delete for user_id=%s: %v", userID, err)
        return err
    }
    logger.Debug("DeleteSubscriptionsByUserID: deleted subscriptions for user_id=%s", userID)
    return nil
}