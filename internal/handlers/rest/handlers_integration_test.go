//go:build integration

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"Rest-user-agregator/internal/authentication"
	"Rest-user-agregator/internal/cache"
	"Rest-user-agregator/internal/database"
	"Rest-user-agregator/internal/models"
	"Rest-user-agregator/internal/service"
	"Rest-user-agregator/pkg/logger"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	if err := godotenv.Load("../../../.env.test"); err != nil {
		log.Println("WARNING: .env.test not found, using env vars")
	}
	log.Println("LOG_LEVEL from env.test:", os.Getenv("LOG_LEVEL"))
	logger.Init(os.Getenv("LOG_PATH"), os.Getenv("LOG_LEVEL"))

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "postgres://postgres:mysecret@localhost:5432/subscriptions?sslmode=disable"
		log.Println("WARNING: DB_PATH not set, using default")
	} else {
		log.Println("INFO: Using DB_PATH from .env")
	}

	err := database.Init(dbPath)
	if err != nil {
		panic("Failed to init DB: " + err.Error())
	}
	if err := database.RunMigrations(); err != nil {
		panic("Failed to run migrations: " + err.Error())
	}
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	if err := cache.InitRedis(redisAddr, "", 0); err != nil {
		log.Println("WARNING: Redis not available, cache disabled")
	}
	defer database.Close()

	if err := database.CreateTestTable(); err != nil {
		panic("Failed to create table: " + err.Error())
	}
	if err := database.CleanTestTable(); err != nil {
		panic("Failed to clean table: " + err.Error())
	}
	os.Exit(m.Run())
}

func setupTestHandler() (*Handler, *TemplateHandler) {
    repo := database.NewPostgresRepo()
    templateSvc := service.NewTemplateService(repo)
    svc := service.NewSubscriptionService(repo, repo)
    return NewHandler(svc, nil), NewTemplateHandler(templateSvc)
}

// ============================================================
// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ДЛЯ ТЕСТОВ
// ============================================================

// addAdminContext — добавляет роль admin и стандартный userID 
// (используется в большинстве тестов)
func addAdminContext(req *http.Request) *http.Request {
    ctx := context.WithValue(req.Context(), authentication.RoleKey, "admin")
    ctx = context.WithValue(ctx, authentication.UserIDKey, "550e8400-e29b-41d4-a716-446655440000")
    return req.WithContext(ctx)
}

// addAdminContextWithUser — добавляет роль admin и произвольный userID
// (используется в тестах, где нужно создать подписки для разных пользователей)
func addAdminContextWithUser(req *http.Request, userID string) *http.Request {
    ctx := context.WithValue(req.Context(), authentication.RoleKey, "admin")
    ctx = context.WithValue(ctx, authentication.UserIDKey, userID)
    return req.WithContext(ctx)
}

// addTestContext — добавляет произвольные userID и role в контекст
// (используется для имитации авторизованного пользователя)
func addTestContext(req *http.Request, userID, role string) *http.Request {
    ctx := req.Context()
    ctx = context.WithValue(ctx, authentication.UserIDKey, userID)
    ctx = context.WithValue(ctx, authentication.RoleKey, role)
    return req.WithContext(ctx)
}


func createTestUser(t *testing.T, userID string) string {
	db := database.GetDB()
	if db == nil {
		t.Fatal("Database not initialized")
	}
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE id = $1", userID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check user existence: %v", err)
	}
	if count == 0 {
		hashedPassword := "$2a$10$8dXxWmxnKk59pdXdy44l/eb4g1PnaFenHN3B.4lLR4bRy4ZL4xjK."
		_, err = db.Exec(
			"INSERT INTO users (id, email, password_hash, role, created_at) VALUES ($1, $2, $3, $4, NOW())",
			userID,
			"test_"+userID+"@example.com",
			hashedPassword,
			"user",
		)
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}
		t.Logf("Test user created: %s", userID)
	} else {
		t.Logf("Test user already exists: %s", userID)
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "test-secret-key"
	}
	claims := authentication.Claims{
		UserID: userID,
		Email:  "test_" + userID + "@example.com",
		Role:   "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}
	t.Logf("JWT token generated for user: %s", userID)
	return tokenString
}

func getRedisKeys(pattern string) ([]string, error) {
	client := cache.GetClient()
	if client == nil {
		return []string{}, nil
	}
	return client.Keys(context.Background(), pattern).Result()
}

// ============================================================
// 1. ТЕСТ: СОЗДАНИЕ ПОДПИСКИ
// ============================================================
func TestCreateSubscriptionHandler(t *testing.T) {
	t.Cleanup(func() {
		if err := database.CleanTestTable(); err != nil {
			t.Logf("Cleanup failed: %v", err)
		}
	})
	if err := database.CleanTestTable(); err != nil {
		t.Fatalf("Cannot continue test without clean table")
	}
	t.Log("Table cleaned successfully")

	userID := "550e8400-e29b-41d4-a716-446655440000"
	createTestUser(t, userID)
  handler, templateHandler := setupTestHandler()

	// 1. Создаём шаблон через API
	templateBody := `{"service_name":"TestTemplate","price":100}`
	req := httptest.NewRequest("POST", "/api/admin/templates", bytes.NewReader([]byte(templateBody)))
	req = addAdminContext(req)
	w := httptest.NewRecorder()
	templateHandler.CreateTemplateHandler(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create template: %d", w.Code)
	}
	var templateResp map[string]int
	if err := json.NewDecoder(w.Body).Decode(&templateResp); err != nil {
		t.Fatalf("Failed to decode template response: %v", err)
	}
	templateID := templateResp["id"]

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			"success",
			fmt.Sprintf(`{"template_id":%d,"start_date":"09-2028","end_date":"12-2028"}`, templateID),
			http.StatusCreated,
		},
		{
			"empty template_id",
			`{"template_id":0,"start_date":"07-2025"}`,
			http.StatusBadRequest,
		},
		{
			"invalid date",
			fmt.Sprintf(`{"template_id":%d,"start_date":"2025-07"}`, templateID),
			http.StatusBadRequest,
		},
		{
			"invalid JSON",
			`{"template_id":}`,
			http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(tt.body)))
			req = addAdminContext(req)
			w := httptest.NewRecorder()
			handler.CreateSubscriptionHandler(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("got %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

// ============================================================
// 2. ТЕСТ: ПОЛУЧЕНИЕ ПОДПИСКИ ПО ID
// ============================================================
func TestGetSubscriptionHandler(t *testing.T) {
	t.Cleanup(func() {
		if err := database.CleanTestTable(); err != nil {
			t.Logf("Cleanup failed: %v", err)
		}
	})
	if err := database.CleanTestTable(); err != nil {
		t.Fatalf("Cannot continue test without clean table")
	}
	t.Log("Table cleaned successfully")

	userID := "550e8400-e29b-41d4-a716-446655440000"
	createTestUser(t, userID)
    handler, templateHandler := setupTestHandler()

	// 1. Создаём шаблон
	templateBody := `{"service_name":"TestGet","price":100}`
	req := httptest.NewRequest("POST", "/api/admin/templates", bytes.NewReader([]byte(templateBody)))
	req = addAdminContext(req)
	w := httptest.NewRecorder()
	templateHandler.CreateTemplateHandler(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create template: %d", w.Code)
	}
	var templateResp map[string]int
	if err := json.NewDecoder(w.Body).Decode(&templateResp); err != nil {
		t.Fatalf("Failed to decode template response: %v", err)
	}
	templateID := templateResp["id"]

	// 2. Создаём подписку
	createBody := fmt.Sprintf(`{"template_id":%d,"start_date":"09-2026"}`, templateID)
	req = httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(createBody)))
	req = addAdminContext(req)
	w = httptest.NewRecorder()
	handler.CreateSubscriptionHandler(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create subscription: %d", w.Code)
	}
	var resp map[string]int
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	id := resp["id"]

	// 3. Получаем подписку
	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/subscriptions/{id}", http.NoBody)
		req = addAdminContext(req)
		req.SetPathValue("id", strconv.Itoa(id))
		w := httptest.NewRecorder()
		handler.GetSubscriptionHandler(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("got %d, want %d", w.Code, http.StatusOK)
		}
		var sub models.Subscription
		if err := json.NewDecoder(w.Body).Decode(&sub); err != nil {
			t.Fatalf("Failed to decode subscription: %v", err)
		}
		if sub.ServiceName != "TestGet" {
			t.Errorf("expected 'TestGet', got '%s'", sub.ServiceName)
		}
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/subscriptions/{id}", http.NoBody)
		req = addAdminContext(req)
		req.SetPathValue("id", "99999")
		w := httptest.NewRecorder()
		handler.GetSubscriptionHandler(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("got %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/subscriptions/{id}", http.NoBody)
		req = addAdminContext(req)
		req.SetPathValue("id", "abc")
		w := httptest.NewRecorder()
		handler.GetSubscriptionHandler(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

// ============================================================
// 3. ТЕСТ: ОБНОВЛЕНИЕ ПОДПИСКИ
// ============================================================
func TestUpdateSubscriptionHandler(t *testing.T) {
    t.Cleanup(func() {
        if err := database.CleanTestTable(); err != nil {
            t.Logf("Cleanup failed: %v", err)
        }
    })
    if err := database.CleanTestTable(); err != nil {
        t.Fatalf("Cannot continue test without clean table")
    }
    t.Log("Table cleaned successfully")

    userID := "550e8400-e29b-41d4-a716-446655440000"
    createTestUser(t, userID)
    handler, templateHandler := setupTestHandler()

    templateBody := `{"service_name":"BeforeUpdate","price":100}`
    req := httptest.NewRequest("POST", "/api/admin/templates", bytes.NewReader([]byte(templateBody)))
    req = addAdminContext(req)
    w := httptest.NewRecorder()
    templateHandler.CreateTemplateHandler(w, req)
    if w.Code != http.StatusCreated {
        t.Fatalf("Failed to create template: %d", w.Code)
    }
    var templateResp map[string]int
    if err := json.NewDecoder(w.Body).Decode(&templateResp); err != nil {
        t.Fatalf("Failed to decode template response: %v", err)
    }
    templateID := templateResp["id"]

    createBody := fmt.Sprintf(`{"template_id":%d,"start_date":"09-2026"}`, templateID)
    req = httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(createBody)))
    req = addAdminContext(req)
    w = httptest.NewRecorder()
    handler.CreateSubscriptionHandler(w, req)
    if w.Code != http.StatusCreated {
        t.Fatalf("Failed to create subscription: %d", w.Code)
    }
    var resp map[string]int
    if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
        t.Fatalf("Failed to decode response: %v", err)
    }
    id := resp["id"]

    t.Run("success", func(t *testing.T) {
    // ✅ Добавлен user_id
    updateBody := fmt.Sprintf(`{"template_id":%d,"service_name":"BeforeUpdate","user_id":"%s","start_date":"08-2027","end_date":"12-2027"}`, templateID, userID)
    req := httptest.NewRequest("PUT", "/api/subscriptions/{id}", bytes.NewReader([]byte(updateBody)))
    req = addAdminContext(req)
    req.SetPathValue("id", strconv.Itoa(id))
    w := httptest.NewRecorder()
    handler.UpdateSubscriptionHandler(w, req)
    if w.Code != http.StatusOK {
        t.Errorf("got %d, want %d", w.Code, http.StatusOK)
    }
})

    t.Run("invalid id", func(t *testing.T) {
        req := httptest.NewRequest("PUT", "/api/subscriptions/{id}", bytes.NewReader([]byte(`{"start_date":"07-2025"}`)))
        req = addAdminContext(req)
        req.SetPathValue("id", "abc")
        w := httptest.NewRecorder()
        handler.UpdateSubscriptionHandler(w, req)
        if w.Code != http.StatusBadRequest {
            t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
        }
    })
}

// ============================================================
// 4. ТЕСТ: УДАЛЕНИЕ ПОДПИСКИ
// ============================================================
func TestDeleteSubscriptionHandler(t *testing.T) {
	t.Cleanup(func() {
		if err := database.CleanTestTable(); err != nil {
			t.Logf("Cleanup failed: %v", err)
		}
	})
	if err := database.CleanTestTable(); err != nil {
		t.Fatalf("Cannot continue test without clean table")
	}
	t.Log("Table cleaned successfully")

	userID := "550e8400-e29b-41d4-a716-446655440000"
	createTestUser(t, userID)
handler, templateHandler := setupTestHandler()

	templateBody := `{"service_name":"ToDelete","price":100}`
	req := httptest.NewRequest("POST", "/api/admin/templates", bytes.NewReader([]byte(templateBody)))
	req = addAdminContext(req)
	w := httptest.NewRecorder()
	templateHandler.CreateTemplateHandler(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create template: %d", w.Code)
	}
	var templateResp map[string]int
	if err := json.NewDecoder(w.Body).Decode(&templateResp); err != nil {
		t.Fatalf("Failed to decode template response: %v", err)
	}
	templateID := templateResp["id"]

	createBody := fmt.Sprintf(`{"template_id":%d,"start_date":"09-2026"}`, templateID)
	req = httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(createBody)))
	req = addAdminContext(req)
	w = httptest.NewRecorder()
	handler.CreateSubscriptionHandler(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create subscription: %d", w.Code)
	}
	var resp map[string]int
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	id := resp["id"]

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/subscriptions/{id}", http.NoBody)
		req = addAdminContext(req)
		req.SetPathValue("id", strconv.Itoa(id))
		w := httptest.NewRecorder()
		handler.DeleteSubscriptionHandler(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("got %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/subscriptions/{id}", http.NoBody)
		req = addAdminContext(req)
		req.SetPathValue("id", "abc")
		w := httptest.NewRecorder()
		handler.DeleteSubscriptionHandler(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

// ============================================================
// 5. ТЕСТ: СПИСОК ПОДПИСОК
// ============================================================
func TestListSubscriptionsHandler(t *testing.T) {
	t.Cleanup(func() {
		if err := database.CleanTestTable(); err != nil {
			t.Logf("Cleanup failed: %v", err)
		}
	})
	if err := database.CleanTestTable(); err != nil {
		t.Fatalf("Cannot continue test without clean table")
	}
	t.Log("Table cleaned successfully")

	userID := "550e8400-e29b-41d4-a716-446655440000"
	createTestUser(t, userID)
handler, templateHandler := setupTestHandler()

	templateBody := `{"service_name":"ListTest","price":100}`
	req := httptest.NewRequest("POST", "/api/admin/templates", bytes.NewReader([]byte(templateBody)))
	req = addAdminContext(req)
	w := httptest.NewRecorder()
	templateHandler.CreateTemplateHandler(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create template: %d", w.Code)
	}
	var templateResp map[string]int
	if err := json.NewDecoder(w.Body).Decode(&templateResp); err != nil {
		t.Fatalf("Failed to decode template response: %v", err)
	}


// Шаблон 1
templateBody1 := `{"service_name":"ListTest1","price":100}`
req1 := httptest.NewRequest("POST", "/api/admin/templates", bytes.NewReader([]byte(templateBody1)))
req1 = addAdminContext(req1)
w1 := httptest.NewRecorder()
templateHandler.CreateTemplateHandler(w1, req1)
if w1.Code != http.StatusCreated {
    t.Fatalf("Failed to create template 1: %d", w1.Code)
}
var templateResp1 map[string]int
json.NewDecoder(w1.Body).Decode(&templateResp1)
templateID1 := templateResp1["id"]

createBody1 := fmt.Sprintf(`{"template_id":%d,"start_date":"09-2026"}`, templateID1)
reqCreate1 := httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(createBody1)))
reqCreate1 = addAdminContext(reqCreate1)
wCreate1 := httptest.NewRecorder()
handler.CreateSubscriptionHandler(wCreate1, reqCreate1)
if wCreate1.Code != http.StatusCreated {
    t.Fatalf("Failed to create subscription 1: %d", wCreate1.Code)
}

// Шаблон 2
templateBody2 := `{"service_name":"ListTest2","price":200}`
req2 := httptest.NewRequest("POST", "/api/admin/templates", bytes.NewReader([]byte(templateBody2)))
req2 = addAdminContext(req2)
w2 := httptest.NewRecorder()
templateHandler.CreateTemplateHandler(w2, req2)
if w2.Code != http.StatusCreated {
    t.Fatalf("Failed to create template 2: %d", w2.Code)
}
var templateResp2 map[string]int
json.NewDecoder(w2.Body).Decode(&templateResp2)
templateID2 := templateResp2["id"]

createBody2 := fmt.Sprintf(`{"template_id":%d,"start_date":"10-2026"}`, templateID2)
reqCreate2 := httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(createBody2)))
reqCreate2 = addAdminContext(reqCreate2)
wCreate2 := httptest.NewRecorder()
handler.CreateSubscriptionHandler(wCreate2, reqCreate2)
if wCreate2.Code != http.StatusCreated {
    t.Fatalf("Failed to create subscription 2: %d", wCreate2.Code)
}

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/subscriptions", http.NoBody)
		req = addAdminContext(req)
		w := httptest.NewRecorder()
		handler.ListSubscriptionsHandler(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("got %d, want %d", w.Code, http.StatusOK)
		}
		var list []models.Subscription
		if err := json.NewDecoder(w.Body).Decode(&list); err != nil {
			t.Fatalf("Failed to decode list: %v", err)
		}
		if len(list) != 2 {
    t.Errorf("expected 2 subscriptions, got %d", len(list))
}
	})
}

// ============================================================
// 6. ТЕСТ: TOTAL-COST (сокращённая версия)
// ============================================================
func TestGetTotalCostHandler(t *testing.T) {
	t.Cleanup(func() {
		if err := database.CleanTestTable(); err != nil {
			t.Logf("Cleanup failed: %v", err)
		}
	})
	if err := database.CleanTestTable(); err != nil {
		t.Fatalf("Failed to clean table: %v", err)
	}
	t.Log("Table cleaned successfully")

	userID := "550e8400-e29b-41d4-a716-446655440000"
	createTestUser(t, userID)
handler, templateHandler := setupTestHandler()

	// Создаём шаблон
	templateBody := `{"service_name":"CostTest","price":100}`
	req := httptest.NewRequest("POST", "/api/admin/templates", bytes.NewReader([]byte(templateBody)))
	req = addAdminContext(req)
	w := httptest.NewRecorder()
	templateHandler.CreateTemplateHandler(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create template: %d", w.Code)
	}
	var templateResp map[string]int
	if err := json.NewDecoder(w.Body).Decode(&templateResp); err != nil {
		t.Fatalf("Failed to decode template response: %v", err)
	}
	templateID := templateResp["id"]

	// Создаём подписку
	createBody := fmt.Sprintf(`{"template_id":%d,"start_date":"09-2026"}`, templateID)
	req = httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(createBody)))
	req = addAdminContext(req)
	w = httptest.NewRecorder()
	handler.CreateSubscriptionHandler(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create subscription: %d", w.Code)
	}

	// Проверяем total-cost
	t.Run("success", func(t *testing.T) {
		url := "/api/subscriptions/total-cost?start_date=01-2025&end_date=12-2026"
		req := httptest.NewRequest("GET", url, http.NoBody)
		req = addAdminContext(req)
		w := httptest.NewRecorder()
		handler.GetTotalCostHandler(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("got %d, want 200", w.Code)
			return
		}
		var resp map[string]int
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp["total"] <= 0 {
			t.Errorf("expected positive total, got %d", resp["total"])
		}
	})
}

// ============================================================
// 7-8. ТЕСТЫ: ПОЛЬЗОВАТЕЛЬ И АДМИН (сокращённые)
// ============================================================
func TestListSubscriptionsHandler_UserSeesOwn(t *testing.T) {
	t.Cleanup(func() {
		if err := database.CleanTestTable(); err != nil {
			t.Logf("Cleanup failed: %v", err)
		}
	})
	if err := database.CleanTestTable(); err != nil {
		t.Fatalf("Failed to clean table: %v", err)
	}
	t.Log("Table cleaned successfully")

	userID1 := "550e8400-e29b-41d4-a716-446655440000"
	userID2 := "550e8400-e29b-41d4-a716-446655440001"
	createTestUser(t, userID1)
	createTestUser(t, userID2)
handler, templateHandler := setupTestHandler()

	templateBody := `{"service_name":"UserTest","price":100}`
	req := httptest.NewRequest("POST", "/api/admin/templates", bytes.NewReader([]byte(templateBody)))
	req = addAdminContextWithUser(req, userID1)
	w := httptest.NewRecorder()
	templateHandler.CreateTemplateHandler(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create template: %d", w.Code)
	}
	var templateResp map[string]int
	if err := json.NewDecoder(w.Body).Decode(&templateResp); err != nil {
		t.Fatalf("Failed to decode template response: %v", err)
	}
	templateID := templateResp["id"]

	for _, uid := range []string{userID1, userID2} {
		createBody := fmt.Sprintf(`{"template_id":%d,"start_date":"09-2026"}`, templateID)
		req := httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(createBody)))
		req = addAdminContextWithUser(req, uid)
		w := httptest.NewRecorder()
		handler.CreateSubscriptionHandler(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("failed to create subscription for %s: %d", uid, w.Code)
		}
	}

	token1 := createTestUser(t, userID1)
	req = httptest.NewRequest("GET", "/api/subscriptions", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token1)
	req = addTestContext(req, userID1, "user")
	w = httptest.NewRecorder()
	handler.ListSubscriptionsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var list []models.Subscription
	if err := json.NewDecoder(w.Body).Decode(&list); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 subscription for user, got %d", len(list))
	}
}

func TestListSubscriptionsHandler_AdminSeesAll(t *testing.T) {
    t.Cleanup(func() {
        if err := database.CleanTestTable(); err != nil {
            t.Logf("Cleanup failed: %v", err)
        }
    })
    if err := database.CleanTestTable(); err != nil {
        t.Fatalf("Failed to clean table: %v", err)
    }
    t.Log("Table cleaned successfully")

    userID1 := "550e8400-e29b-41d4-a716-446655440000"
    userID2 := "550e8400-e29b-41d4-a716-446655440001"
    createTestUser(t, userID1)
    createTestUser(t, userID2)
    handler, templateHandler := setupTestHandler()

    templateBody := `{"service_name":"AdminTest","price":100}`
    req := httptest.NewRequest("POST", "/api/admin/templates", bytes.NewReader([]byte(templateBody)))
    req = addAdminContext(req)
    w := httptest.NewRecorder()
    templateHandler.CreateTemplateHandler(w, req)
    if w.Code != http.StatusCreated {
        t.Fatalf("Failed to create template: %d", w.Code)
    }
    var templateResp map[string]int
    if err := json.NewDecoder(w.Body).Decode(&templateResp); err != nil {
        t.Fatalf("Failed to decode template response: %v", err)
    }
    templateID := templateResp["id"]

    // ✅ Исправлено: используем addAdminContextWithUser для создания подписок
    for _, uid := range []string{userID1, userID2} {
        createBody := fmt.Sprintf(`{"template_id":%d,"start_date":"09-2026"}`, templateID)
        req := httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(createBody)))
        req = addAdminContextWithUser(req, uid)  // ← передаём правильный userID
        w := httptest.NewRecorder()
        handler.CreateSubscriptionHandler(w, req)
        if w.Code != http.StatusCreated {
            t.Fatalf("failed to create subscription for %s: %d", uid, w.Code)
        }
    }

    // ✅ Исправлено: админ получает список, используя addAdminContext
    req = httptest.NewRequest("GET", "/api/subscriptions", http.NoBody)
    req = addAdminContext(req)  // ← админ видит все подписки
    w = httptest.NewRecorder()
    handler.ListSubscriptionsHandler(w, req)

    if w.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", w.Code)
    }
    var list []models.Subscription
    if err := json.NewDecoder(w.Body).Decode(&list); err != nil {
        t.Fatalf("failed to decode response: %v", err)
    }
    if len(list) != 2 {
        t.Errorf("expected 2 subscriptions for admin, got %d", len(list))
    }
}