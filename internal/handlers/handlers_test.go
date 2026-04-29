// internal/handlers/handlers_test.go
package handlers

import (
	"bytes"
	"encoding/json"
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
	_, err = database.DB.Exec("TRUNCATE subscriptions RESTART IDENTITY")
	if err != nil {
		panic("Failed to truncate table: " + err.Error())
	}
	os.Exit(m.Run()) //os.Exit(...)	Немедленно завершает программу с указанным кодом
	                 //m.Run()	Запускает все тесты в этом пакете и возвращает код (0 — успех, 1 — провал)
}

func TestCreateSubscriptionHandler(t *testing.T) {// Тестирование функции создания подписки
	tests := []struct {   // Создаем структуру для тестов
		name       string // Создаем поле name для названия теста
		body       string // Создаем поле body для тела запроса
		wantStatus int    // Создаем поле wantStatus для ожидаемого статуса ответа
	}{
		{"success", `{"service_name":"Test","price":100,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"07-2025"}`, http.StatusCreated},
		{"empty service_name", `{"service_name":"","price":100,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"07-2025"}`, http.StatusBadRequest},
		{"negative price", `{"service_name":"Test","price":-10,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"07-2025"}`, http.StatusBadRequest},
		{"empty user_id", `{"service_name":"Test","price":100,"user_id":"","start_date":"07-2025"}`, http.StatusBadRequest},
		{"invalid date", `{"service_name":"Test","price":100,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"2025-07"}`, http.StatusBadRequest},
		{"invalid JSON", `{"service_name":}`, http.StatusBadRequest},
	}

	for _, tt := range tests {                  // Создаем цикл для тестов
		t.Run(tt.name, func(t_sub *testing.T) { // t.Run в цикле создаёт под-тесты с именем tt.name из структуры tests
             // t_sub - указатель на структуру testing.T, которая используется для вывода сообщений об ошибках в  t_sub.Errorf
			req := httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(tt.body)))  // Создаем запрос для 
			                                                                                          // тестируемого хэндлера
								// bytes.NewReader([]byte(tt.body))-переводит тело запроса tt.body из структуры tests в байты
		w := httptest.NewRecorder()  // Создаем указатель на структуру ResponseRecorder-структура из стандартной библиотеки 
		                             // Go, пакет net/http/httptest-специально создана для тестирования HTTP-хендлеров.
			CreateSubscriptionHandler(w, req)   // Вызываем тестируемый хэндлер
			if w.Code != tt.wantStatus {        // Сравниваем полученный статус ответа w.Code с ожидаемым tt.wantStatus
				t_sub.Errorf("got %d, want %d", w.Code, tt.wantStatus) //Errorf-метод структуры testing.T,
				                                    //  который выводит сообщение об ошибке в консоль
			} 
		})  
	} 
}

func TestGetSubscriptionHandler(t *testing.T) {// Функция тестирования хэндлера получения подписки
	createBody := `{"service_name":"TestGet","price":100,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"07-2025"}`
	req := httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(createBody)))// Создаем запрос для 
			                                                                                    // тестируемого хэндлера
		w := httptest.NewRecorder()  // Создаем указатель на структуру ResponseRecorderструктура из стандартной библиотеки 
		                             // Go, пакет net/http/httptest-специально создана для тестирования HTTP-хендлеров.
	CreateSubscriptionHandler(w, req)     // Вызываем тестируемую функцию

	var resp map[string]int               // Создаем переменную resp(мапа) для хранения ответа
	json.NewDecoder(w.Body).Decode(&resp) // Декодируем ответ w.Body в переменную resp
	id := resp["id"]// Получаем id из ответа

	t.Run("success", func(t_sub *testing.T) { // Создаем подтест
		req := httptest.NewRequest("GET", "/api/subscriptions/{id}", nil) // Создаем запрос для тестируемого хэндлера
		req.SetPathValue("id", strconv.Itoa(id))                          // устанавливаем значение id в путь запроса 
		                                                                  // "/api/subscriptions/{id}"
			w := httptest.NewRecorder()  // Создаем указатель на структуру ResponseRecorderструктура из стандартной библиотеки 
		                             // Go, пакет net/http/httptest-специально создана для тестирования HTTP-хендлеров.	
									 // Новый w создаётся в каждом под-тесте, чтобы каждый запрос получал чистый ResponseRecorder.														  
		GetSubscriptionHandler(w, req) // Вызываем тестируемую функцию
		if w.Code != http.StatusOK {
			t_sub.Errorf("got %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("not found", func(t_sub *testing.T) {// Создаем подтест с несуществующим id
		req := httptest.NewRequest("GET", "/api/subscriptions/{id}", nil)// Создаем запрос для тестируемого хэндлера
		req.SetPathValue("id", "99999") // Несуществующий id		w := httptest.NewRecorder()
   w := httptest.NewRecorder() //Создаем указатель на структуру ResponseRecorderструктура из стандартной библиотеки 
		                       // Go, пакет net/http/httptest-специально создана для тестирования HTTP-хендлеров.	
							   // Новый w создаётся в каждом под-тесте, чтобы каждый запрос получал чистый ResponseRecorder.
		GetSubscriptionHandler(w, req)  // Вызываем тестируемый хендлер
		if w.Code != http.StatusNotFound {
			t_sub.Errorf("got %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid id", func(t_sub *testing.T) {// Создаем подтест с невалидным id
		req := httptest.NewRequest("GET", "/api/subscriptions/{id}", nil)// Создаем запрос для тестируемого хэндлера
		req.SetPathValue("id", "abc") // Невалидный id
    w := httptest.NewRecorder() // Создаем указатель на структуру ResponsRecorderструктура из стандартной библиотеки 
		                        // Go, пакет net/http/httptest-специально создана для тестирования HTTP-хендлеров.	
								// Новый w создаётся в каждом под-тесте, чтобы каждый запрос получал чистый ResponseRecorder.					
		GetSubscriptionHandler(w, req)
		if w.Code != http.StatusBadRequest {
			t_sub.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestUpdateSubscriptionHandler(t *testing.T) { // Функция тестирования хэндлера обновления подписки
	createBody := `{"service_name":"BeforeUpdate","price":100,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"07-2025"}`
	req := httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(createBody))) //Создаем запрос для
	                                                                                              //тестируемого хэндлера
  w := httptest.NewRecorder() //Создаем указатель на структуру ResponsRecorderструктура из стандартной библиотеки 
		                       // Go, пакет net/http/httptest-специально создана для тестирования HTTP-хендлеров.	
							   // Новый w создаётся в каждом под-тесте, чтобы каждый запрос получал чистый ResponseRecorder.
	CreateSubscriptionHandler(w, req) // Создаем реальную подписку БД, вызывая соотвествующий хендлер

	var resp map[string]int                  // Создаем переменную resp(мапа) для хранения ответа
	json.NewDecoder(w.Body).Decode(&resp)    // Декодируем ответ w.Body в переменную resp
	id := resp["id"] // Получаем id из ответа

	t.Run("success", func(t *testing.T) {  // Создаем подтест с валидными данными
		updateBody := `{"service_name":"AfterUpdate","price":200,"user_id":"550e8400-e29b-41d4-a716-446655440000",
		"start_date":"08-2025","end_date":"12-2025"}` // Валидные данные для обновления
		req := httptest.NewRequest("PUT", "/api/subscriptions/{id}", bytes.NewReader([]byte(updateBody)))// Создаем запрос для 
			                                                                                          // тестируемого хэндлера
		req.SetPathValue("id", strconv.Itoa(id))// Задаем валидный id
  w := httptest.NewRecorder() //Создаем указатель на структуру ResponsRecorderструктура из стандартной библиотеки 
		                       // Go, пакет net/http/httptest-специально создана для тестирования HTTP-хендлеров.	
							   // Новый w создаётся в каждом под-тесте, чтобы каждый запрос получал чистый ResponseRecorder.
		UpdateSubscriptionHandler(w, req) // Вызываем тестируемый хендлер
		if w.Code != http.StatusOK {
			t.Errorf("got %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("invalid id", func(t *testing.T) { // Создаем подтест с невалидным id
		updateBody := `{"service_name":"Test","price":100,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"07-2025"}`
		req := httptest.NewRequest("PUT", "/api/subscriptions/{id}", bytes.NewReader([]byte(updateBody))) // Создаем запрос для 
			                                                                                          // тестируемого хэндлера
		req.SetPathValue("id", "abc")// Задаем невалидный id
 w := httptest.NewRecorder() //Создаем указатель на структуру ResponsRecorderструктура из стандартной библиотеки 
		                       // Go, пакет net/http/httptest-специально создана для тестирования HTTP-хендлеров.	
							   // Новый w создаётся в каждом под-тесте, чтобы каждый запрос получал чистый ResponseRecorder.
		UpdateSubscriptionHandler(w, req)// Вызываем тестируемый хендлер
		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestDeleteSubscriptionHandler(t *testing.T) { // Функция тестирования хэндлера удаления подписки
	createBody := `{"service_name":"ToDelete","price":100,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"07-2025"}`
	req := httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(createBody)))// Создаем запрос для 
			                                                                                    // тестируемого хэндлера
  w := httptest.NewRecorder() //Создаем указатель на структуру ResponsRecorderструктура из стандартной библиотеки 
		                       // Go, пакет net/http/httptest-специально создана для тестирования HTTP-хендлеров.	
							   // Новый w создаётся в каждом под-тесте, чтобы каждый запрос получал чистый ResponseRecorder.
	CreateSubscriptionHandler(w, req) // Создаем реальную подписку БД, вызывая соотвествующий хендлер

	var resp map[string]int// Создаем переменную resp(мапа) для хранения ответа
	json.NewDecoder(w.Body).Decode(&resp)// Декодируем ответ в мапу
	id := resp["id"] // Получаем id из ответа

	t.Run("success", func(t *testing.T) { // Создаем подтест с валидными данными
		req := httptest.NewRequest("DELETE", "/api/subscriptions/{id}", nil) // Создаем запрос для 
			                                                                // тестируемого хэндлера
		req.SetPathValue("id", strconv.Itoa(id)) // Задаем валидный id
w := httptest.NewRecorder() //Создаем указатель на структуру ResponsRecorderструктура из стандартной библиотеки 
		                    // Go, пакет net/http/httptest-специально создана для тестирования HTTP-хендлеров.	
							// Новый w создаётся в каждом под-тесте, чтобы каждый запрос получал чистый ResponseRecorder.
		DeleteSubscriptionHandler(w, req) // Вызываем тестируемый хендлер
		if w.Code != http.StatusOK {
			t.Errorf("got %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("invalid id", func(t *testing.T) { // Создаем подтест с невалидным id
		req := httptest.NewRequest("DELETE", "/api/subscriptions/{id}", nil) // Создаем запрос для 
			                                                                 // тестируемого хэндлера
		req.SetPathValue("id", "abc")// Задаем невалидный id
		w := httptest.NewRecorder()  // Создаем указатель на структуру ResponsRecorderструктура из стандартной библиотеки 
		                             // Go, пакет net/http/httptest-специально создана для тестирования HTTP-хендлеров.


		DeleteSubscriptionHandler(w, req) // Вызываем тестируемый хендлер
		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestListSubscriptionsHandler(t *testing.T) { // Функция тестирования хэндлера получения списка подписок
	// Очистка перед тестом
	database.DB.Exec("DELETE FROM subscriptions WHERE user_id = '550e8400-e29b-41d4-a716-446655440001'")// Удаляем все 
	                                                      // подписки с user_id = 550e8400-e29b-41d4-a716-446655440001

	for i := 1; i <= 3; i++ {// Создаем 3 одинаковые подписки. отличаются только id
		body := `{"service_name":"ListTest","price":100,"user_id":"550e8400-e29b-41d4-a716-446655440001","start_date":"07-2025"}`
		req := httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(body)))// Создаем запрос для 
			                                                                                    // тестируемого хэндлера
w := httptest.NewRecorder() //Создаем указатель на структуру ResponsRecorderструктура из стандартной библиотеки 
		                       // Go, пакет net/http/httptest-специально создана для тестирования HTTP-хендлеров.	
							   // Новый w создаётся в каждом под-тесте, чтобы каждый запрос получал чистый ResponseRecorder.
		CreateSubscriptionHandler(w, req)// Создаем реальную подписку БД, вызывая соотвествующий хендлер
	}

	t.Run("success", func(t *testing.T) {// Создаем подтест с валидными данными
		req := httptest.NewRequest("GET", "/api/subscriptions", nil)// Создаем запрос для 
			                                                         // тестируемого хэндлера
	w := httptest.NewRecorder() //Создаем указатель на структуру ResponsRecorderструктура из стандартной библиотеки 
		                       // Go, пакет net/http/httptest-специально создана для тестирования HTTP-хендлеров.	
							   // Новый w создаётся в каждом под-тесте, чтобы каждый запрос получал чистый ResponseRecorder.
		ListSubscriptionsHandler(w, req)// Вызываем тестируемый хендлер
		if w.Code != http.StatusOK {
			t.Errorf("got %d, want %d", w.Code, http.StatusOK)
		}

		var list []models.Subscription// Создаем пустой список подписок на основе структуры models.Subscription
		json.NewDecoder(w.Body).Decode(&list)//Заполняем list данными из тела ответа
		if len(list) < 3 {// Проверяем, что в списке есть 3 подписки
			t.Errorf("expected at least 3, got %d", len(list))
		}
	})
}

func TestGetTotalCostHandler(t *testing.T) {
	// Очистка перед тестом
	database.DB.Exec("DELETE FROM subscriptions WHERE user_id = '550e8400-e29b-41d4-a716-446655440002'")//Удаляем все 
	                                                      // подписки с user_id = 550e8400-e29b-41d4-a716-446655440001
	bodies := []string{// Создаем список подписок
		`{"service_name":"Cost1","price":100,"user_id":"550e8400-e29b-41d4-a716-446655440002","start_date":"01-2025"}`,
		`{"service_name":"Cost2","price":200,"user_id":"550e8400-e29b-41d4-a716-446655440002","start_date":"02-2025"}`,
		`{"service_name":"Cost3","price":300,"user_id":"550e8400-e29b-41d4-a716-446655440002","start_date":"03-2025"}`,
	}
	for _, body := range bodies {//Цикл по  подпискам
		req := httptest.NewRequest("POST", "/api/subscriptions", bytes.NewReader([]byte(body)))//Создаем запрос для 
		                                                                                       // тестируемого хэндлера
		w := httptest.NewRecorder()//Создаем указатель на структуру ResponsRecorderструктура из стандартной библиотеки
		CreateSubscriptionHandler(w, req)//Вызываем тестируемый хендлер
	}

	t.Run("total cost", func(t *testing.T) {// Создаем подтест с валидными данными
		req := httptest.NewRequest("GET", "/api/subscriptions/total-cost?user_id=550e8400-e29b-41d4-a716-446655440002", nil)
w := httptest.NewRecorder() //Создаем указатель на структуру ResponsRecorderструктура из стандартной библиотеки 
		                       // Go, пакет net/http/httptest-специально создана для тестирования HTTP-хендлеров.	
							   // Новый w создаётся в каждом под-тесте, чтобы каждый запрос получал чистый ResponseRecorder.
		GetTotalCostHandler(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("got %d, want %d", w.Code, http.StatusOK)
		}

		var resp map[string]int// Создаем пустой список подписок на основе структуры models.Subscription
		json.NewDecoder(w.Body).Decode(&resp)//Заполняем resp данными из тела ответа
		if resp["total"] != 600 {// Проверяем, что общая стоимость равна 600
			t.Errorf("expected 600, got %d", resp["total"])
		}
	})

	t.Run("with date filter", func(t *testing.T) {// Создаем подтест с валидными данными  с фильтром
	                                              //  по дате(1 запись  из 3х не попадает в диапазон)
		req := httptest.NewRequest("GET", "/api/subscriptions/total-cost?user_id=550e8400-e29b-41d4-a716-446655440002&start_date=02-2025&end_date=03-2025", nil)
w := httptest.NewRecorder() //Создаем указатель на структуру ResponsRecorderструктура из стандартной библиотеки 
		                       // Go, пакет net/http/httptest-специально создана для тестирования HTTP-хендлеров.	
							   // Новый w создаётся в каждом под-тесте, чтобы каждый запрос получал чистый ResponseRecorder.
		GetTotalCostHandler(w, req)//Вызываем тестируемый хендлер
		if w.Code != http.StatusOK {// Проверяем, что статус ответа 200
			t.Errorf("got %d, want %d", w.Code, http.StatusOK)
		}

		var resp map[string]int// Создаем пустой список подписок на основе структуры models.Subscription
		json.NewDecoder(w.Body).Decode(&resp)//Заполняем resp данными из тела ответа
		if resp["total"] != 500 {
			t.Errorf("expected 500, got %d", resp["total"])
		}
	})
}
