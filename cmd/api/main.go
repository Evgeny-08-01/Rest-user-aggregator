package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Evgeny-08-01/Rest-user-aggregator/internal/database"
	"github.com/Evgeny-08-01/Rest-user-aggregator/internal/handlers"
	"github.com/joho/godotenv"
	"github.com/Evgeny-08-01/Rest-user-aggregator/pkg/logger"
)

func main() {
	// 1. Загружаем .env файл
	  logger.Init("app.log")
	// 1. Загружаем .env файл
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found")
	}

	// 2. Получаем путь к БД ИЗ .env
	databasePath := os.Getenv("DB_PATH")
	if databasePath == "" {
		databasePath = "./subscriptions.db"
	}

	// 3. Подключаемся к БД
	err = database.Init(databasePath)  // Подключение к БД
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	
	// Откладываем закрытие БД до завершения программы
	defer database.Close()
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
	// 4. Запускаем миграции
	err = runMigrations()
	if err != nil {
		log.Println("Migrations error:", err)
	}
	// 5. Роутер (switch для URL)
	mux := http.NewServeMux()

	// 6. CRUDL операции
	mux.HandleFunc("POST /api/subscriptions", handlers.CreateSubscriptionHandler)
	mux.HandleFunc("GET /api/subscriptions/{id}", handlers.GetSubscriptionHandler)
	mux.HandleFunc("PUT /api/subscriptions/{id}", handlers.UpdateSubscriptionHandler)
	mux.HandleFunc("DELETE /api/subscriptions/{id}", handlers.DeleteSubscriptionHandler)
	mux.HandleFunc("GET /api/subscriptions", handlers.ListSubscriptionsHandler)
	mux.HandleFunc("GET /api/subscriptions/total-cost", handlers.GetTotalCostHandler)

	// 7. Получаем порт из .env
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	// 8. Запускаем сервер
   log.Printf("Server starting on port %s", port)
	//  запуск HTTP сервера
	err = http.ListenAndServe(":"+port, mux)
	if err != nil {
	  log.Fatal("Server failed:", err)
	}
}

func runMigrations() error {
    migrationSQL, err := os.ReadFile("migrations/000001_create_subscriptions_table.up.sql")
    if err != nil {
        return err
    }
    _, err = database.DB.Exec(string(migrationSQL))
    return err
}