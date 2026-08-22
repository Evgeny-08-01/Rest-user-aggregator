// ============================================================
// ПАКЕТ: cache
// ============================================================
// Это инструмент (обёртка) для работы с Redis.
// ============================================================
// Что делает пакет:
// Функция	             Что делает
// Init()	             Подключается к Redis
// Get()	             Читает из Redis
// Set()	             Пишет в Redis
// Delete()	             Удаляет из Redis
// DeleteByPattern()	 Удаляет по шаблону
// GetClient()	         Возвращает клиент Redis для тестов
// ============================================================
package cache

import (
	"context"
	"fmt"
	"time" // Для установки TTL (время жизни кеша)

	"Rest-user-agregator/pkg/logger"

	"github.com/redis/go-redis/v9"
)

// ============================================================
// 1. ГЛОБАЛЬНАЯ ПЕРЕМЕННАЯ (клиент Redis)
// ============================================================
// client — указатель на подключение к Redis.
// Доступна только внутри пакета (приватная).
// Используется всеми функциями пакета.
// ============================================================
var client *redis.Client

// ============================================================
// 2. ИНИЦИАЛИЗАЦИЯ ПОДКЛЮЧЕНИЯ К REDIS
// ============================================================
// InitRedis — создаёт подключение к Redis и проверяет его работоспособность.
//
// ПАРАМЕТРЫ:
//   - addr: адрес Redis (например, "localhost:6379" или "redis:6379")
//   - password: пароль (оставляем пустым, если нет)
//   - db: номер базы данных (0 по умолчанию)
//
// ВОЗВРАЩАЕТ:
//   - error: ошибка, если не удалось подключиться
//
// КОГДА ВЫЗЫВАТЬ:
//   - При старте сервера (в main.go)
// ============================================================
var enabled bool

func InitRedis(addr, password string, db int) error {

	// 1. СОЗДАЁМ КЛИЕНТ REDIS
	//    redis.NewClient — создаёт структуру с настройками.
	//    Физическое подключение откроется при первом запросе (Ping).
	client = redis.NewClient(&redis.Options{
		Addr:     addr,     // Адрес: хост:порт
		Password: password, // Пароль (если есть)
		DB:       db,       // Номер базы (по умолчанию 0)
	})

	// 2. ПРОВЕРЯЕМ ПОДКЛЮЧЕНИЕ
	//    Создаём контекст с таймаутом 3 секунды.
	//    Если Redis не ответит за 3 секунды — считаем, что он недоступен.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // Освобождаем ресурсы контекста

	// Ping — отправляет команду PING в Redis.
	// Если Redis отвечает PONG — подключение работает.
	if err := client.Ping(ctx).Err(); err != nil {
		logger.Error("Redis connection failed: %v", err)
		return err
	}

	// 3. ЛОГИРУЕМ УСПЕХ
	enabled = true
	logger.Info("Redis connected successfully: %s", addr)
	return nil
}

// ============================================================
// 3. ПОЛУЧЕНИЕ ДАННЫХ ИЗ КЕША
// ============================================================
// Get — читает значение по ключу и возвращает его как int.
//
// ПАРАМЕТРЫ:
//   - ctx: контекст (для таймаутов и отмены)
//   - key: строка — ключ в Redis (например, "total:user:date")
//
// ВОЗВРАЩАЕТ:
//   - int: значение (0, если ключа нет)
//   - error: ошибка, если Redis не отвечает
//
// ОСОБЕННОСТИ:
//   - Если ключ не найден — возвращает 0, nil (без ошибки)
//   - redis.Nil — специальная ошибка, означающая "ключ не найден"
// ============================================================
func Get(ctx context.Context, key string) (int, error) {
	if !enabled || client == nil {
		logger.Debug("Cache disabled or client not initialized, skipping operation")
		return 0, nil
	}
	// client.Get — выполняет команду GET в Redis.
	// .Int() — преобразует ответ в int.
	// Если ключа нет — возвращает ошибку redis.Nil.
	val, err := client.Get(ctx, key).Int()

	// Проверяем: если ключ не найден — возвращаем 0, без ошибки
	if err == redis.Nil {
		return 0, nil
	}
	// Если другая ошибка (например, Redis упал) — возвращаем её
	return val, err
}

// ============================================================
// 4. СОХРАНЕНИЕ ДАННЫХ В КЕШ
// ============================================================
// Set — сохраняет значение в Redis с указанным временем жизни.
//
// ПАРАМЕТРЫ:
//   - ctx: контекст (для таймаутов и отмены)
//   - key: строка — ключ в Redis
//   - value: int — значение (например, 3600)
//   - ttl: время жизни (например, 5*time.Minute)
//
// ВОЗВРАЩАЕТ:
//   - error: ошибка, если не удалось сохранить
//
// КОГДА ИСПОЛЬЗОВАТЬ:
//   - После тяжёлого запроса к БД, чтобы закешировать результат
// ============================================================
func Set(ctx context.Context, key string, value int, ttl time.Duration) error {
	if !enabled || client == nil {
		logger.Debug("Cache disabled or client not initialized, skipping operation")
		return nil
	}
	// client.Set — выполняет команду SET в Redis.
	// Сохраняет ключ → значение с автоматическим удалением через ttl.
	return client.Set(ctx, key, value, ttl).Err()
}

// ============================================================
// 5. УДАЛЕНИЕ ОДНОГО КЛЮЧА
// ============================================================
// Delete — удаляет один ключ из Redis.
//
// ПАРАМЕТРЫ:
//   - ctx: контекст
//   - key: строка — ключ для удаления
//
// ВОЗВРАЩАЕТ:
//   - error: ошибка, если не удалось удалить
//
// КОГДА ИСПОЛЬЗОВАТЬ:
//   - Когда нужно удалить конкретный кеш (например, при обновлении подписки)
// ============================================================
func Delete(ctx context.Context, key string) error {
	if !enabled || client == nil {
		logger.Debug("Cache disabled or client not initialized, skipping operation")
		return nil
	}
	return client.Del(ctx, key).Err()
}

// ============================================================
// 6. УДАЛЕНИЕ ВСЕХ КЛЮЧЕЙ ПО ШАБЛОНУ
// ============================================================
// DeleteByPattern — удаляет все ключи, соответствующие шаблону.
//
// ПАРАМЕТРЫ:
//   - ctx: контекст
//   - pattern: шаблон (например, "total:user:*" — удалит все ключи для пользователя)
//
// ВОЗВРАЩАЕТ:
//   - error: ошибка, если не удалось удалить
//
// КОГДА ИСПОЛЬЗОВАТЬ:
//   - Когда пользователь обновил подписку — нужно удалить все его кеши
// ============================================================
func DeleteByPattern(ctx context.Context, pattern string) error {
	if !enabled || client == nil {
		logger.Debug("Cache disabled or client not initialized, skipping operation")
		return nil
	}
	iter := client.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		if err := client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}

	return iter.Err()
}

// ============================================================
// 7. ПОЛУЧЕНИЕ КЛИЕНТА REDIS ДЛЯ ТЕСТОВ
// ============================================================
// GetClient — возвращает клиент Redis для тестов.
// Используется в интеграционных тестах для проверки ключей.
// ============================================================
func GetClient() *redis.Client {
	return client
}
// ============================================================
// 8. СТРУКТУРА RedisCache — РЕАЛИЗАЦИЯ ИНТЕРФЕЙСА Cache
// ============================================================
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache — конструктор RedisCache
func NewRedisCache() *RedisCache {
	return &RedisCache{
		client: client,
	}
}

// Get — реализация интерфейса Cache
func (r *RedisCache) Get(ctx context.Context, key string) (int, error) {
	if r.client == nil {
		return 0, nil
	}
	val, err := r.client.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

// Set — реализация интерфейса Cache
func (r *RedisCache) Set(ctx context.Context, key string, value int, ttl time.Duration) error {
	if r.client == nil {
		return nil
	}
	return r.client.Set(ctx, key, value, ttl).Err()
}

// Delete — реализация интерфейса Cache
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	if r.client == nil {
		return nil
	}
	return r.client.Del(ctx, key).Err()
}

// Keys — реализация интерфейса Cache
func (r *RedisCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	if r.client == nil {
		return []string{}, nil
	}
	return r.client.Keys(ctx, pattern).Result()
}
// PingWithContext проверяет соединение с Redis с контекстом
func PingWithContext(ctx context.Context) error {
	if client == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return client.Ping(ctx).Err()
}
// ============================================================
// 9. ЗАКРЫТИЕ ПОДКЛЮЧЕНИЯ К REDIS
// ============================================================
// Close — закрывает соединение с Redis.
// Безопасно вызывать даже если клиент не инициализирован.
// ============================================================
func Close() error {
    if client == nil {
        return nil
    }
    err := client.Close()
    client = nil
    return err
}