package testutils

import (
	"errors"
	"testing"
	"time"

	"Rest-user-agregator/internal/cache"
	"Rest-user-agregator/internal/database"
	"Rest-user-agregator/pkg/logger"

	"github.com/stretchr/testify/assert"
)

// ============================================================
// ТЕСТЫ ДЛЯ analyzeError
// ============================================================

func TestAnalyzeError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: "UNKNOWN_ERROR",
		},
		{
			name:     "connection refused",
			err:      errors.New("dial tcp 127.0.0.1:6379: connect: connection refused"),
			expected: "INFRASTRUCTURE_ERROR",
		},
		{
			name:     "no such host",
			err:      errors.New("lookup redis: no such host"),
			expected: "INFRASTRUCTURE_ERROR",
		},
		{
			name:     "i/o timeout",
			err:      errors.New("i/o timeout"),
			expected: "INFRASTRUCTURE_ERROR",
		},
		{
			name:     "authentication failed",
			err:      errors.New("pq: password authentication failed"),
			expected: "CONFIGURATION_ERROR",
		},
		{
			name:     "access denied",
			err:      errors.New("access denied"),
			expected: "CONFIGURATION_ERROR",
		},
		{
			name:     "invalid data",
			err:      errors.New("invalid input syntax for type uuid"),
			expected: "APPLICATION_ERROR",
		},
		{
			name:     "not found",
			err:      errors.New("record not found"),
			expected: "APPLICATION_ERROR",
		},
		{
			name:     "unknown error",
			err:      errors.New("something went wrong"),
			expected: "UNKNOWN_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzeError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================
// ТЕСТЫ ДЛЯ getHint
// ============================================================

func TestGetHint(t *testing.T) {
	tests := []struct {
		name      string
		component string
		err       error
		expected  string
	}{
		{
			name:      "nil error",
			component: "Redis",
			err:       nil,
			expected:  "Check logs for details",
		},
		{
			name:      "connection refused",
			component: "Redis",
			err:       errors.New("connection refused"),
			expected:  "Redis is not running. Check CI services.",
		},
		{
			name:      "authentication failed",
			component: "PostgreSQL",
			err:       errors.New("authentication failed"),
			expected:  "Wrong credentials for PostgreSQL. Check .env.",
		},
		{
			name:      "timeout",
			component: "HTTP Server",
			err:       errors.New("timeout"),
			expected:  "HTTP Server is too slow. Increase timeout.",
		},
		{
			name:      "unknown error",
			component: "Redis",
			err:       errors.New("unknown"),
			expected:  "Check Redis logs.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getHint(tt.component, tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================
// ТЕСТЫ ДЛЯ checkComponent (с моками)
// ============================================================

func TestCheckComponent_Success(t *testing.T) {
	mockCheck := func() error {
		return nil
	}

	checkComponent(t, "TestComponent", 1*time.Second, mockCheck)
	t.Log("✅ checkComponent succeeded with mock")
}

func TestCheckComponent_Timeout(t *testing.T) {
	mockCheck := func() error {
		return errors.New("not ready")
	}

	defer func() {
		if r := recover(); r != nil {
			t.Log("✅ checkComponent correctly panicked on timeout")
		} else {
			t.Error("❌ checkComponent should panic on timeout")
		}
	}()

	checkComponent(t, "TestComponent", 100*time.Millisecond, mockCheck)
}

// ============================================================
// ТЕСТЫ ДЛЯ AssertReady
// ============================================================

func TestAssertReady_Success(t *testing.T) {
	mockCheck := func() error {
		return nil
	}

	AssertReady(t, "TestComponent", 1*time.Second, mockCheck)
	t.Log("✅ AssertReady succeeded with mock")
}

func TestAssertReady_Timeout(t *testing.T) {
	mockCheck := func() error {
		return errors.New("not ready")
	}

	defer func() {
		if r := recover(); r != nil {
			t.Log("✅ AssertReady correctly panicked on timeout")
		} else {
			t.Error("❌ AssertReady should panic on timeout")
		}
	}()

	AssertReady(t, "TestComponent", 100*time.Millisecond, mockCheck)
}

// ============================================================
// ТЕСТЫ ДЛЯ CheckAllDependencies
// ============================================================

func TestCheckAllDependencies_Signature(t *testing.T) {
	// ✅ Инициализируем Redis
	if err := cache.InitRedis("localhost:6379", "", 0); err != nil {
		logger.Warn("Redis init warning: %v", err)
	}

	// ✅ Инициализируем PostgreSQL
	if err := database.Init("postgres://postgres:1771@localhost:5432/subscriptions?sslmode=disable"); err != nil {
		logger.Warn("DB init warning: %v", err)
	}

	// ✅ Проверяем, что CheckAllDependencies работает
	CheckAllDependencies(t, 15*time.Second, "8087", "50051")
	t.Log("✅ CheckAllDependencies works")
}