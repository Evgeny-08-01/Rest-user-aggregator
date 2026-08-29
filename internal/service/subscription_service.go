// package service реализует бизнес логику сервисов
package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"Rest-user-agregator/internal/cache"
	"Rest-user-agregator/internal/models"
	"Rest-user-agregator/internal/repository"
	"Rest-user-agregator/pkg/logger"

	"github.com/redis/go-redis/v9"
)

//var ErrCannotChangeStartDate = errors.New("cannot change start_date that is today or in the past")

type SubscriptionService struct {
	repo  repository.SubscriptionRepository
	templateRepo repository.TemplateRepository
	cache cache.Cache
}

// NewSubscriptionService — конструктор сервиса
func NewSubscriptionService(
    repo repository.SubscriptionRepository,
    templateRepo repository.TemplateRepository,
) *SubscriptionService {
    return &SubscriptionService{
        repo:         repo,
        templateRepo: templateRepo,
        cache:        cache.NewRedisCache(),
    }
}

// GetTotalCost — рассчитывает суммарную стоимость подписок за указанный период.
// Параметры:
//   - userID: ID пользователя (пусто для админа → все подписки)
//   - serviceName: фильтр по названию сервиса (опционально)
//   - startDate: дата начала в формате MM-YYYY (обязательно)
//   - endDate: дата окончания в формате MM-YYYY (обязательно)
// Возвращает:
//   - int: суммарная стоимость
//   - error: ErrStartDateRequired если startDate пустой
//   - error: ErrInvalidDateRange если startDate > endDate
//   - error: ошибки парсинга дат или БД
func (s *SubscriptionService) GetTotalCost(ctx context.Context, userID, serviceName, startDate, endDate string) (int, error) {
    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()

    // ============================================================
    // 1. ВАЛИДАЦИЯ ВХОДНЫХ ДАННЫХ
    // ============================================================
    if startDate == "" {
        return 0, ErrStartDateRequired
    }
    if endDate == "" {
        return 0, fmt.Errorf("end_date is required")
    }

    // ============================================================
    // 2. ПАРСИНГ ДАТ
    // ============================================================
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

    if err := validateDateRange(startDateTimeDB, endDateTimeDB); err != nil {
        logger.Warn("GetTotalCost: invalid date range: startDate=%s > endDate=%s", startDate, endDate)
        return 0, fmt.Errorf("start_date > end_date")
    }

    // ============================================================
    // 3. АДМИН — КЕШ ОТКЛЮЧАЕМ
    // ============================================================
    if userID == "" {
        logger.Debug("GetTotalCost: admin request (userID empty), skipping cache")
        return s.repo.GetTotalCost(ctx, userID, serviceName, startDateTimeDB, endDateTimeDB)
    }

    // ============================================================
    // 4. ПОЛУЧАЕМ ВЕРСИЮ КЕША
    // ============================================================
    version, err := s.repo.GetCacheUserVersion(ctx, userID)
    if err != nil {
        logger.Warn("GetTotalCost: failed to get cache version for user %s: %v, skipping cache", userID, err)
        return s.repo.GetTotalCost(ctx, userID, serviceName, startDateTimeDB, endDateTimeDB)
    }

    // ============================================================
    // 5. РАБОТА С КЕШЕМ
    // ============================================================
    cacheKey := fmt.Sprintf("total:v%d:%s:%s:%s:%s",
        version,
        userID,
        serviceName,
        startDate,
        endDate,
    )

    cachedValue, err := s.cache.Get(ctx, cacheKey)
    if err == nil && cachedValue > 0 {
        logger.Debug("GetTotalCost: cache hit for key %s, value=%d", cacheKey, cachedValue)
        return cachedValue, nil
    }

    if err != nil && err != redis.Nil {
        logger.Warn("GetTotalCost: Redis error for key %s: %v, falling back to DB", cacheKey, err)
    }

    // ============================================================
    // 6. КЕШ-ПРОМАХ: ИДЁМ В БД
    // ============================================================
    logger.Debug("GetTotalCost: cache miss for key %s, querying DB", cacheKey)
    total, err := s.repo.GetTotalCost(ctx, userID, serviceName, startDateTimeDB, endDateTimeDB)
    if err != nil {
        logger.Error("GetTotalCost: DB query failed: %v", err)
        return 0, err
    }

    // ============================================================
    // 7. СОХРАНЯЕМ В КЕШ
    // ============================================================
    const ttl = 5 * time.Minute
    if err := s.cache.Set(ctx, cacheKey, total, ttl); err != nil {
        logger.Warn("GetTotalCost: failed to cache result for key %s: %v", cacheKey, err)
    } else {
        logger.Debug("GetTotalCost: cached result for key %s, total=%d, ttl=%v", cacheKey, total, ttl)
    }

    return total, nil
}

// CreateSubscription — создаёт новую подписку с полной валидацией
// Параметры:
//   - templateID: ID шаблона (обязательно, > 0)
//   - userID: ID пользователя из JWT (обязательно, не пустой)
//   - startDate: дата начала в формате MM-YYYY (обязательно)
//   - endDate: дата окончания в формате MM-YYYY (опционально)
// Возвращает:
//   - int: ID созданной подписки
//   - error: ошибка валидации или БД
func (s *SubscriptionService) CreateSubscription(ctx context.Context, templateID int, userID string, startDate, endDate string) (int, error) {
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    // ============================================================
    // 1. ВАЛИДАЦИЯ ВХОДНЫХ ДАННЫХ (кастомные ошибки)
    // ============================================================

    if templateID <= 0 {
        return 0, ErrTemplateIDRequired
    }
    if userID == "" {
        return 0, ErrUserIDRequired
    }
    if startDate == "" {
        return 0, ErrStartDateRequired
    }

    // ============================================================
    // 2. БИЗНЕС-ЛОГИКА
    // ============================================================

    // 2.1. Получаем шаблон
    template, err := s.templateRepo.GetTemplateByID(ctx, templateID)
    if err != nil {
        return 0, fmt.Errorf("%w: %v", ErrTemplateNotFound, err)
    }
    if template == nil {
        return 0, ErrTemplateNotFound
    }

    // 2.2. Парсим даты
    startDateParsed, err := parseDate(startDate)
    if err != nil {
        return 0, fmt.Errorf("%w: %s", ErrInvalidDateFormat, err.Error())
    }
    var endDateParsed *time.Time
    if endDate != "" {
        parsed, err := parseDate(endDate)
        if err != nil {
            return 0, fmt.Errorf("%w: %s", ErrInvalidDateFormat, err.Error())
        }
        endDateParsed = &parsed
    }

    // 2.3. Проверяем, что start_date не в прошлом
    if !canChangeStartDate(startDateParsed) {
        return 0, ErrCannotChangeStartDate
    }

    // 2.4. Создаём подписку
    sub := models.Subscription{
        ServiceName: template.ServiceName,
        Price:       template.Price,
        UserID:      userID,
        TemplateID:  templateID,
    }

    id, err := s.repo.CreateSubscription(ctx, sub, startDateParsed, endDateParsed)
    if err != nil && err.Error() == "duplicate" {
        return 0, ErrTemplateHasSubscriptions
    }
    if err != nil {
        return 0, err
    }
    return id, nil
}

// UpdateSubscription — обновляет подписку с полной валидацией и проверкой прав
// Параметры:
//   - sub: обновлённые данные подписки (должен содержать ID, UserID, StartDate, EndDate)
//   - role: роль текущего пользователя (из JWT)
// Возвращает:
//   - error: ErrInvalidID если sub.ID <= 0
//   - error: ErrUserIDRequired если sub.UserID пустой
//   - error: ErrStartDateRequired если sub.StartDate пустой
//   - error: ErrPermissionDenied если пользователь не владелец и не админ
//   - error: sql.ErrNoRows если подписка не найдена
//   - error: ErrCannotChangeStartDate если start_date в прошлом
//   - error: другие ошибки (парсинг дат, БД)
func (s *SubscriptionService) UpdateSubscription(ctx context.Context, sub models.Subscription, role string) error {
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    // ============================================================
    // 1. ВАЛИДАЦИЯ ВХОДНЫХ ДАННЫХ (кастомные ошибки)
    // ============================================================
    if sub.ID <= 0 {
        return ErrInvalidID
    }
    if sub.UserID == "" {
        return ErrUserIDRequired
    }
    if sub.StartDate == "" {
        return ErrStartDateRequired
    }

    // ============================================================
    // 2. БИЗНЕС-ЛОГИКА
    // ============================================================

    // 2.1. Получаем существующую подписку из БД для проверки владельца
    existing, err := s.repo.GetSubscriptionByID(ctx, sub.ID)
    if err != nil {
        return err
    }
    if existing == nil {
        return sql.ErrNoRows
    }

    // 2.2. Проверка прав доступа
    //     - Админ может обновлять любые подписки
    //     - Обычный пользователь — только свои
    if role != "admin" && existing.UserID != sub.UserID {
        return ErrPermissionDenied
    }

    // 2.3. Парсим даты
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

    // 2.4. Проверяем, что start_date не в прошлом
    if !canChangeStartDate(startDate) {
        return ErrCannotChangeStartDate
    }

    // 2.5. Вызываем репозиторий для обновления
    return s.repo.UpdateSubscription(ctx, sub, startDate, endDate)
}
// GetSubscriptionByID — получает подписку по ID с полной валидацией и проверкой прав
// Параметры:
//   - id: ID подписки (обязательно, > 0)
//   - userID: ID пользователя из JWT (обязательно, не пустой)
//   - role: роль пользователя (admin/user)
// Возвращает:
//   - *models.Subscription: данные подписки
//   - error: ErrInvalidID если id <= 0
//   - error: ErrUserIDRequired если userID пустой
//   - error: ErrPermissionDenied если пользователь не владелец и не админ
//   - error: другие ошибки БД
func (s *SubscriptionService) GetSubscriptionByID(ctx context.Context, id int, userID string, role string) (*models.Subscription, error) {
    // ============================================================
    // 1. ВАЛИДАЦИЯ ВХОДНЫХ ДАННЫХ (кастомные ошибки)
    // ============================================================
    if id <= 0 {
        return nil, ErrInvalidID
    }
    if userID == "" {
        return nil, ErrUserIDRequired
    }

    // ============================================================
    // 2. БИЗНЕС-ЛОГИКА
    // ============================================================

    // 2.1. Получаем подписку из репозитория
    sub, err := s.repo.GetSubscriptionByID(ctx, id)
    if err != nil {
        return nil, err
    }
    if sub == nil {
        return nil, nil
    }

    // 2.2. Проверка прав: админ видит всё, обычный пользователь — только свои
    if role != "admin" && sub.UserID != userID {
        return nil, ErrPermissionDenied
    }

    return sub, nil
}
// DeleteSubscription — удаляет подписку по ID с полной валидацией и проверкой прав
// Параметры:
//   - id: ID подписки (обязательно, > 0)
//   - userID: ID пользователя из JWT (обязательно, не пустой)
//   - role: роль пользователя (admin/user)
// Возвращает:
//   - error: ErrInvalidID если id <= 0
//   - error: ErrUserIDRequired если userID пустой
//   - error: ErrPermissionDenied если пользователь не владелец и не админ
//   - error: sql.ErrNoRows если подписка не найдена
//   - error: другие ошибки БД
func (s *SubscriptionService) DeleteSubscription(ctx context.Context, id int, userID string, role string) error {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    // ============================================================
    // 1. ВАЛИДАЦИЯ ВХОДНЫХ ДАННЫХ (кастомные ошибки)
    // ============================================================
    if id <= 0 {
        return ErrInvalidID
    }
    if userID == "" {
        return ErrUserIDRequired
    }

    // ============================================================
    // 2. БИЗНЕС-ЛОГИКА
    // ============================================================

    // 2.1. Получаем существующую подписку из БД для проверки владельца
    existing, err := s.repo.GetSubscriptionByID(ctx, id)
    if err != nil {
        return err
    }
    if existing == nil {
        return sql.ErrNoRows
    }

    // 2.2. Проверка прав доступа
    //     - Админ может удалять любые подписки
    //     - Обычный пользователь — только свои
    if role != "admin" && existing.UserID != userID {
        return ErrPermissionDenied
    }

    // 2.3. Вызываем репозиторий для удаления
    return s.repo.DeleteSubscription(ctx, id)
}

// ListSubscriptions — получает список подписок с пагинацией и проверкой прав
// Параметры:
//   - userID: ID пользователя из JWT (обязательно, если не админ)
//   - role: роль пользователя (admin/user)
//   - limit: лимит записей
//   - offset: смещение
// Возвращает:
//   - []models.Subscription: список подписок
//   - error: ErrUserIDRequired если userID пустой и роль не admin
//   - error: ошибка БД
func (s *SubscriptionService) ListSubscriptions(ctx context.Context, userID, role string, limit, offset int) ([]models.Subscription, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    // ============================================================
    // 1. ВАЛИДАЦИЯ ВХОДНЫХ ДАННЫХ
    // ============================================================
    // Админ может видеть все подписки (userID не нужен)
    // Обычный пользователь должен иметь userID
    if role != "admin" && userID == "" {
        return nil, ErrUserIDRequired
    }

    // ============================================================
    // 2. БИЗНЕС-ЛОГИКА
    // ============================================================
    // Админ видит всё
    if role == "admin" {
        return s.repo.ListSubscriptions(ctx, "", limit, offset)
    }

    // Обычный пользователь видит только свои
    return s.repo.ListSubscriptions(ctx, userID, limit, offset)
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

// SetCache — устанавливает кеш для сервиса (используется в тестах)
func (s *SubscriptionService) SetCache(c cache.Cache) {
	s.cache = c
}

func canChangeStartDate(startDate time.Time) bool {
    now := time.Now()
    start := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, time.Local)
    current := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
    return start.After(current) || start.Equal(current)
}
/*func canChangeStartDate(startDate time.Time) bool {
	now := time.Now().In(time.Local)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	start := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.Local)
	return start.After(today)
}*/