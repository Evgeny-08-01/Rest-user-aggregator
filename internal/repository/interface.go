// Package repository - интерфейс для работы с БД
package repository

import (
	"context"
	"time"

	"Rest-user-agregator/internal/models"
)

// SubscriptionRepository - список всех методов для работы с БД
type SubscriptionRepository interface {
    CreateSubscription(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) (int, error)
    GetSubscriptionByID(ctx context.Context, id int) (*models.Subscription, error)
    UpdateSubscription(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) error
    DeleteSubscription(ctx context.Context, id int) error
    ListSubscriptions(ctx context.Context, limit, offset int) ([]models.Subscription, error)
    GetTotalCost(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error) 
}
// ============================================================
// UserRepository - интерфейс для работы с пользователями
// ============================================================
type UserRepository interface {
	CreateUser(ctx context.Context, user models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
}