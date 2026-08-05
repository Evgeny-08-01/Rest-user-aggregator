package main

// ============================================================
// 1. ИМПОРТЫ
// ============================================================

import (
	"context" // ВСТАВКА ФРОНТЕНД                                                               // ВСТАВКА ФРОНТЕНД
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "Rest-user-agregator/docs"
	"Rest-user-agregator/internal/database"
	"Rest-user-agregator/internal/handlers"
	"Rest-user-agregator/internal/service"
	"Rest-user-agregator/pkg/logger"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/swaggo/http-swagger"
)

// @title Subscription API
// @version 1.0
// @SERVER_PORT=8080
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
// 3. ОСНОВНАЯ ЛОГИКА (СБОРКА И ЗАПУСК)
// ============================================================
// Структура проекта
//
//	 func run() error {
//	 1. .env
//	 2. Логгер
//	 3. БД
//	 4. Миграции
//	 5. Сервер
//	   return nil
//	}
func run() error {
	// 1. Загружаем .env
	loadEnv()
	// 2. Инициализация моего Логгера
	initLogger()
	// 3. Инициализация БД
	if err := initDB(); err != nil {
		return fmt.Errorf("DB init: %w", err)
	}
	defer database.Close() // Откладываем закрытие БД до завершения программы
//	logger.Info("Database connected successfully") -дублирование с db.init

	// 4. Миграции
	if err := runMigrations(); err != nil {
		return fmt.Errorf("migrations: %w", err)
	}

	// 5. Сервер
	if err := startServer(); err != nil {
		return fmt.Errorf("server: %w", err)
	}

	return nil
}

// ============================================================
// 4. ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ (каждая делает что то одно)
// ============================================================

// 4.1 Загрузка .env
func loadEnv() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("[WARN] .env file not found, using default values")
	} else {
		log.Println("[INFO] .env file loaded successfully")
		log.Printf("[INFO] DB_PATH from .env: %s", os.Getenv("DB_PATH"))
	}
}

// 4.2 Инициализация моего логгера(читаем уровень из .env)
func initLogger() {
	logLevel := os.Getenv("LOG_LEVEL") // читает переменную окружения LOG_LEVEL
	logPath := os.Getenv("LOG_PATH")   // читает переменную окружения LOG_PATH
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

// 4.3 Подключение к БД
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

// 4.4 Миграции
func runMigrations() error {
	if shouldRollback() {
		return rollbackMigrations()
	}
	return applyMigrations()
}

// 4.5 Запуск сервера
func startServer() error {
	repo := database.NewPostgresRepo() // экземпляр репозитория, содержащий пул соединений и указатель на БД,
	//  содержит методы работы с БД. NewPostgresRepo-конструктор над PostgresRepo
	// PostgresRepo- структура и содежит поле: db *sql.DB
	svc := service.NewSubscriptionService(repo)
	handler := handlers.NewHandler(svc) // экземпляр хендлера, содержащий экземпляр репозитория repo для работы с БД,
	// содержит методы обработки HTTP-запросов.
	// NewHandler — конструктор, создающий экземпляр Handler.
	// Handler — структура с полем Repo(тип интерфейс) repository.SubscriptionRepository(интерфейс).
	// !!!! Таким образом  handler содержит методы обработки запросов и подключение к БД

	mux := http.NewServeMux() // Создаем роутер-switch для URL

	// ============================================================
	// ФРОНТЕНД
	// ============================================================
	// ============================================================
	// 1. СТАТИКА (CSS, JS) — ТОЛЬКО ДЛЯ ЭТИХ ПАПОК- идет из index.html
	// ============================================================
	mux.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./web/css")))) // ВСТАВКА ФРОНТЕНД
	mux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./web/js"))))    // ВСТАВКА ФРОНТЕНД

	// ============================================================
	// 2. HTML (КОРЕНЬ) — С ВОЗМОЖНОСТЬЮ ДОБАВИТЬ ЛОГИКУ
	// ============================================================
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { // ВСТАВКА ФРОНТЕНД
		// Здесь можно добавить:
		// - Проверку авторизации
		// - Логирование
		// - Выбор другой страницы (login.html / index.html)
		http.ServeFile(w, r, "./web/index.html") // ВСТАВКА ФРОНТЕНД
	})

	// ============================================================
	// CORS MIDDLEWARE
	// ============================================================
	corsMiddleware := func(next http.HandlerFunc) http.HandlerFunc { // ВСТАВКА ФРОНТЕНД
		return func(w http.ResponseWriter, r *http.Request) { // ВСТАВКА ФРОНТЕНД
			w.Header().Set("Access-Control-Allow-Origin", "*")                                // ВСТАВКА ФРОНТЕНД
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS") // ВСТАВКА ФРОНТЕНД
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")                    // ВСТАВКА ФРОНТЕНД
			if r.Method == "OPTIONS" {                                                        // ВСТАВКА ФРОНТЕНД
				w.WriteHeader(http.StatusOK) // ВСТАВКА ФРОНТЕНД
				return                       // ВСТАВКА ФРОНТЕНД
			} // ВСТАВКА ФРОНТЕНД
			next(w, r) // ВСТАВКА ФРОНТЕНД
		} // ВСТАВКА ФРОНТЕНД

	} // ВСТАВКА ФРОНТЕНД
	// ============================================================
	// 3. ЭНДПОИНТ /api/config — ОТДАЁТ АДРЕС СЕРВЕРА ДЛЯ ФРОНТЕНДА
	// ============================================================
	// Фронтенд загружает адрес бэкенда динамически, чтобы не было
	// жёсткой привязки к localhost:8080. Это позволяет менять порт
	// или хост без пересборки фронтенда.
	//
	// Пример ответа:
	//   {"apiBase": "http://localhost:8080/api"}
	// ============================================================
	mux.HandleFunc("GET /api/config", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// 1. Получаем порт из переменной окружения (или 8080 по умолчанию)
		port := os.Getenv("SERVER_PORT")
		if port == "" {
			port = "8080"
		}

		// 2. Формируем объект с адресом бэкенда
		config := map[string]string{
			"apiBase": "http://localhost:" + port + "/api",
		}

		// 3. Устанавливаем заголовок, что отдаём JSON
		w.Header().Set("Content-Type", "application/json")

		// 4. Кодируем объект в JSON и отправляем клиенту
		if err := json.NewEncoder(w).Encode(config); err != nil {
			logger.Error("Failed to encode config: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}))
		// ============================================================
	// 4. ЛОГИН (заглушка для теста)
	// ============================================================
	// Этот эндпоинт обрабатывает POST /api/login.
	// Принимает JSON с полями username и password.
	// Возвращает токен (заглушка) и роль пользователя.
	//
	// Пример запроса:
	//   {"username": "admin", "password": "admin"}
	//
	// Пример ответа:
	//   {"token": "fake-jwt-token", "username": "admin", "role": "user"}
	// ============================================================
	mux.HandleFunc("POST /api/login", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// 1. Проверяем, что метод POST
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// 2. Читаем тело запроса
		var creds struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			logger.Warn("Login: invalid JSON: %v", err)
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// 3. Проверяем логин и пароль (заглушка)
		// В реальном проекте здесь был бы запрос к БД и сравнение хешей
		if (creds.Username == "admin" && creds.Password == "admin") ||
		   (creds.Username == "user" && creds.Password == "user") {

			// 4. Генерируем фейковый JWT-токен (для теста)
			token := "fake-jwt-token-for-" + creds.Username

			// 5. Определяем роль
			role := "user"
			if creds.Username == "admin" {
				role = "admin"
			}

			// 6. Отправляем ответ
			w.Header().Set("Content-Type", "application/json")
			response := map[string]string{
				"token":    token,
				"username": creds.Username,
				"role":     role,
			}
			if err := json.NewEncoder(w).Encode(response); err != nil {
				logger.Error("Login: failed to encode response: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		// 7. Если логин/пароль не подошли
		logger.Warn("Login: invalid credentials for user: %s", creds.Username)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
	}))
	// CRUDL операции
	mux.HandleFunc("POST    /api/subscriptions", corsMiddleware(handlers.LoggingMiddleware(handler.CreateSubscriptionHandler)))      // ИЗМЕНЕНО ПОД ФРОНТЕНД
	mux.HandleFunc("GET     /api/subscriptions/{id}", corsMiddleware(handlers.LoggingMiddleware(handler.GetSubscriptionHandler)))    // ИЗМЕНЕНО ПОД ФРОНТЕНД
	mux.HandleFunc("PUT     /api/subscriptions/{id}", corsMiddleware(handlers.LoggingMiddleware(handler.UpdateSubscriptionHandler))) // ИЗМЕНЕНО ПОД ФРОНТЕНД
	mux.HandleFunc("DELETE  /api/subscriptions/{id}", corsMiddleware(handlers.LoggingMiddleware(handler.DeleteSubscriptionHandler))) // ИЗМЕНЕНО ПОД ФРОНТЕНД
	mux.HandleFunc("GET     /api/subscriptions", corsMiddleware(handlers.LoggingMiddleware(handler.ListSubscriptionsHandler)))       // ИЗМЕНЕНО ПОД ФРОНТЕНД
	mux.HandleFunc("GET     /api/subscriptions/total-cost", corsMiddleware(handlers.LoggingMiddleware(handler.GetTotalCostHandler))) // ИЗМЕНЕНО ПОД ФРОНТЕНД
	mux.HandleFunc("GET     /swagger/", httpSwagger.WrapHandler)
	// ============================================================
	// HEALTHCHECK (для Docker)
	// ============================================================
	mux.HandleFunc("GET /health", handlers.LoggingMiddleware(handlers.HealthHandler))

	//  Получаем порт из .env
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
		logger.Warn("SERVER_PORT not set, using default 8080")
	}
	// Создаем HTTP сервер с таймаутами
	srv := &http.Server{ // указатель на структуру http.Server.... поля структуры:

		Addr:         ":" + port,       // адрес и порт, на котором сервер будет слушать запросы
		Handler:      mux,              // роутер, который будет обрабатывать входящие запросы
		ReadTimeout:  5 * time.Second,  // максимальное время на чтение всего запроса (заголовки + тело) — защита от медленных клиентов
		WriteTimeout: 10 * time.Second, // максимальное время на запись ответа — защита от зависших хендлеров
		IdleTimeout:  15 * time.Second, // максимальное время жизни keep-alive соединения без новых запросов
	}
	// Запускаем сервер в горутине
	go func() {
		logger.Info("Server starting on port %s", port)
		if err2 := srv.ListenAndServe(); err2 != nil && err2 != http.ErrServerClosed { // Обработка ошибок сервера
			logger.Error("Server failed: %v", err2)
			os.Exit(1) // Завершаем программу по аварии с кодом 1
		}
	}()

	//  Graceful shutdown (ожидание сигнала на отключение)
	quit := make(chan os.Signal, 1)                      // Создаем канал для ожидания сигналов
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // Наблюдаем за сигналами SIGINT и SIGTERM
	<-quit                                               // Блокируем до получения сигнала

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 11*time.Second) //      Контекст с таймаутом на завершение (11 секунд > WriteTimeout)
	defer cancel()

	//  Останавливаем сервер по сигналу от контекста
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown: %v", err)
		os.Exit(1)
	}

	logger.Info("Server exited properly")
	return nil
}

// Реализуем логику для rolling back
func shouldRollback() bool {
	return len(os.Args) > 1 && os.Args[1] == "-down"
}
func rollbackMigrations() error {
	if err := database.RollbackMigrations(); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}
	logger.Info("Migration rolled back")
	return nil
}
func applyMigrations() error {
	if err := database.RunMigrations(); err != nil {
		logger.Warn("Migrations warning (maybe already applied): %v", err)
	}
	return nil
}
