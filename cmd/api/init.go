package main

import (
	"Rest-user-agregator/internal/cache"
	"Rest-user-agregator/internal/database"
	"Rest-user-agregator/pkg/logger"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// 1. Загрузка .env
func loadEnv() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("[WARN] .env file not found, using default values")
	} else {
		log.Println("[INFO] .env file loaded successfully")
		log.Printf("[INFO] DB_PATH from .env: %s", os.Getenv("DB_PATH"))

		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			log.Println("[WARN] JWT_SECRET not set in .env file")
		} else {
			log.Printf("[INFO] JWT_SECRET loaded (length: %d)", len(jwtSecret))
		}
	}
}

// 2. Инициализация моего логгера(читаем уровень из .env)
func initLogger() {
	logLevel := os.Getenv("LOG_LEVEL") // читает переменную окружения LOG_LEVEL
	logPath := os.Getenv("LOG_PATH")   // читает переменную окружения LOG_PATH
	//	loggerType := os.Getenv("LOGGER")
	if logLevel == "" {
		logLevel = "info" //   default level
	}

	if logPath == "" {
		if os.Getenv("ENV") == "docker" {
			logPath = "/var/log/app/app.log" // путь к файлу с логами для Docker, когда поднимаем Docker-контейнер...docker compose up
			//  или CI/CD pipeline
		} else {
			logPath = "./logs/app.log" // путь к файлу с логами для локальной разработки ,без Docker... go run main.go
		}
	}
	logger.Init(logPath, logLevel)
	logger.Info("Starting Subscription API server")
}

// 3. initPprof — ЗАПУСК PROFILER ПРИ НЕОБХОДИМОСТИ
func initPprof() {
	if os.Getenv("PPROF_ENABLED") != "true" {
		return
	}
	go func() {
		logger.Info("pprof server starting on localhost:6060")
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			log.Printf("pprof server error: %v", err)
		}
	}()
}

// 4. Подключение к БД REDIS (можем работать и без кэша, но медленно, особенно для total cost хэндлера)
func initRedis() error {
	// ============================================================
	// 1. ЧИТАЕМ АДРЕС REDIS ИЗ ПЕРЕМЕННОЙ ОКРУЖЕНИЯ
	// ============================================================
	// redisAddr — адрес Redis в формате "хост:порт"
	// Примеры:
	//   - "localhost:6379" — для локальной разработки (без Docker)
	//   - "redis:6379"     — для запуска в Docker (сервис называется "redis")
	//
	// Переменная задаётся в .env файле:
	//   REDIS_ADDR=redis:6379
	// ============================================================
	redisAddr := os.Getenv("REDIS_ADDR")

	// Если переменная не задана — используем значение по умолчанию
	// localhost:6379 — стандартный адрес для Redis при локальной установке
	if redisAddr == "" {
		redisAddr = "localhost:6379"
		// Логируем предупреждение, что используем значение по умолчанию
		// Это не ошибка, просто напоминание для разработчика
		logger.Warn("REDIS_ADDR not set, using default: %s", redisAddr)
	} else {
		// Если переменная задана — логируем, что используем её
		logger.Debug("REDIS_ADDR: %s", redisAddr)
	}

	// ============================================================
	// 2. ПЫТАЕМСЯ ПОДКЛЮЧИТЬСЯ К REDIS
	// ============================================================
	// Вызываем функцию InitRedis из пакета cache.
	// Параметры:
	//   - redisAddr: адрес Redis (хост:порт)
	//   - ""        : пароль (у нас нет пароля, оставляем пустым)
	//   - 0         : номер базы данных (по умолчанию 0)
	//
	// InitRedis внутри себя:
	//   1. Создаёт клиент Redis
	//   2. Отправляет PING для проверки подключения
	//   3. Если PING успешен — возвращает nil (успех)
	//   4. Если нет — возвращает ошибку
	// ============================================================
	if err := cache.InitRedis(redisAddr, "", 0); err != nil {
		// ============================================================
		// 3. ОБРАБОТКА ОШИБКИ ПОДКЛЮЧЕНИЯ
		// ============================================================
		// Если Redis недоступен — мы НЕ останавливаем сервер.
		// Почему:
		//   - Кеш — это оптимизация, а не критическая часть
		//   - Без кеша сервер всё ещё работает (но медленнее)
		//   - Пользователь не должен видеть ошибку из-за того, что Redis упал
		//
		// Что делаем:
		//   - Логируем предупреждение, что работаем без кеша
		//   - Возвращаем ошибку в run(), но там она обрабатывается как WARN
		// ============================================================

		logger.Warn("Redis unavailable, continuing without cache")
		return err
	}

	// ============================================================
	// 4. УСПЕШНОЕ ПОДКЛЮЧЕНИЕ
	// ============================================================
	// Если подключение успешно — логируем успех.
	// После этой строки кеш готов к использованию.
	// ============================================================
	logger.Info("Redis initialized successfully on %s", redisAddr)
	return nil
}

// 5. Подключение к БД PostgreSQL
func initDB() error {
	databasePath := os.Getenv("DB_PATH") // Получаем путь к БД ИЗ .env
	if databasePath == "" {
		databasePath = "postgres://postgres:mysecret@db:5432/subscriptions?sslmode=disable" // если не получили, то ставим default
		logger.Warn("DB_PATH not set, using default")
	}
	// 2. Если мы внутри Docker (ENV=docker), используем `db`
	// Если локально — оставляем как есть (уже localhost)
	// Определяем окружение
	if os.Getenv("ENV") == "docker" {
		// В Docker заменяем localhost на db
		databasePath = strings.Replace(databasePath, "localhost", "db", 1)
		logger.Info("Running in Docker, using 'db' host")
	} else {
		// Локально — заменяем db на localhost (если вдруг в .env прописан db)
		databasePath = strings.Replace(databasePath, "db", "localhost", 1)
		logger.Info("Running locally, using localhost")
	}

	err := database.Init(databasePath) // Подключение к БД
	if err != nil {
		logger.Warn("DB_PATH not set")
		return fmt.Errorf("DB_PATH not set: %w", err)
	}
	return nil
}

// 6. Миграции
func runMigrations() error {
	if shouldRollback() {
		return rollbackMigrations()
	}
	return applyMigrations()
}
