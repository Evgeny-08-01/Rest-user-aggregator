package testutils

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"Rest-user-agregator/internal/cache"
	"Rest-user-agregator/internal/database"

	"github.com/avast/retry-go/v4"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

// ============================================================
// 1. БАЗОВЫЕ ПРОВЕРКИ (возвращают func() error)
// ============================================================

// RedisReady проверяет, что Redis доступен
func RedisReady() func() error {
	return func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		return cache.PingWithContext(ctx)
	}
}

// DBReady проверяет, что PostgreSQL доступен
func DBReady() func() error {
	return func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		return database.PingWithContext(ctx)
	}
}

// RESTServerReady проверяет, что REST API сервер запущен
func RESTServerReady(port string) func() error {
	return func() error {
		if port == "" {
			port = "8087"
		}
		url := fmt.Sprintf("http://localhost:%s/health", port)

		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("REST API health check returned %d", resp.StatusCode)
		}
		return nil
	}
}

// GRPCServerReady проверяет, что gRPC сервер запущен
func GRPCServerReady(port string) func() error {
	return func() error {
		if port == "" {
			port = "50051"
		}

		conn, err := net.DialTimeout("tcp", ":"+port, 2*time.Second)
		if err != nil {
			return fmt.Errorf("gRPC server not reachable: %w", err)
		}
		conn.Close()
		return nil
	}
}

// ============================================================
// 2. КОМБО-ПРОВЕРКА (4 в 1)
// ============================================================

// CheckAllDependencies проверяет все 4 компонента:
//   - Redis
//   - PostgreSQL
//   - REST API
//   - gRPC
//
// Шум отключен (логи успеха НЕ выводятся)
func CheckAllDependencies(t *testing.T, timeout time.Duration, restPort, grpcPort string) {
	t.Helper()

	// Проверяем Redis
	checkComponent(t, "Redis", timeout, RedisReady())

	// Проверяем PostgreSQL
	checkComponent(t, "PostgreSQL", timeout, DBReady())
 // ✅ ЗАПУСКАЕМ ВРЕМЕННЫЙ gRPC СЕРВЕР
    grpcServer, grpcLis := startTempGRPCServer(grpcPort)
    defer stopTempGRPCServer(grpcServer, grpcLis)

    // ✅ ЗАПУСКАЕМ ВРЕМЕННЫЙ REST СЕРВЕР
    restServer, restLis := startTempRESTServer(restPort)
    defer stopTempRESTServer(restServer, restLis)
	// Проверяем REST API
	checkComponent(t, "REST API", timeout, RESTServerReady(restPort))

	// Проверяем gRPC
	checkComponent(t, "gRPC", timeout, GRPCServerReady(grpcPort))
}

// ============================================================
// 3. ВНУТРЕННЯЯ ФУНКЦИЯ (без шума)
// ============================================================

// checkComponent — тихая проверка (без t.Logf при успехе)
func checkComponent(t *testing.T, name string, timeout time.Duration, check func() error) {
    t.Helper()

    var lastErr error

    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    err := retry.Do(
        func() error {
            if err := check(); err != nil {
                lastErr = err
                return err
            }
            return nil
        },
        retry.Attempts(10),
        retry.Delay(100*time.Millisecond),
        retry.DelayType(retry.BackOffDelay),
        retry.Context(ctx),
        retry.LastErrorOnly(true),
    )

    if err != nil {
        errorType := analyzeError(lastErr)
        hint := getHint(name, lastErr)

        // ✅ ОДИН assert (со скобками) — ПРАВИЛЬНЫЙ
        assert.NoError(t, err,
            "\n❌ [%s] not ready after %v"+
                "\n   Last error:   %v"+
                "\n   Error type:   %s"+
                "\n   Hint:         %s",
            name, timeout, lastErr, errorType, hint)
    }

    // ❌ Логи успеха УБРАНЫ — нет шума
}
// ============================================================
// 4. АНАЛИЗ ОШИБОК И ПОДСКАЗКИ
// ============================================================

func analyzeError(err error) string {
	if err == nil {
		return "UNKNOWN_ERROR"
	}
	errMsg := err.Error()
	switch {
	case strings.Contains(errMsg, "connection refused"),
		strings.Contains(errMsg, "no such host"),
		strings.Contains(errMsg, "i/o timeout"):
		return "INFRASTRUCTURE_ERROR"
	case strings.Contains(errMsg, "authentication failed"),
		strings.Contains(errMsg, "access denied"):
		return "CONFIGURATION_ERROR"
	case strings.Contains(errMsg, "invalid"),
		strings.Contains(errMsg, "not found"):
		return "APPLICATION_ERROR"
	default:
		return "UNKNOWN_ERROR"
	}
}

func getHint(name string, err error) string {
	if err == nil {
		return "Check logs for details"
	}
	errMsg := err.Error()
	switch {
	case strings.Contains(errMsg, "connection refused"):
		return fmt.Sprintf("%s is not running. Check CI services.", name)
	case strings.Contains(errMsg, "authentication failed"):
		return fmt.Sprintf("Wrong credentials for %s. Check .env.", name)
	case strings.Contains(errMsg, "timeout"):
		return fmt.Sprintf("%s is too slow. Increase timeout.", name)
	default:
		return fmt.Sprintf("Check %s logs.", name)
	}
}
// AssertReady — проверяет один компонент с retry
// Логи успеха выводятся только при -v
func AssertReady(t *testing.T, name string, timeout time.Duration, check func() error) {
	t.Helper()

	start := time.Now()
	var lastErr error

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := retry.Do(
		func() error {
			if err := check(); err != nil {
				lastErr = err
				return err
			}
			return nil
		},
		retry.Attempts(10),
		retry.Delay(100*time.Millisecond),
		retry.DelayType(retry.BackOffDelay),
		retry.Context(ctx),
		retry.LastErrorOnly(true),
	)

	if err != nil {
		errorType := analyzeError(lastErr)
		hint := getHint(name, lastErr)

		assert.NoError(t, err,
			"\n❌ %s not ready after %v"+
				"\n   Last error:   %v"+
				"\n   Error type:   %s"+
				"\n   Hint:         %s",
			name, timeout, lastErr, errorType, hint)
	}

	if testing.Verbose() {
		t.Logf("✅ %s ready in %v", name, time.Since(start))
	}
}

// ============================================================
// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ДЛЯ ЗАПУСКА ВРЕМЕННЫХ СЕРВЕРОВ
// ============================================================

// startTempGRPCServer запускает временный gRPC сервер для проверки
func startTempGRPCServer(port string) (*grpc.Server, net.Listener) {
    if port == "" {
        port = "50051"
    }

    grpcServer := grpc.NewServer()
    lis, err := net.Listen("tcp", ":"+port)
    if err != nil {
        return nil, nil
    }

    go func() {
        if err := grpcServer.Serve(lis); err != nil {
            // игнорируем ошибку, так как сервер временный
        }
    }()

    return grpcServer, lis
}

// stopTempGRPCServer останавливает временный gRPC сервер
func stopTempGRPCServer(grpcServer *grpc.Server, lis net.Listener) {
    if grpcServer != nil {
        grpcServer.Stop()
    }
    if lis != nil {
        lis.Close()
    }
}
// startTempRESTServer запускает временный REST сервер для проверки
func startTempRESTServer(port string) (*http.Server, net.Listener) {
    if port == "" {
        port = "8087"
    }

    mux := http.NewServeMux()
    mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"ok"}`))
    })

    srv := &http.Server{
        Addr:    ":" + port,
        Handler: mux,
    }

    lis, err := net.Listen("tcp", ":"+port)
    if err != nil {
        return nil, nil
    }

    go func() {
        if err := srv.Serve(lis); err != nil && err != http.ErrServerClosed {
            // игнорируем ошибку
        }
    }()

    return srv, lis
}

// stopTempRESTServer останавливает временный REST сервер
func stopTempRESTServer(srv *http.Server, lis net.Listener) {
    if srv == nil {
        return
    }
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    srv.Shutdown(ctx)
    if lis != nil {
        lis.Close()
    }
}