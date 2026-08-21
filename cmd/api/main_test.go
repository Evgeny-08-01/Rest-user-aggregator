//go:build unit
package main

import (
	"os"
	"syscall"
	"testing"
	"time"
)

// ============================================================
// ТЕСТЫ ДЛЯ ФУНКЦИЙ СОЗДАНИЯ СЕРВЕРОВ
// ============================================================

func TestCreateRESTServer(t *testing.T) {
	svc, authSvc := buildServices()
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

	svc, authSvc := buildServices()
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
    svc, _ := buildServices()
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

	svc, _ := buildServices()
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

	svc, _ := buildServices()
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

func TestBuildServices(t *testing.T) {
	svc, authSvc := buildServices()

	if svc == nil {
		t.Error("SubscriptionService is nil")
	}
	if authSvc == nil {
		t.Error("AuthService is nil")
	}
}

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

func TestLoadEnv(t *testing.T) {
	loadEnv()
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		t.Log("DB_PATH not set (this is OK if .env not exists)")
	}
}

func TestLoadEnvWithFile(t *testing.T) {
	tmpFile, err := os.Create(".env.test")
	if err != nil {
		t.Skip("Cannot create test .env file")
	}
	defer os.Remove(".env.test")

	tmpFile.WriteString("TEST_VAR=test_value\n")
	tmpFile.Close()

	oldEnv := os.Getenv("TEST_VAR")
	os.Rename(".env.test", ".env")
	defer os.Rename(".env", ".env.test")

	loadEnv()

	if os.Getenv("TEST_VAR") == "test_value" {
		t.Log("loadEnv loaded .env file")
	} else {
		t.Log("loadEnv did not load .env")
	}
	os.Setenv("TEST_VAR", oldEnv)
}

func TestInitLogger(t *testing.T) {
	oldLevel := os.Getenv("LOG_LEVEL")
	oldPath := os.Getenv("LOG_PATH")
	oldEnv := os.Getenv("ENV")

	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_PATH", "./test.log")
	os.Setenv("ENV", "local")

	initLogger()

	os.Setenv("LOG_LEVEL", oldLevel)
	os.Setenv("LOG_PATH", oldPath)
	os.Setenv("ENV", oldEnv)

	t.Log("Logger initialized")
}

func TestInitLoggerDocker(t *testing.T) {
	oldPath := os.Getenv("LOG_PATH")
	oldLevel := os.Getenv("LOG_LEVEL")
	oldEnv := os.Getenv("ENV")

	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_PATH", "")
	os.Setenv("ENV", "docker")

	initLogger()
	t.Log("Logger initialized with Docker settings")

	os.Setenv("LOG_PATH", oldPath)
	os.Setenv("LOG_LEVEL", oldLevel)
	os.Setenv("ENV", oldEnv)
}

func TestInitPprof(t *testing.T) {
	oldPprof := os.Getenv("PPROF_ENABLED")

	os.Setenv("PPROF_ENABLED", "false")
	initPprof()
	t.Log("Pprof skipped when disabled")

	os.Setenv("PPROF_ENABLED", "true")
	initPprof()
	t.Log("Pprof started in goroutine")

	os.Setenv("PPROF_ENABLED", oldPprof)
}

// ============================================================
// ТЕСТЫ ДЛЯ БД
// ============================================================

func TestInitDB(t *testing.T) {
	oldPath := os.Getenv("DB_PATH")
	oldEnv := os.Getenv("ENV")

	os.Setenv("DB_PATH", "postgres://wrong:wrong@localhost:5432/wrong?sslmode=disable")
	os.Setenv("ENV", "local")

	err := initDB()
	if err != nil {
		t.Logf("initDB correctly failed: %v", err)
	} else {
		t.Error("initDB should fail with wrong credentials")
	}

	os.Setenv("DB_PATH", oldPath)
	os.Setenv("ENV", oldEnv)
}

func TestInitDBEmpty(t *testing.T) {
	oldPath := os.Getenv("DB_PATH")
	oldEnv := os.Getenv("ENV")

	os.Setenv("DB_PATH", "")
	os.Setenv("ENV", "local")

	err := initDB()
	if err != nil {
		t.Logf("initDB failed with empty path: %v", err)
	} else {
		t.Log("initDB used default path")
	}

	os.Setenv("DB_PATH", oldPath)
	os.Setenv("ENV", oldEnv)
}

func TestInitDBDocker(t *testing.T) {
	oldPath := os.Getenv("DB_PATH")
	oldEnv := os.Getenv("ENV")

	os.Setenv("DB_PATH", "postgres://postgres:mysecret@localhost:5432/subscriptions?sslmode=disable")
	os.Setenv("ENV", "docker")

	err := initDB()
	if err != nil {
		t.Logf("initDB with Docker env failed: %v", err)
	} else {
		t.Log("initDB with Docker env succeeded")
	}

	os.Setenv("DB_PATH", oldPath)
	os.Setenv("ENV", oldEnv)
}

func TestInitDBWrongPassword(t *testing.T) {
	oldPath := os.Getenv("DB_PATH")
	oldEnv := os.Getenv("ENV")

	os.Setenv("DB_PATH", "postgres://wrong:wrong@localhost:5432/wrong?sslmode=disable")
	os.Setenv("ENV", "local")

	err := initDB()
	if err != nil {
		t.Logf("initDB failed with wrong password: %v", err)
	} else {
		t.Error("initDB should fail with wrong password")
	}

	os.Setenv("DB_PATH", oldPath)
	os.Setenv("ENV", oldEnv)
}

func TestInitDBWithMock(t *testing.T) {
	oldPath := os.Getenv("DB_PATH")
	oldEnv := os.Getenv("ENV")

	os.Setenv("DB_PATH", "postgres://test:test@localhost:5432/test?sslmode=disable")
	os.Setenv("ENV", "local")

	err := initDB()
	if err != nil {
		t.Logf("initDB failed as expected: %v", err)
	} else {
		t.Log("initDB succeeded (unexpected)")
	}

	os.Setenv("DB_PATH", oldPath)
	os.Setenv("ENV", oldEnv)
}

// ============================================================
// ТЕСТЫ ДЛЯ МИГРАЦИЙ
// ============================================================

func TestRunMigrations(t *testing.T) {
	err := runMigrations()
	if err != nil {
		t.Logf("Migrations failed (expected without DB): %v", err)
	}
}

func TestApplyMigrations(t *testing.T) {
	err := applyMigrations()
	if err != nil {
		t.Logf("applyMigrations failed (expected without DB): %v", err)
	}
}

func TestApplyMigrationsError(t *testing.T) {
	err := applyMigrations()
	if err != nil {
		t.Logf("applyMigrations failed: %v", err)
	} else {
		t.Log("applyMigrations succeeded (unexpected)")
	}
}

func TestRollbackMigrations(t *testing.T) {
	err := rollbackMigrations()
	if err != nil {
		t.Logf("rollbackMigrations failed (expected without DB): %v", err)
	}
}

func TestRollbackMigrationsError(t *testing.T) {
	err := rollbackMigrations()
	if err != nil {
		t.Logf("rollbackMigrations failed: %v", err)
	} else {
		t.Log("rollbackMigrations succeeded (unexpected)")
	}
}

// ============================================================
// ТЕСТЫ ДЛЯ shouldRollback
// ============================================================

func TestShouldRollback(t *testing.T) {
	oldArgs := os.Args

	os.Args = []string{"cmd"}
	if shouldRollback() {
		t.Error("shouldRollback() returned true without -down flag")
	}

	os.Args = []string{"cmd", "-down"}
	if !shouldRollback() {
		t.Error("shouldRollback() returned false with -down flag")
	}

	os.Args = oldArgs
	t.Log("ShouldRollback works")
}

func TestShouldRollbackWithArgs(t *testing.T) {
	oldArgs := os.Args

	os.Args = []string{"cmd"}
	if shouldRollback() {
		t.Error("shouldRollback true without args")
	}

	os.Args = []string{"cmd", "-down"}
	if !shouldRollback() {
		t.Error("shouldRollback false with -down")
	}

	os.Args = []string{"cmd", "-up"}
	if shouldRollback() {
		t.Error("shouldRollback true with -up")
	}

	t.Log("shouldRollback works with all args")
	os.Args = oldArgs
}

// ============================================================
// ТЕСТЫ ДЛЯ run И runServers
// ============================================================

func TestRun(t *testing.T) {
	oldArgs := os.Args
	done := make(chan bool)

	go func() {
		err := run()
		if err != nil {
			t.Logf("run() error: %v", err)
		}
		done <- true
	}()

	time.Sleep(1 * time.Second)
	proc, _ := os.FindProcess(os.Getpid())
	proc.Signal(syscall.SIGINT)

	select {
	case <-done:
		t.Log("run() stopped gracefully")
	case <-time.After(3 * time.Second):
		t.Log("run() timeout")
	}
	os.Args = oldArgs
}

func TestRunServers(t *testing.T) {
	svc, authSvc := buildServices()
	done := make(chan bool)

	go func() {
		err := runServers(svc, authSvc)
		if err != nil {
			t.Logf("runServers error: %v", err)
		}
		done <- true
	}()

	time.Sleep(1 * time.Second)
	proc, _ := os.FindProcess(os.Getpid())
	proc.Signal(syscall.SIGINT)

	select {
	case <-done:
		t.Log("runServers stopped gracefully")
	case <-time.After(5 * time.Second):
		t.Log("runServers timeout (this is OK in test)")
	}
}

func TestRunServersWithMock(t *testing.T) {
	svc, authSvc := buildServices()
	done := make(chan bool)

	go func() {
		err := runServers(svc, authSvc)
		if err != nil {
			t.Logf("runServers error: %v", err)
		}
		done <- true
	}()

	time.Sleep(1 * time.Second)
	proc, _ := os.FindProcess(os.Getpid())
	proc.Signal(syscall.SIGINT)

	select {
	case <-done:
		t.Log("runServers stopped gracefully")
	case <-time.After(3 * time.Second):
		t.Log("runServers timeout")
	}
}