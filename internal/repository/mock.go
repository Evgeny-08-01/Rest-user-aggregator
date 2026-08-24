package repository

import (
	"context"
	"errors"
	"time"

	"Rest-user-agregator/internal/models"
)

type MockSubRepo struct {
	CreateMock       func(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) (int, error)
	GetSubByIDMock   func(ctx context.Context, id int) (*models.Subscription, error)
	UpdateSubMock    func(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) error
	DeleteSubMock    func(ctx context.Context, id int) error
	ListSubMock      func(ctx context.Context, userID string, limit, offset int) ([]models.Subscription, error)
	GetTotalCostMock func(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error)
	GetCacheUserVersionMock      func(ctx context.Context, userID string) (int, error)
	IncrementCacheUserVersionMock func(ctx context.Context, userID string) error
}

func (m *MockSubRepo) CreateSubscription(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) (int, error) {
	if m.CreateMock != nil {
		return m.CreateMock(ctx, sub, startDate, endDate)
	}
	return 0, errors.New("CreateMock not mocked")
}

func (m *MockSubRepo) GetSubscriptionByID(ctx context.Context, id int) (*models.Subscription, error) {
	if m.GetSubByIDMock != nil {
		return m.GetSubByIDMock(ctx, id)
	}
	return nil, errors.New("GetSubByIDMock not mocked")
}

func (m *MockSubRepo) UpdateSubscription(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) error {
	if m.UpdateSubMock != nil {
		return m.UpdateSubMock(ctx, sub, startDate, endDate)
	}
	return errors.New("UpdateSubMock not mocked")
}

func (m *MockSubRepo) DeleteSubscription(ctx context.Context, id int) error {
	if m.DeleteSubMock != nil {
		return m.DeleteSubMock(ctx, id)
	}
	return errors.New("DeleteSubMock not mocked")
}

func (m *MockSubRepo) ListSubscriptions(ctx context.Context, userID string, limit, offset int) ([]models.Subscription, error) {
	if m.ListSubMock != nil {
		return m.ListSubMock(ctx, userID, limit, offset)
	}
	return nil, errors.New("ListSubMock not mocked")
}

func (m *MockSubRepo) GetTotalCost(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error) {
	if m.GetTotalCostMock != nil {
		return m.GetTotalCostMock(ctx, userID, serviceName, startDate, endDate)
	}
	return 0, errors.New("GetTotalCostMock not mocked")
}

func (m *MockSubRepo) GetCacheUserVersion(ctx context.Context, userID string) (int, error) {
	if m.GetCacheUserVersionMock != nil {
		return m.GetCacheUserVersionMock(ctx, userID)
	}
	return 1, nil
}

func (m *MockSubRepo) IncrementCacheUserVersion(ctx context.Context, userID string) error {
	if m.IncrementCacheUserVersionMock != nil {
		return m.IncrementCacheUserVersionMock(ctx, userID)
	}
	return nil
}
