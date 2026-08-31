//go:build unit

package service

import (
	"context"
	"time"

	"Rest-user-agregator/internal/models"
)

type mockCache struct {
	getFunc    func(ctx context.Context, key string) (int, error)
	setFunc    func(ctx context.Context, key string, value int, ttl time.Duration) error
	deleteFunc func(ctx context.Context, key string) error
	keysFunc   func(ctx context.Context, pattern string) ([]string, error)
}

func (m *mockCache) Get(ctx context.Context, key string) (int, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, key)
	}
	return 0, nil
}

func (m *mockCache) Set(ctx context.Context, key string, value int, ttl time.Duration) error {
	if m.setFunc != nil {
		return m.setFunc(ctx, key, value, ttl)
	}
	return nil
}

func (m *mockCache) Delete(ctx context.Context, key string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, key)
	}
	return nil
}

func (m *mockCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	if m.keysFunc != nil {
		return m.keysFunc(ctx, pattern)
	}
	return []string{}, nil
}

type mockTemplateRepo struct {
	getTemplateByIDFunc   func(ctx context.Context, id int) (*models.Template, error)
	getTemplateByNameFunc func(ctx context.Context, serviceName string) (*models.Template, error)
	createTemplateFunc    func(ctx context.Context, serviceName string, price int) (int, error)
	listTemplatesFunc     func(ctx context.Context) ([]models.Template, error)
	updateTemplateFunc    func(ctx context.Context, id int, serviceName string, price int) error
	deleteTemplateFunc    func(ctx context.Context, id int) error
}

func (m *mockTemplateRepo) CreateTemplate(ctx context.Context, serviceName string, price int) (int, error) {
	if m.createTemplateFunc != nil {
		return m.createTemplateFunc(ctx, serviceName, price)
	}
	return 1, nil
}

func (m *mockTemplateRepo) ListTemplates(ctx context.Context) ([]models.Template, error) {
	if m.listTemplatesFunc != nil {
		return m.listTemplatesFunc(ctx)
	}
	return []models.Template{}, nil
}

func (m *mockTemplateRepo) GetTemplateByID(ctx context.Context, id int) (*models.Template, error) {
	if m.getTemplateByIDFunc != nil {
		return m.getTemplateByIDFunc(ctx, id)
	}
	return &models.Template{ID: id, ServiceName: "TestTemplate", Price: 100}, nil
}

func (m *mockTemplateRepo) GetTemplateByName(ctx context.Context, serviceName string) (*models.Template, error) {
	if m.getTemplateByNameFunc != nil {
		return m.getTemplateByNameFunc(ctx, serviceName)
	}
	return nil, nil
}

func (m *mockTemplateRepo) UpdateTemplate(ctx context.Context, id int, serviceName string, price int) error {
	if m.updateTemplateFunc != nil {
		return m.updateTemplateFunc(ctx, id, serviceName, price)
	}
	return nil
}

func (m *mockTemplateRepo) DeleteTemplate(ctx context.Context, id int) error {
	if m.deleteTemplateFunc != nil {
		return m.deleteTemplateFunc(ctx, id)
	}
	return nil
}

type mockRepo struct {
	createFunc                    func(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) (int, error)
	getByIDFunc                   func(ctx context.Context, id int) (*models.Subscription, error)
	updateFunc                    func(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) error
	deleteFunc                    func(ctx context.Context, id int) error
	listFunc                      func(ctx context.Context, userID string, limit, offset int) ([]models.Subscription, error)
	getTotalCostFunc              func(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error)
	getCacheUserVersionFunc       func(ctx context.Context, userID string) (int, error)
	incrementCacheUserVersionFunc func(ctx context.Context, userID string) error
}

func (m *mockRepo) CreateSubscription(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) (int, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, sub, startDate, endDate)
	}
	return 1, nil
}

func (m *mockRepo) GetSubscriptionByID(ctx context.Context, id int) (*models.Subscription, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return &models.Subscription{ID: id, ServiceName: "Test"}, nil
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

func (m *mockRepo) ListSubscriptions(ctx context.Context, userID string, limit, offset int) ([]models.Subscription, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, userID, limit, offset)
	}
	return []models.Subscription{}, nil
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
