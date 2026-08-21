//go:build unit

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"Rest-user-agregator/internal/models"
	"Rest-user-agregator/internal/repository"
	"Rest-user-agregator/internal/service"
    "Rest-user-agregator/internal/authentication"
)

// ============================================================
// ЮНИТ-ТЕСТЫ С МОКАМИ (БЕЗ БД)
// ============================================================
func TestCreate_Mock(t *testing.T) {
    // 1. Создаём мок-репозиторий
    mockRepo := &repository.MockSubRepo{}

    // 2. Настраиваем поведение мока для успешного создания
    mockRepo.CreateMock = func(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) (int, error) {
        // Здесь мы имитируем успешное создание подписки
        // Возвращаем id=1 и nil (без ошибки)
        return 1, nil
    }

    // 3. Создаём хендлер с моком (вместо реальной БД)
    svc := service.NewSubscriptionService(mockRepo)
    handler := NewHandler(svc, nil)

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{"success", `{"service_name":"Test","price":100,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"07-2025"}`, http.StatusCreated},
		{"empty service_name", `{"service_name":"","price":100,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"07-2025"}`, http.StatusBadRequest},
		{"negative price", `{"service_name":"Test","price":-10,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"07-2025"}`, http.StatusBadRequest},
		{"empty user_id", `{"service_name":"Test","price":100,"user_id":"","start_date":"07-2025"}`, http.StatusBadRequest},
		{"invalid date", `{"service_name":"Test","price":100,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"2025-07"}`, http.StatusBadRequest},
		{"invalid JSON", `{"service_name":}`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(tt.body)))
           ctx := context.WithValue(req.Context(), authentication.RoleKey, "admin")
           ctx = context.WithValue(ctx, authentication.UserIDKey, "550e8400-e29b-41d4-a716-446655440000")
           req = req.WithContext(ctx)
			w := httptest.NewRecorder()
			handler.CreateSubscriptionHandler(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("got %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

//МОК ДЛЯ ФУНКЦИИ GET  (получение по ID) 
func TestGet_Mock(t *testing.T) {
    mockRepo := &repository.MockSubRepo{}
    mockRepo.GetSubByIDMock = func(ctx context.Context, id int) (*models.Subscription, error) {
        if id == 1 {  // ← только для id=1 возвращаем подписку
            return &models.Subscription{
                ID:          1,
                ServiceName: "Test Service",
                Price:       100,
                UserID:      "550e8400-e29b-41d4-a716-446655440000",
                StartDate:   "07-2025",
                EndDate:     "",
            }, nil
        }
        return nil, nil  // ← для всех остальных — not found
    }
    svc := service.NewSubscriptionService(mockRepo)
    handler := NewHandler(svc, nil)

    tests := []struct {
        name       string
        id         string
        wantStatus int
    }{
        {"success", "1", http.StatusOK},
        {"not found", "99999", http.StatusNotFound},
        {"invalid id", "abc", http.StatusBadRequest},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("GET", "/api/subscriptions/"+tt.id, nil)
            ctx := context.WithValue(req.Context(), authentication.RoleKey, "admin")
            req = req.WithContext(ctx)
            req.SetPathValue("id", tt.id)
            w := httptest.NewRecorder()
            handler.GetSubscriptionHandler(w, req)
            if w.Code != tt.wantStatus {
                t.Errorf("got %d, want %d", w.Code, tt.wantStatus)
            }
        })
    }
}

//МОК для Update (обновление)
func TestUpdate_Mock(t *testing.T) { 
    mockRepo := &repository.MockSubRepo{}
    mockRepo.UpdateSubMock = func(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) error {
        return nil // всегда успех
    }
    svc := service.NewSubscriptionService(mockRepo)
    handler := NewHandler(svc, nil)

    tests := []struct {
        name       string
        id         string
        body       string
        wantStatus int
    }{
        {"success", "1", `{"service_name":"Updated","price":200,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"08-2025","end_date":"12-2025"}`, http.StatusOK},
        {"invalid id", "abc", `{"service_name":"Test","price":100,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"07-2025"}`, http.StatusBadRequest},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("PUT", "/api/subscriptions/"+tt.id, bytes.NewReader([]byte(tt.body)))
            req.SetPathValue("id", tt.id)
            ctx := context.WithValue(req.Context(), authentication.RoleKey, "admin")
            req = req.WithContext(ctx)
            w := httptest.NewRecorder()
            handler.UpdateSubscriptionHandler(w, req)
            if w.Code != tt.wantStatus {
                t.Errorf("got %d, want %d", w.Code, tt.wantStatus)
            }
        })
    }
}

// МОК для Delete (удаление)
func TestDelete_Mock(t *testing.T) {  
    mockRepo := &repository.MockSubRepo{}
    mockRepo.DeleteSubMock = func(ctx context.Context, id int) error {
        return nil // всегда успех
    }
    svc := service.NewSubscriptionService(mockRepo)
    handler := NewHandler(svc, nil)

    tests := []struct {
        name       string
        id         string
        wantStatus int
    }{
        {"success", "1", http.StatusOK},
        {"invalid id", "abc", http.StatusBadRequest},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("DELETE", "/api/subscriptions/"+tt.id, nil)
            req.SetPathValue("id", tt.id)
            ctx := context.WithValue(req.Context(), authentication.RoleKey, "admin")
            req = req.WithContext(ctx)
            w := httptest.NewRecorder()
            handler.DeleteSubscriptionHandler(w, req)
            if w.Code != tt.wantStatus {
                t.Errorf("got %d, want %d", w.Code, tt.wantStatus)
            }
        })
    }
}

//МОК для List (список)
func TestList_Mock(t *testing.T) { 
    mockRepo := &repository.MockSubRepo{}
    mockRepo.ListSubMock = func(ctx context.Context, limit, offset int) ([]models.Subscription, error) {
        return []models.Subscription{
            {ID: 1, ServiceName: "Test1", Price: 100, UserID: "550e8400-e29b-41d4-a716-446655440000", StartDate: "07-2025"},
            {ID: 2, ServiceName: "Test2", Price: 200, UserID: "550e8400-e29b-41d4-a716-446655440000", StartDate: "08-2025"},
        }, nil
    }
    svc := service.NewSubscriptionService(mockRepo)
    handler := NewHandler(svc, nil)

    req := httptest.NewRequest("GET", "/api/subscriptions?limit=10&offset=0", nil)
    ctx := context.WithValue(req.Context(), authentication.RoleKey, "admin")
     req = req.WithContext(ctx)
    w := httptest.NewRecorder()
    handler.ListSubscriptionsHandler(w, req)

    if w.Code != http.StatusOK {
        t.Errorf("got %d, want %d", w.Code, http.StatusOK)
    }

    var list []models.Subscription
    json.NewDecoder(w.Body).Decode(&list)
    if len(list) != 2 {
        t.Errorf("expected 2, got %d", len(list))
    }
}
//МОК TotalCost (суммарная стоимость)
func TestTotalCost_Mock(t *testing.T) { 
    mockRepo := &repository.MockSubRepo{}
    mockRepo.GetTotalCostMock = func(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error) {
        return 1000, nil
    }

    svc := service.NewSubscriptionService(mockRepo)
    handler := NewHandler(svc, nil)

    req := httptest.NewRequest("GET", "/api/subscriptions/total-cost?user_id=test&start_date=01-2025&end_date=12-2025", nil)
    ctx := context.WithValue(req.Context(), authentication.RoleKey, "admin")
    req = req.WithContext(ctx)
    w := httptest.NewRecorder()
    handler.GetTotalCostHandler(w, req)

    if w.Code != http.StatusOK {
        t.Errorf("got %d, want %d", w.Code, http.StatusOK)
    }

    var resp map[string]int
    json.NewDecoder(w.Body).Decode(&resp)
    if resp["total"] != 1000 {
        t.Errorf("expected 1000, got %d", resp["total"])
    }
}
// ========== ВСПОМОГАТЕЛЬНЫЕ ТЕСТЫ (НЕ ИЗМЕНИЛИСЬ) ==========

func TestValidateSubscription(t *testing.T) {
    tests := []struct {
        name    string
        sub     models.Subscription
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid subscription",
            sub: models.Subscription{
                ServiceName: "Yandex Plus",
                Price:       400,
                UserID:      "550e8400-e29b-41d4-a716-446655440000",
                StartDate:   "07-2025",
                EndDate:     "",
            },
            wantErr: false,
        },
        {
            name: "empty service_name",
            sub: models.Subscription{
                ServiceName: "",
                Price:       400,
                UserID:      "550e8400-e29b-41d4-a716-446655440000",
                StartDate:   "07-2025",
                EndDate:     "",
            },
            wantErr: true,
            errMsg:  "service_name is required",
        },
        {
            name: "negative price",
            sub: models.Subscription{
                ServiceName: "Test",
                Price:       -100,
                UserID:      "550e8400-e29b-41d4-a716-446655440000",
                StartDate:   "07-2025",
                EndDate:     "",
            },
            wantErr: true,
            errMsg:  "price cant be negative value",
        },
        {
            name: "empty user_id",
            sub: models.Subscription{
                ServiceName: "Test",
                Price:       100,
                UserID:      "",
                StartDate:   "07-2025",
                EndDate:     "",
            },
            wantErr: true,
            errMsg:  "user_id is required",
        },
        {
            name: "invalid UUID format",
            sub: models.Subscription{
                ServiceName: "Test",
                Price:       100,
                UserID:      "not-a-uuid",
                StartDate:   "07-2025",
                EndDate:     "",
            },
            wantErr: true,
            errMsg:  "user_id: not valid-UUID",
        },
        {
            name: "invalid start_date format",
            sub: models.Subscription{
                ServiceName: "Test",
                Price:       100,
                UserID:      "550e8400-e29b-41d4-a716-446655440000",
                StartDate:   "2025-07",
                EndDate:     "",
            },
            wantErr: true,
            errMsg:  "start_date must be in format MM-YYYY",
        },
        {
            name: "invalid end_date format",
            sub: models.Subscription{
                ServiceName: "Test",
                Price:       100,
                UserID:      "550e8400-e29b-41d4-a716-446655440000",
                StartDate:   "07-2025",
                EndDate:     "2025-12",
            },
            wantErr: true,
            errMsg:  "end_date must be in format MM-YYYY",
        },
        {
            name: "valid with end_date",
            sub: models.Subscription{
                ServiceName: "Test",
                Price:       100,
                UserID:      "550e8400-e29b-41d4-a716-446655440000",
                StartDate:   "07-2025",
                EndDate:     "12-2025",
            },
            wantErr: false,
        },
        {
            name: "zero price",
            sub: models.Subscription{
                ServiceName: "Test",
                Price:       0,
                UserID:      "550e8400-e29b-41d4-a716-446655440000",
                StartDate:   "07-2025",
                EndDate:     "",
            },
            wantErr: false,
        },
        {
            name: "valid UUID with spaces",
            sub: models.Subscription{
                ServiceName: "Test",
                Price:       100,
                UserID:      "550e8400-e29b-41d4-a716-446655440000",
                StartDate:   "07-2025",
                EndDate:     "",
            },
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateSubscription(tt.sub)
            if (err != nil) != tt.wantErr {
                t.Errorf("validateSubscription() error = %v, wantErr %v", err, tt.wantErr)
            }
            if tt.wantErr && err != nil && err.Error() != tt.errMsg {
                t.Errorf("expected error message '%s', got '%s'", tt.errMsg, err.Error())
            }
        })
    }
}

func TestParseJSON(t *testing.T) { 
    t.Run("valid JSON", func(t *testing.T) {
        body := `{"service_name":"Test","price":100}`
        req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte(body)))

        var data map[string]interface{}
        err := parseJSON(req, &data)
        if err != nil {
            t.Errorf("parseJSON failed: %v", err)
        }
        if data["service_name"] != "Test" {
            t.Errorf("expected 'Test', got '%v'", data["service_name"])
        }
        if data["price"] != float64(100) {
            t.Errorf("expected 100, got '%v'", data["price"])
        }
    })

    t.Run("invalid JSON", func(t *testing.T) {
        req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte(`{invalid`)))
        var data map[string]interface{}
        err := parseJSON(req, &data)
        if err == nil {
            t.Error("expected error, got nil")
        }
    })

    t.Run("empty body", func(t *testing.T) {
        req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte(``)))
        var data map[string]interface{}
        err := parseJSON(req, &data)
        if err == nil {
            t.Error("expected error, got nil")
        }
    })

    t.Run("JSON with null", func(t *testing.T) {
        body := `{"service_name":null}`
        req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte(body)))

        var data map[string]interface{}
        err := parseJSON(req, &data)
        if err != nil {
            t.Errorf("parseJSON failed: %v", err)
        }
        if data["service_name"] != nil {
            t.Errorf("expected nil, got '%v'", data["service_name"])
        }
    })
}

func TestIsValidDate(t *testing.T) {
    tests := []struct {
        name  string
        date  string
        valid bool
    }{
        {"valid date", "07-2025", true},
        {"valid date December", "12-2025", true},
        {"valid date January", "01-2025", true},
        {"invalid format", "2025-07", false},
        {"invalid month 13", "13-2025", false},
        {"invalid month 00", "00-2025", false},
        {"invalid year 2 digits", "07-25", false},
        {"invalid year 5 digits", "07-20255", false},
        {"empty string", "", false},
        {"no separator", "072025", false},
        {"month as word", "Jan-2025", false},
        {"year before 1900", "07-1899", false},
        {"year after 2100", "07-2101", false},
        {"February valid", "02-2024", true},
        {"month with leading zero", "05-2025", true},
        {"year exactly 1900", "01-1900", true},
        {"year exactly 2100", "12-2100", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := isValidDate(tt.date)
            if result != tt.valid {
                t.Errorf("isValidDate(%q) = %v, want %v", tt.date, result, tt.valid)
            }
        })
    }
}
// ============================================================
// ТЕСТ: HealthHandler
// ============================================================
// Проверяет, что HealthHandler возвращает {"status":"ok"}
// Используется для Docker healthcheck
// ============================================================
func TestHealthHandler(t *testing.T) {
    t.Log("Starting TestHealthHandler")

    req := httptest.NewRequest("GET", "/health", nil)
    w := httptest.NewRecorder()

    HealthHandler(w, req)

    if w.Code != http.StatusOK {
        t.Logf("Expected status 200, got %d", w.Code)
        t.Errorf("Expected 200, got %d", w.Code)
    }

    var resp map[string]string
    if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
        t.Logf("Failed to decode response: %v", err)
        t.Fatalf("Failed to decode: %v", err)
    }

    if resp["status"] != "ok" {
        t.Logf("Expected status 'ok', got '%s'", resp["status"])
        t.Errorf("Expected 'ok', got '%s'", resp["status"])
    }
    t.Log("HealthHandler returned correct response")
}

// ============================================================
// ТЕСТ: LoggingMiddleware
// ============================================================
// Проверяет, что LoggingMiddleware не блокирует запросы
// ============================================================
func TestLoggingMiddleware(t *testing.T) {
    t.Log("Starting TestLoggingMiddleware")

    next := func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }

    handler := LoggingMiddleware(next)
    req := httptest.NewRequest("GET", "/test", nil)
    w := httptest.NewRecorder()

    handler(w, req)

    if w.Code != http.StatusOK {
        t.Logf("Expected 200, got %d", w.Code)
        t.Errorf("Expected 200, got %d", w.Code)
    }
    t.Log("LoggingMiddleware passed request successfully")
}
