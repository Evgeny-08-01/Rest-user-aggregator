package main

// ============================================================
// 1. ИМПОРТЫ
// ============================================================

import (
	"context" // ВСТАВКА ФРОНТЕНД
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "Rest-user-agregator/docs"
	"Rest-user-agregator/internal/authentication"
	"Rest-user-agregator/internal/database"
	"Rest-user-agregator/internal/handlers"
	"Rest-user-agregator/internal/middleware"
	"Rest-user-agregator/internal/service"
	"Rest-user-agregator/pkg/logger"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/swaggo/http-swagger"
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
	
defer func() {
    if err := database.Close(); err != nil {
        logger.Error("failed to close db: %v", err)
    }
}()
	

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
	// ============================================================
	// 3. Инициализация БД (УЖЕ БЫЛО)
	// ============================================================
	repo := database.NewPostgresRepo() // экземпляр репозитория, содержащий пул соединений и указатель на БД,
	// содержит методы работы с БД. NewPostgresRepo-конструктор над PostgresRepo
	// PostgresRepo- структура и содежит поле: db *sql.DB

	// ============================================================
	// 3.1 СОЗДАНИЕ СЕРВИСОВ
	// ============================================================
	svc := service.NewSubscriptionService(repo) // Сервис для работы с подписками (CRUD и total-cost)
	authService := service.NewAuthService(repo) // НОВЫЙ СЕРВИС: Сервис для авторизации (регистрация, логин, JWT)

	// ============================================================
	// 4. ИНИЦИАЛИЗАЦИЯ ХЭНДЛЕРА
	// ============================================================
	// Раньше хэндлер принимал только сервис подписок.
	// Теперь передаём оба сервиса: для подписок и для авторизации.
	// ============================================================
	handler := handlers.NewHandler(svc, authService) // экземпляр хендлера, содержащий экземпляр репозитория repo для работы с БД,
	// содержит методы обработки HTTP-запросов.
	// NewHandler — конструктор, создающий экземпляр Handler.
	// Handler — структура с полями Service (интерфейс) и AuthService (интерфейс).
	// !!!! Таким образом handler содержит методы обработки запросов и подключение к БД

	mux := http.NewServeMux() // Создаем роутер-switch для URL
    mux.Handle("/metrics", promhttp.Handler())
	// ============================================================
	// ФРОНТЕНД
	// ============================================================
	// ============================================================
	// 1. СТАТИКА (CSS, JS) — ТОЛЬКО ДЛЯ ЭТИХ ПАПОК- идет из index.html
	// ============================================================
	mux.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./web/css")))) // ВСТАВКА ФРОНТЕНД
	mux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./web/js"))))    // ВСТАВКА ФРОНТЕНД
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
	mux.HandleFunc("GET /api/config", middleware.MetricsMiddleware(middleware.CorsMiddleware(func(w http.ResponseWriter, r *http.Request) {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8087"
	}
	config := map[string]string{
		"apiBase": "http://localhost:" + port + "/api",
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(config); err != nil {
		logger.Error("Failed to encode config: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
})))
	// ============================================================
	// 4. АВТОРИЗАЦИЯ (регистрация и вход) — ОТКРЫТЫЕ ЭНДПОИНТЫ
	// ============================================================
	// Эти эндпоинты обрабатывают запросы на создание нового пользователя
	// и вход в систему с получением JWT-токена.
	//
	// Регистрация:  POST /api/register
	//   Тело:       { "email": "user@example.com", "password": "123456", "role": "user" }
	//   Ответ:      201 Created { "message": "User registered successfully" }
	//
	// Логин:        POST /api/login
	//   Тело:       { "email": "user@example.com", "password": "123456" }
	//   Ответ:      200 OK { "token": "jwt_token", "email": "...", "role": "..." }
	//
	// Оба эндпоинта используют CORS middleware (разрешает запросы с фронтенда)
	// и LoggingMiddleware (логирует каждый запрос).
	// ============================================================
	mux.HandleFunc("POST /api/register", middleware.CorsMiddleware(handlers.LoggingMiddleware(handler.RegistrationHandler)))
	mux.HandleFunc("POST /api/login", middleware.CorsMiddleware(handlers.LoggingMiddleware(handler.LoginHandler)))

	// ============================================================
	// 5. CRUDL операции (ЗАЩИЩЕНЫ АВТОРИЗАЦИЕЙ)
	// ============================================================
	// Все эндпоинты /api/subscriptions защищены middleware AuthMiddleware,
	// который проверяет JWT-токен и добавляет user_id в контекст.
	// ============================================================
mux.HandleFunc("POST /api/subscriptions",
	authentication.AuthMiddleware(
		middleware.CorsMiddleware(
			middleware.MetricsMiddleware(
				handlers.LoggingMiddleware(
					handler.CreateSubscriptionHandler)))))

mux.HandleFunc("GET /api/subscriptions/{id}", 
authentication.AuthMiddleware(
	middleware.CorsMiddleware(
		middleware.MetricsMiddleware(
			handlers.LoggingMiddleware(
				handler.GetSubscriptionHandler)))))

mux.HandleFunc("PUT /api/subscriptions/{id}", 
authentication.AuthMiddleware(
	middleware.CorsMiddleware(
		middleware.MetricsMiddleware(
			handlers.LoggingMiddleware(
				handler.UpdateSubscriptionHandler)))))

mux.HandleFunc("DELETE /api/subscriptions/{id}", 
authentication.AuthMiddleware(
	middleware.CorsMiddleware(
		middleware.MetricsMiddleware(
			handlers.LoggingMiddleware(
				handler.DeleteSubscriptionHandler)))))

mux.HandleFunc("GET /api/subscriptions", 
authentication.AuthMiddleware(
	middleware.CorsMiddleware(
		middleware.MetricsMiddleware(
			handlers.LoggingMiddleware(
				handler.ListSubscriptionsHandler)))))

mux.HandleFunc("GET /api/subscriptions/total-cost", 
authentication.AuthMiddleware(
	middleware.CorsMiddleware(
		middleware.MetricsMiddleware(
			handlers.LoggingMiddleware(
				handler.GetTotalCostHandler)))))
	// ============================================================
	// 6. ПУБЛИЧНЫЕ ЭНДПОИНТЫ (БЕЗ АВТОРИЗАЦИИ)
	// ============================================================

	// Swagger документация
	mux.HandleFunc("GET /swagger/", httpSwagger.WrapHandler)

	// Healthcheck (для Docker)
mux.HandleFunc("GET /health", middleware.MetricsMiddleware(
	                                 middleware.CorsMiddleware(
		                                   handlers.LoggingMiddleware(
			                                      handlers.HealthHandler))))
// ============================================================
	// 2. HTML (КОРЕНЬ) — С ВОЗМОЖНОСТЬЮ ДОБАВИТЬ ЛОГИКУ
	// ============================================================
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {                      // ВСТАВКА ФРОНТЕНД
		// Здесь можно добавить:
		// - Проверку авторизации
		// - Логирование
		// - Выбор другой страницы (login.html / index.html)
		http.ServeFile(w, r, "./web/index.html")                                            // ВСТАВКА ФРОНТЕНД
	})

	// ============================================================
	// 7. ЗАПУСК СЕРВЕРА
	// ============================================================
	// Получаем порт из .env
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

	// ============================================================
	// 8. GRACEFUL SHUTDOWN (ожидание сигнала на отключение)
	// ============================================================
	quit := make(chan os.Signal, 1)                      // Создаем канал для ожидания сигналов
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // Наблюдаем за сигналами SIGINT и SIGTERM
	<-quit                                               // Блокируем до получения сигнала

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 11*time.Second) // Контекст с таймаутом на завершение (11 секунд > WriteTimeout)
	defer cancel()

	// Останавливаем сервер по сигналу от контекста
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown: %v", err)
		os.Exit(1)
	}

	logger.Info("Server exited properly")
	return nil
}


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