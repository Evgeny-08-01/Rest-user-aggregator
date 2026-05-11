// Package repository - интерфейс для работы с БД
package repository

import (
    "context"
    "github.com/Evgeny-08-01/Rest-user-aggregator/internal/models"
)

// SubscriptionRepository - список всех методов для работы с БД
type SubscriptionRepository interface {
    CreateMtd(ctx context.Context, sub models.Subscription) (int, error)
    GetByIDMtd(ctx context.Context, id int) (*models.Subscription, error)
    UpdateMtd(ctx context.Context, sub models.Subscription) error
    DeleteMtd(ctx context.Context, id int) error
    ListMtd(ctx context.Context, limit, offset int) ([]models.Subscription, error)
    GetTotalCostMtd(ctx context.Context, userID, serviceName, startDate, endDate string) (int, error)
}