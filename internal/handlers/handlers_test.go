// internal/handlers/handlers_test.go
package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/Evgeny-08-01/Rest-user-aggregator/internal/database"
	"github.com/Evgeny-08-01/Rest-user-aggregator/internal/models"
	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	godotenv.Load("../.env")

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "postgres://postgres:mysecret@localhost:5432/subscriptions?sslmode=disable"
	}

	err := database.Init(dbPath)
	if err != nil {
		panic("Failed to init DB: " + err.Error())
	}
	defer database.Close()

	// Очистка таблицы перед тестами
_, err = database.GetDB().Exec("TRUNCATE subscriptions RESTART IDENTITY")
	if err != nil {
		panic("Failed to truncate table: " + err.Error())
	}
	os.Exit(m.Run())
}

// Создаёт handler для тестов (один раз, но можно и в каждом тесте)
func setupTestHandler() *Handler {
	repo := database.NewPostgresRepo()
	return NewHandler(repo)
}

func TestCreateSubscriptionHandler(t *testing.T) {
	handler := setupTestHandler()

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
			w := httptest.NewRecorder()
			handler.CreateSubscriptionHandler(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("got %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestGetSubscriptionHandler(t *testing.T) {
	handler := setupTestHandler()

	// Создаём тестовую подписку
	createBody := `{"service_name":"TestGet","price":100,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"07-2025"}`
	req := httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(createBody)))
	w := httptest.NewRecorder()
	handler.CreateSubscriptionHandler(w, req)

	var resp map[string]int
	json.NewDecoder(w.Body).Decode(&resp)
	id := resp["id"]

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/subscriptions/{id}", nil)
		req.SetPathValue("id", strconv.Itoa(id))
		w := httptest.NewRecorder()
		handler.GetSubscriptionHandler(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("got %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/subscriptions/{id}", nil)
		req.SetPathValue("id", "99999")
		w := httptest.NewRecorder()
		handler.GetSubscriptionHandler(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("got %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/subscriptions/{id}", nil)
		req.SetPathValue("id", "abc")
		w := httptest.NewRecorder()
		handler.GetSubscriptionHandler(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestUpdateSubscriptionHandler(t *testing.T) {
	handler := setupTestHandler()

	// Создаём тестовую подписку
	createBody := `{"service_name":"BeforeUpdate","price":100,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"07-2025"}`
	req := httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(createBody)))
	w := httptest.NewRecorder()
	handler.CreateSubscriptionHandler(w, req)

	var resp map[string]int
	json.NewDecoder(w.Body).Decode(&resp)
	id := resp["id"]

	t.Run("success", func(t *testing.T) {
		updateBody := `{"service_name":"AfterUpdate","price":200,"user_id":"550e8400-e29b-41d4-a716-446655440000",
		"start_date":"08-2025","end_date":"12-2025"}`
		req := httptest.NewRequest("PUT", "/api/subscriptions/{id}", bytes.NewReader([]byte(updateBody)))
		req.SetPathValue("id", strconv.Itoa(id))
		w := httptest.NewRecorder()
		handler.UpdateSubscriptionHandler(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("got %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		updateBody := `{"service_name":"Test","price":100,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"07-2025"}`
		req := httptest.NewRequest("PUT", "/api/subscriptions/{id}", bytes.NewReader([]byte(updateBody)))
		req.SetPathValue("id", "abc")
		w := httptest.NewRecorder()
		handler.UpdateSubscriptionHandler(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestDeleteSubscriptionHandler(t *testing.T) {
	handler := setupTestHandler()

	// Создаём тестовую подписку
	createBody := `{"service_name":"ToDelete","price":100,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"07-2025"}`
	req := httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(createBody)))
	w := httptest.NewRecorder()
	handler.CreateSubscriptionHandler(w, req)

	var resp map[string]int
	json.NewDecoder(w.Body).Decode(&resp)
	id := resp["id"]

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/subscriptions/{id}", nil)
		req.SetPathValue("id", strconv.Itoa(id))
		w := httptest.NewRecorder()
		handler.DeleteSubscriptionHandler(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("got %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/subscriptions/{id}", nil)
		req.SetPathValue("id", "abc")
		w := httptest.NewRecorder()
		handler.DeleteSubscriptionHandler(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestListSubscriptionsHandler(t *testing.T) {
	handler := setupTestHandler()

	// Очистка перед тестом
database.GetDB().Exec("DELETE FROM subscriptions WHERE user_id = '550e8400-e29b-41d4-a716-446655440001'")

	// Создаём 3 тестовые подписки
	for i := 1; i <= 3; i++ {
		body := `{"service_name":"ListTest","price":100,"user_id":"550e8400-e29b-41d4-a716-446655440001","start_date":"07-2025"}`
		req := httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(body)))
		w := httptest.NewRecorder()
		handler.CreateSubscriptionHandler(w, req)
	}

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/subscriptions", nil)
		w := httptest.NewRecorder()
		handler.ListSubscriptionsHandler(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("got %d, want %d", w.Code, http.StatusOK)
		}

		var list []models.Subscription
		json.NewDecoder(w.Body).Decode(&list)
		if len(list) < 3 {
			t.Errorf("expected at least 3, got %d", len(list))
		}
	})
}

func TestGetTotalCostHandler(t *testing.T) {
	handler := setupTestHandler()

	userID := "550e8400-e29b-41d4-a716-446655440002"
	// Очистка перед тестом
database.GetDB().Exec("DELETE FROM subscriptions WHERE user_id = $1", userID)

	// Подготовка тестовых данных
	bodies := []struct {
		name      string
		price     int
		startDate string
		endDate   string
	}{
		{"Cost1", 100, "01-2025", ""},
		{"Cost2", 200, "02-2025", ""},
		{"Cost3", 300, "03-2025", ""},
	}

	for _, b := range bodies {
		body := fmt.Sprintf(`{"service_name":"%s","price":%d,"user_id":"%s","start_date":"%s","end_date":"%s"}`,
			b.name, b.price, userID, b.startDate, b.endDate)
		req := httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(body)))
		w := httptest.NewRecorder()
		handler.CreateSubscriptionHandler(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("failed to create test subscription %s: %d", b.name, w.Code)
		}
	}

	// Тестовые сценарии
	tests := []struct {
		name        string
		userID      string
		serviceName string
		startDate   string
		endDate     string
		expected    int
	}{
		{"full year", userID, "", "01-2025", "12-2025", 6400},
		{"feb-mar", userID, "", "02-2025", "03-2025", 900},
		{"only feb", userID, "", "02-2025", "02-2025", 300},
		{"only mar", userID, "", "03-2025", "03-2025", 600},
		{"jan-mar", userID, "", "01-2025", "03-2025", 1000},
		{"feb-jun", userID, "", "02-2025", "06-2025", 2700},
		{"jun-dec", userID, "", "06-2025", "12-2025", 4200},
		{"apr-sep", userID, "", "04-2025", "09-2025", 3600},
		{"full year Cost1", userID, "Cost1", "01-2025", "12-2025", 1200},
		{"full year Cost2", userID, "Cost2", "01-2025", "12-2025", 2200},
		{"full year Cost3", userID, "Cost3", "01-2025", "12-2025", 3000},
		{"feb-mar Cost2", userID, "Cost2", "02-2025", "03-2025", 400},
		{"jan-mar Cost3", userID, "Cost3", "01-2025", "03-2025", 300},
		{"single month Jan", userID, "Cost1", "01-2025", "01-2025", 100},
		{"single month Feb", userID, "Cost2", "02-2025", "02-2025", 200},
		{"invalid period", userID, "", "12-2025", "01-2025", -1},
		{"unknown service", userID, "NoSuchService", "01-2025", "12-2025", 0},
		{"empty user and service", "", "", "01-2025", "12-2025", 10400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/subscriptions/total-cost?user_id=%s&service_name=%s&start_date=%s&end_date=%s",
				tt.userID, tt.serviceName, tt.startDate, tt.endDate)
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			handler.GetTotalCostHandler(w, req)

			if tt.expected == -1 {
				if w.Code != http.StatusBadRequest {
					t.Errorf("expected 400, got %d", w.Code)
				}
				return
			}

			if w.Code != http.StatusOK {
				t.Errorf("got %d, want 200", w.Code)
				return
			}

			var resp map[string]int
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if resp["total"] != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, resp["total"])
			}
		})
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