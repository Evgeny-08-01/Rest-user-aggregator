// Package repository - интерфейс для работы с БД
package repository

import (
	"context"
	"time"

	"Rest-user-agregator/internal/models"
)

// SubscriptionRepository - список всех методов для работы с БД
type SubscriptionRepository interface {
	//  Методы для подписок(таблица subscriptions)
	CreateSubscription(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) (int, error)
	GetSubscriptionByID(ctx context.Context, id int) (*models.Subscription, error)
	UpdateSubscription(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) error
	DeleteSubscription(ctx context.Context, id int) error
	ListSubscriptions(ctx context.Context, userID string, limit, offset int) ([]models.Subscription, error)
	GetTotalCost(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error)
	// ============================================================
	// МЕТОДЫ ДЛЯ РАБОТЫ С КЕШЕМ (таблица cache_control_user)
	// ============================================================
	// GetCacheUserVersion — возвращает текущую версию кеша пользователя.
	// Если записи нет — создаёт со значением 1.
	// IncrementCacheUserVersion — увеличивает версию кеша пользователя на 1.
	// Вызывается при Create/Update/Delete подписки.
	GetCacheUserVersion(ctx context.Context, userID string) (int, error)
	IncrementCacheUserVersion(ctx context.Context, userID string) error
}

// ============================================================
// UserRepository - интерфейс для работы с пользователями
// ============================================================
type UserRepository interface {
	CreateUser(ctx context.Context, user models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
}

// ============================================================
// МЕТОДЫ ДЛЯ ШАБЛОНОВ (subscription_templates)
// ============================================================
type TemplateRepository interface {
	CreateTemplate(ctx context.Context, serviceName string, price int) (int, error)
	ListTemplates(ctx context.Context) ([]models.Template, error)
	GetTemplateByID(ctx context.Context, id int) (*models.Template, error)
	GetTemplateByName(ctx context.Context, serviceName string) (*models.Template, error)
	UpdateTemplate(ctx context.Context, id int, serviceName string, price int) error
	DeleteTemplate(ctx context.Context, id int) error
}
