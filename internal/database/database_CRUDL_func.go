package database

// Файл database_CRUDL_func-файл с методами CRUDL
import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"Rest-user-agregator/internal/models"
	"Rest-user-agregator/pkg/logger"
)

// CreateSubscription : 1 Метод== добавляет подписку в конец БД
// возвращает id+error
func (r *PostgresRepo) CreateSubscription(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) (int, error) {
// Проверка на дубликат
var count int
err := r.db.QueryRowContext(ctx,
    "SELECT COUNT(*) FROM subscriptions WHERE user_id = $1 AND template_id = $2",
    sub.UserID, sub.TemplateID,
).Scan(&count)
if err != nil {
    return 0, err
}
if count > 0 {
    return 0, errors.New("duplicate")
}	
    var id int
    query := `INSERT INTO subscriptions (user_id, template_id, start_date, end_date) VALUES ($1,$2,$3,$4) RETURNING id`
	err = r.db.QueryRowContext(ctx, query,
    sub.UserID,
    sub.TemplateID, // нужно добавить TemplateID в models.Subscription
    startDate,
    endDate,
).Scan(&id)

	if err != nil {
		logger.Error("CreateSubscription: failed to insert subscription (service=%s, user_id=%s): %v",
			sub.ServiceName, sub.UserID, err)
		return 0, err
	}

	logger.Debug("CreateSubscription: successfully created subscription id=%d for user_id=%s, service=%s",
		id, sub.UserID, sub.ServiceName)
	// инвалидируем кеш пользователя сразу после окончания создания подписки
	if err := r.IncrementCacheUserVersion(ctx, sub.UserID); err != nil {
		logger.Warn("CreateSubscription: failed to increment cache version for user %s: %v", sub.UserID, err)
	}
	return id, nil
}

// GetSubscriptionByID : 2 Метод==  получение подписки по ID***************** Read
func (r *PostgresRepo) GetSubscriptionByID(ctx context.Context, id int) (*models.Subscription, error) {
    query := `
        SELECT 
            s.id,
            s.user_id,
            s.start_date,
            s.end_date,
            t.service_name,
            t.price
        FROM subscriptions s
        LEFT JOIN subscription_templates t ON s.template_id = t.id
        WHERE s.id = $1
    `

    var sub models.Subscription
    var startDate time.Time
    var endDate sql.NullTime
    var serviceName sql.NullString
    var price sql.NullInt64

    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &sub.ID,
        &sub.UserID,
        &startDate,
        &endDate,
        &serviceName,
        &price,
    )
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        logger.Error("GetSubscriptionByID: scan failed for id=%d: %v", id, err)
        return nil, err
    }

    sub.StartDate = startDate.Format("01-2006")
    if endDate.Valid {
        sub.EndDate = endDate.Time.Format("01-2006")
    }

    if serviceName.Valid {
        sub.ServiceName = serviceName.String
    } else {
        sub.ServiceName = "Удалённый шаблон"
    }

    if price.Valid {
        sub.Price = int(price.Int64)
    } else {
        sub.Price = 0
    }

    return &sub, nil
}

// UpdateSubscription : 3 Метод== обновление подписки
// Принимает уже готовые startDate и endDate (time.Time)
func (r *PostgresRepo) UpdateSubscription(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) error {
/*	query := `UPDATE subscriptions SET user_id = $1, start_date = $2, end_date = $3 WHERE id = $4`

	result, err := r.db.ExecContext(ctx, query,
		sub.UserID,
		startDate,
		endDate,
		sub.ID,
	)*/
 query := `UPDATE subscriptions SET start_date = $1, end_date = $2 WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query,
		startDate,
		endDate,
		sub.ID,
	)   
	if err != nil {
		logger.Error("UpdateSubscription: exec failed for id %d: %v", sub.ID, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("UpdateSubscription: RowsAffected failed for id %d: %v", sub.ID, err)
		return err
	}
	if rowsAffected == 0 {
		logger.Warn("UpdateSubscription: no rows affected for id %d", sub.ID)
		return sql.ErrNoRows
	}

	logger.Debug("UpdateSubscription: successfully updated subscription id %d", sub.ID)
	// инвалидируем кеш пользователя сразу после окончания обновления подписки
	if err := r.IncrementCacheUserVersion(ctx, sub.UserID); err != nil {
		logger.Warn("UpdateSubscription: failed to increment cache version for user %s: %v", sub.UserID, err)
	} else {
		logger.Debug("IncrementCacheUserVersion  err==nil")
	}
	logger.Debug("UpdateSubscription: ctx=%v, sub.UserID=%s, sub.ID=%d", ctx, sub.UserID, sub.ID)
	return nil
}

// DeleteSubscription : 4 Метод== -  удаляет подписку по ID     *************** Delete
func (r *PostgresRepo) DeleteSubscription(ctx context.Context, id int) error {
	// 1. ПОЛУЧАЕМ user_id ДО УДАЛЕНИЯ (для инвалидации кеша)
	var userID string
	err := r.db.QueryRowContext(ctx, "SELECT user_id FROM subscriptions WHERE id = $1", id).Scan(&userID)
	if err != nil {
		logger.Warn("DeleteSubscription: failed to get user_id for subscription %d: %v", id, err)
		return err
	}

	// 2. ВЫПОЛНЯЕМ УДАЛЕНИЕ
	query := `DELETE FROM subscriptions WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		logger.Error("DeleteSubscription: exec failed for id %d: %v", id, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Warn("DeleteSubscription: RowsAffected failed for id %d: %v", id, err)
		return err
	}
	if rowsAffected == 0 {
		logger.Warn("DeleteSubscription: no rows affected for id %d", id)
		return sql.ErrNoRows
	}

	logger.Debug("DeleteSubscription: successfully deleted subscription id %d", id)

	// 3. ИНВАЛИДИРУЕМ КЕШ ПОЛЬЗОВАТЕЛЯ (user_id уже получен до удаления)
	if err := r.IncrementCacheUserVersion(ctx, userID); err != nil {
		logger.Warn("DeleteSubscription: failed to increment cache version for user %s: %v", userID, err)
	}

	return nil
}
// ============================================================
// ListSubscriptions — получение списка подписок с пагинацией
// ============================================================
// Что делает:
//   - Возвращает список подписок с данными из шаблонов (название, цена)
//   - Если userID == "" — возвращает все подписки (для админа)
//   - Если userID != "" — возвращает только подписки этого пользователя
//
// АЛИАСЫ:
//   - s — subscriptions (подписки)
//   - t — subscription_templates (шаблоны)
// ============================================================
func (r *PostgresRepo) ListSubscriptions(ctx context.Context, userID string, limit, offset int) ([]models.Subscription, error) {
    // ============================================================
    // 1. SQL-ЗАПРОС С ВЕТВЛЕНИЕМ
    // ============================================================
    var query string
    var args []any

    if userID == "" {
        // Админ: все подписки (без WHERE)
        query = `
            SELECT 
                s.id,           -- ID подписки
                s.user_id,      -- ID пользователя
                s.start_date,   -- Дата начала
                s.end_date,     -- Дата окончания
                t.service_name, -- Название из шаблона
                t.price         -- Цена из шаблона
            FROM 
                subscriptions AS s
            LEFT JOIN 
                subscription_templates AS t ON s.template_id = t.id
            ORDER BY 
                s.id
            LIMIT $1 OFFSET $2
        `
        args = []any{limit, offset}
    } else {
        // Пользователь: только свои подписки
        query = `
            SELECT 
                s.id,           -- ID подписки
                s.user_id,      -- ID пользователя
                s.start_date,   -- Дата начала
                s.end_date,     -- Дата окончания
                t.service_name, -- Название из шаблона
                t.price         -- Цена из шаблона
            FROM 
                subscriptions AS s
            LEFT JOIN 
                subscription_templates AS t ON s.template_id = t.id
            WHERE 
                s.user_id = $1
            ORDER BY 
                s.id
            LIMIT $2 OFFSET $3
        `
        args = []any{userID, limit, offset}
    }

    // ============================================================
    // 2. ВЫПОЛНЯЕМ ЗАПРОС
    // ============================================================
    rows, err := r.db.QueryContext(ctx, query, args...)
    if err != nil {
        logger.Error("ListSubscriptions: query failed: %v", err)
        return nil, err
    }
    defer rows.Close()

    // ============================================================
    // 3. ПРОХОДИМ ПО РЕЗУЛЬТАТАМ
    // ============================================================
    var subscriptions []models.Subscription

    for rows.Next() {
        var sub models.Subscription
        var startDate time.Time
        var endDate sql.NullTime
        var serviceName sql.NullString
        var price sql.NullInt64

        // Сканируем поля
        err := rows.Scan(
            &sub.ID,
            &sub.UserID,
            &startDate,
            &endDate,
            &serviceName,
            &price,
        )
        if err != nil {
            logger.Error("ListSubscriptions: scan failed: %v", err)
            return nil, err
        }

        // ============================================================
        // 4. ФОРМАТИРУЕМ ДАТЫ (MM-YYYY)
        // ============================================================
        sub.StartDate = startDate.Format("01-2006")
        if endDate.Valid {
            sub.EndDate = endDate.Time.Format("01-2006")
        }

        // ============================================================
        // 5. НАЗВАНИЕ И ЦЕНА (ЕСЛИ ШАБЛОН ЕСТЬ)
        // ============================================================
        if serviceName.Valid {
            sub.ServiceName = serviceName.String
        } else {
            sub.ServiceName = "Удалённый шаблон"
        }

        if price.Valid {
            sub.Price = int(price.Int64)
        } else {
            sub.Price = 0
        }

        // Добавляем в результат
        subscriptions = append(subscriptions, sub)
    }

    // ============================================================
    // 6. ПРОВЕРЯЕМ ОШИБКИ ПОСЛЕ ИТЕРАЦИИ
    // ============================================================
    if err = rows.Err(); err != nil {
        logger.Error("ListSubscriptions: rows error: %v", err)
        return nil, err
    }

    logger.Debug("ListSubscriptions: successfully fetched %d subscriptions (limit=%d, offset=%d)",
        len(subscriptions), limit, offset)
    return subscriptions, nil
}
// GetTotalCost -: 6 Метод возвращает суммарную стоимость подписок за период с фильтрацией
func (r *PostgresRepo) GetTotalCost(ctx context.Context, userID, serviceName string, startDate, endDate time.Time) (int, error) {
    // Базовый запрос: суммарная стоимость подписок за период
    query := `
        SELECT COALESCE(SUM(subscription_templates.price * (EXTRACT(MONTH FROM AGE(
            LEAST(COALESCE(subscriptions.end_date, 'infinity'), $2),
            GREATEST(subscriptions.start_date, $1)
        )) + 1)), 0)
        FROM subscriptions
        INNER JOIN subscription_templates ON subscriptions.template_id = subscription_templates.id
        WHERE subscriptions.start_date <= $2 
          AND (subscriptions.end_date IS NULL OR subscriptions.end_date >= $1)`

    // Аргументы: $1 = startDate, $2 = endDate
    args := []any{startDate, endDate}

    // Фильтр по user_id (если передан)
    if userID != "" {
        query += " AND subscriptions.user_id = $" + strconv.Itoa(len(args)+1)
        args = append(args, userID)
    }

    // Фильтр по service_name (если передан)
    if serviceName != "" {
        // Экранируем спецсимволы LIKE: % и _
        safeServiceName := strings.ReplaceAll(serviceName, "%", "\\%")
        safeServiceName = strings.ReplaceAll(safeServiceName, "_", "\\_")

        // ILIKE — регистронезависимый поиск
        // ESCAPE '\' — защита от спецсимволов
        query += " AND subscription_templates.service_name ILIKE '%' || $" + strconv.Itoa(len(args)+1) + " || '%' ESCAPE '\\'"
        args = append(args, safeServiceName)
    }

    var total int
    logger.Debug("GetTotalCost: query=%s, args=%v", query, args)
    err := r.db.QueryRowContext(ctx, query, args...).Scan(&total)
    if err != nil {
        logger.Error("GetTotalCost: query failed with userID=%s, serviceName=%s, startDate=%s, endDate=%s: %v",
            userID, serviceName, startDate, endDate, err)
        return 0, err
    }

    logger.Debug("GetTotalCost: successfully calculated total cost=%d for userID=%s, serviceName=%s, period=%s to %s",
        total, userID, serviceName, startDate, endDate)
    return total, nil
}