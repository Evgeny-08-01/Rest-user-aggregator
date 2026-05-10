package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/Evgeny-08-01/Rest-user-aggregator/docs"
	"github.com/Evgeny-08-01/Rest-user-aggregator/internal/database"
	"github.com/Evgeny-08-01/Rest-user-aggregator/internal/handlers"
	"github.com/Evgeny-08-01/Rest-user-aggregator/pkg/logger"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/swaggo/http-swagger"
)

// @title Subscription API
// @version 1.0
// @SERVER_PORT=8080
// @BasePath /api
func main() {
	// 1. Загружаем .env файл
	err := godotenv.Load("./.env")
	if err != nil {
		 logger.Warn(".env file not found, using default values") 
	}

 // 2. Инициализируем логгер (читаем уровень из .env)
 
    logLevel := os.Getenv("LOG_LEVEL")
    if logLevel == "" {
        logLevel = "info"
    }
logPath := "/root/app.log"
if os.Getenv("ENV") != "docker" {
    logPath = "app.log"
}	
logger.Init(logPath, logLevel)
logger.Info("Starting Subscription API server") 

	// 2. Получаем путь к БД ИЗ .env
databasePath := os.Getenv("DB_PATH")
if databasePath == "" {
    databasePath = "postgres://postgres:mysecret@db:5432/subscriptions?sslmode=disable"
	logger.Warn("DB_PATH not set, using default")
}
	// 3. Подключаемся к БД
	err = database.Init(databasePath) // Подключение к БД
	if err != nil {
		 logger.Fatal("Failed to connect to database: %v", err) 
	}

	// Откладываем закрытие БД до завершения программы
	defer database.Close()
	logger.Info("Database connected successfully") 
	// 4. Проверяем на наличие миграций
	if len(os.Args) > 1 && os.Args[1] == "-down" {
		downSQL, err2 := os.ReadFile("migrations/000001_create_subscriptions_table.down.sql")
		if err2 != nil {
			log.Fatal("Failed to read down migration:", err2)
		}
		_, err = database.DB.Exec(string(downSQL))
		if err != nil {
			log.Fatal("Failed to rollback migration:", err)
		}
		log.Println("Migration rolled back")
		return
	}
	// 5. Запускаем миграции
	err = runMigrations()  // пользовательская функция (см. ниже)
	if err != nil {
		logger.Warn("Migrations warning (maybe already applied): %v", err)
	}
	// 6. Роутер (switch для URL)
	mux := http.NewServeMux()

	// 7. CRUDL операции
	mux.HandleFunc("POST    /api/subscriptions",               handlers.LoggingMiddleware(handlers.CreateSubscriptionHandler))
	mux.HandleFunc("GET     /api/subscriptions/{id}",          handlers.LoggingMiddleware(handlers.GetSubscriptionHandler))
	mux.HandleFunc("PUT     /api/subscriptions/{id}",          handlers.LoggingMiddleware(handlers.UpdateSubscriptionHandler))
	mux.HandleFunc("DELETE  /api/subscriptions/{id}",          handlers.LoggingMiddleware(handlers.DeleteSubscriptionHandler))
	mux.HandleFunc("GET     /api/subscriptions",               handlers.LoggingMiddleware(handlers.ListSubscriptionsHandler))
	mux.HandleFunc("GET     /api/subscriptions/total-cost",    handlers.LoggingMiddleware(handlers.GetTotalCostHandler))
	mux.HandleFunc("GET     /swagger/",                        httpSwagger.WrapHandler)
	// 8. Получаем порт из .env
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
		logger.Warn("SERVER_PORT not set, using default 8080") 
	}
   // 9. HTTP сервер с таймаутами 
    srv := &http.Server{
        Addr:         ":" + port,
        Handler:      mux,
        ReadTimeout:  5  * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  15 * time.Second,
    }
 // 10. Запускаем сервер в горутине
    go func() {
             logger.Info("Server starting on port %s", port) 
        if err2 := srv.ListenAndServe(); err2 != nil && err2 != http.ErrServerClosed {
           logger.Error("Server failed: %v", err2)
			os.Exit(1) // Завершаем программу с кодом 1
        }
    }()

// 11. Graceful shutdown (ожидание сигнала)
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

logger.Info("Shutting down server...")
// 12. Контекст с таймаутом на завершение (5 секунд)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

// 13. Останавливаем сервер
    if err := srv.Shutdown(ctx); err != nil {
       logger.Error("Server forced to shutdown: %v", err)
	   os.Exit(1)
    }

   logger.Info("Server exited properly")
}

func runMigrations() error {
	migrationSQL, err := os.ReadFile("migrations/000001_create_subscriptions_table.up.sql")
	if err != nil {
		return err
	}
	_, err = database.DB.Exec(string(migrationSQL))
	return err
}