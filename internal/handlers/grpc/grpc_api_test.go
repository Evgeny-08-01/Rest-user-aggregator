//go:build unit

package grpcserver

import (
	"context"
	"testing"
	"time"

	pb "Rest-user-agregator/proto/subscription"
	"Rest-user-agregator/internal/models"
	"Rest-user-agregator/internal/service"
	"github.com/stretchr/testify/assert"
)

// ============================================================
// МОК-РЕПОЗИТОРИЙ
// ============================================================
type mockRepo struct {
	getByIDFunc       func(ctx context.Context, id int) (*models.Subscription, error)
	listFunc          func(ctx context.Context, limit, offset int) ([]models.Subscription, error)
	createFunc        func(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) (int, error)
	updateFunc        func(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) error
	deleteFunc        func(ctx context.Context, id int) error
	getTotalCostFunc  func(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error)
	getCacheUserVersionFunc    func(ctx context.Context, userID string) (int, error)
	incrementCacheUserVersionFunc func(ctx context.Context, userID string) error
}

func (m *mockRepo) GetSubscriptionByID(ctx context.Context, id int) (*models.Subscription, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return &models.Subscription{ID: id, ServiceName: "Test"}, nil
}

func (m *mockRepo) ListSubscriptions(ctx context.Context, limit, offset int) ([]models.Subscription, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, limit, offset)
	}
	return []models.Subscription{}, nil
}

func (m *mockRepo) CreateSubscription(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) (int, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, sub, startDate, endDate)
	}
	return 1, nil
}

func (m *mockRepo) UpdateSubscription(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, sub, startDate, endDate)
	}
	return nil
}

func (m *mockRepo) DeleteSubscription(ctx context.Context, id int) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockRepo) GetTotalCost(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error) {
	if m.getTotalCostFunc != nil {
		return m.getTotalCostFunc(ctx, userID, serviceName, startDate, endDate)
	}
	return 0, nil
}

func (m *mockRepo) GetCacheUserVersion(ctx context.Context, userID string) (int, error) {
	if m.getCacheUserVersionFunc != nil {
		return m.getCacheUserVersionFunc(ctx, userID)
	}
	return 1, nil
}

func (m *mockRepo) IncrementCacheUserVersion(ctx context.Context, userID string) error {
	if m.incrementCacheUserVersionFunc != nil {
		return m.incrementCacheUserVersionFunc(ctx, userID)
	}
	return nil
}

// ============================================================
// ТЕСТЫ
// ============================================================

// 1. ПОЛУЧЕНИЕ ПОДПИСКИ ПО ID
func TestGetSubscription(t *testing.T) {
	repo := &mockRepo{
		getByIDFunc: func(ctx context.Context, id int) (*models.Subscription, error) {
			return &models.Subscription{
				ID:          id,
				ServiceName: "Test gRPC",
				Price:       100,
				UserID:      "test-user",
				StartDate:   "01-2025",
				EndDate:     "12-2025",
			}, nil
		},
	}

	svc := service.NewSubscriptionService(repo)
	server := NewSubscriptionServer(svc)

	req := &pb.GetRequest{Id: 1}
	resp, err := server.GetSubscription(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Test gRPC", resp.ServiceName)
	assert.Equal(t, int32(100), resp.Price)
}

// 2. СПИСОК ПОДПИСОК
func TestGetSubscriptions(t *testing.T) {
	repo := &mockRepo{
		listFunc: func(ctx context.Context, limit, offset int) ([]models.Subscription, error) {
			return []models.Subscription{
				{ID: 1, ServiceName: "Test1", Price: 100},
				{ID: 2, ServiceName: "Test2", Price: 200},
			}, nil
		},
	}

	svc := service.NewSubscriptionService(repo)
	server := NewSubscriptionServer(svc)

	req := &pb.GetSubscriptionsRequest{Limit: 10, Offset: 0}
	resp, err := server.GetSubscriptions(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Subscriptions, 2)
	assert.Equal(t, int32(1), resp.Subscriptions[0].Id)
}

// 3. СОЗДАНИЕ ПОДПИСКИ
func TestCreateSubscription(t *testing.T) {
	repo := &mockRepo{
		createFunc: func(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) (int, error) {
			return 5, nil
		},
	}

	svc := service.NewSubscriptionService(repo)
	server := NewSubscriptionServer(svc)

	req := &pb.CreateRequest{
		ServiceName: "Test Create",
		Price:       150,
		UserId:      "test-user",
		StartDate:   "01-2025",
		EndDate:     "12-2025",
	}
	resp, err := server.CreateSubscription(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(5), resp.Id)
}

// 4. ОБНОВЛЕНИЕ ПОДПИСКИ
func TestUpdateSubscription(t *testing.T) {
	repo := &mockRepo{
		updateFunc: func(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) error {
			return nil
		},
	}

	svc := service.NewSubscriptionService(repo)
	server := NewSubscriptionServer(svc)

	req := &pb.UpdateRequest{
		Id:          1,
		ServiceName: "Updated",
		Price:       200,
		UserId:      "test-user",
		StartDate:   "01-2025",
		EndDate:     "12-2025",
	}
	resp, err := server.UpdateSubscription(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
// 5. УДАЛЕНИЕ ПОДПИСКИ
func TestDeleteSubscription(t *testing.T) {
	repo := &mockRepo{
		deleteFunc: func(ctx context.Context, id int) error {
			return nil
		},
	}

	svc := service.NewSubscriptionService(repo)
	server := NewSubscriptionServer(svc)

	req := &pb.GetRequest{Id: 1}
	resp, err := server.DeleteSubscription(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// 6. ОБЩАЯ СТОИМОСТЬ
func TestGetTotalCost(t *testing.T) {
	repo := &mockRepo{
		getTotalCostFunc: func(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error) {
			return 1500, nil
		},
	}

	svc := service.NewSubscriptionService(repo)
	server := NewSubscriptionServer(svc)

	req := &pb.TotalCostRequest{
		UserId:      "test-user",
		ServiceName: "",
		StartDate:   "01-2025",
		EndDate:     "12-2025",
	}
	resp, err := server.GetTotalCost(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(1500), resp.Total)
}