// ============================================================
// ПАКЕТ: cache
// ============================================================
// Интерфейс Cache для использования в сервисах и тестах
// ============================================================
package cache

import (
	"context"
	"time"
)

// Cache — интерфейс для работы с кешем.
// Используется в сервисах (SubscriptionService) для кеширования результатов.
// Позволяет подменять реальный Redis на мок в тестах.
type Cache interface {
	Get(ctx context.Context, key string) (int, error)
	Set(ctx context.Context, key string, value int, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Keys(ctx context.Context, pattern string) ([]string, error)
}
