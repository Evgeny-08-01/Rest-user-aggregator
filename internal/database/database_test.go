//go:build integration

package database

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"Rest-user-agregator/internal/models"
	"Rest-user-agregator/pkg/logger"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	if err := godotenv.Load("../../.env.test"); err != nil {
		logger.Warn(".env.test not found, using env vars")
	}
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		panic("DB_PATH not set, cannot run tests")
	}

	err := Init(dbPath)
	if err != nil {
		panic("Failed to init DB: " + err.Error())
	}
	defer func() {
		if err := Close(); err != nil {
			logger.Warn("Failed to close database: %v", err)
		}
	}()

	if err := DropTestTable(); err != nil {
		panic("Failed to drop table: " + err.Error())
	}
	if err := CreateTestTable(); err != nil {
		panic("Failed to create test table: " + err.Error())
	}
	if err := CleanTestTable(); err != nil {
		panic("Failed to clean table: " + err.Error())
	}

	code := m.Run()
	os.Exit(code)
}

func createTestUser(t *testing.T, userID string) {
	t.Cleanup(func() {
		if err := CleanTestTable(); err != nil {
			t.Logf("Cleanup failed: %v", err)
		}
	})
	repo := NewPostgresRepo()
	ctx := context.Background()

	user := models.User{
		ID:       userID,
		Email:    fmt.Sprintf("test_%s@mail.com", userID[:8]),
		Password: "hash",
		Role:     "user",
	}

	err := repo.CreateUser(ctx, user)
	if err != nil && !strings.Contains(err.Error(), "duplicate key") {
		t.Fatalf("CreateUser failed: %v", err)
	}
}

func TestCleanTestTable(t *testing.T) {
	t.Cleanup(func() {
		if err := CleanTestTable(); err != nil {
			t.Logf("Cleanup failed: %v", err)
		}
	})

	repo := NewPostgresRepo()
	ctx := context.Background()
	userID := "550e8400-e29b-41d4-a716-446655440000"
	createTestUser(t, userID)

	templateID, err := repo.CreateTemplate(ctx, "Netflix", 500)
	if err != nil {
		t.Fatalf("CreateTemplate failed: %v", err)
	}

	sub := models.Subscription{
		UserID:     userID,
		TemplateID: templateID,
		StartDate:  "01-2025",
	}
	startDate, _ := time.Parse("01-2006", sub.StartDate)
	id, err := repo.CreateSubscription(ctx, sub, startDate, nil)
	if err != nil {
		t.Fatalf("CreateSubscription failed: %v", err)
	}

	if err := CleanTestTable(); err != nil {
		t.Fatalf("CleanTestTable failed: %v", err)
	}

	saved, err := repo.GetSubscriptionByID(ctx, id)
	if err != nil {
		t.Fatalf("GetSubscriptionByID failed: %v", err)
	}
	if saved != nil {
		t.Error("Subscription still exists after clean")
	}
}

func TestCreateUser(t *testing.T) {
	if err := DropTestTable(); err != nil {
		t.Fatalf("DropTestTable failed: %v", err)
	}
	if err := CreateTestTable(); err != nil {
		t.Fatalf("CreateTestTable failed: %v", err)
	}
	t.Cleanup(func() {
		if err := CleanTestTable(); err != nil {
			t.Logf("Cleanup failed: %v", err)
		}
	})
	if err := CleanTestTable(); err != nil {
		t.Fatalf("CleanTestTable failed: %v", err)
	}
	repo := NewPostgresRepo()

	user := models.User{
		ID:       uuid.New().String(),
		Email:    fmt.Sprintf("test_%d@mail.com", time.Now().UnixNano()),
		Password: "hash123",
		Role:     "user",
	}
	err := repo.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	var count int
	err = repo.db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM users WHERE email = $1", user.Email).Scan(&count)
	if err != nil {
		t.Fatalf("Direct query failed: %v", err)
	}

	saved, err := repo.GetUserByEmail(context.Background(), user.Email)
	if err != nil {
		t.Fatalf("GetUserByEmail failed: %v", err)
	}
	if saved == nil {
		t.Fatal("User not found after creation")
	}
	if saved.Email != user.Email {
		t.Errorf("Expected %s, got %s", user.Email, saved.Email)
	}
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	t.Cleanup(func() {
		if err := CleanTestTable(); err != nil {
			t.Logf("Cleanup failed: %v", err)
		}
	})
	if err := CleanTestTable(); err != nil {
		t.Fatalf("CleanTestTable failed: %v", err)
	}

	repo := NewPostgresRepo()
	user, err := repo.GetUserByEmail(context.Background(), "notfound@mail.com")
	if err != nil {
		t.Fatalf("GetUserByEmail failed: %v", err)
	}
	if user != nil {
		t.Error("Expected nil, got user")
	}
}

func TestDeleteSubscriptionsByUserID(t *testing.T) {
	t.Cleanup(func() {
		if err := CleanTestTable(); err != nil {
			t.Logf("Cleanup failed: %v", err)
		}
	})
	if err := CleanTestTable(); err != nil {
		t.Fatalf("CleanTestTable failed: %v", err)
	}

	repo := NewPostgresRepo()
	ctx := context.Background()
	userID := "550e8400-e29b-41d4-a716-446655440000"
	createTestUser(t, userID)

	// Создаём ПЕРВЫЙ шаблон
	templateID1, err := repo.CreateTemplate(ctx, "TestDelete1", 100)
	if err != nil {
		t.Fatalf("CreateTemplate 1 failed: %v", err)
	}

	// Создаём ВТОРОЙ шаблон
	templateID2, err := repo.CreateTemplate(ctx, "TestDelete2", 200)
	if err != nil {
		t.Fatalf("CreateTemplate 2 failed: %v", err)
	}

	// Создаём ПЕРВУЮ подписку
	sub1 := models.Subscription{
		UserID:     userID,
		TemplateID: templateID1,
		StartDate:  "01-2025",
	}
	startDate, _ := time.Parse("01-2006", sub1.StartDate)
	_, err = repo.CreateSubscription(ctx, sub1, startDate, nil)
	if err != nil {
		t.Fatalf("CreateSubscription 1 failed: %v", err)
	}

	// Создаём ВТОРУЮ подписку (с другим шаблоном)
	sub2 := models.Subscription{
		UserID:     userID,
		TemplateID: templateID2,
		StartDate:  "02-2025",
	}
	startDate2, _ := time.Parse("01-2006", sub2.StartDate)
	_, err = repo.CreateSubscription(ctx, sub2, startDate2, nil)
	if err != nil {
		t.Fatalf("CreateSubscription 2 failed: %v", err)
	}

	// Проверяем, что создалось 2 подписки
	list, err := repo.ListSubscriptions(ctx, "", 10, 0)
	if err != nil {
		t.Fatalf("ListSubscriptions failed: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("Expected 2 subscriptions, got %d", len(list))
	}

	// Удаляем все подписки пользователя
	err = DeleteSubscriptionsByUserID(userID)
	if err != nil {
		t.Fatalf("DeleteSubscriptionsByUserID failed: %v", err)
	}

	// Проверяем, что подписок не осталось
	list, err = repo.ListSubscriptions(ctx, "", 10, 0)
	if err != nil {
		t.Fatalf("ListSubscriptions failed: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("Expected 0 subscriptions, got %d", len(list))
	}
}

func TestGetDB(t *testing.T) {
	t.Cleanup(func() {
		if err := CleanTestTable(); err != nil {
			t.Logf("Cleanup failed: %v", err)
		}
	})
	db := GetDB()
	if db == nil {
		t.Error("GetDB returned nil")
	}
}

func TestCreateUser_Success(t *testing.T) {
	t.Cleanup(func() {
		if err := CleanTestTable(); err != nil {
			t.Logf("Cleanup failed: %v", err)
		}
	})
	if err := CleanTestTable(); err != nil {
		t.Fatalf("Failed to clean table: %v", err)
	}

	repo := NewPostgresRepo()
	user := models.User{
		ID:       "550e8400-e29b-41d4-a716-446655440000",
		Email:    "testuser@example.com",
		Password: "hashed_password_here",
		Role:     "user",
	}

	err := repo.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	saved, err := repo.GetUserByEmail(context.Background(), user.Email)
	if err != nil {
		t.Fatalf("GetUserByEmail failed: %v", err)
	}
	if saved == nil {
		t.Fatal("User not found after creation")
	}
}

func TestCreateSubscription_Success(t *testing.T) {
	t.Cleanup(func() {
		if err := CleanTestTable(); err != nil {
			t.Logf("Cleanup failed: %v", err)
		}
	})
	if err := CleanTestTable(); err != nil {
		t.Fatalf("Failed to clean table: %v", err)
	}

	repo := NewPostgresRepo()
	ctx := context.Background()
	userID := "550e8400-e29b-41d4-a716-446655440000"
	createTestUser(t, userID)

	templateID, err := repo.CreateTemplate(ctx, "Netflix", 500)
	if err != nil {
		t.Fatalf("CreateTemplate failed: %v", err)
	}

	sub := models.Subscription{
		UserID:     userID,
		TemplateID: templateID,
		StartDate:  "07-2025",
	}
	startDate, _ := time.Parse("01-2006", sub.StartDate)

	id, err := repo.CreateSubscription(ctx, sub, startDate, nil)
	if err != nil {
		t.Fatalf("CreateSubscription failed: %v", err)
	}
	if id <= 0 {
		t.Errorf("Expected positive ID, got %d", id)
	}
}

func TestGetSubscriptionByID_Success(t *testing.T) {
	t.Cleanup(func() {
		if err := CleanTestTable(); err != nil {
			t.Logf("Cleanup failed: %v", err)
		}
	})
	if err := CleanTestTable(); err != nil {
		t.Fatalf("Failed to clean table: %v", err)
	}

	repo := NewPostgresRepo()
	ctx := context.Background()
	userID := "550e8400-e29b-41d4-a716-446655440000"
	createTestUser(t, userID)

	templateID, err := repo.CreateTemplate(ctx, "Netflix", 500)
	if err != nil {
		t.Fatalf("CreateTemplate failed: %v", err)
	}

	sub := models.Subscription{
		UserID:     userID,
		TemplateID: templateID,
		StartDate:  "01-2025",
	}
	startDate, _ := time.Parse("01-2006", sub.StartDate)
	id, err := repo.CreateSubscription(ctx, sub, startDate, nil)
	if err != nil {
		t.Fatalf("CreateSubscription failed: %v", err)
	}

	saved, err := repo.GetSubscriptionByID(ctx, id)
	if err != nil {
		t.Fatalf("GetSubscriptionByID failed: %v", err)
	}
	if saved == nil {
		t.Fatal("Subscription not found")
	}
	if saved.ServiceName != "Netflix" {
		t.Errorf("Expected 'Netflix', got '%s'", saved.ServiceName)
	}
	if saved.Price != 500 {
		t.Errorf("Expected 500, got %d", saved.Price)
	}
}

func TestUpdateSubscription_Success(t *testing.T) {
	t.Cleanup(func() {
		if err := CleanTestTable(); err != nil {
			t.Logf("Cleanup failed: %v", err)
		}
	})
	if err := CleanTestTable(); err != nil {
		t.Fatalf("Failed to clean table: %v", err)
	}

	repo := NewPostgresRepo()
	ctx := context.Background()
	userID := "550e8400-e29b-41d4-a716-446655440000"
	createTestUser(t, userID)

	templateID, err := repo.CreateTemplate(ctx, "Netflix", 500)
	if err != nil {
		t.Fatalf("CreateTemplate failed: %v", err)
	}

	sub := models.Subscription{
		UserID:     userID,
		TemplateID: templateID,
		StartDate:  "01-2025",
	}
	startDate, _ := time.Parse("01-2006", sub.StartDate)
	id, err := repo.CreateSubscription(ctx, sub, startDate, nil)
	if err != nil {
		t.Fatalf("CreateSubscription failed: %v", err)
	}

	err = repo.UpdateTemplate(ctx, templateID, "Netflix Premium", 700)
	if err != nil {
		t.Fatalf("UpdateTemplate failed: %v", err)
	}

	saved, err := repo.GetSubscriptionByID(ctx, id)
	if err != nil {
		t.Fatalf("GetSubscriptionByID failed: %v", err)
	}
	if saved.ServiceName != "Netflix Premium" {
		t.Errorf("Expected 'Netflix Premium', got '%s'", saved.ServiceName)
	}
	if saved.Price != 700 {
		t.Errorf("Expected 700, got %d", saved.Price)
	}
}

func TestDeleteSubscription_Success(t *testing.T) {
	t.Cleanup(func() {
		if err := CleanTestTable(); err != nil {
			t.Logf("Cleanup failed: %v", err)
		}
	})
	if err := CleanTestTable(); err != nil {
		t.Fatalf("Failed to clean table: %v", err)
	}

	repo := NewPostgresRepo()
	ctx := context.Background()
	userID := "550e8400-e29b-41d4-a716-446655440000"
	createTestUser(t, userID)

	templateID, err := repo.CreateTemplate(ctx, "ToDelete", 100)
	if err != nil {
		t.Fatalf("CreateTemplate failed: %v", err)
	}

	sub := models.Subscription{
		UserID:     userID,
		TemplateID: templateID,
		StartDate:  "01-2025",
	}
	startDate, _ := time.Parse("01-2006", sub.StartDate)
	id, err := repo.CreateSubscription(ctx, sub, startDate, nil)
	if err != nil {
		t.Fatalf("CreateSubscription failed: %v", err)
	}

	err = repo.DeleteSubscription(ctx, id)
	if err != nil {
		t.Fatalf("DeleteSubscription failed: %v", err)
	}

	saved, err := repo.GetSubscriptionByID(ctx, id)
	if err != nil {
		t.Fatalf("GetSubscriptionByID failed: %v", err)
	}
	if saved != nil {
		t.Error("Subscription still exists after deletion")
	}
}

func TestRunMigrations(t *testing.T) {
	t.Log("Starting TestRunMigrations")
	t.Log("Cleaning tables before migration test")
	if err := CleanTestTable(); err != nil {
		t.Fatalf("CleanTestTable failed: %v", err)
	}
	t.Log("Tables cleaned successfully")
	t.Log("Running migrations...")
	err := RunMigrations()
	if err != nil {
		t.Errorf("RunMigrations failed: %v", err)
		return
	}
	t.Log("Migrations applied successfully")
}

func TestRollbackMigrations(t *testing.T) {
	t.Log("Starting TestRollbackMigrations")
	t.Log("Cleaning tables before rollback test")
	if err := CleanTestTable(); err != nil {
		t.Fatalf("CleanTestTable failed: %v", err)
	}
	t.Log("Tables cleaned successfully")
	t.Log("Applying migrations before rollback...")
	if err := RunMigrations(); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}
	t.Log("Migrations applied, proceeding to rollback")
	t.Log("Rolling back migrations...")
	err := RollbackMigrations()
	if err != nil {
		t.Errorf("RollbackMigrations failed: %v", err)
		return
	}
	t.Log("Migrations rolled back successfully")
}
