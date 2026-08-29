package main

// ============================================================
// 1. ИМПОРТЫ
// ============================================================

import (

	"fmt"
	"log"
	_ "net/http/pprof"


	_ "Rest-user-agregator/docs"
	"Rest-user-agregator/internal/database"
	"Rest-user-agregator/internal/service"
	"Rest-user-agregator/pkg/logger"


	_ "github.com/lib/pq"

//	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Subscription API
// @version 1.0
// @SERVER_PORT=8087
// @BasePath /api
// ============================================================
// 2. ТОЧКА ВХОДА
// ============================================================
func main() {
	if err := run(); err != nil {
		log.Fatalf("Fatal error: %v", err)
	}
}

// ============================================================
// 3. ОСНОВНАЯ ЛОГИКА ПРОЕКТА
// ============================================================
//
// Структура с выносом всей логики в отдельные функции:
//
//	func run() error {
//	  1. .env
//	  2. Логгер
//	  3. Redis (кеш — опционально)
//	  4. БД (PostgreSQL — критично)
//	  5. Миграции
//	  6. Создание зависимостей (repo, services) — общие для REST и gRPC
//	  7. Запуск приложения (startApplication)
//	     - ПАРАЛЛЕЛЬНЫЙ ЗАПУСК (REST + gRPC в горутинах с WaitGroup)
//	     - Ожидание сигналов остановки (SIGINT, SIGTERM)
//	     - Graceful shutdown (оба сервера с таймаутом)
//	     - Ожидание завершения всех горутин (WaitGroup)
//	  return nil
//	}
func run() error {
	// 1. Загружаем .env (ошибки внутри)
	loadEnv()

	// 2. Инициализация логгера (ошибки внутри)
	initLogger()
	logger.Info("Starting Subscription API server")

	// 3. Инициализация Pprof (ошибки внутри)
	initPprof()

	// 4. Redis (не критичен)
	if err := initRedis(); err != nil {
		logger.Warn("Redis unavailable, continuing without cache: %v", err)
	}

	// 5. Инициализация БД (критична)
	if err := initDB(); err != nil {
		return fmt.Errorf("DB init: %w", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			logger.Error("failed to close db: %v", err)
		}
	}()

	// 6. Миграции
	if err := runMigrations(); err != nil {
		return fmt.Errorf("migrations: %w", err)
	}

	// 7. Создание зависимостей
svc, authSvc, templateSvc := buildServices()

	// 8. Запуск серверов

return runServers(svc, authSvc, templateSvc)
}

// ============================================================
// 4.  ФУНКЦИИ
// ============================================================



// 7. buildServices — СОЗДАЁТ ВСЕ ЗАВИСИМОСТИ
// ============================================================
func buildServices() (*service.SubscriptionService, *service.AuthService, *service.TemplateService) {
    repo := database.NewPostgresRepo()
    templateRepo := database.NewPostgresRepo()
    svc := service.NewSubscriptionService(repo, templateRepo)
    authService := service.NewAuthService(repo)
    templateService := service.NewTemplateService(templateRepo)
    return svc, authService, templateService
}

