// ============================================================
// ФАЙЛ: internal/service/subscription_service_test.go
// ============================================================
// НАЗНАЧЕНИЕ: Юнит-тесты для сервиса подписок (subscription_service.go)
// ТЕГ: //go:build unit — запускается с флагом -tags=unit
//
// ЧТО ПРОВЕРЯЕТСЯ:
//   1. CreateSubscription — создание подписки через сервис
//   2. GetSubscriptionByID — получение подписки по ID
//   3. UpdateSubscription — обновление подписки
//   4. DeleteSubscription — удаление подписки
//   5. ListSubscriptions — получение списка подписок
//   6. GetTotalCost — расчёт суммарной стоимости
//
// ПОЧЕМУ МОКИ:
//   - Тесты не используют реальную БД → быстрые
//   - Не зависят от окружения → стабильные
//   - Проверяют только бизнес-логику сервиса
//
// КАК ЗАПУСТИТЬ:
//   go test ./internal/service -tags=unit -v
// ============================================================

//go:build unit

package service

import (
	"context"       // Для передачи контекста в методы
	"testing"       // Стандартный пакет для тестов
	"time"          // Для работы с датами
	"errors"



    "Rest-user-agregator/internal/cache"
    "Rest-user-agregator/internal/repository"
    "Rest-user-agregator/internal/service"
    "github.com/go-redis/redis/v9"
	"Rest-user-agregator/internal/models"       // Модели данных
)

// ============================================================
// МОК-РЕПОЗИТОРИЙ (заменяет реальную БД в тестах)
// ============================================================
// Что делает:
//   - Имитирует работу с БД без реального подключения
//   - Возвращает заранее заданные значения
//   - Позволяет тестировать сервис изолированно
//
// Почему используется:
//   - Тесты не зависят от PostgreSQL
//   - Тесты работают быстро (миллисекунды)
//   - Можно проверить все сценарии (успех, ошибка)
// ============================================================
type mockRepo struct {
	// Каждая функция-заглушка вызывается при соответствующем методе репозитория
	createFunc        func(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) (int, error)
	getByIDFunc       func(ctx context.Context, id int) (*models.Subscription, error)
	updateFunc        func(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) error
	deleteFunc        func(ctx context.Context, id int) error
	listFunc          func(ctx context.Context, limit, offset int) ([]models.Subscription, error)
	getTotalCostFunc  func(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error)
}

// ============================================================
// РЕАЛИЗАЦИЯ МЕТОДОВ ИНТЕРФЕЙСА SubscriptionRepository
// ============================================================
// Каждый метод проверяет, задана ли соответствующая функция-заглушка.
// Если задана — вызывает её и возвращает результат.
// Если не задана — возвращает значение по умолчанию или ошибку.
// ============================================================
var (
    originalCacheGet = cache.Get
    originalCacheSet = cache.Set
)
// CreateSubscription — заглушка для создания подписки
func (m *mockRepo) CreateSubscription(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) (int, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, sub, startDate, endDate)
	}
	return 1, nil // По умолчанию возвращаем ID = 1 и nil ошибку
}

// GetSubscriptionByID — заглушка для получения подписки по ID
func (m *mockRepo) GetSubscriptionByID(ctx context.Context, id int) (*models.Subscription, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	// По умолчанию возвращаем подписку с указанным ID
	return &models.Subscription{ID: id, ServiceName: "Test"}, nil
}

// UpdateSubscription — заглушка для обновления подписки
func (m *mockRepo) UpdateSubscription(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, sub, startDate, endDate)
	}
	return nil // По умолчанию — успех (nil ошибка)
}

// DeleteSubscription — заглушка для удаления подписки
func (m *mockRepo) DeleteSubscription(ctx context.Context, id int) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil // По умолчанию — успех
}

// ListSubscriptions — заглушка для получения списка подписок
func (m *mockRepo) ListSubscriptions(ctx context.Context, limit, offset int) ([]models.Subscription, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, limit, offset)
	}
	return []models.Subscription{}, nil // По умолчанию — пустой список
}

// GetTotalCost — заглушка для расчёта суммарной стоимости
func (m *mockRepo) GetTotalCost(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error) {
	if m.getTotalCostFunc != nil {
		return m.getTotalCostFunc(ctx, userID, serviceName, startDate, endDate)
	}
	return 0, nil // По умолчанию — 0
}

// ============================================================
// 1. ТЕСТ: СОЗДАНИЕ ПОДПИСКИ (CreateSubscription)
// ============================================================
// Что проверяет:
//   - Сервис правильно вызывает репозиторий
//   - Сервис возвращает ID из репозитория
//   - Сервис не меняет данные при передаче
//
// ПОЧЕМУ ВАЖНО:
//   - Без этого теста мы не знаем, работает ли создание подписок
//   - Проверяет передачу данных между сервисом и репозиторием
// ============================================================
func TestCreateSubscription(t *testing.T) {
	// 1. СОЗДАЁМ МОК-РЕПОЗИТОРИЙ
	//    Указываем, что при вызове CreateSubscription вернуть ID = 5
	repo := &mockRepo{
		createFunc: func(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) (int, error) {
			// Проверяем, что переданные данные не пустые
			if sub.ServiceName == "" {
				t.Error("ServiceName is empty")
			}
			return 5, nil // Возвращаем ID = 5 и nil ошибку
		},
	}
	// 2. СОЗДАЁМ СЕРВИС С МОКОМ
	//    Внедряем мок-репозиторий вместо реального
	svc := NewSubscriptionService(repo)

	// 3. ПОДГОТАВЛИВАЕМ ТЕСТОВЫЕ ДАННЫЕ
	sub := models.Subscription{
		ServiceName: "Test",
		Price:       100,
		UserID:      "550e8400-e29b-41d4-a716-446655440000",
		StartDate:   "07-2025",
	}

	// 4. ВЫЗЫВАЕМ МЕТОД СЕРВИСА
	id, err := svc.CreateSubscription(context.Background(), sub)

	// 5. ПРОВЕРЯЕМ РЕЗУЛЬТАТ
	if err != nil {
		t.Fatalf("CreateSubscription failed: %v", err)
	}
	if id != 5 {
		t.Errorf("Expected ID 5, got %d", id)
	}
}

// ============================================================
// 2. ТЕСТ: ПОЛУЧЕНИЕ ПОДПИСКИ ПО ID (GetSubscriptionByID)
// ============================================================
// Что проверяет:
//   - Сервис правильно вызывает репозиторий с переданным ID
//   - Сервис возвращает подписку из репозитория
//   - Данные подписки соответствуют ожидаемым
// ============================================================
func TestGetSubscriptionByID(t *testing.T) {
	// 1. СОЗДАЁМ МОК-РЕПОЗИТОРИЙ
	//    При вызове GetSubscriptionByID возвращаем подписку с ID = 10
	repo := &mockRepo{
		getByIDFunc: func(ctx context.Context, id int) (*models.Subscription, error) {
			// Проверяем, что передан правильный ID
			if id != 10 {
				t.Errorf("Expected ID 10, got %d", id)
			}
			return &models.Subscription{
				ID:          id,
				ServiceName: "TestGet",
				Price:       100,
			}, nil
		},
	}
	svc := NewSubscriptionService(repo)

	// 2. ВЫЗЫВАЕМ МЕТОД СЕРВИСА
	sub, err := svc.GetSubscriptionByID(context.Background(), 10)

	// 3. ПРОВЕРЯЕМ РЕЗУЛЬТАТ
	if err != nil {
		t.Fatalf("GetSubscriptionByID failed: %v", err)
	}
	if sub == nil {
		t.Fatal("Expected subscription, got nil")
	}
	if sub.ID != 10 {
		t.Errorf("Expected ID 10, got %d", sub.ID)
	}
	if sub.ServiceName != "TestGet" {
		t.Errorf("Expected 'TestGet', got '%s'", sub.ServiceName)
	}
}

// ============================================================
// 3. ТЕСТ: ОБНОВЛЕНИЕ ПОДПИСКИ (UpdateSubscription)
// ============================================================
// Что проверяет:
//   - Сервис правильно вызывает репозиторий с обновлёнными данными
//   - Сервис возвращает ошибку, если репозиторий вернул ошибку
// ============================================================
func TestUpdateSubscription(t *testing.T) {
	// 1. СОЗДАЁМ МОК-РЕПОЗИТОРИЙ
	//    При вызове UpdateSubscription проверяем переданные данные
	repo := &mockRepo{
		updateFunc: func(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) error {
			// Проверяем, что передан правильный ID и данные
			if sub.ID != 1 {
				t.Errorf("Expected ID 1, got %d", sub.ID)
			}
			if sub.ServiceName != "Updated" {
				t.Errorf("Expected 'Updated', got '%s'", sub.ServiceName)
			}
			return nil // Успех
		},
	}
	svc := NewSubscriptionService(repo)

	// 2. ПОДГОТАВЛИВАЕМ ТЕСТОВЫЕ ДАННЫЕ
	sub := models.Subscription{
		ID:          1,
		ServiceName: "Updated",
		Price:       200,
		UserID:      "550e8400-e29b-41d4-a716-446655440000",
		StartDate:   "08-2025",
		EndDate:     "12-2025",
	}

	// 3. ВЫЗЫВАЕМ МЕТОД СЕРВИСА
	err := svc.UpdateSubscription(context.Background(), sub)

	// 4. ПРОВЕРЯЕМ РЕЗУЛЬТАТ
	if err != nil {
		t.Fatalf("UpdateSubscription failed: %v", err)
	}
}

// ============================================================
// 4. ТЕСТ: УДАЛЕНИЕ ПОДПИСКИ (DeleteSubscription)
// ============================================================
// Что проверяет:
//   - Сервис правильно вызывает репозиторий с переданным ID
//   - Сервис возвращает ошибку, если удаление не удалось
// ============================================================
func TestDeleteSubscription(t *testing.T) {
	// 1. СОЗДАЁМ МОК-РЕПОЗИТОРИЙ
	repo := &mockRepo{
		deleteFunc: func(ctx context.Context, id int) error {
			// Проверяем, что передан правильный ID
			if id != 1 {
				t.Errorf("Expected ID 1, got %d", id)
			}
			return nil // Успех
		},
	}
	svc := NewSubscriptionService(repo)

	// 2. ВЫЗЫВАЕМ МЕТОД СЕРВИСА
	err := svc.DeleteSubscription(context.Background(), 1)

	// 3. ПРОВЕРЯЕМ РЕЗУЛЬТАТ
	if err != nil {
		t.Fatalf("DeleteSubscription failed: %v", err)
	}
}

// ============================================================
// 5. ТЕСТ: ПОЛУЧЕНИЕ СПИСКА ПОДПИСОК (ListSubscriptions)
// ============================================================
// Что проверяет:
//   - Сервис правильно вызывает репозиторий с параметрами пагинации
//   - Сервис возвращает список подписок из репозитория
// ============================================================
func TestListSubscriptions(t *testing.T) {
	// 1. СОЗДАЁМ МОК-РЕПОЗИТОРИЙ
	//    При вызове ListSubscriptions возвращаем список из 2 подписок
	repo := &mockRepo{
		listFunc: func(ctx context.Context, limit, offset int) ([]models.Subscription, error) {
			// Проверяем параметры пагинации
			if limit != 10 {
				t.Errorf("Expected limit 10, got %d", limit)
			}
			if offset != 0 {
				t.Errorf("Expected offset 0, got %d", offset)
			}
			return []models.Subscription{
				{ID: 1, ServiceName: "Test1"},
				{ID: 2, ServiceName: "Test2"},
			}, nil
		},
	}
	svc := NewSubscriptionService(repo)

	// 2. ВЫЗЫВАЕМ МЕТОД СЕРВИСА
	list, err := svc.ListSubscriptions(context.Background(), 10, 0)

	// 3. ПРОВЕРЯЕМ РЕЗУЛЬТАТ
	if err != nil {
		t.Fatalf("ListSubscriptions failed: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("Expected 2 subscriptions, got %d", len(list))
	}
	if list[0].ID != 1 {
		t.Errorf("Expected first ID 1, got %d", list[0].ID)
	}
}

// ============================================================
// 6. ТЕСТ: РАСЧЁТ СУММАРНОЙ СТОИМОСТИ (GetTotalCost)
// ============================================================
// Что проверяет:
//   - Сервис правильно вызывает репозиторий с параметрами
//   - Сервис возвращает сумму из репозитория
// ============================================================
func TestGetTotalCost(t *testing.T) {
	// 1. СОЗДАЁМ МОК-РЕПОЗИТОРИЙ
	//    При вызове GetTotalCost возвращаем 1500
	repo := &mockRepo{
		getTotalCostFunc: func(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error) {
			// Проверяем, что переданные параметры соответствуют ожиданиям
			if userID != "user1" {
				t.Errorf("Expected userID 'user1', got '%s'", userID)
			}
			if serviceName != "" {
				t.Errorf("Expected empty serviceName, got '%s'", serviceName)
			}
			return 1500, nil
		},
	}
	svc := NewSubscriptionService(repo)

	// 2. ВЫЗЫВАЕМ МЕТОД СЕРВИСА
	total, err := svc.GetTotalCost(
		context.Background(),
		"user1",   // userID
		"",        // serviceName (пусто)
		"01-2025", // startDate
		"12-2025", // endDate
	)

	// 3. ПРОВЕРЯЕМ РЕЗУЛЬТАТ
	if err != nil {
		t.Fatalf("GetTotalCost failed: %v", err)
	}
	if total != 1500 {
		t.Errorf("Expected 1500, got %d", total)
	}
}
// ============================================================
// ТЕСТЫ КЕШИРОВАНИЯ ДЛЯ GetTotalCost (ЮНИТ-ТЕСТЫ С МОКАМИ)
// ============================================================
//
// Эти тесты проверяют логику работы кеша без реального Redis.
// Все вызовы cache.Get и cache.Set заменены на моки.
// ============================================================

// TestGetTotalCost_CacheHit — проверяет, что при наличии кеша
// возвращается значение из Redis, а БД не вызывается.
//
// Сценарий:
//   1. Мокаем cache.Get на возврат 500 (кеш-попадание).
//   2. Мокаем репозиторий, чтобы он падал, если его вызовут.
//   3. Вызываем GetTotalCost.
//   4. Проверяем, что вернулось 500 и репозиторий не вызывался.
func TestGetTotalCost_CacheHit(t *testing.T) {
    // 1. СОЗДАЁМ МОК РЕПОЗИТОРИЯ
    mockRepo := &repository.MockSubRepo{}
    
    // 2. НАСТРАИВАЕМ МОК: ЕСЛИ БУДЕТ ВЫЗОВ — ТЕСТ УПАДЁТ
    //    Это гарантирует, что при кеш-попадании мы НЕ идём в БД.
    mockRepo.GetTotalCostMock = func(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error) {
        t.Error("Repository should not be called on cache hit")
        return 0, nil
    }

    // 3. МОКАЕМ cache.Get (возвращаем значение 500)
    //    cacheGetMock — переменная в пакете cache, которую мы подменяем в тесте.
    //    В реальном коде это делается через интерфейс, но для простоты используем
    //    глобальную переменную-заглушку.
    cache.Get = func(ctx context.Context, key string) (int, error) {
        return 500, nil
    }
    defer func() { cache.Get = originalCacheGet }() // Восстанавливаем оригинал

    // 4. ВЫЗЫВАЕМ GetTotalCost
    svc := service.NewSubscriptionService(mockRepo)
    result, err := svc.GetTotalCost(
        context.Background(),
        "test-user",
        "test-service",
        "01-2025",
        "12-2025",
    )

    // 5. ПРОВЕРЯЕМ РЕЗУЛЬТАТ
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    if result != 500 {
        t.Errorf("Expected 500, got %d", result)
    }
}

// TestGetTotalCost_CacheMiss — проверяет, что при кеш-промахе
// идём в БД и сохраняем результат в Redis.
//
// Сценарий:
//   1. Мокаем cache.Get на возврат ошибки redis.Nil (ключа нет).
//   2. Мокаем репозиторий на возврат 1000.
//   3. Мокаем cache.Set на успех.
//   4. Вызываем GetTotalCost.
//   5. Проверяем, что результат = 1000 и cache.Set был вызван.
func TestGetTotalCost_CacheMiss(t *testing.T) {
    mockRepo := &repository.MockSubRepo{}
    mockRepo.GetTotalCostMock = func(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error) {
        return 1000, nil
    }

    // 1. Мокаем cache.Get на ошибку redis.Nil (ключа нет)
    cache.Get = func(ctx context.Context, key string) (int, error) {
        return 0, redis.Nil
    }
    defer func() { cache.Get = originalCacheGet }()

    // 2. Мокаем cache.Set на успех (сохраняем результат)
    cacheSetCalled := false
    cache.Set = func(ctx context.Context, key string, value int, ttl time.Duration) error {
        cacheSetCalled = true
        return nil
    }
    defer func() { cache.Set = originalCacheSet }()

    // 3. ВЫЗЫВАЕМ GetTotalCost
    svc := service.NewSubscriptionService(mockRepo)
    result, err := svc.GetTotalCost(
        context.Background(),
        "test-user",
        "test-service",
        "01-2025",
        "12-2025",
    )

    // 4. ПРОВЕРЯЕМ РЕЗУЛЬТАТ
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    if result != 1000 {
        t.Errorf("Expected 1000, got %d", result)
    }
    if !cacheSetCalled {
        t.Error("Expected cache.Set to be called, but it wasn't")
    }
}

// TestGetTotalCost_RedisError — проверяет, что при ошибке Redis
// (не redis.Nil, а реальная ошибка) мы идём в БД и не падаем.
//
// Сценарий:
//   1. Мокаем cache.Get на ошибку (таймаут).
//   2. Мокаем репозиторий на успех.
//   3. Вызываем GetTotalCost.
//   4. Проверяем, что результат из БД, а ошибка Redis залогирована.
func TestGetTotalCost_RedisError(t *testing.T) {
    mockRepo := &repository.MockSubRepo{}
    mockRepo.GetTotalCostMock = func(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error) {
        return 2000, nil
    }

    // 1. Мокаем cache.Get на ошибку (таймаут)
    cache.Get = func(ctx context.Context, key string) (int, error) {
        return 0, errors.New("redis timeout")
    }
    defer func() { cache.Get = originalCacheGet }()

    // 2. ВЫЗЫВАЕМ GetTotalCost
    svc := service.NewSubscriptionService(mockRepo)
    result, err := svc.GetTotalCost(
        context.Background(),
        "test-user",
        "test-service",
        "01-2025",
        "12-2025",
    )

    // 3. ПРОВЕРЯЕМ РЕЗУЛЬТАТ (должны получить из БД)
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    if result != 2000 {
        t.Errorf("Expected 2000, got %d", result)
    }
}

// TestGetTotalCost_VersionError — проверяет, что при ошибке получения
// версии из БД кеш отключается и мы идём напрямую в БД.
//
// Сценарий:
//   1. Мокаем GetCacheUserVersion на ошибку.
//   2. Мокаем репозиторий на успех.
//   3. Вызываем GetTotalCost.
//   4. Проверяем, что результат из БД, а cache.Get НЕ вызывался.
func TestGetTotalCost_VersionError(t *testing.T) {
    mockRepo := &repository.MockSubRepo{}
    mockRepo.GetTotalCostMock = func(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error) {
        return 3000, nil
    }
    // Мокаем ошибку получения версии
    mockRepo.GetCacheUserVersionMock = func(ctx context.Context, userID string) (int, error) {
        return 0, errors.New("db error")
    }

    // 1. cache.Get НЕ ДОЛЖЕН вызываться — если вызовется, тест упадёт
    cache.Get = func(ctx context.Context, key string) (int, error) {
        t.Error("cache.Get should not be called when version retrieval fails")
        return 0, nil
    }
    defer func() { cache.Get = originalCacheGet }()

    // 2. ВЫЗЫВАЕМ GetTotalCost
    svc := service.NewSubscriptionService(mockRepo)
    result, err := svc.GetTotalCost(
        context.Background(),
        "test-user",
        "test-service",
        "01-2025",
        "12-2025",
    )

    // 3. ПРОВЕРЯЕМ РЕЗУЛЬТАТ (из БД)
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    if result != 3000 {
        t.Errorf("Expected 3000, got %d", result)
    }
}

// TestGetTotalCost_InvalidDate — проверяет обработку невалидных дат.
//
// Сценарий:
//   1. Передаём невалидную start_date.
//   2. Ожидаем ошибку, БД и кеш не трогаем.
func TestGetTotalCost_InvalidDate(t *testing.T) {
    mockRepo := &repository.MockSubRepo{}

    // 1. ВЫЗЫВАЕМ GetTotalCost с невалидной датой
    svc := service.NewSubscriptionService(mockRepo)
    result, err := svc.GetTotalCost(
        context.Background(),
        "test-user",
        "test-service",
        "invalid-date", // невалидный формат
        "12-2025",
    )

    // 2. ПРОВЕРЯЕМ, ЧТО ВЕРНУЛАСЬ ОШИБКА
    if err == nil {
        t.Error("Expected error for invalid date, got nil")
    }
    if result != 0 {
        t.Errorf("Expected 0, got %d", result)
    }
}