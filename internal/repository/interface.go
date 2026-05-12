// Package repository - интерфейс для работы с БД
package repository

import (
    "context"
    "github.com/Evgeny-08-01/Rest-user-aggregator/internal/models"
)

// SubscriptionRepository - список всех методов для работы с БД
type SubscriptionRepository interface {
    CreateSubscription(ctx context.Context, sub models.Subscription) (int, error)
    GetSubscriptionByID(ctx context.Context, id int) (*models.Subscription, error)
    UpdateSubscription(ctx context.Context, sub models.Subscription) error
    DeleteSubscription(ctx context.Context, id int) error
    ListSubscriptions(ctx context.Context, limit, offset int) ([]models.Subscription, error)
    GetTotalCost(ctx context.Context, userID, serviceName, startDate, endDate string) (int, error)
}