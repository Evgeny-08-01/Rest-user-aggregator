// Package database-инициализация базы данных
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"Rest-user-agregator/internal/cache"
	"Rest-user-agregator/internal/repository"
	"Rest-user-agregator/pkg/logger"

	_ "github.com/lib/pq"
)

// db -пакетная переменная (уровня пакета) с соединением с БД. Доступна только внутри пакета database (приватная).
var db *sql.DB

/// Init открывает соединение с БД и настраивает пул соединений
func Init(databasePath string) error {
	// 1. ОТКРЫВАЕМ СОЕДИНЕНИЕ (структуру)
	//    sql.Open НЕ создаёт физическое соединение, только структуру *sql.DB.
	//    Физическое соединение откроется при первом запросе (Ping или Query).
	var err error
	db, err = sql.Open("postgres", databasePath)
	if err != nil {
		// Ошибка может быть только при невалидном DSN (строка подключения)
		logger.Error("Failed to open database connection: %v", err)
		return err
	}

	// ============================================================
	// 2. НАСТРОЙКА ПУЛА СОЕДИНЕНИЙ (КРИТИЧЕСКИ ВАЖНО)
	// ============================================================
	// Без этих настроек Go использует значения по умолчанию:
	//   - MaxOpenConns = 0 (не ограничено) → может создать тысячи соединений
	//   - MaxIdleConns = 2 (держит только 2) → остальные закрываются сразу
	//
	// При 200 RPS это приводит к:
	//   - 200 открытий/закрытий в секунду → нагрузка на CPU
	//   - Исчерпание файловых дескрипторов → ошибки "too many open files"
	//   - Падение сервера при нагрузке
	// ============================================================

	// SetMaxOpenConns — максимальное количество ОДНОВРЕМЕННЫХ соединений.
	// 50 означает: одновременно может быть не больше 50 активных запросов к БД.
	// Остальные запросы будут ждать в очереди, пока освободится соединение.
	// Почему 50: при 200 RPS и времени запроса 10 мс нужно ~2 соединения,
	// но с запасом для пиковых нагрузок берём 50.
	db.SetMaxOpenConns(50)

	// SetMaxIdleConns — количество "спящих" соединений в пуле.
	// 25 означает: даже если запросов нет, держим 25 открытых соединений.
	// Это экономит время на открытие новых соединений при следующем запросе.
	// Почему 25: половина от MaxOpenConns — баланс между скоростью и памятью.
	db.SetMaxIdleConns(25)

	// SetConnMaxLifetime — максимальное ВРЕМЯ ЖИЗНИ соединения.
	// 5 минут: через 5 минут соединение закрывается (даже если используется).
	// Зачем: PostgreSQL перезагружается, обновляет настройки, закрывает старые соединения.
	// Если Go будет держать соединение дольше 5 минут — БД его принудительно закроет,
	// а Go получит ошибку "bad connection".
	db.SetConnMaxLifetime(5 * time.Minute)

	// SetConnMaxIdleTime — максимальное время ПРОСТОЯ соединения.
	// 3 минуты: если соединение не использовалось 3 минуты — оно закрывается.
	// Зачем: освобождать ресурсы, когда нагрузка упала (ночью, выходные).
	// Если не закрывать — память БД будет занята даже в простое.
	db.SetConnMaxIdleTime(3 * time.Minute)

	// Логируем настройки, чтобы в логах было видно, что пул настроен
	logger.Info("Database connection pool configured: MaxOpen=50, MaxIdle=25, MaxLifetime=5m, MaxIdleTime=3m")

	// ============================================================
	// 3. ПРОВЕРЯЕМ, ЧТО БД ДОСТУПНА
	// ============================================================
	// Ping открывает ФИЗИЧЕСКОЕ соединение и отправляет тестовый запрос.
	// Если БД не запущена — вернёт ошибку (и сервер упадёт при старте).
	err = db.Ping()
	if err != nil {
		logger.Error("Failed to ping database: %v", err)
		return err
	}

	// Всё хорошо, БД доступна
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
func CreateTestTable() error {
    if db == nil {
        return errors.New("database not initialized")
    }

    // Создаём users
    _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            email TEXT UNIQUE NOT NULL,
            password_hash TEXT NOT NULL,
            role TEXT NOT NULL DEFAULT 'user',
            created_at TIMESTAMP DEFAULT NOW()
        )
    `)
    if err != nil {
        logger.Error("CreateTestTable: failed to create users: %v", err)
        return err
    }

    // Создаём subscriptions
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS subscriptions (
            id SERIAL PRIMARY KEY,
            service_name VARCHAR(255) NOT NULL,
            price INTEGER NOT NULL,
            user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            start_date DATE NOT NULL,
            end_date DATE
        )
    `)
    if err != nil {
        logger.Error("CreateTestTable: failed to create subscriptions: %v", err)
        return err
    }

    // Создаём cache_control_user
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS cache_control_user (
            user_id UUID PRIMARY KEY,
            version INT NOT NULL DEFAULT 1
        )
    `)
    if err != nil {
        logger.Error("CreateTestTable: failed to create cache_control_user: %v", err)
        return err
    }

    logger.Debug("CreateTestTable: all tables created or already exists")
    return nil
}
// DropTestTable — удаляет тестовые таблицы, если они существуют
func DropTestTable() error {
    if db == nil {
        return errors.New("database not initialized")
    }

    // Удаляем таблицы в обратном порядке (из-за зависимостей)
    _, err := db.Exec("DROP TABLE IF EXISTS cache_control_user CASCADE")
    if err != nil {
        logger.Error("DropTestTable: failed to drop cache_control_user: %v", err)
        return err
    }

    _, err = db.Exec("DROP TABLE IF EXISTS users CASCADE")
    if err != nil {
        logger.Error("DropTestTable: failed to drop users: %v", err)
        return err
    }

    _, err = db.Exec("DROP TABLE IF EXISTS subscriptions CASCADE")
    if err != nil {
        logger.Error("DropTestTable: failed to drop subscriptions: %v", err)
        return err
    }

    logger.Debug("DropTestTable: all tables dropped successfully")
    return nil
}
// CleanTestTable очищает таблицу перед тестами
func CleanTestTable() error {
    if db == nil {
        err := errors.New("database not initialized")
        logger.Error("CleanTestTable: %v", err)
        return err
    }
    _, err := db.Exec("TRUNCATE subscriptions, users, cache_control_user RESTART IDENTITY")
    if err != nil {
        logger.Error("CleanTestTable: failed to truncate tables: %v", err)
        return err
    }

    // Очищаем Redis: удаляем все ключи с префиксом total:
    client := cache.GetClient()
    if client != nil {
        ctx := context.Background()
        iter := client.Scan(ctx, 0, "total:*", 0).Iterator()
        for iter.Next(ctx) {
            if err := client.Del(ctx, iter.Val()).Err(); err != nil {
                logger.Warn("CleanTestTable: failed to delete key %s: %v", iter.Val(), err)
            }
        }
        if err := iter.Err(); err != nil {
            logger.Warn("CleanTestTable: Redis scan error: %v", err)
        } else {
            logger.Debug("CleanTestTable: Redis keys with prefix total:* deleted")
        }
    }

    logger.Debug("CleanTestTable: tables truncated successfully")
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
// PingWithContext проверяет соединение с PostgreSQL с контекстом
func PingWithContext(ctx context.Context) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	return db.PingContext(ctx)
}