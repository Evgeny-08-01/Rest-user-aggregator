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

//g o:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"Rest-user-agregator/internal/models"
)

// ============================================================
// МОК ДЛЯ КЕША (заменяет реальный Redis в тестах)
// ============================================================
type mockCache struct {
    getFunc func(ctx context.Context, key string) (int, error)
    setFunc func(ctx context.Context, key string, value int, ttl time.Duration) error
    deleteFunc func(ctx context.Context, key string) error
    keysFunc func(ctx context.Context, pattern string) ([]string, error)
}

// Get — реализация интерфейса Cache
func (m *mockCache) Get(ctx context.Context, key string) (int, error) {
    if m.getFunc != nil {
        return m.getFunc(ctx, key)
    }
    return 0, nil
}

// Set — реализация интерфейса Cache
func (m *mockCache) Set(ctx context.Context, key string, value int, ttl time.Duration) error {
    if m.setFunc != nil {
        return m.setFunc(ctx, key, value, ttl)
    }
    return nil
}

// Delete — реализация интерфейса Cache
func (m *mockCache) Delete(ctx context.Context, key string) error {
    if m.deleteFunc != nil {
        return m.deleteFunc(ctx, key)
    }
    return nil
}

// Keys — реализация интерфейса Cache
func (m *mockCache) Keys(ctx context.Context, pattern string) ([]string, error) {
    if m.keysFunc != nil {
        return m.keysFunc(ctx, pattern)
    }
    return []string{}, nil
}
// ============================================================
// МОК-РЕПОЗИТОРИЙ (заменяет реальную БД в тестах)
// ============================================================
type mockRepo struct {
	// Каждая функция-заглушка вызывается при соответствующем методе репозитория
	createFunc        func(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) (int, error)
	getByIDFunc       func(ctx context.Context, id int) (*models.Subscription, error)
	updateFunc        func(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) error
	deleteFunc        func(ctx context.Context, id int) error
	listFunc          func(ctx context.Context, userID string, limit, offset int) ([]models.Subscription, error)
	getTotalCostFunc  func(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error)
	getCacheUserVersionFunc    func(ctx context.Context, userID string) (int, error)
	incrementCacheUserVersionFunc func(ctx context.Context, userID string) error
}

// ============================================================
// РЕАЛИЗАЦИЯ МЕТОДОВ ИНТЕРФЕЙСА SubscriptionRepository
// ============================================================

// CreateSubscription — заглушка для создания подписки
func (m *mockRepo) CreateSubscription(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) (int, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, sub, startDate, endDate)
	}
	return 1, nil
}

// GetSubscriptionByID — заглушка для получения подписки по ID
func (m *mockRepo) GetSubscriptionByID(ctx context.Context, id int) (*models.Subscription, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return &models.Subscription{ID: id, ServiceName: "Test"}, nil
}

// UpdateSubscription — заглушка для обновления подписки
func (m *mockRepo) UpdateSubscription(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, sub, startDate, endDate)
	}
	return nil
}

// DeleteSubscription — заглушка для удаления подписки
func (m *mockRepo) DeleteSubscription(ctx context.Context, id int) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

// ListSubscriptions — заглушка для получения списка подписок
func (m *mockRepo) ListSubscriptions(ctx context.Context, userID string, limit, offset int) ([]models.Subscription, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, userID, limit, offset)
	}
	return []models.Subscription{}, nil
}

// GetTotalCost — заглушка для расчёта суммарной стоимости
func (m *mockRepo) GetTotalCost(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error) {
	if m.getTotalCostFunc != nil {
		return m.getTotalCostFunc(ctx, userID, serviceName, startDate, endDate)
	}
	return 0, nil
}

// GetCacheUserVersion — заглушка для получения версии кеша
func (m *mockRepo) GetCacheUserVersion(ctx context.Context, userID string) (int, error) {
	if m.getCacheUserVersionFunc != nil {
		return m.getCacheUserVersionFunc(ctx, userID)
	}
	return 1, nil
}

// IncrementCacheUserVersion — заглушка для инкремента версии кеша
func (m *mockRepo) IncrementCacheUserVersion(ctx context.Context, userID string) error {
	if m.incrementCacheUserVersionFunc != nil {
		return m.incrementCacheUserVersionFunc(ctx, userID)
	}
	return nil
}

// ============================================================
// 1. ТЕСТ: СОЗДАНИЕ ПОДПИСКИ (CreateSubscription)
// ============================================================
func TestCreateSubscription(t *testing.T) {
	repo := &mockRepo{
		createFunc: func(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) (int, error) {
			if sub.ServiceName == "" {
				t.Error("ServiceName is empty")
			}
			return 5, nil
		},
	}
	svc := NewSubscriptionService(repo)

	sub := models.Subscription{
		ServiceName: "Test",
		Price:       100,
		UserID:      "550e8400-e29b-41d4-a716-446655440000",
		StartDate:   "07-2025",
	}

	id, err := svc.CreateSubscription(context.Background(), sub)

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
func TestGetSubscriptionByID(t *testing.T) {
	repo := &mockRepo{
		getByIDFunc: func(ctx context.Context, id int) (*models.Subscription, error) {
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

	sub, err := svc.GetSubscriptionByID(context.Background(), 10)

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
func TestUpdateSubscription(t *testing.T) {
    repo := &mockRepo{
        updateFunc: func(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) error {
            if sub.ID != 1 {
                t.Errorf("Expected ID 1, got %d", sub.ID)
            }
            if sub.ServiceName != "Updated" {
                t.Errorf("Expected 'Updated', got '%s'", sub.ServiceName)
            }
            return nil
        },
    }
    svc := NewSubscriptionService(repo)

    sub := models.Subscription{
        ID:          1,
        ServiceName: "Updated",
        Price:       200,
        UserID:      "550e8400-e29b-41d4-a716-446655440000",
        StartDate:   "12-2026", // ← гарантированно будущая дата
        EndDate:     "12-2027",
    }

    err := svc.UpdateSubscription(context.Background(), sub,"user")

    if err != nil {
        t.Fatalf("UpdateSubscription failed: %v", err)
    }
}
func TestListSubscriptions_AdminSeesAll(t *testing.T) {
    repo := &mockRepo{
        listFunc: func(ctx context.Context, userID string, limit, offset int) ([]models.Subscription, error) {
            if userID != "" {
                t.Errorf("Expected empty userID for admin, got '%s'", userID)
            }
            return []models.Subscription{{ID: 1}}, nil
        },
    }
    svc := NewSubscriptionService(repo)

    list, err := svc.ListSubscriptions(context.Background(), "", "admin", 10, 0)
    if err != nil {
        t.Fatalf("ListSubscriptions failed: %v", err)
    }
    if len(list) != 1 {
        t.Errorf("Expected 1 subscription, got %d", len(list))
    }
}
func TestListSubscriptions_UserSeesOwn(t *testing.T) {
    repo := &mockRepo{
        listFunc: func(ctx context.Context, userID string, limit, offset int) ([]models.Subscription, error) {
            if userID != "user-123" {
                t.Errorf("Expected userID 'user-123', got '%s'", userID)
            }
            return []models.Subscription{{ID: 2}}, nil
        },
    }
    svc := NewSubscriptionService(repo)

    list, err := svc.ListSubscriptions(context.Background(), "user-123", "user", 10, 0)
    if err != nil {
        t.Fatalf("ListSubscriptions failed: %v", err)
    }
    if len(list) != 1 {
        t.Errorf("Expected 1 subscription, got %d", len(list))
    }
}
func TestListSubscriptions_UserWithoutID_Error(t *testing.T) {
    svc := NewSubscriptionService(&mockRepo{})

    _, err := svc.ListSubscriptions(context.Background(), "", "user", 10, 0)
    if err == nil {
        t.Error("Expected error for empty userID, got nil")
    }
    if err.Error() != "user_id is required" {
        t.Errorf("Expected 'user_id is required', got '%s'", err.Error())
    }
}
// ============================================================
// 4. ТЕСТ: УДАЛЕНИЕ ПОДПИСКИ (DeleteSubscription)
// ============================================================
func TestDeleteSubscription(t *testing.T) {
	repo := &mockRepo{
		deleteFunc: func(ctx context.Context, id int) error {
			if id != 1 {
				t.Errorf("Expected ID 1, got %d", id)
			}
			return nil
		},
	}
	svc := NewSubscriptionService(repo)

	err := svc.DeleteSubscription(context.Background(), 1)

	if err != nil {
		t.Fatalf("DeleteSubscription failed: %v", err)
	}
}

// ============================================================
// 5. ТЕСТ: ПОЛУЧЕНИЕ СПИСКА ПОДПИСОК (ListSubscriptions)
// ============================================================
func TestListSubscriptions(t *testing.T) {
    repo := &mockRepo{
        listFunc: func(ctx context.Context, userID string, limit, offset int) ([]models.Subscription, error) {
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

    // ✅ Админ → видит все подписки
    list, err := svc.ListSubscriptions(context.Background(), "", "admin", 10, 0)
    if err != nil {
        t.Fatalf("ListSubscriptions failed: %v", err)
    }
    if len(list) != 2 {
        t.Errorf("Expected 2 subscriptions, got %d", len(list))
    }
    if list[0].ID != 1 {
        t.Errorf("Expected first ID 1, got %d", list[0].ID)
    }

    // ✅ Пользователь → видит только свои
    list, err = svc.ListSubscriptions(context.Background(), "user-123", "user", 10, 0)
    if err != nil {
        t.Fatalf("ListSubscriptions failed: %v", err)
    }
    if len(list) != 2 {
        t.Errorf("Expected 2 subscriptions, got %d", len(list))
    }
}

// ============================================================
// 6. ТЕСТ: РАСЧЁТ СУММАРНОЙ СТОИМОСТИ (GetTotalCost)
// ============================================================
func TestGetTotalCost(t *testing.T) {
	repo := &mockRepo{
		getTotalCostFunc: func(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error) {
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

	total, err := svc.GetTotalCost(
		context.Background(),
		"user1",
		"",
		"01-2025",
		"12-2025",
	)

	if err != nil {
		t.Fatalf("GetTotalCost failed: %v", err)
	}
	if total != 1500 {
		t.Errorf("Expected 1500, got %d", total)
	}
}



// ============================================================
// ТЕСТЫ КЕШИРОВАНИЯ 
// ============================================================

// TestGetTotalCost_CacheHit — проверяет, что при наличии кеша
// возвращается значение из Redis, а БД не вызывается.
func TestGetTotalCost_CacheHit(t *testing.T) {
    mockRepo := &mockRepo{
        getTotalCostFunc: func(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error) {
            t.Error("Repository should not be called on cache hit")
            return 0, nil
        },
        getCacheUserVersionFunc: func(ctx context.Context, userID string) (int, error) {
            return 1, nil
        },
    }

    mockCache := &mockCache{
        getFunc: func(ctx context.Context, key string) (int, error) {
            return 500, nil // кеш-попадание
        },
    }

    svc := NewSubscriptionService(mockRepo)
    svc.SetCache(mockCache)

    result, err := svc.GetTotalCost(
        context.Background(),
        "test-user",
        "test-service",
        "01-2025",
        "12-2025",
    )

    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    if result != 500 {
        t.Errorf("Expected 500, got %d", result)
    }
}

// TestGetTotalCost_CacheMiss — проверяет, что при кеш-промахе
// идём в БД и сохраняем результат в Redis.
func TestGetTotalCost_CacheMiss(t *testing.T) {
    mockRepo := &mockRepo{
        getTotalCostFunc: func(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error) {
            return 1000, nil // БД возвращает 1000
        },
        getCacheUserVersionFunc: func(ctx context.Context, userID string) (int, error) {
            return 1, nil
        },
    }

    cacheSetCalled := false
    mockCache := &mockCache{
        getFunc: func(ctx context.Context, key string) (int, error) {
            return 0, nil // кеш-промах (ключа нет)
        },
        setFunc: func(ctx context.Context, key string, value int, ttl time.Duration) error {
            cacheSetCalled = true
            return nil
        },
    }

    svc := NewSubscriptionService(mockRepo)
    svc.SetCache(mockCache)

    result, err := svc.GetTotalCost(
        context.Background(),
        "test-user",
        "test-service",
        "01-2025",
        "12-2025",
    )

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
func TestGetTotalCost_RedisError(t *testing.T) {
    mockRepo := &mockRepo{
        getTotalCostFunc: func(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error) {
            return 2000, nil
        },
        getCacheUserVersionFunc: func(ctx context.Context, userID string) (int, error) {
            return 1, nil
        },
    }

    mockCache := &mockCache{
        getFunc: func(ctx context.Context, key string) (int, error) {
            return 0, errors.New("redis timeout") // ошибка Redis
        },
    }

    svc := NewSubscriptionService(mockRepo)
    svc.SetCache(mockCache)

    result, err := svc.GetTotalCost(
        context.Background(),
        "test-user",
        "test-service",
        "01-2025",
        "12-2025",
    )

    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    if result != 2000 {
        t.Errorf("Expected 2000, got %d", result)
    }
}

// TestGetTotalCost_VersionError — проверяет, что при ошибке получения
// версии из БД кеш отключается и мы идём напрямую в БД.
func TestGetTotalCost_VersionError(t *testing.T) {
    mockRepo := &mockRepo{
        getTotalCostFunc: func(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error) {
            return 3000, nil
        },
        getCacheUserVersionFunc: func(ctx context.Context, userID string) (int, error) {
            return 0, errors.New("db error") // ошибка получения версии
        },
    }

    // cache.Get НЕ ДОЛЖЕН вызываться — если вызовется, тест упадёт
    mockCache := &mockCache{
        getFunc: func(ctx context.Context, key string) (int, error) {
            t.Error("cache.Get should not be called when version retrieval fails")
            return 0, nil
        },
    }

    svc := NewSubscriptionService(mockRepo)
    svc.SetCache(mockCache)

    result, err := svc.GetTotalCost(
        context.Background(),
        "test-user",
        "test-service",
        "01-2025",
        "12-2025",
    )

    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    if result != 3000 {
        t.Errorf("Expected 3000, got %d", result)
    }
}

// TestGetTotalCost_InvalidDate — проверяет обработку невалидных дат.
func TestGetTotalCost_InvalidDate(t *testing.T) {
    mockRepo := &mockRepo{
        getTotalCostFunc: func(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error) {
            t.Error("Repository should not be called on invalid date")
            return 0, nil
        },
    }

    svc := NewSubscriptionService(mockRepo)

    result, err := svc.GetTotalCost(
        context.Background(),
        "test-user",
        "test-service",
        "invalid-date", // невалидный формат
        "12-2025",
    )

    if err == nil {
        t.Error("Expected error for invalid date, got nil")
    }
    if result != 0 {
        t.Errorf("Expected 0, got %d", result)
    }
}
