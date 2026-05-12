// Package database-инициализация базы данных
package database

import (
	"database/sql"


	"github.com/Evgeny-08-01/Rest-user-aggregator/internal/repository"
	"github.com/Evgeny-08-01/Rest-user-aggregator/pkg/logger"
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
		db = nil
	if err != nil {
		logger.Error("Failed to close database connection: %v", err)
		return err
	}
	logger.Debug("Database connection closed")
	return nil
}

// PostgresRepo структура - реализует интерфейс SubscriptionRepository
type PostgresRepo struct {
    db *sql.DB
}
// ЯВНАЯ ПРОВЕРКА: гарантирует, что PostgresRepo реализует интерфейс repository.SubscriptionRepository
var _ repository.SubscriptionRepository = (*PostgresRepo)(nil)

// NewPostgresRepo - конструктор
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
