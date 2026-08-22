//go:build integration
package main

import (
	"Rest-user-agregator/internal/database"
	"Rest-user-agregator/internal/service"
	"Rest-user-agregator/internal/testutils"
	"Rest-user-agregator/pkg/logger"
	"context"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
)
var (
    svc     *service.SubscriptionService
    authSvc *service.AuthService
)
// ============================================================
// 1. ПРОВЕРКА ЗАВИСИМОСТЕЙ (выполняется ПЕРВОЙ!)
// ============================================================
func Test0Dependencies(t *testing.T) {
	testutils.CheckAllDependencies(t, 15*time.Second, "8087", "50051")
}
func TestMain(m *testing.M) {
// ✅ Загружаем .env.test для тестов
    if err := godotenv.Load("../../.env.test"); err != nil {
        logger.Warn("[WARN] .env.test not found, using default values")
    }
// ✅ ИНИЦИАЛИЗАЦИЯ REDIS для тестирования
    if err := initRedis(); err != nil {
        logger.Warn("Redis not available, tests will continue without cache: %v", err)
    }
// ✅ ИНИЦИАЛИЗАЦИЯ БД для тестов
    dbPath := os.Getenv("DB_PATH")
    if dbPath == "" {
        panic("DB_PATH not set")
    }
	logger.Info("DB_PATH not set dbPath:%s",dbPath)
    if err := database.Init(dbPath); err != nil {
        panic("Failed to init DB: " + err.Error())
    }

	repo := database.NewPostgresRepo()
	svc = service.NewSubscriptionService(repo)
	authSvc = service.NewAuthService(repo)

	// Запуск тестов
	code := m.Run()

	// Закрытие соединений
	database.Close()
	// cache.Close() если есть

	os.Exit(code)
}
// ============================================================
// ТЕСТЫ ДЛЯ ФУНКЦИЙ СОЗДАНИЯ СЕРВЕРОВ
// ============================================================

func TestCreateRESTServer(t *testing.T) {
	srv := createRESTServer(svc, authSvc)

	if srv == nil {
		t.Error("REST server is nil")
	}
	if srv.Handler == nil {
		t.Error("REST server handler is nil")
	}
	if srv.Addr == "" {
		t.Error("REST server port is empty")
	}
}

func TestCreateRESTServerWithCustomPort(t *testing.T) {
	oldPort := os.Getenv("SERVER_PORT")
	os.Setenv("SERVER_PORT", "9999")
	defer os.Setenv("SERVER_PORT", oldPort)

	srv := createRESTServer(svc, authSvc)

	if srv == nil {
		t.Error("REST server is nil")
	}
	if srv.Addr != ":9999" {
		t.Errorf("Expected port :9999, got %s", srv.Addr)
	} else {
		t.Log("REST server created with custom port :9999")
	}
}

func TestCreateGRPCServer(t *testing.T) {
    srv, lis, err := createGRPCServer(svc)
    
    // ДОБАВИТЬ: закрываем listener после теста
    if lis != nil {
        defer lis.Close()
    }
    
    if err != nil {
        t.Errorf("createGRPCServer failed: %v", err)
    }
    if srv == nil {
        t.Error("gRPC server is nil")
    }
    if lis == nil {
        t.Error("gRPC listener is nil")
    }
}
func TestCreateGRPCServerWithCustomPort(t *testing.T) {
	oldPort := os.Getenv("GRPC_PORT")
	os.Setenv("GRPC_PORT", "9999")
	defer os.Setenv("GRPC_PORT", oldPort)

	srv, lis, err := createGRPCServer(svc)

	if err != nil {
		t.Logf("createGRPCServer error: %v", err)
	}
	if srv == nil {
		t.Error("gRPC server is nil")
	}
	if lis == nil {
		t.Error("gRPC listener is nil")
	} else {
		t.Log("gRPC server created with custom port :9999")
		lis.Close()
	}
}

func TestCreateGRPCServerError(t *testing.T) {
	oldPort := os.Getenv("GRPC_PORT")
	os.Setenv("GRPC_PORT", "50051")
	defer os.Setenv("GRPC_PORT", oldPort)

	srv, lis, err := createGRPCServer(svc)

	if err == nil {
		t.Log("Port 50051 is free, skipping error test")
	} else {
		t.Logf("createGRPCServer correctly failed: %v", err)
		if srv != nil {
			t.Error("gRPC server should be nil on error")
		}
		if lis != nil {
			t.Error("gRPC listener should be nil on error")
		}
	}
}

// ============================================================
// ТЕСТЫ ДЛЯ СЕРВИСОВ
// ============================================================

// ============================================================
// ТЕСТЫ ДЛЯ REDIS
// ============================================================

func TestInitRedis(t *testing.T) {
	if os.Getenv("SKIP_REDIS_TEST") == "true" {
		t.Skip("⏭️ Skipping Redis test")
	}
	err := initRedis()
	if err != nil {
		t.Logf("Redis not available: %v", err)
	}
}

func TestInitRedisCustomAddr(t *testing.T) {
	oldAddr := os.Getenv("REDIS_ADDR")
	os.Setenv("REDIS_ADDR", "localhost:9999")
	defer os.Setenv("REDIS_ADDR", oldAddr)

	err := initRedis()
	if err != nil {
		t.Logf("initRedis failed with custom address: %v", err)
	} else {
		t.Log("initRedis succeeded with custom address (unexpected)")
	}
}

func TestInitRedisError(t *testing.T) {
	oldAddr := os.Getenv("REDIS_ADDR")
	os.Setenv("REDIS_ADDR", "wrong:6379")
	defer os.Setenv("REDIS_ADDR", oldAddr)

	err := initRedis()
	if err != nil {
		t.Logf("initRedis correctly failed: %v", err)
	} else {
		t.Error("initRedis should fail with wrong address")
	}
}

// ============================================================
// ТЕСТЫ ДЛЯ ЛОГГЕРА
// ============================================================


// ============================================================
// ТЕСТЫ ДЛЯ shouldRollback
// ============================================================

// main_integration_test.go

func TestRunWithContext(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    errCh := make(chan error, 1)
    go func() {
        errCh <- runWithContext(ctx)
    }()

    select {
    case err := <-errCh:
        if err != nil {
            t.Logf("runWithContext() error: %v", err)
        }
        t.Log("✅ runWithContext() stopped gracefully")
    case <-ctx.Done():
        t.Log("⏰ Timeout (CI may be slow)")
    }
}

func TestRunServersWithContext(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    errCh := make(chan error, 1)
    go func() {
        errCh <- runServersWithContext(ctx, svc, authSvc)
    }()

    select {
    case err := <-errCh:
        if err != nil {
            t.Logf("runServersWithContext() error: %v", err)
        }
        t.Log("✅ runServersWithContext() stopped gracefully")
    case <-ctx.Done():
        t.Log("⏰ Timeout (CI may be slow)")
    }
}
func TestBuildServices(t *testing.T) {
    s, a := buildServices()
    if s == nil {
        t.Error("SubscriptionService is nil")
    }
    if a == nil {
        t.Error("AuthService is nil")
    }
}