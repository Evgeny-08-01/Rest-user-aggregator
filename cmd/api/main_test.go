package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

// ============================================================
// ЗАГРУЗКА .env.test (только для тестов)
// ============================================================
func loadTestEnv() error {
	// Загружаем ТОЛЬКО .env.test — если его нет, тест пропускается
	err := godotenv.Load("../../.env.test")
	if err != nil {
		return err
	}
	return nil
}

// ============================================================
// ТЕСТ ДЛЯ initDB()
// ============================================================
func TestInitDB(t *testing.T) {
    if err := loadTestEnv(); err != nil {
        t.Skip("Skipping test: .env.test not found")
    }

	// Сохраняем оригинальное значение DB_PATH
	originalDBPath := os.Getenv("DB_PATH")
	// Восстанавливаем оригинальное значение DB_PATH после теста
	defer os.Setenv("DB_PATH", originalDBPath)

	// Сценарий 1: неправильный путь → ошибка
	os.Setenv("DB_PATH", "invalid")
	err := initDB()
	if err == nil {
		fmt.Println("111")
			      t.Fatal("expected error for invalid DB_PATH, got nil")
	}
	fmt.Println("222")
	t.Log("Variant 1 passed: invalid path returned error")

    
	// Сценарий 2: правильный путь → если БД не запущена, тест падает
    os.Setenv("DB_PATH", originalDBPath)
    err = initDB()
    if err != nil {
		fmt.Println("333")
        t.Fatalf("database not running: %v", err)
    }
}

// ============================================================
// ТЕСТ ДЛЯ run()
// ============================================================
func TestRun(t *testing.T) {
	// Загружаем .env.test
	if err := loadTestEnv(); err != nil {
		t.Skip("Skipping test: .env.test not found")
	}

	// Сохраняем оригинальные значения
    origDB := os.Getenv("DB_PATH")
    origLogLevel := os.Getenv("LOG_LEVEL")
    origLogPath := os.Getenv("LOG_PATH")

    // Меняем переменные
    os.Setenv("DB_PATH", "invalid")
    os.Setenv("LOG_LEVEL", "info")
    os.Setenv("LOG_PATH", "./logs/test.log")

    // Откладываем восстановление (каждый на отдельной строке)
    defer os.Setenv("DB_PATH", origDB)
    defer os.Setenv("LOG_LEVEL", origLogLevel)
    defer os.Setenv("LOG_PATH", origLogPath)

	// Запускаем run() — она должна вернуть ошибку до запуска сервера
    err := run()
    if err == nil {
        t.Fatal("expected error for invalid DB_PATH, got nil")
    }

    t.Logf("run() returned expected error: %v", err)
}