// Package database-инициализация базы данных
package database

import (
	"context"
	"database/sql"
	"log"

	"github.com/Evgeny-08-01/Rest-user-aggregator/internal/models"
	"github.com/Evgeny-08-01/Rest-user-aggregator/internal/repository"
	_ "github.com/lib/pq" 
)

// db -пакетная переменная (уровня пакета) с соединением с БД. Доступна только внутри пакета database (приватная).
var db *sql.DB

// Init открывает соединение с БД и проверяет его работоспособность
func Init(databasePath string) error {
	var err error
	db, err = sql.Open("postgres", databasePath)
	if err != nil {
		return err
	}
	//  Проверяем подключение
	err = db.Ping()
	if err != nil {
		return err
	}
	log.Println("Database connected")
	return nil
}

// Close закрывает соединение с базой данных
func Close() error {
	if db == nil {
		return nil
	}
	err := db.Close()
	db = nil
	return err
}
// PostgresRepo структура - реализует интерфейс SubscriptionRepository
type PostgresRepo struct {
    db *sql.DB
}
// ЯВНАЯ ПРОВЕРКА: гарантирует, что PostgresRepo реализует интерфейс repository.SubscriptionRepository
var _ repository.SubscriptionRepository = (*PostgresRepo)(nil)

// NewPostgresRepo - конструктор
func NewPostgresRepo() *PostgresRepo {
    return &PostgresRepo{db: db}
}
// CreateMtd - вызывает существующую функцию CreateSubscription
func (r *PostgresRepo) CreateMtd(ctx context.Context, sub models.Subscription) (int, error) {
    return CreateSubscription(ctx, sub)
}

// GetByIDMtd - вызывает существующую функцию GetSubscriptionByID
func (r *PostgresRepo) GetByIDMtd(ctx context.Context, id int) (*models.Subscription, error) {
    return GetSubscriptionByID(ctx, id)
}

// UpdateMtd - вызывает существующую функцию UpdateSubscription
func (r *PostgresRepo) UpdateMtd(ctx context.Context, sub models.Subscription) error {
    return UpdateSubscription(ctx, sub)
}

// DeleteMtd - вызывает существующую функцию DeleteSubscription
func (r *PostgresRepo) DeleteMtd(ctx context.Context, id int) error {
    return DeleteSubscription(ctx, id)
}

// ListMtd - вызывает существующую функцию ListSubscriptions
func (r *PostgresRepo) ListMtd(ctx context.Context, limit, offset int) ([]models.Subscription, error) {
    return ListSubscriptions(ctx, limit, offset)
}

// GetTotalCostMtd - вызывает существующую функцию GetTotalCost
func (r *PostgresRepo) GetTotalCostMtd(ctx context.Context, userID, serviceName, startDate, endDate string) (int, error) {
    return GetTotalCost(ctx, userID, serviceName, startDate, endDate)
}
// GetDB - для тестов возвращает соединение с БД
func GetDB() *sql.DB {
    return db
}
