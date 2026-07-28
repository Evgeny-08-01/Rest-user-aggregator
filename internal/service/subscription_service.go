// package service реализует бизнес логику сервисов
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Evgeny-08-01/Rest-user-agregator/internal/models"
	"github.com/Evgeny-08-01/Rest-user-agregator/internal/repository"
	"github.com/Evgeny-08-01/Rest-user-agregator/pkg/logger"
)
type SubscriptionService struct {
    repo repository.SubscriptionRepository
}
// NewSubscriptionService — конструктор сервиса
func NewSubscriptionService(repo repository.SubscriptionRepository) *SubscriptionService {
    return &SubscriptionService{repo: repo}
}

// GetTotalCost — рассчитывает суммарную стоимость подписок за указанный период.
// Параметры:
//   - ctx: контекст для управления временем жизни запроса
//   - userID: идентификатор пользователя (опционально, фильтр)
//   - serviceName: название сервиса (опционально, фильтр)
//   - startDate: дата начала периода в формате MM-YYYY (обязательно)
//   - endDate: дата окончания периода в формате MM-YYYY (опционально)
//
// Возвращает:
//   - total: суммарная стоимость (int)
//   - error: ошибка, если парсинг дат не удался или диапазон невалидный
func (s *SubscriptionService) GetTotalCost(ctx context.Context, userID, serviceName, startDate, endDate string) (int, error) {
    // 1. Парсинг дат
    startDateTimeDB, err := parseDate(startDate)
    if err != nil {
        logger.Warn("GetTotalCost: failed to parse startDate %s: %v", startDate, err)
        return 0, fmt.Errorf("invalid startDate: %w", err)
    }

    endDateTimeDB, err := parseEndDate(endDate)
    if err != nil {
        logger.Warn("GetTotalCost: failed to parse endDate %s: %v", endDate, err)
        return 0, fmt.Errorf("invalid endDate: %w", err)
    }

    // 2. Валидация диапазона
    if err := validateDateRange(startDateTimeDB, endDateTimeDB); err != nil {
        logger.Warn("GetTotalCost: invalid date range: startDate=%s > endDate=%s", startDate, endDate)
        return 0, err
    }

    // 3. Вызов репозитория
    return s.repo.GetTotalCost(ctx, userID, serviceName, startDateTimeDB, endDateTimeDB)
}
// CreateSubscription — бизнес-логика создания новой подписки.
// Параметры:  - ctx: контекст для управления временем жизни запроса
//             - sub: структура с данными новой подписки (ServiceName, Price, UserID, StartDate, EndDate)
// Логика:
//   1. Парсит startDate из строки (формат MM-YYYY) в time.Time
//   2. Если endDate указан — парсит его в time.Time
//   3. Вызывает репозиторий для создания записи в БД
// Возвращает: - id: идентификатор созданной подписки (int), - error: ошибка, если парсинг не удался или создание не удалось
 func (s *SubscriptionService) CreateSubscription(ctx context.Context, sub models.Subscription) (int, error) {
    startDate, err := parseDate(sub.StartDate)
    if err != nil {
        return 0, fmt.Errorf("invalid start_date: %w", err)
    }

    var endDate *time.Time
    if sub.EndDate != "" {
        parsed, err := parseDate(sub.EndDate)
        if err != nil {
            return 0, fmt.Errorf("invalid end_date: %w", err)
        }
        endDate = &parsed
    }

    return s.repo.CreateSubscription(ctx, sub, startDate, endDate)
}
// UpdateSubscription — бизнес-логика обновления подписки.
// Параметры:   - ctx: контекст для управления временем жизни запроса
//              - sub: структура с новыми данными подписки (поля: ID, ServiceName, Price, UserID, StartDate, EndDate)
// Логика:
//   1. Парсит startDate из строки (формат MM-YYYY) в time.Time
//   2. Если endDate указан — парсит его в time.Time
//   3. Вызывает репозиторий для обновления записи в БД
// Возвращает: - error: ошибка, если парсинг не удался или обновление не удалось
func (s *SubscriptionService) UpdateSubscription(ctx context.Context, sub models.Subscription) error {
    startDate, err := parseDate(sub.StartDate)
    if err != nil {
        return fmt.Errorf("invalid start_date: %w", err)
    }

    var endDate *time.Time
    if sub.EndDate != "" {
        parsed, err := parseDate(sub.EndDate)
        if err != nil {
            return fmt.Errorf("invalid end_date: %w", err)
        }
        endDate = &parsed
    }

    return s.repo.UpdateSubscription(ctx, sub, startDate, endDate)
}
// GetSubscriptionByID — бизнес-логика получения подписки по ID.
// Параметры:
//   - ctx: контекст для управления временем жизни запроса
//   - id: идентификатор подписки (int)
// Логика:
//   1. Вызывает репозиторий для получения записи из БД
//   2. Возвращает nil, nil если подписка не найдена
// Возвращает:
//   - *models.Subscription: структура с данными подписки (или nil, если не найдена)
//   - error: ошибка, если запрос к БД не удался
func (s *SubscriptionService) GetSubscriptionByID(ctx context.Context, id int) (*models.Subscription, error) {
    return s.repo.GetSubscriptionByID(ctx, id)
}

// DeleteSubscription — бизнес-логика удаления подписки по ID.
// Параметры:
//   - ctx: контекст для управления временем жизни запроса
//   - id: идентификатор подписки (int)
// Логика:
//   1. Вызывает репозиторий для удаления записи из БД
//   2. Возвращает sql.ErrNoRows если подписка не найдена
// Возвращает:
//   - error: ошибка, если удаление не удалось или запись не найдена
func (s *SubscriptionService) DeleteSubscription(ctx context.Context, id int) error {
    return s.repo.DeleteSubscription(ctx, id)
}

// ListSubscriptions — бизнес-логика получения списка подписок с пагинацией.
// Параметры:
//   - ctx: контекст для управления временем жизни запроса
//   - limit: максимальное количество записей (должен быть > 0)
//   - offset: сдвиг от начала (должен быть >= 0)
// Логика:
//   1. Вызывает репозиторий для получения списка из БД
//   2. Возвращает пустой слайс, если записи не найдены
// Возвращает:
//   - []models.Subscription: слайс подписок (может быть пустым)
//   - error: ошибка, если запрос к БД не удался
func (s *SubscriptionService) ListSubscriptions(ctx context.Context, limit, offset int) ([]models.Subscription, error) {
    return s.repo.ListSubscriptions(ctx, limit, offset)
}
// ============================================================
// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ДЛЯ ПАРСИНГА ДАТ
// ============================================================

// parseDate парсит строку в формате MM-YYYY в time.Time
func parseDate(dateStr string) (time.Time, error) {
    if dateStr == "" {
        return time.Time{}, fmt.Errorf("date is empty")
    }
    return time.Parse("01-2006", dateStr)
}

// parseEndDate парсит endDate, превращает в последний день месяца или 2100-01-01
func parseEndDate(dateStr string) (time.Time, error) {
    if dateStr == "" {
        return time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC), nil
    }
    parsed, err := time.Parse("01-2006", dateStr)
    if err != nil {
        return time.Time{}, err
    }
    return time.Date(parsed.Year(), parsed.Month()+1, 0, 0, 0, 0, 0, time.UTC), nil
}

// validateDateRange проверяет, что startDate <= endDate
func validateDateRange(start, end time.Time) error {
    if start.After(end) {
        return fmt.Errorf("start_date > end_date")
    }
    return nil
}