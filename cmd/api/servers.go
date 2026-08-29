package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"


	"Rest-user-agregator/internal/database"
	handlers "Rest-user-agregator/internal/handlers/rest"
	"Rest-user-agregator/internal/middleware"
	"Rest-user-agregator/internal/service"
	"Rest-user-agregator/pkg/logger"
	pb "Rest-user-agregator/proto/subscription"
	grpcserver "Rest-user-agregator/internal/handlers/grpc"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

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

func runServers(svc *service.SubscriptionService, authService *service.AuthService, templateSvc *service.TemplateService) error {
	// 1. СОЗДАНИЕ СЕРВЕРОВ
	restServer := createRESTServer(svc, authService)              // настройка REST сервера
	grpcServer, lis, err := createGRPCServer(svc, templateSvc)    // настройка GRPC сервер
	if err != nil {
		return fmt.Errorf("failed to start gPRC server: %w", err)
	}

	// 2. ПАРАЛЛЕЛЬНЫЙ ЗАПУСК
	var wg sync.WaitGroup        // Счётчик горутин
	chErr := make(chan error, 2) // Канал для ошибок от серверов

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

// Запуск сервера REST API
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
// 1. СТАТИКА И ФРОНТЕНД
// ============================================================
// 1. Статика (CSS, JS, картинки)
mux.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./web/css"))))
mux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./web/js"))))

// 2. HTML-страницы
mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    switch r.URL.Path {
    case "/", "/index.html":
        http.ServeFile(w, r, "./web/index.html")
    case "/templates.html":
        http.ServeFile(w, r, "./web/templates.html")
    default:
        http.NotFound(w, r)
    }
})

// ============================================================
// 2. ПУБЛИЧНЫЕ ЭНДПОИНТЫ (без авторизации)
// ============================================================
mux.HandleFunc("POST /api/register", middleware.CorsMiddleware(handlers.LoggingMiddleware(handler.RegistrationHandler)))
mux.HandleFunc("POST /api/login", middleware.CorsMiddleware(handlers.LoggingMiddleware(handler.LoginHandler)))
mux.HandleFunc("GET /api/config", middleware.CorsMiddleware(handlers.LoggingMiddleware(handler.ConfigHandler)))
mux.HandleFunc("GET /health", middleware.CorsMiddleware(handlers.LoggingMiddleware(handlers.HealthHandler)))
mux.HandleFunc("GET /swagger/", httpSwagger.WrapHandler)
mux.Handle("/metrics", promhttp.Handler())

// ============================================================
// 3. ПОДПИСКИ (требуют авторизацию)
// ============================================================
mux.HandleFunc("POST /api/subscriptions", handlers.WrapHandler(handler.CreateSubscriptionHandler))
mux.HandleFunc("GET /api/subscriptions", handlers.WrapHandler(handler.ListSubscriptionsHandler))
mux.HandleFunc("GET /api/subscriptions/{id}", handlers.WrapHandler(handler.GetSubscriptionHandler))
mux.HandleFunc("PUT /api/subscriptions/{id}", handlers.WrapHandler(handler.UpdateSubscriptionHandler))
mux.HandleFunc("DELETE /api/subscriptions/{id}", handlers.WrapHandler(handler.DeleteSubscriptionHandler))
mux.HandleFunc("GET /api/subscriptions/total-cost", handlers.WrapHandler(handler.GetTotalCostHandler))
// СОЗДАЁМ ШАБЛОННЫЙ ХЕНДЛЕР
repo := database.NewPostgresRepo()
templateService := service.NewTemplateService(repo)
templateHandler := handlers.NewTemplateHandler(templateService)
// ============================================================
// 4. ШАБЛОНЫ (требуют авторизацию + проверку роли внутри)
// ============================================================
mux.HandleFunc("POST /api/admin/templates", handlers.WrapHandler(templateHandler.CreateTemplateHandler))
mux.HandleFunc("GET /api/templates", handlers.WrapHandler(templateHandler.ListTemplatesHandler))
mux.HandleFunc("GET /api/templates/{id}", handlers.WrapHandler(templateHandler.GetTemplateHandler))
mux.HandleFunc("PUT /api/admin/templates/{id}", handlers.WrapHandler(templateHandler.UpdateTemplateHandler))
mux.HandleFunc("DELETE /api/admin/templates/{id}", handlers.WrapHandler(templateHandler.DeleteTemplateHandler))

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
func createGRPCServer(svc *service.SubscriptionService, templateSvc *service.TemplateService) (*grpc.Server, net.Listener, error) {
	// 1. ПОРТ ИЗ .env
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
		logger.Warn("GRPC_PORT not set, using default: %s", grpcPort)
	}

	// 2. СОЗДАНИЕ gRPC-ОБРАБОТЧИКА
	 grpcHandler := grpcserver.NewSubscriptionServer(svc, templateSvc)

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
		return nil, nil, err
	}

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
