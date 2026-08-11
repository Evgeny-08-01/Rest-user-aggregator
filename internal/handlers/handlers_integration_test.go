//go:build integration

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"Rest-user-agregator/internal/database"
	"Rest-user-agregator/internal/models"
	"Rest-user-agregator/internal/service"
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
	ctx := context.WithValue(req.Context(), "role", "admin")
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
	handler := setupTestHandler()

	// 1. СОЗДАЁМ ТЕСТОВУЮ ПОДПИСКУ
	createBody := `{"service_name":"TestGet","price":100,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"07-2025"}`
	req := httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(createBody)))
	req = addAdminContext(req)
	w := httptest.NewRecorder()
	handler.CreateSubscriptionHandler(w, req)

	// 2. ИЗВЛЕКАЕМ ID СОЗДАННОЙ ПОДПИСКИ
	var resp map[string]int
	json.NewDecoder(w.Body).Decode(&resp)
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
	handler := setupTestHandler()

	// 1. СОЗДАЁМ ТЕСТОВУЮ ПОДПИСКУ
	createBody := `{"service_name":"BeforeUpdate","price":100,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"07-2025"}`
	req := httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(createBody)))
	req = addAdminContext(req)
	w := httptest.NewRecorder()
	handler.CreateSubscriptionHandler(w, req)

	// 2. ИЗВЛЕКАЕМ ID
	var resp map[string]int
	json.NewDecoder(w.Body).Decode(&resp)
	id := resp["id"]

	// 3. СУБТЕСТ: УСПЕШНОЕ ОБНОВЛЕНИЕ
	t.Run("success", func(t *testing.T) {
		// Новые данные: другое название, другая цена, дата окончания
		updateBody := `{"service_name":"AfterUpdate","price":200,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"08-2025","end_date":"12-2025"}`
		req := httptest.NewRequest("PUT", "/api/subscriptions/{id}", bytes.NewReader([]byte(updateBody)))
		req = addAdminContext(req)
		req.SetPathValue("id", strconv.Itoa(id))
		w := httptest.NewRecorder()
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
	handler := setupTestHandler()

	// 1. СОЗДАЁМ ПОДПИСКУ
	createBody := `{"service_name":"ToDelete","price":100,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"07-2025"}`
	req := httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(createBody)))
	req = addAdminContext(req)
	w := httptest.NewRecorder()
	handler.CreateSubscriptionHandler(w, req)

	// 2. ИЗВЛЕКАЕМ ID
	var resp map[string]int
	json.NewDecoder(w.Body).Decode(&resp)
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

	// ID пользователя для теста
	userID := "550e8400-e29b-41d4-a716-446655440002"

	// 1. ОЧИЩАЕМ ДАННЫЕ ПОЛЬЗОВАТЕЛЯ
	if err := database.DeleteSubscriptionsByUserID(userID); err != nil {
		t.Fatalf("Failed to clean test data: %v", err)
	}

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
		{"empty user and service", "", "", "01-2025", "12-2025", 10400},
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