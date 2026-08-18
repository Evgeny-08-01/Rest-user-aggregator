//g o:build integration

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

	"Rest-user-agregator/internal/authentication" //для аутентификации пользователей       // для работы с Redis
	"Rest-user-agregator/internal/cache"
	"Rest-user-agregator/internal/database"
	"Rest-user-agregator/internal/models"
	"Rest-user-agregator/internal/service"
	"Rest-user-agregator/pkg/logger"

	"github.com/golang-jwt/jwt/v5" // для генерации JWT
	"github.com/joho/godotenv"
)

// ============================================================
// TestMain — настройка окружения перед запуском всех тестов
// ============================================================
// Эта функция выполняется ОДИН РАЗ перед всеми тестами в пакете.
// Она подготавливает базу данных и окружение для всех тестов.
//
// ПОРЯДОК ДЕЙСТВИЙ:
//   1. Загружаем переменные окружения из .env.test
//   2. Подключаемся к PostgreSQL
//   3. Создаём таблицу (если её нет)
//   4. Очищаем таблицу от старых данных
//   5. Запускаем все тесты
//   6. Закрываем соединение с БД после всех тестов
// ============================================================
func TestMain(m *testing.M) {
	// 1. ЗАГРУЖАЕМ ПЕРЕМЕННЫЕ ОКРУЖЕНИЯ
	//    Файл .env.test лежит в корне проекта (на два уровня выше)
	//    Если файл не найден — тесты пропускаются (но не падают)
	godotenv.Load("../../.env.test")
	if err := godotenv.Load("../../.env.test"); err != nil {
    log.Println("WARNING: .env.test not found, using env vars")
}
	log.Println("LOG_LEVEL from env.test:", os.Getenv("LOG_LEVEL"))
	logger.Init(os.Getenv("LOG_PATH"), os.Getenv("LOG_LEVEL"))
	
	// 2. ПОЛУЧАЕМ СТРОКУ ПОДКЛЮЧЕНИЯ К БД
	//    Берём из переменной окружения DB_PATH
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		// Если не задана — используем стандартную для локальной разработки
		dbPath = "postgres://postgres:mysecret@localhost:5432/subscriptions?sslmode=disable"
	}

	// 3. ПОДКЛЮЧАЕМСЯ К БАЗЕ ДАННЫХ
	err := database.Init(dbPath)
	if err != nil {
		// Если не удалось подключиться — тесты не имеют смысла
		// Используем panic, чтобы остановить выполнение
		panic("Failed to init DB: " + err.Error())
	}
// Инициализация Redis
redisAddr := os.Getenv("REDIS_ADDR")
if redisAddr == "" {
    redisAddr = "localhost:6379"
}
if err := cache.InitRedis(redisAddr, "", 0); err != nil {
    log.Println("WARNING: Redis not available, cache disabled")
}
	// 4. ЗАКРЫВАЕМ СОЕДИНЕНИЕ ПОСЛЕ ТЕСТОВ
	//    defer означает: "выполни эту функцию в конце работы TestMain"
	//    Это гарантирует, что БД закроется даже если тесты упадут
	defer database.Close()

	// 5. СОЗДАЁМ ТАБЛИЦУ
	//    Если таблица subscriptions уже есть — ничего не делает
	if err := database.CreateTestTable(); err != nil {
		panic("Failed to create table: " + err.Error())
	}

	// 6. ОЧИЩАЕМ ТАБЛИЦУ
	//    Удаляем все записи и сбрасываем счётчик ID
	//    Это нужно, чтобы каждый тест начинался с пустой БД
	if err := database.CleanTestTable(); err != nil {
		panic("Failed to clean table: " + err.Error())
	}
	// 7. ЗАПУСКАЕМ ВСЕ ТЕСТЫ
	//    m.Run() запускает все тесты в этом пакете
	//    os.Exit() передаёт код выхода (0 = успех, 1 = ошибка)
	os.Exit(m.Run())
}

// ============================================================
// setupTestHandler — создаёт экземпляр Handler для тестов
// ============================================================
// Handler — это структура, которая содержит все HTTP-обработчики.
// Для тестов нам нужен реальный Handler, который работает с реальной БД.
//
// Что делает:
//   1. Создаёт репозиторий (работа с БД через PostgreSQL)
//   2. Создаёт сервис (бизнес-логика)
//   3. Возвращает Handler с этими зависимостями
//
// Почему AuthService = nil:
//   В интеграционных тестах мы проверяем работу с подписками,
//   а не авторизацию. Поэтому сервис авторизации не нужен.
// ============================================================
func setupTestHandler() *Handler {
	repo := database.NewPostgresRepo()
	svc := service.NewSubscriptionService(repo)
	return NewHandler(svc, nil)
}

// ============================================================
// addAdminContext — добавляет роль admin в контекст запроса
// ============================================================
// Все защищённые эндпоинты требуют авторизации.
// В контекст запроса мы добавляем роль "admin", чтобы хендлер
// пропустил запрос без ошибки 403 Forbidden.
//
// КАК РАБОТАЕТ:
//   1. req.Context() — получаем контекст запроса
//   2. context.WithValue() — добавляем туда пару ("role", "admin")
//   3. req.WithContext() — создаём новый запрос с обновлённым контекстом
//
// ПОЧЕМУ ADMIN:
//   В тестах мы используем роль admin, чтобы иметь полный доступ
//   к созданию, обновлению и удалению подписок.
//   Если бы мы использовали роль user — некоторые операции были бы запрещены.
// ============================================================
func addAdminContext(req *http.Request) *http.Request {
	ctx := context.WithValue(req.Context(), authentication.RoleKey, "admin")
	ctx = context.WithValue(ctx, authentication.UserIDKey, "550e8400-e29b-41d4-a716-446655440000")
	return req.WithContext(ctx)
}


// ============================================================
// 1. ТЕСТ: СОЗДАНИЕ ПОДПИСКИ (POST /api/subscriptions)
// ============================================================
// ЧТО ПРОВЕРЯЕТ:
//   - Успешное создание подписки (201 Created)
//   - Ошибки валидации (400 Bad Request):
//     * Пустое название сервиса
//     * Отрицательная цена
//     * Пустой user_id
//     * Неверный формат даты
//     * Невалидный JSON
//
// КАК РАБОТАЕТ:
//   1. Создаём HTTP-запрос с телом (JSON)
//   2. Добавляем роль admin в контекст
//   3. Отправляем запрос хендлеру
//   4. Проверяем, что статус ответа совпадает с ожидаемым
// ============================================================
func TestCreateSubscriptionHandler(t *testing.T) {
	 if err := database.CleanTestTable(); err != nil {
        t.Logf("Failed to clean table: %v", err)
        t.Fatalf("Cannot continue test without clean table")
    }
    t.Log("Table cleaned successfully")
	handler := setupTestHandler()

	// Таблица тестов: название, тело запроса, ожидаемый статус
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		// Успешное создание
		{
			"success",
			`{"service_name":"Test","price":100,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"07-2025"}`,
			http.StatusCreated,
		},
		// Пустое название → ошибка валидации
		{
			"empty service_name",
			`{"service_name":"","price":100,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"07-2025"}`,
			http.StatusBadRequest,
		},
		// Отрицательная цена → ошибка валидации
		{
			"negative price",
			`{"service_name":"Test","price":-10,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"07-2025"}`,
			http.StatusBadRequest,
		},
		// Пустой user_id → ошибка валидации
		{
			"empty user_id",
			`{"service_name":"Test","price":100,"user_id":"","start_date":"07-2025"}`,
			http.StatusBadRequest,
		},
		// Неверный формат даты → ошибка валидации (должно быть MM-YYYY)
		{
			"invalid date",
			`{"service_name":"Test","price":100,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"2025-07"}`,
			http.StatusBadRequest,
		},
		// Невалидный JSON → ошибка парсинга
		{
			"invalid JSON",
			`{"service_name":}`,
			http.StatusBadRequest,
		},
	}

	// Запускаем каждый тест
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаём POST-запрос с телом
			req := httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(tt.body)))
			// Добавляем роль admin
			req = addAdminContext(req)
			// Создаём recorder для ответа
			w := httptest.NewRecorder()
			// Вызываем хендлер
			ctx := context.WithValue(req.Context(), authentication.UserIDKey, "550e8400-e29b-41d4-a716-446655440000")
            ctx = context.WithValue(ctx, authentication.RoleKey, "admin")
            req = req.WithContext(ctx)
			handler.CreateSubscriptionHandler(w, req)
			// Проверяем статус
			if w.Code != tt.wantStatus {
				t.Errorf("got %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

// ============================================================
// 2. ТЕСТ: ПОЛУЧЕНИЕ ПОДПИСКИ ПО ID (GET /api/subscriptions/{id})
// ============================================================
// ЧТО ПРОВЕРЯЕТ:
//   - Успешное получение существующей подписки (200 OK)
//   - Ошибка 404 для несуществующего ID
//   - Ошибка 400 для невалидного ID (например, "abc")
//
// КАК РАБОТАЕТ:
//   1. Создаём тестовую подписку
//   2. Получаем её ID из ответа
//   3. Запрашиваем подписку по ID
//   4. Проверяем, что статус = 200 OK
//   5. Проверяем, что ответ содержит правильные данные
// ============================================================
func TestGetSubscriptionHandler(t *testing.T) {
 if err := database.CleanTestTable(); err != nil {
        t.Logf("Failed to clean table: %v", err)
        t.Fatalf("Cannot continue test without clean table")
    }
    t.Log("Table cleaned successfully")
	handler := setupTestHandler()

	// 1. СОЗДАЁМ ТЕСТОВУЮ ПОДПИСКУ
	createBody := `{"service_name":"TestGet","price":100,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"07-2025"}`
	req := httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(createBody)))
	req = addAdminContext(req)
	w := httptest.NewRecorder()
	handler.CreateSubscriptionHandler(w, req)

	// 2. ИЗВЛЕКАЕМ ID СОЗДАННОЙ ПОДПИСКИ
	var resp map[string]int
if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
    t.Fatalf("Failed to decode response: %v", err)
}
	id := resp["id"]

	// 3. СУБТЕСТ: УСПЕШНОЕ ПОЛУЧЕНИЕ
	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/subscriptions/{id}", nil)
		req = addAdminContext(req)
		req.SetPathValue("id", strconv.Itoa(id))
		w := httptest.NewRecorder()
		handler.GetSubscriptionHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("got %d, want %d", w.Code, http.StatusOK)
		}
	})

	// 4. СУБТЕСТ: ПОДПИСКА НЕ НАЙДЕНА
	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/subscriptions/{id}", nil)
		req = addAdminContext(req)
		req.SetPathValue("id", "99999") // такого ID нет в БД
		w := httptest.NewRecorder()
		ctx := context.WithValue(req.Context(), authentication.UserIDKey, "550e8400-e29b-41d4-a716-446655440000")
        ctx = context.WithValue(ctx, authentication.RoleKey, "admin")
        req = req.WithContext(ctx)     
		handler.GetSubscriptionHandler(w, req)

		// Ожидаем 404 Not Found
		if w.Code != http.StatusNotFound {
			t.Errorf("got %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	// 5. СУБТЕСТ: НЕВАЛИДНЫЙ ID
	t.Run("invalid id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/subscriptions/{id}", nil)
		req = addAdminContext(req)
		req.SetPathValue("id", "abc") // не число
		w := httptest.NewRecorder()
		handler.GetSubscriptionHandler(w, req)

		// Ожидаем 400 Bad Request
		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

// ============================================================
// 3. ТЕСТ: ОБНОВЛЕНИЕ ПОДПИСКИ (PUT /api/subscriptions/{id})
// ============================================================
// ЧТО ПРОВЕРЯЕТ:
//   - Успешное обновление существующей подписки (200 OK)
//   - Ошибка 400 для невалидного ID
//
// КАК РАБОТАЕТ:
//   1. Создаём подписку
//   2. Обновляем её через PUT
//   3. Проверяем статус ответа
// ============================================================
func TestUpdateSubscriptionHandler(t *testing.T) {
	 if err := database.CleanTestTable(); err != nil {
        t.Logf("Failed to clean table: %v", err)
        t.Fatalf("Cannot continue test without clean table")
    }
    t.Log("Table cleaned successfully")
	handler := setupTestHandler()

	// 1. СОЗДАЁМ ТЕСТОВУЮ ПОДПИСКУ
	createBody := `{"service_name":"BeforeUpdate","price":100,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"07-2025"}`
	req := httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(createBody)))
	req = addAdminContext(req)
	w := httptest.NewRecorder()
	ctx := context.WithValue(req.Context(), authentication.UserIDKey, "550e8400-e29b-41d4-a716-446655440000")
    ctx = context.WithValue(ctx, authentication.RoleKey, "admin")
    req = req.WithContext(ctx) 
	handler.CreateSubscriptionHandler(w, req)

	// 2. ИЗВЛЕКАЕМ ID
	var resp map[string]int
if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
    t.Fatalf("Failed to decode response: %v", err)
}
	id := resp["id"]

	// 3. СУБТЕСТ: УСПЕШНОЕ ОБНОВЛЕНИЕ
	t.Run("success", func(t *testing.T) {
		// Новые данные: другое название, другая цена, дата окончания
		updateBody := `{"service_name":"AfterUpdate","price":200,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"08-2025","end_date":"12-2025"}`
		req := httptest.NewRequest("PUT", "/api/subscriptions/{id}", bytes.NewReader([]byte(updateBody)))
		req = addAdminContext(req)
		req.SetPathValue("id", strconv.Itoa(id))
		w := httptest.NewRecorder()
		ctx := context.WithValue(req.Context(), authentication.UserIDKey, "550e8400-e29b-41d4-a716-446655440000")
        ctx = context.WithValue(ctx, authentication.RoleKey, "admin")
        req = req.WithContext(ctx)         
		handler.UpdateSubscriptionHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("got %d, want %d", w.Code, http.StatusOK)
		}
	})

	// 4. СУБТЕСТ: НЕВАЛИДНЫЙ ID
	t.Run("invalid id", func(t *testing.T) {
		updateBody := `{"service_name":"Test","price":100,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"07-2025"}`
		req := httptest.NewRequest("PUT", "/api/subscriptions/{id}", bytes.NewReader([]byte(updateBody)))
		req = addAdminContext(req)
		req.SetPathValue("id", "abc") // не число
		w := httptest.NewRecorder()
		handler.UpdateSubscriptionHandler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

// ============================================================
// 4. ТЕСТ: УДАЛЕНИЕ ПОДПИСКИ (DELETE /api/subscriptions/{id})
// ============================================================
// ЧТО ПРОВЕРЯЕТ:
//   - Успешное удаление существующей подписки (200 OK)
//   - Ошибка 400 для невалидного ID
//
// КАК РАБОТАЕТ:
//   1. Создаём подписку
//   2. Удаляем её через DELETE
//   3. Проверяем статус ответа
// ============================================================
func TestDeleteSubscriptionHandler(t *testing.T) {
	 if err := database.CleanTestTable(); err != nil {
        t.Logf("Failed to clean table: %v", err)
        t.Fatalf("Cannot continue test without clean table")
    }
    t.Log("Table cleaned successfully")
	handler := setupTestHandler()

	// 1. СОЗДАЁМ ПОДПИСКУ
	createBody := `{"service_name":"ToDelete","price":100,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"07-2025"}`
	req := httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(createBody)))
	req = addAdminContext(req)
	w := httptest.NewRecorder()
	handler.CreateSubscriptionHandler(w, req)

	// 2. ИЗВЛЕКАЕМ ID
	var resp map[string]int
if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
    t.Fatalf("Failed to decode response: %v", err)
}
	id := resp["id"]

	// 3. СУБТЕСТ: УСПЕШНОЕ УДАЛЕНИЕ
	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/subscriptions/{id}", nil)
		req = addAdminContext(req)
		req.SetPathValue("id", strconv.Itoa(id))
		w := httptest.NewRecorder()
		handler.DeleteSubscriptionHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("got %d, want %d", w.Code, http.StatusOK)
		}
	})

	// 4. СУБТЕСТ: НЕВАЛИДНЫЙ ID
	t.Run("invalid id", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/subscriptions/{id}", nil)
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
// 5. ТЕСТ: ПОЛУЧЕНИЕ СПИСКА ПОДПИСОК (GET /api/subscriptions)
// ============================================================
// ЧТО ПРОВЕРЯЕТ:
//   - Успешное получение списка (200 OK)
//   - В списке есть минимум 3 записи (мы создаём 3 подписки)
//   - Пагинация работает (limit/offset)
//
// КАК РАБОТАЕТ:
//   1. Очищаем данные пользователя
//   2. Создаём 3 подписки
//   3. Запрашиваем список
//   4. Проверяем, что в ответе минимум 3 записи
// ============================================================
func TestListSubscriptionsHandler(t *testing.T) {
	 if err := database.CleanTestTable(); err != nil {
        t.Logf("Failed to clean table: %v", err)
        t.Fatalf("Cannot continue test without clean table")
    }
    t.Log("Table cleaned successfully")
	handler := setupTestHandler()

	// 1. ОЧИЩАЕМ ДАННЫЕ ПОЛЬЗОВАТЕЛЯ
	//    Удаляем все подписки с этим user_id, чтобы не было конфликтов
	if err := database.DeleteSubscriptionsByUserID("550e8400-e29b-41d4-a716-446655440001"); err != nil {
		t.Fatalf("Failed to clean test data: %v", err)
	}

	// 2. СОЗДАЁМ 3 ТЕСТОВЫЕ ПОДПИСКИ
	for i := 1; i <= 3; i++ {
		body := `{"service_name":"ListTest","price":100,"user_id":"550e8400-e29b-41d4-a716-446655440001","start_date":"07-2025"}`
		req := httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(body)))
		req = addAdminContext(req)
		w := httptest.NewRecorder()
		handler.CreateSubscriptionHandler(w, req)
	}

	// 3. СУБТЕСТ: ПОЛУЧАЕМ СПИСОК
	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/subscriptions", nil)
		req = addAdminContext(req)
		w := httptest.NewRecorder()
		handler.ListSubscriptionsHandler(w, req)

		// Проверяем статус
		if w.Code != http.StatusOK {
			t.Errorf("got %d, want %d", w.Code, http.StatusOK)
		}

		// Проверяем, что в списке не меньше 3 записей
		var list []models.Subscription
		json.NewDecoder(w.Body).Decode(&list)
		if len(list) < 3 {
			t.Errorf("expected at least 3, got %d", len(list))
		}
	})
}

// ============================================================
// 6. ТЕСТ: СУММАРНАЯ СТОИМОСТЬ (GET /api/subscriptions/total-cost)
// ============================================================
// ЧТО ПРОВЕРЯЕТ:
//   - Правильность расчёта суммарной стоимости за период
//   - Фильтрацию по user_id (если передан)
//   - Фильтрацию по service_name (если передан)
//   - Обработку невалидного периода (start_date > end_date)
//   - Обработку несуществующего сервиса (возвращает 0)
//
// КАК РАБОТАЕТ:
//   1. Создаём 3 подписки с разными ценами и датами
//   2. Для каждого тестового сценария отправляем GET-запрос с параметрами
//   3. Сравниваем ответ бэкенда с ожидаемым значением
//
// ПОЧЕМУ ТАКИЕ ЦИФРЫ В ОЖИДАНИЯХ:
//   Подписки:
//     - Cost1: 100 ₽, началась в январе 2025, без окончания
//     - Cost2: 200 ₽, началась в феврале 2025, без окончания
//     - Cost3: 300 ₽, началась в марте 2025, без окончания
//
//   Формула в БД считает стоимость так:
//     price * (количество месяцев от start_date до end_date)
//
//   Эти ожидания были проверены на работающем бэкенде
//   и отражают реальное поведение формулы.
// ============================================================
func TestGetTotalCostHandler(t *testing.T) {
	handler := setupTestHandler()
    // Очищаем ВСЮ таблицу перед тестом
    if err := database.CleanTestTable(); err != nil {
        t.Fatalf("Failed to clean table: %v", err)
    }
    t.Log("Table cleaned successfully")
	
	// ID пользователя для теста
	userID := "550e8400-e29b-41d4-a716-446655440002"


	// 2. СОЗДАЁМ 3 ПОДПИСКИ
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
		req = addAdminContext(req)
		w := httptest.NewRecorder()
		handler.CreateSubscriptionHandler(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("failed to create test subscription %s: %d", b.name, w.Code)
		}
	}

	// 3. ТЕСТОВЫЕ СЦЕНАРИИ
	//
	// Ожидаемые значения — это те числа, которые возвращает бэкенд.
	// Они были проверены эмпирически и отражают реальную логику формулы.
	tests := []struct {
		name        string
		userID      string
		serviceName string
		startDate   string
		endDate     string
		expected    int
	}{
		// Полный год: сумма всех подписок за январь-декабрь
		{"full year", userID, "", "01-2025", "12-2025", 6400},

		// Февраль-март: только подписки 2 и 3
		{"feb-mar", userID, "", "02-2025", "03-2025", 900},

		// Только февраль
		{"only feb", userID, "", "02-2025", "02-2025", 300},

		// Только март
		{"only mar", userID, "", "03-2025", "03-2025", 600},

		// Январь-март
		{"jan-mar", userID, "", "01-2025", "03-2025", 1000},

		// Февраль-июнь
		{"feb-jun", userID, "", "02-2025", "06-2025", 2700},

		// Июнь-декабрь
		{"jun-dec", userID, "", "06-2025", "12-2025", 4200},

		// Апрель-сентябрь
		{"apr-sep", userID, "", "04-2025", "09-2025", 3600},

		// Фильтр по названию: только Cost1 за год
		{"full year Cost1", userID, "Cost1", "01-2025", "12-2025", 1200},

		// Фильтр по названию: только Cost2 за год
		{"full year Cost2", userID, "Cost2", "01-2025", "12-2025", 2200},

		// Фильтр по названию: только Cost3 за год
		{"full year Cost3", userID, "Cost3", "01-2025", "12-2025", 3000},

		// Фильтр по названию: Cost2 за февраль-март
		{"feb-mar Cost2", userID, "Cost2", "02-2025", "03-2025", 400},

		// Фильтр по названию: Cost3 за январь-март
		{"jan-mar Cost3", userID, "Cost3", "01-2025", "03-2025", 300},

		// Одиночные месяцы
		{"single month Jan", userID, "Cost1", "01-2025", "01-2025", 100},
		{"single month Feb", userID, "Cost2", "02-2025", "02-2025", 200},

		// Невалидный период: start_date > end_date → ошибка
		{"invalid period", userID, "", "12-2025", "01-2025", -1},

		// Неизвестный сервис → 0
		{"unknown service", userID, "NoSuchService", "01-2025", "12-2025", 0},

		// Без фильтров: все подписки всех пользователей
		{"empty user and service", "", "", "01-2025", "12-2025", 6400},
	}

	// 4. ЗАПУСКАЕМ ВСЕ СЦЕНАРИИ
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Формируем URL с параметрами
			url := fmt.Sprintf("/api/subscriptions/total-cost?user_id=%s&service_name=%s&start_date=%s&end_date=%s",
				tt.userID, tt.serviceName, tt.startDate, tt.endDate)

			// Создаём GET-запрос
			req := httptest.NewRequest("GET", url, nil)
			req = addAdminContext(req)

			// Отправляем запрос
			w := httptest.NewRecorder()
			handler.GetTotalCostHandler(w, req)

			// Если ожидаем ошибку 400
			if tt.expected == -1 {
				if w.Code != http.StatusBadRequest {
					t.Errorf("expected 400, got %d", w.Code)
				}
				return
			}

			// Проверяем статус 200 OK
			if w.Code != http.StatusOK {
				t.Errorf("got %d, want 200", w.Code)
				return
			}

			// Декодируем JSON-ответ
			var resp map[string]int
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			// Проверяем сумму
			if resp["total"] != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, resp["total"])
			}
		})
	}
}
// ============================================================
// ИНТЕГРАЦИОННЫЕ ТЕСТЫ КЕШИРОВАНИЯ
// ============================================================
//
// Эти тесты проверяют реальную работу с Redis через API.
// Требуют запущенного Redis и чистой БД.
// ============================================================
// ============================================================
// createTestUser — создаёт тестового пользователя в БД и возвращает JWT-токен.
// ============================================================
// Зачем:
//   - Для интеграционных тестов нужен реальный пользователь с валидным токеном.
//   - Без этого хендлеры возвращают 401 Unauthorized.
//
// Как работает:
//   1. Проверяем, есть ли пользователь в БД.
//   2. Если нет — создаём.
//   3. Генерируем JWT-токен с ролью "user".
//   4. Возвращаем токен.
//
// Параметры:
//   - t: указатель на тест (для логирования и Fail)
//   - userID: UUID пользователя (фиксированный для тестов)
//
// Возвращает:
//   - string: JWT-токен
// ============================================================
func createTestUser(t *testing.T, userID string) string {
    // 1. ПОДКЛЮЧАЕМСЯ К БД (используем глобальный db из пакета database)
    db := database.GetDB()
    if db == nil {
        t.Fatal("Database not initialized")
    }

    // 2. ПРОВЕРЯЕМ, ЕСТЬ ЛИ ПОЛЬЗОВАТЕЛЬ
    var count int
    err := db.QueryRow("SELECT COUNT(*) FROM users WHERE id = $1", userID).Scan(&count)
    if err != nil {
        t.Fatalf("Failed to check user existence: %v", err)
    }

    // 3. ЕСЛИ ПОЛЬЗОВАТЕЛЯ НЕТ — СОЗДАЁМ
    if count == 0 {
        // Хеш пароля для "password123" (bcrypt)
        // Сгенерирован заранее, чтобы не вычислять в тесте
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

    // 4. ГЕНЕРИРУЕМ JWT-ТОКЕН
    //    Используем тот же секрет, что и в .env (JWT_SECRET)
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        secret = "test-secret-key" // fallback для тестов
    }

    // Создаём claims (данные токена)
    claims := authentication.Claims{
        UserID: userID,
        Email:  "test_" + userID + "@example.com",
        Role:   "admin",
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    // Создаём токен
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(secret))
    if err != nil {
        t.Fatalf("Failed to generate JWT token: %v", err)
    }

    t.Logf("JWT token generated for user: %s", userID)
    return tokenString
}
// ============================================================
// TestCacheInvalidationAfterCreate
// ============================================================
// Проверяет инвалидацию кеша после создания подписки.
// Сценарий:
//   1. Делаем GET total-cost (создаётся кеш с v1).
//   2. Запоминаем результат.
//   3. Создаём подписку (POST) → версия инкрементится до 2.
//   4. Делаем GET total-cost снова.
//   5. Проверяем, что результат ИЗМЕНИЛСЯ (данные обновились).
//   6. Проверяем, что в Redis есть ключ с v2, а старый v1 игнорируется.
// ============================================================
func TestCacheInvalidationAfterCreate(t *testing.T) {
    // 1. ОЧИЩАЕМ ТАБЛИЦЫ ПЕРЕД ТЕСТОМ
    if err := database.CleanTestTable(); err != nil {
        t.Fatalf("CleanTestTable failed: %v", err)
    }
    t.Log("Table cleaned successfully")

    handler := setupTestHandler()
    userID := "550e8400-e29b-41d4-a716-446655440000"

    // 2. СОЗДАЁМ ПОЛЬЗОВАТЕЛЯ И ПОЛУЧАЕМ ТОКЕН
    token := createTestUser(t, userID)
    t.Logf("Token generated")

    // 3. ДЕЛАЕМ GET total-cost (создаём кеш v1)
    t.Log("Step 1: GET total-cost (create cache v1),userID:%w",userID)
    req := httptest.NewRequest("GET", "/api/subscriptions/total-cost?user_id="+userID+"&start_date=01-2025&end_date=12-2025", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    req = addTestContext(req, userID, "user")
    w := httptest.NewRecorder()
    handler.GetTotalCostHandler(w, req)
    if w.Code != http.StatusOK {
        t.Fatalf("GET total-cost failed: %d", w.Code)
    }

    // 4. ЗАПОМИНАЕМ ПЕРВЫЙ РЕЗУЛЬТАТ
    var resp1 map[string]int
    if err := json.NewDecoder(w.Body).Decode(&resp1); err != nil {
        t.Fatalf("Failed to decode first response: %v", err)
    }
    firstTotal := resp1["total"]
    t.Logf("First total: %d", firstTotal)

    // 5. ПРОВЕРЯЕМ REDIS — ДОЛЖЕН БЫТЬ КЛЮЧ С v1
    t.Log("Step 2: Checking Redis for v1 keys")
    keys1, err := getRedisKeys("total:v1:*")
    if err != nil {
        t.Logf("Redis check warning: %v", err)
    }
    if len(keys1) == 0 {
        t.Error("Expected at least one Redis key with v1, got none,keys1%w:",keys1)
    } else {
        t.Logf("Found Redis keys with v1: %d", len(keys1))
        for _, key := range keys1 {
            t.Logf("  - %s", key)
        }
    }

    // 6. СОЗДАЁМ ПОДПИСКУ (POST)
    t.Log("Step 3: Create subscription (POST)")
    createBody := `{"service_name":"TestCache","price":100,"user_id":"` + userID + `","start_date":"01-2025","end_date":"12-2025"}`
    req = httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(createBody)))
    req.Header.Set("Authorization", "Bearer "+token)
    req = addTestContext(req, userID, "user")
    w = httptest.NewRecorder()
    handler.CreateSubscriptionHandler(w, req)

    if w.Code != http.StatusCreated {
        t.Fatalf("Create subscription failed: %d", w.Code)
    }
    t.Log("Subscription created successfully")

    // 7. ДЕЛАЕМ GET total-cost СНОВА
    t.Log("Step 4: GET total-cost again")
    req = httptest.NewRequest("GET", "/api/subscriptions/total-cost?user_id="+userID+"&start_date=01-2025&end_date=12-2025", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    req = addTestContext(req, userID, "user")
    w = httptest.NewRecorder()
    handler.GetTotalCostHandler(w, req)

    if w.Code != http.StatusOK {
        t.Fatalf("GET total-cost after create failed: %d", w.Code)
    }

    // 8. ЗАПОМИНАЕМ ВТОРОЙ РЕЗУЛЬТАТ
    var resp2 map[string]int
    if err := json.NewDecoder(w.Body).Decode(&resp2); err != nil {
        t.Fatalf("Failed to decode second response: %v", err)
    }
    secondTotal := resp2["total"]
    t.Logf("Second total: %d", secondTotal)

    // 9. ПРОВЕРЯЕМ, ЧТО РЕЗУЛЬТАТ ИЗМЕНИЛСЯ
    if secondTotal <= firstTotal {
        t.Errorf("Expected total to increase after creating subscription, got %d <= %d", secondTotal, firstTotal)
    } else {
        t.Logf("Total increased correctly: %d -> %d", firstTotal, secondTotal)
    }

    // 10. ПРОВЕРЯЕМ REDIS — ДОЛЖЕН БЫТЬ КЛЮЧ С v2
    t.Log("Step 5: Checking Redis for v2 keys")
    keys2, err := getRedisKeys("total:v2:*")
    if err != nil {
        t.Logf("Redis check warning: %v", err)
    }
    if len(keys2) == 0 {
        t.Error("Expected at least one Redis key with v2, got none")
    } else {
        t.Logf("Found Redis keys with v2: %d", len(keys2))
        for _, key := range keys2 {
            t.Logf("  - %s", key)
        }
    }

    // 11. ПРОВЕРЯЕМ, ЧТО КЕШ ОБНОВИЛСЯ
    if len(keys1) > 0 && len(keys2) > 0 {
        t.Log("Both v1 and v2 keys exist — v1 is ignored because version changed")
    } else {
        t.Log("Expected both v1 and v2 keys to exist")
    }
}




// getRedisKeys — возвращает все ключи Redis по паттерну.
func getRedisKeys(pattern string) ([]string, error) {
    client := cache.GetClient()
    if client == nil {
        return []string{}, nil
    }
    return client.Keys(context.Background(), pattern).Result()
}

// addTestContext — добавляет user_id и role в контекст запроса.
func addTestContext(req *http.Request, userID, role string) *http.Request {
    ctx := req.Context()
    ctx = context.WithValue(ctx, authentication.UserIDKey, userID)
    ctx = context.WithValue(ctx, authentication.RoleKey, role)
    return req.WithContext(ctx)
}
// ============================================================
// TestCacheInvalidationAfterUpdate
// ============================================================
// Проверяет инвалидацию кеша после обновления подписки.
// Сценарий:
//   1. Создаём подписку (POST).
//   2. Делаем GET total-cost (создаётся кеш с v1) — запоминаем результат.
//   3. Обновляем подписку (PUT).
//   4. Делаем GET total-cost снова — результат должен измениться.
//   5. Проверяем, что в Redis есть ключ с v2.
// ============================================================
func TestCacheInvalidationAfterUpdate(t *testing.T) {
    // 1. ОЧИЩАЕМ ТАБЛИЦЫ ПЕРЕД ТЕСТОМ
    if err := database.CleanTestTable(); err != nil {
        t.Fatalf("CleanTestTable failed: %v", err)
    }
    t.Log("Table cleaned successfully")

    handler := setupTestHandler()
    userID := "550e8400-e29b-41d4-a716-446655440000"
    token := createTestUser(t, userID)

    // 2. СОЗДАЁМ ПОДПИСКУ (POST)
    t.Log("Step 1: Create subscription (POST)")
    createBody := `{"service_name":"BeforeUpdate","price":100,"user_id":"` + userID + `","start_date":"01-2025","end_date":"12-2025"}`
    req := httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(createBody)))
    req.Header.Set("Authorization", "Bearer "+token)
    req = addTestContext(req, userID, "user")
    w := httptest.NewRecorder()
    handler.CreateSubscriptionHandler(w, req)

    if w.Code != http.StatusCreated {
        t.Fatalf("Create subscription failed: %d", w.Code)
    }
    t.Log("Subscription created successfully")

    // 3. ПОЛУЧАЕМ ID
    var createResp map[string]int
    if err := json.NewDecoder(w.Body).Decode(&createResp); err != nil {
        t.Fatalf("Failed to decode create response: %v", err)
    }
    subID := createResp["id"]
    t.Logf("Subscription ID: %d", subID)

    // 4. ДЕЛАЕМ GET total-cost (создаём кеш v1) — ЗАПОМИНАЕМ РЕЗУЛЬТАТ
    t.Log("Step 2: GET total-cost (create cache v1)")
    req = httptest.NewRequest("GET", "/api/subscriptions/total-cost?user_id="+userID+"&start_date=01-2025&end_date=12-2025", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    req = addTestContext(req, userID, "user")
    w = httptest.NewRecorder()
    handler.GetTotalCostHandler(w, req)

    if w.Code != http.StatusOK {
        t.Fatalf("GET total-cost before update failed: %d", w.Code)
    }

    var resp1 map[string]int
    if err := json.NewDecoder(w.Body).Decode(&resp1); err != nil {
        t.Fatalf("Failed to decode first response: %v", err)
    }
    firstTotal := resp1["total"]
    t.Logf("First total: %d", firstTotal)

    // 5. ПРОВЕРЯЕМ v1
    t.Log("Step 3: Checking Redis for v1 keys")
    keysV1, err := getRedisKeys("total:v1:*")
    if err != nil {
        t.Logf("Redis check warning: %v", err)
    }
    if len(keysV1) == 0 {
        t.Error("Expected at least one Redis key with v1, got none")
    }

    // 6. ОБНОВЛЯЕМ ПОДПИСКУ (PUT) — МЕНЯЕМ ЦЕНУ, ЧТОБЫ ИЗМЕНИЛСЯ TOTAL
    t.Log("Step 4: Update subscription (PUT) — change price from 100 to 300")
    updateBody := `{"service_name":"AfterUpdate","price":300,"user_id":"` + userID + `","start_date":"01-2025","end_date":"12-2025"}`
    req = httptest.NewRequest("PUT", "/api/subscriptions/{id}", bytes.NewReader([]byte(updateBody)))
	req.SetPathValue("id", strconv.Itoa(subID))
    req.Header.Set("Authorization", "Bearer "+token)
    req = addTestContext(req, userID, "admin")
    w = httptest.NewRecorder()
    handler.UpdateSubscriptionHandler(w, req)

    if w.Code != http.StatusOK {
        t.Fatalf("Update subscription failed: %d", w.Code)
    }
    t.Log("Subscription updated successfully")

    // 7. ДЕЛАЕМ GET total-cost СНОВА
    t.Log("Step 5: GET total-cost after update")
    req = httptest.NewRequest("GET", "/api/subscriptions/total-cost?user_id="+userID+"&start_date=01-2025&end_date=12-2025", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    req = addTestContext(req, userID, "user")
    w = httptest.NewRecorder()
    handler.GetTotalCostHandler(w, req)

    if w.Code != http.StatusOK {
        t.Fatalf("GET total-cost after update failed: %d", w.Code)
    }

    var resp2 map[string]int
    if err := json.NewDecoder(w.Body).Decode(&resp2); err != nil {
        t.Fatalf("Failed to decode second response: %v", err)
    }
    secondTotal := resp2["total"]
    t.Logf("Second total: %d", secondTotal)

    // 8. ПРОВЕРЯЕМ, ЧТО РЕЗУЛЬТАТ ИЗМЕНИЛСЯ
    if secondTotal == firstTotal {
        t.Errorf("Expected total to change after update, got same: %d", firstTotal)
    } else {
        t.Logf("Total changed correctly: %d -> %d", firstTotal, secondTotal)
    }

    // 9. ПРОВЕРЯЕМ v2
    t.Log("Step 6: Checking Redis for v2 keys")
    keysV2, err := getRedisKeys("total:v2:*")
    if err != nil {
        t.Logf("Redis check warning: %v", err)
    }
    if len(keysV2) == 0 {
        t.Error("Expected at least one Redis key with v2, got none")
    }

    t.Log("Both v1 and v2 keys exist — v1 is ignored because version changed")
}
// ============================================================
// TestCacheInvalidationAfterDelete
// ============================================================
// Проверяет инвалидацию кеша после удаления подписки.
// Сценарий:
//   1. Создаём подписку (POST).
//   2. Делаем GET total-cost (создаётся кеш с v1) — запоминаем результат.
//   3. Удаляем подписку (DELETE).
//   4. Делаем GET total-cost снова — результат должен уменьшиться.
//   5. Проверяем, что в Redis есть ключ с v2.
// ============================================================
func TestCacheInvalidationAfterDelete(t *testing.T) {
    // 1. ОЧИЩАЕМ ТАБЛИЦЫ ПЕРЕД ТЕСТОМ
    if err := database.CleanTestTable(); err != nil {
        t.Fatalf("CleanTestTable failed: %v", err)
    }
    t.Log("Table cleaned successfully")

    handler := setupTestHandler()
    userID := "550e8400-e29b-41d4-a716-446655440000"
    token := createTestUser(t, userID)

    // 2. СОЗДАЁМ ПОДПИСКУ (POST)
    t.Log("Step 1: Create subscription (POST)")
    createBody := `{"service_name":"ToDelete","price":100,"user_id":"` + userID + `","start_date":"01-2025","end_date":"12-2025"}`
    req := httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(createBody)))
    req.Header.Set("Authorization", "Bearer "+token)
    req = addTestContext(req, userID, "user")
    w := httptest.NewRecorder()
    handler.CreateSubscriptionHandler(w, req)

    if w.Code != http.StatusCreated {
        t.Fatalf("Create subscription failed: %d", w.Code)
    }
    t.Log("Subscription created successfully")

    // 3. ПОЛУЧАЕМ ID
    var createResp map[string]int
    if err := json.NewDecoder(w.Body).Decode(&createResp); err != nil {
        t.Fatalf("Failed to decode create response: %v", err)
    }
    subID := createResp["id"]
    t.Logf("Subscription ID: %d", subID)

    // 4. ДЕЛАЕМ GET total-cost (создаём кеш v1) — ЗАПОМИНАЕМ РЕЗУЛЬТАТ
    t.Log("Step 2: GET total-cost (create cache v1)")
    req = httptest.NewRequest("GET", "/api/subscriptions/total-cost?user_id="+userID+"&start_date=01-2025&end_date=12-2025", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    req = addTestContext(req, userID, "user")
    w = httptest.NewRecorder()
    handler.GetTotalCostHandler(w, req)

    if w.Code != http.StatusOK {
        t.Fatalf("GET total-cost before delete failed: %d", w.Code)
    }

    var resp1 map[string]int
    if err := json.NewDecoder(w.Body).Decode(&resp1); err != nil {
        t.Fatalf("Failed to decode first response: %v", err)
    }
    firstTotal := resp1["total"]
    t.Logf("First total: %d", firstTotal)

    // 5. ПРОВЕРЯЕМ v1
    t.Log("Step 3: Checking Redis for v1 keys")
    keysV1, err := getRedisKeys("total:v1:*")
    if err != nil {
        t.Logf("Redis check warning: %v", err)
    }
    if len(keysV1) == 0 {
        t.Error("Expected at least one Redis key with v1, got none")
    }

    // 6. УДАЛЯЕМ ПОДПИСКУ (DELETE)
    t.Log("Step 4: Delete subscription (DELETE)")
    req = httptest.NewRequest("DELETE", "/api/subscriptions/{id}", nil)
	req.SetPathValue("id", strconv.Itoa(subID))
    req.Header.Set("Authorization", "Bearer "+token)
    req = addTestContext(req, userID, "admin")
    w = httptest.NewRecorder()
    handler.DeleteSubscriptionHandler(w, req)

    if w.Code != http.StatusOK {
        t.Fatalf("Delete subscription failed: %d", w.Code)
    }
    t.Log("Subscription deleted successfully")

    // 7. ДЕЛАЕМ GET total-cost СНОВА
    t.Log("Step 5: GET total-cost after delete")
    req = httptest.NewRequest("GET", "/api/subscriptions/total-cost?user_id="+userID+"&start_date=01-2025&end_date=12-2025", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    req = addTestContext(req, userID, "user")
    w = httptest.NewRecorder()
    handler.GetTotalCostHandler(w, req)

    if w.Code != http.StatusOK {
        t.Fatalf("GET total-cost after delete failed: %d", w.Code)
    }

    var resp2 map[string]int
    if err := json.NewDecoder(w.Body).Decode(&resp2); err != nil {
        t.Fatalf("Failed to decode second response: %v", err)
    }
    secondTotal := resp2["total"]
    t.Logf("Second total: %d", secondTotal)

    // 8. ПРОВЕРЯЕМ, ЧТО РЕЗУЛЬТАТ ИЗМЕНИЛСЯ (УМЕНЬШИЛСЯ)
    if secondTotal >= firstTotal {
        t.Errorf("Expected total to decrease after deletion, got %d >= %d", secondTotal, firstTotal)
    } else {
        t.Logf("Total decreased correctly: %d -> %d", firstTotal, secondTotal)
    }

    // 9. ПРОВЕРЯЕМ v2
    t.Log("Step 6: Checking Redis for v2 keys")
    keysV2, err := getRedisKeys("total:v2:*")
    if err != nil {
        t.Logf("Redis check warning: %v", err)
    }
    if len(keysV2) == 0 {
        t.Error("Expected at least one Redis key with v2, got none")
    }

    t.Log("Both v1 and v2 keys exist — v1 is ignored because version changed")
}
