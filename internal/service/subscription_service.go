// package service реализует бизнес логику сервисов
package service

import (
	"context"
	"fmt"
	"time"

	"Rest-user-agregator/internal/cache"
	"Rest-user-agregator/internal/models"
	"Rest-user-agregator/internal/repository"
	"Rest-user-agregator/pkg/logger"

	"github.com/redis/go-redis/v9"
)
type SubscriptionService struct {
    repo repository.SubscriptionRepository
	cache cache.Cache
}
// NewSubscriptionService — конструктор сервиса

func NewSubscriptionService(repo repository.SubscriptionRepository) *SubscriptionService {
    return &SubscriptionService{
        repo:  repo,
        cache: cache.NewRedisCache(), // ← инициализируем реальным кешем
    }
}
// GetTotalCost — рассчитывает суммарную стоимость подписок за указанный период.
// Параметры:
//   - ctx: контекст для управления временем жизни запроса
//   - userID: идентификатор пользователя (обязательный, извлекается из JWT)
//   - serviceName: название сервиса (опционально, фильтр)
//   - startDate: дата начала периода в формате MM-YYYY (обязательно)
//   - endDate: дата окончания периода в формате MM-YYYY (опционально)
//
// Логика работы с кешем (Redis):
//  1. Получаем текущую версию кеша пользователя из таблицы cache_control_user в БД.
//     - Версия хранится в БД, чтобы обеспечить 100% консистентность.
//     - Если версию получить не удалось (ошибка БД) — кеш отключается, идём в БД.
//  2. Строим ключ для Redis с учётом версии и всех параметров запроса:
//       total:v{version}:{userID}:{serviceName}:{startDate}:{endDate}
//     - Пример: total:v3:123:yandex:01-2025:12-2025
//     - Если serviceName пустой — подставляется пустая строка.
//  3. Проверяем наличие кеша в Redis.
//     - Если есть — возвращаем (кеш-попадание).
//     - Если нет — идём в БД, сохраняем результат в Redis с TTL.
//
// Почему версия в БД:
//   - При изменении подписок (Create/Update/Delete) мы инкрементим версию в БД.
//   - Если Redis упал и восстановился — старые ключи не читаются (версия изменилась).
//   - Если БД cache_control_user недоступна — кеш отключается, идём напрямую в БД.
//   - Это даёт 100% гарантию, что пользователь никогда не увидит устаревшие данные.
//
// Возвращает:
//   - total: суммарная стоимость (int)
//   - error: ошибка, если парсинг дат не удался или диапазон невалидный
func (s *SubscriptionService) GetTotalCost(ctx context.Context, userID, serviceName, startDate, endDate string) (int, error) {
	// 1. Устанавливаем таймаут на выполнение всего запроса (кеш + БД)
	//    Если операция займёт больше 3 секунд — контекст отменится.
	//    3 секунды достаточно для БД (50-200 мс) + кеш (1 мс),
	//    но защищает от "зависших" запросов под нагрузкой.
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// ============================================================
	// 2. ПАРСИНГ ДАТ (обязательная валидация перед любыми действиями)
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

	// Проверяем, что start_date <= end_date
	if err := validateDateRange(startDateTimeDB, endDateTimeDB); err != nil {
		logger.Warn("GetTotalCost: invalid date range: startDate=%s > endDate=%s", startDate, endDate)
		return 0, err
	}

	// ============================================================
	// 3. ПОЛУЧАЕМ ВЕРСИЮ КЕША ИЗ БД
	// ============================================================
	// Версия хранится в БД, чтобы гарантировать консистентность.
	// Если БД недоступна — мы не используем кеш, идём напрямую в БД.
	version, err := s.repo.GetCacheUserVersion(ctx, userID)
	if err != nil {
		// При любой ошибке (таблица не найдена, соединение разорвано, таймаут) —
		// выключаем кеш и идём напрямую в БД.
		logger.Warn("GetTotalCost: failed to get cache version for user %s: %v, skipping cache", userID, err)
		return s.repo.GetTotalCost(ctx, userID, serviceName, startDateTimeDB, endDateTimeDB)
	}

	// ============================================================
	// 4. СТРОИМ КЛЮЧ ДЛЯ REDIS
	// ============================================================
	// Ключ включает ВСЕ параметры запроса и версию:
	//   total:v{version}:{userID}:{serviceName}:{startDate}:{endDate}
	//
	// Это гарантирует, что:
	//   - При изменении данных (инкремент версии) старый кеш перестаёт читаться.
	//   - Разные комбинации фильтров кешируются отдельно.
	//   - Пустой serviceName — допустимо.
	cacheKey := fmt.Sprintf("total:v%d:%s:%s:%s:%s",
		version,
		userID,
		serviceName,
		startDate,
		endDate,
	)

	// ============================================================
	// 5. ПРОВЕРЯЕМ НАЛИЧИЕ КЕША В REDIS
	// ============================================================
	// Если Redis доступен и ключ существует — возвращаем значение.
	// При ошибке Redis (недоступен, таймаут) — игнорируем кеш, идём в БД.
	cachedValue, err := s.cache.Get(ctx, cacheKey)
		logger.Debug("cacheKey:%s, ctx:%s",cacheKey,ctx )
	if err == nil && cachedValue > 0 {
		// Кеш-попадание: возвращаем сохранённое значение.
		logger.Debug("GetTotalCost: cache hit for key %s, value=%d", cacheKey, cachedValue)
		return cachedValue, nil
	}

	// Если Redis вернул ошибку (не redis.Nil, а реальная ошибка подключения) —
	// логируем предупреждение, но продолжаем работу через БД.
	// redis.Nil — это не ошибка, а штатное "ключа нет" (кеш-промах).
	if err != nil && err != redis.Nil {
		logger.Warn("GetTotalCost: Redis error for key %s: %v, falling back to DB", cacheKey, err)
	}

	// ============================================================
	// 6. КЕШ-ПРОМАХ: ИДЁМ В БД
	// ============================================================
	// Кеш-промах — штатная ситуация (первый запрос, TTL истёк, данные изменились).
	// Логируем Debug (не Warn), потому что это нормальная работа.
	logger.Debug("GetTotalCost: cache miss for key %s, querying DB", cacheKey)
	total, err := s.repo.GetTotalCost(ctx, userID, serviceName, startDateTimeDB, endDateTimeDB)
	if err != nil {
		logger.Error("GetTotalCost: DB query failed: %v", err)
		return 0, err
	}

	// ============================================================
	// 7. СОХРАНЯЕМ РЕЗУЛЬТАТ В REDIS (если доступен)
	// ============================================================
	// Сохраняем с TTL = 5 минут.
	// TTL достаточно короткий, чтобы даже при сбое инвалидации
	// данные автоматически протухли через 5 минут.
	const ttl = 5 * time.Minute
	if err := s.cache.Set(ctx, cacheKey, total, ttl); err != nil {
		// Если Redis недоступен — логируем ошибку, но не останавливаем работу.
		// Пользователь всё равно получит корректные данные из БД.
		logger.Warn("GetTotalCost: failed to cache result for key %s: %v", cacheKey, err)
	} else {
		logger.Debug("GetTotalCost: cached result for key %s, total=%d, ttl=%v", cacheKey, total, ttl)
	}

	// Возвращаем результат из БД
	return total, nil
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
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()
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
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()
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
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
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
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()
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
// SetCache — устанавливает кеш для сервиса (используется в тестах)
func (s *SubscriptionService) SetCache(c cache.Cache) {
    s.cache = c
}