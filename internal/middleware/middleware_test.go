// ============================================================
// ФАЙЛ: internal/middleware/middleware_test.go
// ============================================================
// НАЗНАЧЕНИЕ: Юнит-тесты для middleware
// ТЕГ: //go:build unit — запускается с флагом -tags=unit
//
// ЧТО ПРОВЕРЯЕТСЯ:
//   1. CorsMiddleware — добавляет CORS-заголовки
//   2. MetricsMiddleware — собирает метрики
//
// ПОЧЕМУ ЭТИ ТЕСТЫ ВАЖНЫ:
//   - CorsMiddleware — критичен для работы фронтенда
//   - MetricsMiddleware — критичен для мониторинга
//   - Без них непонятно, работают ли middleware
//
// КАК ЗАПУСТИТЬ:
//   go test ./internal/middleware -tags=unit -v
// ============================================================

//go:build unit

package middleware

import (
	"net/http"          // HTTP-запросы и ответы
	"net/http/httptest" // Тестовый HTTP-сервер (запись ответов)
	"os"                // Для работы с переменными окружения
	"testing"           // Стандартный пакет для тестов
)

// ============================================================
// 1. ТЕСТ: CorsMiddleware
// ============================================================
// Что проверяет:
//   - Добавляет заголовок Access-Control-Allow-Origin
//   - Добавляет заголовок Access-Control-Allow-Methods
//   - Добавляет заголовок Access-Control-Allow-Headers
//   - Добавляет заголовок Access-Control-Allow-Credentials
//   - Обрабатывает OPTIONS-запросы (возвращает 200 OK)
//
// ПОЧЕМУ ВАЖНО:
//   - Без CORS фронтенд не сможет обращаться к API
//   - Если CORS настроен неправильно — запросы будут блокироваться браузером
//
// ============================================================
func TestCorsMiddleware(t *testing.T) {
	// 1. УСТАНАВЛИВАЕМ ПЕРЕМЕННУЮ ОКРУЖЕНИЯ
	//    CorsMiddleware использует SERVER_PORT для формирования Origin
	//    Устанавливаем порт 8087 (как в проекте)
	os.Setenv("SERVER_PORT", "8087")
	// Отложенная очистка: удаляем переменную после теста
	defer os.Unsetenv("SERVER_PORT")

	// 2. СОЗДАЁМ ТЕСТОВЫЙ ХЕНДЛЕР
	//    Это простой хендлер, который возвращает статус 200 OK
	//    Мы будем оборачивать его в CorsMiddleware
	next := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	// 3. ОБОРАЧИВАЕМ В CORSMIDDLEWARE
	//    Теперь любой запрос сначала проходит через CorsMiddleware,
	//    а потом попадает в наш хендлер
	handler := CorsMiddleware(next)

	// 4. ТЕСТОВЫЕ СЦЕНАРИИ
	//    Проверяем два типа запросов: GET и OPTIONS
	tests := []struct {
		name     string // Название теста
		method   string // HTTP-метод (GET, OPTIONS)
		path     string // Путь (не важен, но нужен для запроса)
		wantCode int    // Ожидаемый статус-код
	}{
		{
			name:     "GET request",
			method:   "GET",
			path:     "/api/test",
			wantCode: http.StatusOK,
		},
		{
			name:     "OPTIONS request (preflight)",
			method:   "OPTIONS",
			path:     "/api/test",
			wantCode: http.StatusOK,
		},
	}

	// 5. ЗАПУСКАЕМ ВСЕ ТЕСТЫ
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 5.1 СОЗДАЁМ HTTP-ЗАПРОС
			req := httptest.NewRequest(tt.method, tt.path, nil)

			// 5.2 СОЗДАЁМ RECORDER ДЛЯ ЗАПИСИ ОТВЕТА
			w := httptest.NewRecorder()

			// 5.3 ВЫЗЫВАЕМ ХЕНДЛЕР
			handler(w, req)

			// 5.4 ПРОВЕРЯЕМ СТАТУС-КОД
			if w.Code != tt.wantCode {
				t.Errorf("expected status %d, got %d", tt.wantCode, w.Code)
			}

			// 5.5 ПРОВЕРЯЕМ CORS-ЗАГОЛОВКИ
			//    Access-Control-Allow-Origin — разрешённый источник
			origin := w.Header().Get("Access-Control-Allow-Origin")
			if origin != "http://localhost:8087" {
				t.Errorf("expected 'http://localhost:8087', got '%s'", origin)
			}

			//    Access-Control-Allow-Methods — разрешённые методы
			methods := w.Header().Get("Access-Control-Allow-Methods")
			if methods == "" {
				t.Error("Access-Control-Allow-Methods header missing")
			}

			//    Access-Control-Allow-Headers — разрешённые заголовки
			headers := w.Header().Get("Access-Control-Allow-Headers")
			if headers == "" {
				t.Error("Access-Control-Allow-Headers header missing")
			}

			//    Access-Control-Allow-Credentials — разрешены ли куки
			credentials := w.Header().Get("Access-Control-Allow-Credentials")
			if credentials != "true" {
				t.Errorf("expected 'true', got '%s'", credentials)
			}
		})
	}
}

// ============================================================
// 2. ТЕСТ: CorsMiddleware с кастомным портом
// ============================================================
// Проверяет, что CorsMiddleware использует порт из переменной окружения
//
// ПОЧЕМУ ВАЖНО:
//   - Сервер может работать не на 8087, а на другом порту
//   - CORS должен разрешать тот порт, на котором работает сервер
//
// ============================================================
func TestCorsMiddleware_CustomPort(t *testing.T) {
	// 1. УСТАНАВЛИВАЕМ ДРУГОЙ ПОРТ
	os.Setenv("SERVER_PORT", "9090")
	defer os.Unsetenv("SERVER_PORT")

	// 2. СОЗДАЁМ ХЕНДЛЕР
	next := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	handler := CorsMiddleware(next)

	// 3. ВЫПОЛНЯЕМ ЗАПРОС
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	// 4. ПРОВЕРЯЕМ, ЧТО ORIGIN ИСПОЛЬЗУЕТ ПОРТ ИЗ ПЕРЕМЕННОЙ
	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin != "http://localhost:9090" {
		t.Errorf("expected 'http://localhost:9090', got '%s'", origin)
	}
}

// ============================================================
// 3. ТЕСТ: MetricsMiddleware
// ============================================================
// Что проверяет:
//   - MetricsMiddleware не блокирует запросы
//   - Запрос проходит через middleware и достигает хендлера
//   - Статус-код перехватывается правильно
//
// ПОЧЕМУ ВАЖНО:
//   - Без MetricsMiddleware не будет метрик RPS, задержек, ошибок
//   - Middleware не должен мешать работе хендлера
//
// ============================================================
func TestMetricsMiddleware(t *testing.T) {
	// 1. СОЗДАЁМ ТЕСТОВЫЙ ХЕНДЛЕР
	//    Возвращает статус 201 Created
	next := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}

	// 2. ОБОРАЧИВАЕМ В METRICSMIDDLEWARE
	handler := MetricsMiddleware(next)

	// 3. СОЗДАЁМ ЗАПРОС
	req := httptest.NewRequest("POST", "/api/subscriptions", nil)
	w := httptest.NewRecorder()

	// 4. ВЫЗЫВАЕМ ХЕНДЛЕР
	handler(w, req)

	// 5. ПРОВЕРЯЕМ, ЧТО СТАТУС ПРАВИЛЬНЫЙ
	//    Middleware должен пропустить статус от хендлера
	if w.Code != http.StatusCreated {
		t.Errorf("expected %d, got %d", http.StatusCreated, w.Code)
	}
}

// ============================================================
// 4. ТЕСТ: MetricsMiddleware с разными статусами
// ============================================================
// Проверяет, что MetricsMiddleware корректно перехватывает
// разные статус-коды.
//
// ПОЧЕМУ ВАЖНО:
//   - В метрики записывается статус-код
//   - Если статус перехватывается неправильно — метрики будут неверными
//
// ============================================================
func TestMetricsMiddleware_StatusCodes(t *testing.T) {
	// 1. ТЕСТОВЫЕ СЦЕНАРИИ: разные статус-коды
	tests := []struct {
		name       string // Название теста
		statusCode int    // Статус, который вернёт хендлер
	}{
		{"OK", http.StatusOK},
		{"Created", http.StatusCreated},
		{"Bad Request", http.StatusBadRequest},
		{"Not Found", http.StatusNotFound},
		{"Internal Server Error", http.StatusInternalServerError},
	}

	// 2. ЗАПУСКАЕМ ВСЕ ТЕСТЫ
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 2.1 СОЗДАЁМ ХЕНДЛЕР С УКАЗАННЫМ СТАТУСОМ
			next := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}

			// 2.2 ОБОРАЧИВАЕМ
			handler := MetricsMiddleware(next)

			// 2.3 ВЫПОЛНЯЕМ ЗАПРОС
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			handler(w, req)

			// 2.4 ПРОВЕРЯЕМ СТАТУС
			if w.Code != tt.statusCode {
				t.Errorf("expected %d, got %d", tt.statusCode, w.Code)
			}
		})
	}
}

// ============================================================
// 5. ТЕСТ: MetricsMiddleware с разными методами
// ============================================================
// Проверяет, что MetricsMiddleware работает с любыми HTTP-методами.
//
// ПОЧЕМУ ВАЖНО:
//   - Метрики собираются для всех методов
//   - Если метод не поддерживается — метрики будут неполными
//
// ============================================================
func TestMetricsMiddleware_Methods(t *testing.T) {
	// 1. ВСЕ HTTP-МЕТОДЫ, КОТОРЫЕ ИСПОЛЬЗУЮТСЯ В ПРОЕКТЕ
	methods := []string{"GET", "POST", "PUT", "DELETE"}

	// 2. ЗАПУСКАЕМ ТЕСТ ДЛЯ КАЖДОГО МЕТОДА
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			// 2.1 СОЗДАЁМ ХЕНДЛЕР
			next := func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}

			// 2.2 ОБОРАЧИВАЕМ
			handler := MetricsMiddleware(next)

			// 2.3 ВЫПОЛНЯЕМ ЗАПРОС
			req := httptest.NewRequest(method, "/test", nil)
			w := httptest.NewRecorder()
			handler(w, req)

			// 2.4 ПРОВЕРЯЕМ СТАТУС
			if w.Code != http.StatusOK {
				t.Errorf("expected %d, got %d", http.StatusOK, w.Code)
			}
		})
	}
}
