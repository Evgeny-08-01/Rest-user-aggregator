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
	"sync"
	"syscall"
	"time"

	pb "Rest-user-agregator/proto/subscription"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	_ "Rest-user-agregator/docs"
	"Rest-user-agregator/internal/authentication"
	"Rest-user-agregator/internal/cache"
	"Rest-user-agregator/internal/database"
	"Rest-user-agregator/internal/handlers/grpc"
	handlers "Rest-user-agregator/internal/handlers/rest"
	"Rest-user-agregator/internal/middleware"
	"Rest-user-agregator/internal/service"
	"Rest-user-agregator/pkg/logger"
	"net"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
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
//   func run() error {
//     1. .env
//     2. Логгер
//     3. Redis (кеш — опционально)
//     4. БД (PostgreSQL — критично)
//     5. Миграции
//     6. Создание зависимостей (repo, services) — общие для REST и gRPC
//     7. Запуск приложения (startApplication)
//        - ПАРАЛЛЕЛЬНЫЙ ЗАПУСК (REST + gRPC в горутинах с WaitGroup)
//        - Ожидание сигналов остановки (SIGINT, SIGTERM)
//        - Graceful shutdown (оба сервера с таймаутом)
//        - Ожидание завершения всех горутин (WaitGroup)
//     return nil
//   }
func run() error {
    // 1. Загружаем .env (ошибки внутри)
    loadEnv()

    // 2. Инициализация логгера (ошибки внутри)
    initLogger()
    logger.Info("Starting Subscription API server")

	// 3. Инициализация логгера (ошибки внутри)
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
    svc, authSvc := buildServices()

    // 8. Запуск серверов
    return runServers(svc, authSvc)
}

// ============================================================
// 4.  ФУНКЦИИ 
// ============================================================

// 1. Загрузка .env
func loadEnv() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("[WARN] .env file not found, using default values")
	} else {
		log.Println("[INFO] .env file loaded successfully")
		log.Printf("[INFO] DB_PATH from .env: %s", os.Getenv("DB_PATH"))
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

// 7. buildServices — СОЗДАЁТ ВСЕ ЗАВИСИМОСТИ
// ============================================================
func buildServices() (*service.SubscriptionService, *service.AuthService) {
    repo := database.NewPostgresRepo()
    svc := service.NewSubscriptionService(repo)
    authService := service.NewAuthService(repo)
    return svc, authService
}



// 8. runServers — ЗАПУСК И УПРАВЛЕНИЕ ЖИЗНЕННЫМ ЦИКЛОМ СЕРВЕРОВ
// ============================================================
// 1. Создаёт REST и gRPC серверы
// 2. Запускает их параллельно (WaitGroup)
// 3. Ожидает сигнал остановки (SIGINT/SIGTERM)
// 4. Graceful shutdown (30 секунд)
// ============================================================
// Вызывается из run() после создания всех зависимостей.
// Создаёт REST и gRPC серверы, запускает их параллельно,
// ожидает сигнал остановки и выполняет graceful shutdown.
// ============================================================

func runServers(svc *service.SubscriptionService, authService *service.AuthService) error {
    // 1. СОЗДАНИЕ СЕРВЕРОВ
    restServer := createRESTServer(svc, authService) // настройка REST сервера
    grpcServer, lis,err := createGRPCServer(svc)         // настройка GRPC сервер
  if err != nil{
	return fmt.Errorf("failed to start gPRC server: %w", err)
	}

    // 2. ПАРАЛЛЕЛЬНЫЙ ЗАПУСК
    var wg sync.WaitGroup          // Счётчик горутин
    chErr := make(chan error, 2)  // Канал для ошибок от серверов

    // Запуск REST API в горутине
    wg.Add(1)
    go func() {
        defer wg.Done()
        logger.Info("REST API server starting...")
        // ListenAndServe блокирует выполнение, пока сервер не остановится
        if err := restServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            chErr <- fmt.Errorf("REST API: %w", err)
        }
    }()

    // Запуск gRPC API в горутине
    wg.Add(1)
    go func() {
        defer wg.Done()
        logger.Info("gRPC API server starting...")
        // Serve блокирует выполнение, пока сервер не остановится
        if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
            chErr <- fmt.Errorf("gRPC API: %w", err)
        }
    }()

    // 3. ОЖИДАНИЕ СИГНАЛА ОСТАНОВКИ
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    // Блокируем выполнение, пока не придёт сигнал или ошибка
    select {
    case <-quit:
        logger.Info("Shutdown signal received")
    case err := <-chErr:
        logger.Error("Server error: %v", err)
    }

    // 4. GRACEFUL SHUTDOWN (30 секунд на завершение)
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Остановка REST — не принимает новые запросы, ждёт завершения текущих
   logger.Info("Stopping REST API...")
if err := restServer.Shutdown(ctx); err != nil {
    logger.Error("REST shutdown error: %v", err)
} else {
    logger.Info("REST API stopped gracefully") 
}

    // Остановка gRPC — ждёт завершения текущих RPC-вызовов
    logger.Info("Stopping gRPC API...")
    done := make(chan struct{})
    go func() {
        grpcServer.GracefulStop() // Блокирует выполнение, ждёт завершения всех RPC
        close(done)
    }()

    // Если gRPC не завершился за 30 секунд — принудительно останавливаем
    select {
    case <-done:
        logger.Info("gRPC stopped gracefully")
    case <-ctx.Done():
        logger.Warn("gRPC timeout, forcing stop")
        grpcServer.Stop()
    }

    // 5. ЖДЁМ ЗАВЕРШЕНИЯ ВСЕХ ГОРУТИН
    wg.Wait()
    logger.Info("All servers stopped")

    return nil
}

//   Запуск сервера REST API
func createRESTServer(svc *service.SubscriptionService, authService *service.AuthService) *http.Server {

	// 1. ИНИЦИАЛИЗАЦИЯ ХЭНДЛЕРА
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

// ============================================================
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
		
return srv
}


// createGRPCServer — СОЗДАЁТ gRPC СЕРВЕР (без запуска)
// ============================================================
// Принимает сервис, создаёт gRPC-обработчик, регистрирует его
// и возвращает готовый gRPC сервер и слушатель (listener).
// Порт читается из .env (GRPC_PORT), по умолчанию 50051.
// ============================================================
func createGRPCServer(svc *service.SubscriptionService) (*grpc.Server, net.Listener,error) {
    // 1. ПОРТ ИЗ .env
    grpcPort := os.Getenv("GRPC_PORT")
    if grpcPort == "" {
        grpcPort = "50051"
        logger.Warn("GRPC_PORT not set, using default: %s", grpcPort)
    }

    // 2. СОЗДАНИЕ gRPC-ОБРАБОТЧИКА
    grpcHandler := grpcserver.NewSubscriptionServer(svc)

    // 3. СОЗДАНИЕ gRPC-СЕРВЕРА
    grpcServer := grpc.NewServer()

    // 4. РЕГИСТРАЦИЯ СЕРВИСА
    pb.RegisterSubscriptionServiceServer(grpcServer, grpcHandler)
if os.Getenv("ENV") != "production" {
        reflection.Register(grpcServer)
    }
    // 5. СОЗДАНИЕ СЛУШАТЕЛЯ (listener)
    lis, err := net.Listen("tcp", ":"+grpcPort)
    if err != nil {
        logger.Error("gRPC listen failed on port %s: %v", grpcPort, err)
        return nil, nil, err}
    

    logger.Info("gRPC server created on port %s", grpcPort)
    return grpcServer, lis, err
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
// runWithContext — для тестов (завершается по контексту, а не по сигналу)
//nolint:unused 
func runWithContext(ctx context.Context) error {
    loadEnv()
    initLogger()
    initPprof()

    if err := initRedis(); err != nil {
        logger.Warn("Redis unavailable: %v", err)
     } else {
        // ✅ ЕСЛИ ОТКРЫЛСЯ — ЗАКРЫВАЕМ ПРИ ВЫХОДЕ
        defer func() {
            if err := cache.Close(); err != nil {
                logger.Warn("Redis close error: %v", err)
            }
        }()
    }

    if err := initDB(); err != nil {
        return fmt.Errorf("DB init: %w", err)
    }
    defer database.Close()

    if err := runMigrations(); err != nil {
        return fmt.Errorf("migrations: %w", err)
    }

    svc, authSvc := buildServices()
    return runServersWithContext(ctx, svc, authSvc)
}
// runServersWithContext — для тестов (завершается по ctx.Done())
//nolint:unused 
func runServersWithContext(ctx context.Context, svc *service.SubscriptionService, authService *service.AuthService) error {
    restServer := createRESTServer(svc, authService)
    grpcServer, lis, err := createGRPCServer(svc)
    if err != nil {
        return fmt.Errorf("failed to start gRPC server: %w", err)
    }

    var wg sync.WaitGroup
    chErr := make(chan error, 2)

    // Запуск REST
    wg.Add(1)
    go func() {
        defer wg.Done()
        logger.Info("REST API server starting...")
        if err := restServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            chErr <- fmt.Errorf("REST API: %w", err)
        }
    }()

    // Запуск gRPC
    wg.Add(1)
    go func() {
        defer wg.Done()
        logger.Info("gRPC API server starting...")
        if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
            chErr <- fmt.Errorf("gRPC API: %w", err)
        }
    }()

    // Ждём сигнал ИЛИ отмену контекста
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    select {
    case <-quit:
        logger.Info("Shutdown signal received")
    case <-ctx.Done():
        logger.Info("Context canceled, shutting down...")
    case err := <-chErr:
        logger.Error("Server error: %v", err)
    }

    // Graceful shutdown (как в runServers)
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    logger.Info("Stopping REST API...")
    if err := restServer.Shutdown(shutdownCtx); err != nil {
        logger.Error("REST shutdown error: %v", err)
    }

    logger.Info("Stopping gRPC API...")
    done := make(chan struct{})
    go func() {
        grpcServer.GracefulStop()
        close(done)
    }()

    select {
    case <-done:
        logger.Info("gRPC stopped gracefully")
    case <-shutdownCtx.Done():
        logger.Warn("gRPC timeout, forcing stop")
        grpcServer.Stop()
    }

    wg.Wait()
    logger.Info("All servers stopped")
    return nil
}




