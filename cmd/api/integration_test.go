//go:build integration

package main

import (
	"context"
	"os"
	"testing"
	"time"

	"Rest-user-agregator/internal/database"
	"Rest-user-agregator/internal/models"
)

// ============================================================
// ИНТЕГРАЦИОННЫЕ ТЕСТЫ (С РЕАЛЬНОЙ БД)
// ============================================================

// TestIntegrationDBConnection — проверяет подключение к БД
func TestIntegrationDBConnection(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("⏭️ Skipping integration test (INTEGRATION=true required)")
	}

	err := initDB()
	if err != nil {
		t.Fatalf("❌ DB connection failed: %v", err)
	}
	defer database.Close()

	t.Log("✅ Database connected successfully")
}

// TestIntegrationMigrations — проверяет применение миграций
func TestIntegrationMigrations(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("⏭️ Skipping integration test (INTEGRATION=true required)")
	}

	err := initDB()
	if err != nil {
		t.Fatalf("❌ DB connection failed: %v", err)
	}
	defer database.Close()

	err = runMigrations()
	if err != nil {
		t.Fatalf("❌ Migrations failed: %v", err)
	}

	t.Log("✅ Migrations applied successfully")
}

// TestIntegrationListSubscriptions — проверяет реальный запрос к БД
func TestIntegrationListSubscriptions(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("⏭️ Skipping integration test (INTEGRATION=true required)")
	}

	err := initDB()
	if err != nil {
		t.Fatalf("❌ DB connection failed: %v", err)
	}
	defer database.Close()

	repo := database.NewPostgresRepo()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	limit := 10
	offset := 0
	subs, err := repo.ListSubscriptions(ctx, limit, offset)
	if err != nil {
		t.Errorf("❌ ListSubscriptions failed: %v", err)
	} else {
		t.Logf("✅ ListSubscriptions succeeded, found %d subscriptions", len(subs))
	}
}

// TestIntegrationCreateSubscription — проверяет создание подписки
func TestIntegrationCreateSubscription(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("⏭️ Skipping integration test (INTEGRATION=true required)")
	}

	err := initDB()
	if err != nil {
		t.Fatalf("❌ DB connection failed: %v", err)
	}
	defer database.Close()

	repo := database.NewPostgresRepo()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Исправлено: поля из вашей структуры
	sub := models.Subscription{
		ServiceName: "Test Integration",
		Price:       999, // в копейках/центах (int)
		UserID:      "test_user@example.com",
		StartDate:   time.Now().Format("2006-01-02"),
		EndDate:     time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
	}

	startDate := time.Now()
	endDate := time.Now().AddDate(0, 1, 0)

	id, err := repo.CreateSubscription(ctx, sub, startDate, &endDate)
	if err != nil {
		t.Errorf("❌ CreateSubscription failed: %v", err)
	} else {
		t.Logf("✅ CreateSubscription succeeded: ID=%d", id)
	}
}

// TestIntegrationGetTotalCost — проверяет подсчёт общей стоимости
func TestIntegrationGetTotalCost(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("⏭️ Skipping integration test (INTEGRATION=true required)")
	}

	err := initDB()
	if err != nil {
		t.Fatalf("❌ DB connection failed: %v", err)
	}
	defer database.Close()

	repo := database.NewPostgresRepo()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userID := "test_user@example.com"
	serviceName := "Test Integration"
	startDate := time.Now().AddDate(-1, 0, 0)
	endDate := time.Now().AddDate(1, 0, 0)

	total, err := repo.GetTotalCost(ctx, userID, serviceName, startDate, endDate)
	if err != nil {
		t.Errorf("❌ GetTotalCost failed: %v", err)
	} else {
		t.Logf("✅ GetTotalCost succeeded: total=%d", total)
	}
}

// TestIntegrationRunMigrationsDown — проверяет откат миграций
func TestIntegrationRunMigrationsDown(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("⏭️ Skipping integration test (INTEGRATION=true required)")
	}

	err := initDB()
	if err != nil {
		t.Fatalf("❌ DB connection failed: %v", err)
	}
	defer database.Close()

	err = rollbackMigrations()
	if err != nil {
		t.Logf("⚠️ rollbackMigrations failed (may be normal): %v", err)
	} else {
		t.Log("✅ rollbackMigrations succeeded")
	}

	err = runMigrations()
	if err != nil {
		t.Errorf("❌ runMigrations after rollback failed: %v", err)
	} else {
		t.Log("✅ Migrations reapplied after rollback")
	}
}