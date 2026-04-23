package main

import (
	"log"
	"os"

	"github.com/Evgeny-08-01/Rest-user-aggregator/internal/database"
	"github.com/joho/godotenv"
)

func main() {
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

	// 4. Запускаем миграции
	err = runMigrations(databasePath)
	if err != nil {
		log.Println("Migrations error:", err)
	}

	// 5. Получаем порт из .env
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	// 6. Запускаем сервер
	log.Println("Server starting on port:", port)
	//  запуск HTTP сервера
	// err = http.ListenAndServe(":"+port, nil)
	// if err != nil {
	//     log.Fatal("Server failed:", err)
	// }
}

// Временно пропускаем миграции
func runMigrations(databasePath string) error {
	return nil
}