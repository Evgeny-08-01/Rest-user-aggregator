// ============================================================
// ФАЙЛ: internal/database/database_test.go
// ============================================================
// НАЗНАЧЕНИЕ: Интеграционные тесты для пакета database
// ТЕГ: //go:build integration — запускается с флагом -tags=integration
//
// ЧТО ПРОВЕРЯЕТСЯ:
//   1. CleanTestTable — очистка таблиц
//   2. CreateUser — создание пользователя
//   3. GetUserByEmail — поиск пользователя по email
//   4. DeleteSubscriptionsByUserID — удаление всех подписок пользователя
//
// КАК ЗАПУСТИТЬ:
//   go test ./internal/database -tags=integration -v
//
// ВАЖНО:
//   - Перед запуском должен быть поднят PostgreSQL
//   - Каждый тест начинается с очистки таблиц
// ============================================================

//go:build integration

package database

import (
	"context"       // Для передачи контекста в БД-запросы
	"os"            // Для завершения тестов с кодом
	"testing"       // Стандартный пакет для тестов
	"time"          // Для парсинга дат
	"fmt"

	"Rest-user-agregator/internal/models" // Наши модели данных
	"github.com/google/uuid"
)

// ============================================================
// 1. TestMain — НАСТРОЙКА ПЕРЕД ТЕСТАМИ
// ============================================================
// Выполняется один раз перед всеми тестами в пакете.
//
// ЧТО ДЕЛАЕТ:
//   1. Подключается к тестовой БД
//   2. Закрывает соединение после всех тестов
//   3. Очищает таблицы перед тестами
//   4. Запускает все тесты
// ============================================================
func TestMain(m *testing.M) {
	// 1. ПОДКЛЮЧЕНИЕ К БАЗЕ ДАННЫХ
	//    Используем локальную БД (порт 5432)
	//    Если БД не запущена — тесты упадут
	err := Init("postgres://postgres:1212@localhost:5432/subscriptions?sslmode=disable")
	if err != nil {
		// panic — аварийно завершаем, если БД не доступна
		panic("Failed to init DB: " + err.Error())
	}

	// 2. ОТЛОЖЕННОЕ ЗАКРЫТИЕ СОЕДИНЕНИЯ
	//    Выполнится после всех тестов, даже если они упадут
//	defer Close()

	// 3. ОЧИЩАЕМ ТАБЛИЦЫ ПЕРЕД ТЕСТАМИ
	//    Удаляем все записи, сбрасываем счётчик ID
	if err := CleanTestTable(); err != nil {
		panic("Failed to clean table: " + err.Error())
	}

	// 4. ЗАПУСК ВСЕХ ТЕСТОВ
	//    m.Run() возвращает код выхода (0 — успех, 1 — ошибка)
	code := m.Run()

	// 5. ВЫХОД С КОДОМ
	//    Передаём код результата в систему
	os.Exit(code)
}

// ============================================================
// 2. ТЕСТ: ОЧИСТКА ТАБЛИЦЫ (CleanTestTable)
// ============================================================
// Проверяет, что CleanTestTable удаляет все подписки из БД.
//
// ПОЧЕМУ ВАЖНО:
//   - Эта функция используется перед каждым тестом
//   - Если она не работает — тесты будут влиять друг на друга
// ============================================================
func TestCleanTestTable(t *testing.T) {
	// 1. СОЗДАЁМ РЕПОЗИТОРИЙ
	//    Это объект, через который мы работаем с БД
	repo := NewPostgresRepo()
	ctx := context.Background()

	// 2. СОЗДАЁМ ТЕСТОВУЮ ПОДПИСКУ
	sub := models.Subscription{
		ServiceName: "TestClean",
		Price:       100,
		UserID:      "550e8400-e29b-41d4-a716-446655440000",
		StartDate:   "01-2025",
	}
	startDate, _ := time.Parse("01-2006", sub.StartDate)

	// 3. ВЫЗЫВАЕМ МЕТОД СОЗДАНИЯ
	//    Возвращает ID созданной подписки
	id, err := repo.CreateSubscription(ctx, sub, startDate, nil)
	if err != nil {
		t.Fatalf("CreateSubscription failed: %v", err)
	}
	if id <= 0 {
		t.Fatal("Subscription not created")
	}

	// 4. ОЧИЩАЕМ ТАБЛИЦУ
	//    Должна удалить все подписки
	if err := CleanTestTable(); err != nil {
		t.Fatalf("CleanTestTable failed: %v", err)
	}

	// 5. ПРОВЕРЯЕМ, ЧТО ПОДПИСКА УДАЛЕНА
	//    GetSubscriptionByID должна вернуть nil (не найдено)
	saved, err := repo.GetSubscriptionByID(ctx, id)
	if err != nil {
		t.Fatalf("GetSubscriptionByID failed: %v", err)
	}
	if saved != nil {
		t.Error("Subscription still exists after clean")
	}
}

// ============================================================
// 3. ТЕСТ: СОЗДАНИЕ ПОЛЬЗОВАТЕЛЯ (CreateUser)
// ============================================================
// Проверяет, что пользователь создаётся и потом находится по email.
//
// ПОЧЕМУ ВАЖНО:
//   - Это основа для всей авторизации
//   - Без этого теста мы не знаем, работает ли регистрация
// ============================================================
func TestCreateUser(t *testing.T) {
	// 1. ОЧИЩАЕМ ТАБЛИЦУ ПЕРЕД ТЕСТОМ
	//    Чтобы не было конфликтов с данными от предыдущих тестов
	if err := CleanTestTable(); err != nil {
		t.Fatalf("CleanTestTable failed: %v", err)
	}

	repo := NewPostgresRepo()

	// 2. ПОДГОТАВЛИВАЕМ ТЕСТОВОГО ПОЛЬЗОВАТЕЛЯ
	//    ID — UUID, Email — уникальный, Password — хеш
    user := models.User{
    ID:       uuid.New().String(),
    Email:    fmt.Sprintf("test_%d@mail.com", time.Now().UnixNano()),
    Password: "hash123",
    Role:     "user",
}
	// 3. ВЫЗЫВАЕМ МЕТОД СОЗДАНИЯ
	//    context.Background() — пустой контекст для тестов
	err := repo.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// 4. ПРОВЕРЯЕМ, ЧТО ПОЛЬЗОВАТЕЛЬ СОХРАНИЛСЯ
	//    Ищем по email — если найден, значит создание прошло успешно
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

// ============================================================
// 4. ТЕСТ: ПОИСК ПОЛЬЗОВАТЕЛЯ ПО EMAIL (GetUserByEmail)
// ============================================================
// Проверяет, что GetUserByEmail возвращает nil для несуществующего email.
//
// ПОЧЕМУ ВАЖНО:
//   - Проверяет обработку случая "пользователь не найден"
//   - Без этого теста возможны ошибки при логине
// ============================================================
func TestGetUserByEmail_NotFound(t *testing.T) {
	// 1. ОЧИЩАЕМ ТАБЛИЦУ
	if err := CleanTestTable(); err != nil {
		t.Fatalf("CleanTestTable failed: %v", err)
	}

	repo := NewPostgresRepo()

	// 2. ИЩЕМ НЕСУЩЕСТВУЮЩЕГО ПОЛЬЗОВАТЕЛЯ
	//    Должен вернуть nil, nil (без ошибки)
	user, err := repo.GetUserByEmail(context.Background(), "notfound@mail.com")
	if err != nil {
		t.Fatalf("GetUserByEmail failed: %v", err)
	}
	if user != nil {
		t.Error("Expected nil, got user")
	}
}

// ============================================================
// 5. ТЕСТ: УДАЛЕНИЕ ВСЕХ ПОДПИСОК ПОЛЬЗОВАТЕЛЯ
// ============================================================
// Проверяет, что DeleteSubscriptionsByUserID удаляет все подписки
// только для одного пользователя.
//
// ПОЧЕМУ ВАЖНО:
//   - Используется в интеграционных тестах для изоляции данных
//   - Должен удалять только подписки указанного пользователя
// ============================================================
func TestDeleteSubscriptionsByUserID(t *testing.T) {
	// 1. ОЧИЩАЕМ ТАБЛИЦУ
	if err := CleanTestTable(); err != nil {
		t.Fatalf("CleanTestTable failed: %v", err)
	}

	repo := NewPostgresRepo()
	ctx := context.Background()
	userID := "550e8400-e29b-41d4-a716-446655440000"

	// 2. СОЗДАЁМ 2 ПОДПИСКИ ДЛЯ ПОЛЬЗОВАТЕЛЯ
	for i := 0; i < 2; i++ {
		sub := models.Subscription{
			ServiceName: "TestDelete",
			Price:       100,
			UserID:      userID,
			StartDate:   "01-2025",
		}
		startDate, _ := time.Parse("01-2006", sub.StartDate)
		_, err := repo.CreateSubscription(ctx, sub, startDate, nil)
		if err != nil {
			t.Fatalf("CreateSubscription failed: %v", err)
		}
	}

	// 3. ПРОВЕРЯЕМ, ЧТО ПОДПИСКИ СОЗДАЛИСЬ
	//    В списке должно быть 2 подписки
	list, err := repo.ListSubscriptions(ctx, 10, 0)
	if err != nil {
		t.Fatalf("ListSubscriptions failed: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("Expected 2 subscriptions, got %d", len(list))
	}

	// 4. УДАЛЯЕМ ПОДПИСКИ ПОЛЬЗОВАТЕЛЯ
	//    Должен удалить только подписки с указанным user_id
	err = DeleteSubscriptionsByUserID(userID)
	if err != nil {
		t.Fatalf("DeleteSubscriptionsByUserID failed: %v", err)
	}

	// 5. ПРОВЕРЯЕМ, ЧТО ПОДПИСКИ УДАЛЕНЫ
	//    В списке должно быть 0 подписок
	list, err = repo.ListSubscriptions(ctx, 10, 0)
	if err != nil {
		t.Fatalf("ListSubscriptions failed: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("Expected 0 subscriptions, got %d", len(list))
	}
}

// ============================================================
// ТЕСТ: ПОЛУЧЕНИЕ СОЕДИНЕНИЯ С БД (GetDB)
// ============================================================
// Что проверяет:
//   - GetDB() возвращает не nil
//   - Соединение с БД существует
//
// ПОЧЕМУ ВАЖНО:
//   - GetDB() используется в некоторых тестах и функциях
//   - Если вернёт nil — тесты упадут с паникой
// ============================================================
func TestGetDB(t *testing.T) {
	// 1. ВЫЗЫВАЕМ GetDB()
	db := GetDB()

	// 2. ПРОВЕРЯЕМ, ЧТО СОЕДИНЕНИЕ СУЩЕСТВУЕТ
	if db == nil {
		t.Error("GetDB returned nil")
	}
}

// ============================================================
// НАЗНАЧЕНИЕ: Интеграционные тесты для проверки работы с базой данных
//
// ЧТО ПРОВЕРЯЕТСЯ:
//   1. Создание пользователя (CreateUser)
//   2. Создание подписки (CreateSubscription)
//   3. Получение подписки по ID (GetSubscriptionByID)
//   4. Обновление подписки (UpdateSubscription)
//   5. Удаление подписки (DeleteSubscription)
//


// ============================================================
// 2. ТЕСТ: СОЗДАНИЕ ПОЛЬЗОВАТЕЛЯ
// ============================================================
// Проверяет, что пользователь создаётся в БД и потом находится по email.
//
// ПОЧЕМУ ВАЖНО:
//   Это основа для всей авторизации.
//   Если пользователь не создаётся — логин работать не будет.
// ============================================================
func TestCreateUser_Success(t *testing.T) {
	// 1. ОЧИЩАЕМ ТАБЛИЦУ
	//    Каждый тест должен начинаться с пустой БД,
	//    чтобы не было конфликтов с данными от предыдущих тестов.
	if err := CleanTestTable(); err != nil {
		t.Fatalf("Failed to clean table: %v", err)
	}

	// 2. СОЗДАЁМ РЕПОЗИТОРИЙ
	//    Это объект, через который мы работаем с БД.
	repo := NewPostgresRepo()

	// 3. ПОДГОТАВЛИВАЕМ ТЕСТОВОГО ПОЛЬЗОВАТЕЛЯ
	//    ID — UUID, Email — уникальный, Password — хеш,
	//    Role — 'user' или 'admin'
	user := models.User{
		ID:       "550e8400-e29b-41d4-a716-446655440000",
		Email:    "testuser@example.com",
		Password: "hashed_password_here",
		Role:     "user",
	}

	// 4. ВЫЗЫВАЕМ МЕТОД СОЗДАНИЯ
	//    context.Background() — пустой контекст (подходит для тестов)
	err := repo.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// 5. ПРОВЕРЯЕМ, ЧТО ПОЛЬЗОВАТЕЛЬ СОХРАНИЛСЯ
	//    Ищем по email — если найден, значит создание прошло успешно
	saved, err := repo.GetUserByEmail(context.Background(), user.Email)
	if err != nil {
		t.Fatalf("GetUserByEmail failed: %v", err)
	}
	if saved == nil {
		t.Fatal("User not found after creation")
	}
}

// ============================================================
// 3. ТЕСТ: СОЗДАНИЕ ПОДПИСКИ
// ============================================================
// Проверяет, что подписка создаётся и возвращается ID > 0.
//
// ПОЧЕМУ ВАЖНО:
//   Это основной CRUD-метод.
//   Без него пользователь не сможет добавить подписку.
// ============================================================
func TestCreateSubscription_Success(t *testing.T) {
	// 1. ОЧИЩАЕМ ТАБЛИЦУ
	if err := CleanTestTable(); err != nil {
		t.Fatalf("Failed to clean table: %v", err)
	}

	repo := NewPostgresRepo()
	ctx := context.Background()

	// 2. ПОДГОТАВЛИВАЕМ ДАННЫЕ ДЛЯ ПОДПИСКИ
	sub := models.Subscription{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      "550e8400-e29b-41d4-a716-446655440000",
		StartDate:   "07-2025",
	}

	// 3. ПАРСИМ ДАТУ НАЧАЛА
	//    Из строки "07-2025" в time.Time
	startDate, _ := time.Parse("01-2006", sub.StartDate)

	// 4. ВЫЗЫВАЕМ МЕТОД СОЗДАНИЯ
	//    Передаём: контекст, данные, дату начала, nil (нет даты окончания)
	id, err := repo.CreateSubscription(ctx, sub, startDate, nil)
	if err != nil {
		t.Fatalf("CreateSubscription failed: %v", err)
	}

	// 5. ПРОВЕРЯЕМ, ЧТО ID > 0
	//    В БД ID генерируется автоматически (SERIAL PRIMARY KEY)
	if id <= 0 {
		t.Errorf("Expected positive ID, got %d", id)
	}
}

// ============================================================
// 4. ТЕСТ: ПОЛУЧЕНИЕ ПОДПИСКИ ПО ID
// ============================================================
// Проверяет, что подписка находится по ID и данные совпадают.
//
// ПОЧЕМУ ВАЖНО:
//   Пользователь должен видеть свои подписки.
//   Если этот метод не работает — фронтенд будет пустым.
// ============================================================
func TestGetSubscriptionByID_Success(t *testing.T) {
	// 1. ОЧИЩАЕМ ТАБЛИЦУ
	if err := CleanTestTable(); err != nil {
		t.Fatalf("Failed to clean table: %v", err)
	}

	repo := NewPostgresRepo()
	ctx := context.Background()

	// 2. СОЗДАЁМ ПОДПИСКУ
	sub := models.Subscription{
		ServiceName: "Spotify",
		Price:       250,
		UserID:      "550e8400-e29b-41d4-a716-446655440000",
		StartDate:   "01-2025",
	}
	startDate, _ := time.Parse("01-2006", sub.StartDate)
	id, err := repo.CreateSubscription(ctx, sub, startDate, nil)
	if err != nil {
		t.Fatalf("CreateSubscription failed: %v", err)
	}

	// 3. ПОЛУЧАЕМ ПОДПИСКУ ПО ID
	saved, err := repo.GetSubscriptionByID(ctx, id)
	if err != nil {
		t.Fatalf("GetSubscriptionByID failed: %v", err)
	}

	// 4. ПРОВЕРЯЕМ, ЧТО ДАННЫЕ СОВПАДАЮТ
	if saved == nil {
		t.Fatal("Subscription not found")
	}
	if saved.ServiceName != sub.ServiceName {
		t.Errorf("Expected %s, got %s", sub.ServiceName, saved.ServiceName)
	}
}

// ============================================================
// 5. ТЕСТ: ОБНОВЛЕНИЕ ПОДПИСКИ
// ============================================================
// Проверяет, что подписка обновляется (меняется название и цена).
//
// ПОЧЕМУ ВАЖНО:
//   Пользователь может передумать и изменить подписку.
//   Без этого метода интерфейс станет неполным.
// ============================================================
func TestUpdateSubscription_Success(t *testing.T) {
	// 1. ОЧИЩАЕМ ТАБЛИЦУ
	if err := CleanTestTable(); err != nil {
		t.Fatalf("Failed to clean table: %v", err)
	}

	repo := NewPostgresRepo()
	ctx := context.Background()

	// 2. СОЗДАЁМ ПОДПИСКУ
	sub := models.Subscription{
		ServiceName: "Netflix",
		Price:       500,
		UserID:      "550e8400-e29b-41d4-a716-446655440000",
		StartDate:   "01-2025",
	}
	startDate, _ := time.Parse("01-2006", sub.StartDate)
	id, err := repo.CreateSubscription(ctx, sub, startDate, nil)
	if err != nil {
		t.Fatalf("CreateSubscription failed: %v", err)
	}

	// 3. ОБНОВЛЯЕМ ДАННЫЕ
	sub.ID = id
	sub.ServiceName = "Netflix Premium"
	sub.Price = 700

	// 4. ВЫЗЫВАЕМ МЕТОД ОБНОВЛЕНИЯ
	startDate, _ = time.Parse("01-2006", sub.StartDate)
	err = repo.UpdateSubscription(ctx, sub, startDate, nil)
	if err != nil {
		t.Fatalf("UpdateSubscription failed: %v", err)
	}

	// 5. ПРОВЕРЯЕМ, ЧТО ДАННЫЕ ИЗМЕНИЛИСЬ
	saved, err := repo.GetSubscriptionByID(ctx, id)
	if err != nil {
		t.Fatalf("GetSubscriptionByID failed: %v", err)
	}
	if saved.ServiceName != "Netflix Premium" {
		t.Errorf("Expected Netflix Premium, got %s", saved.ServiceName)
	}
	if saved.Price != 700 {
		t.Errorf("Expected 700, got %d", saved.Price)
	}
}

// ============================================================
// 6. ТЕСТ: УДАЛЕНИЕ ПОДПИСКИ
// ============================================================
// Проверяет, что подписка удаляется и больше не находится в БД.
//
// ПОЧЕМУ ВАЖНО:
//   Пользователь должен иметь возможность удалить подписку.
//   Без этого метода CRUD будет неполным.
// ============================================================
func TestDeleteSubscription_Success(t *testing.T) {
	// 1. ОЧИЩАЕМ ТАБЛИЦУ
	if err := CleanTestTable(); err != nil {
		t.Fatalf("Failed to clean table: %v", err)
	}

	repo := NewPostgresRepo()
	ctx := context.Background()

	// 2. СОЗДАЁМ ПОДПИСКУ
	sub := models.Subscription{
		ServiceName: "ToDelete",
		Price:       100,
		UserID:      "550e8400-e29b-41d4-a716-446655440000",
		StartDate:   "01-2025",
	}
	startDate, _ := time.Parse("01-2006", sub.StartDate)
	id, err := repo.CreateSubscription(ctx, sub, startDate, nil)
	if err != nil {
		t.Fatalf("CreateSubscription failed: %v", err)
	}

	// 3. УДАЛЯЕМ ПОДПИСКУ
	err = repo.DeleteSubscription(ctx, id)
	if err != nil {
		t.Fatalf("DeleteSubscription failed: %v", err)
	}

	// 4. ПРОВЕРЯЕМ, ЧТО ПОДПИСКА НЕ НАХОДИТСЯ
	saved, err := repo.GetSubscriptionByID(ctx, id)
	if err != nil {
		t.Fatalf("GetSubscriptionByID failed: %v", err)
	}
	if saved != nil {
		t.Error("Subscription still exists after deletion")
	}
}

// ============================================================
// ТЕСТ: ПРИМЕНЕНИЕ МИГРАЦИЙ (RunMigrations)
// ============================================================
// Что проверяет:
//   - RunMigrations() не возвращает ошибку
//   - Миграции применяются без паники
//
// ПОЧЕМУ ВАЖНО:
//   - Без миграций таблицы не создаются
//   - Если миграции не работают — приложение не запустится
// ============================================================
func TestRunMigrations(t *testing.T) {
    // 1. ЛОГИРУЕМ НАЧАЛО ТЕСТА
    t.Log("Starting TestRunMigrations")

    // 2. ОЧИЩАЕМ ТАБЛИЦУ ПЕРЕД ТЕСТОМ
    //    Чтобы начать с чистого состояния
    t.Log("Cleaning tables before migration test")
    if err := CleanTestTable(); err != nil {
        t.Logf("CleanTestTable failed: %v", err)
        t.Fatalf("CleanTestTable failed: %v", err)
    }
    t.Log("Tables cleaned successfully")

    // 3. ВЫЗЫВАЕМ RunMigrations()
    //    Должна применить все миграции без ошибок
    t.Log("Running migrations...")
    err := RunMigrations()
    if err != nil {
        t.Logf("RunMigrations failed: %v", err)
        t.Errorf("RunMigrations failed: %v", err)
        return
    }
    t.Log("Migrations applied successfully")
}

// ============================================================
// ТЕСТ: ОТКАТ МИГРАЦИЙ (RollbackMigrations)
// ============================================================
// Что проверяет:
//   - RollbackMigrations() не возвращает ошибку
//   - Откат проходит без паники
//
// ПОЧЕМУ ВАЖНО:
//   - Откат нужен для восстановления базы
//   - Без этого теста непонятно, работает ли откат
// ============================================================
func TestRollbackMigrations(t *testing.T) {
    // 1. ЛОГИРУЕМ НАЧАЛО ТЕСТА
    t.Log("Starting TestRollbackMigrations")

    // 2. ОЧИЩАЕМ ТАБЛИЦУ ПЕРЕД ТЕСТОМ
    t.Log("Cleaning tables before rollback test")
    if err := CleanTestTable(); err != nil {
        t.Logf("CleanTestTable failed: %v", err)
        t.Fatalf("CleanTestTable failed: %v", err)
    }
    t.Log("Tables cleaned successfully")

    // 3. СНАЧАЛА ПРИМЕНЯЕМ МИГРАЦИИ
    t.Log("Applying migrations before rollback...")
    if err := RunMigrations(); err != nil {
        t.Logf("RunMigrations failed: %v", err)
        t.Fatalf("RunMigrations failed: %v", err)
    }
    t.Log("Migrations applied, proceeding to rollback")

    // 4. ВЫЗЫВАЕМ RollbackMigrations()
    //    Должна откатить все миграции без ошибок
    t.Log("Rolling back migrations...")
    err := RollbackMigrations()
    if err != nil {
        t.Logf("RollbackMigrations failed: %v", err)
        t.Errorf("RollbackMigrations failed: %v", err)
        return
    }
    t.Log("Migrations rolled back successfully")
}
