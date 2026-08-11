package database

// Файл database_CRUDL_func-файл с методами CRUDL
import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"Rest-user-agregator/internal/models"
	"Rest-user-agregator/pkg/logger"
)

// CreateSubscription : 1 Метод== добавляет подписку в конец БД
// возвращает id+error
func (r *PostgresRepo) CreateSubscription(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) (int, error) {
    var id int
    query := `INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date) VALUES ($1,$2,$3,$4,$5) RETURNING id`

    err := r.db.QueryRowContext(ctx, query,
        sub.ServiceName,
        sub.Price,
        sub.UserID,
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
    return id, nil
}
// GetSubscriptionByID : 2 Метод==  получение подписки по ID***************** Read
func (r *PostgresRepo) GetSubscriptionByID(ctx context.Context,id int) (*models.Subscription, error) {
	query := `SELECT id, service_name, price, user_id, start_date, end_date  FROM subscriptions WHERE id = $1`
	row := r.db.QueryRowContext(ctx,query, id)
	var sub models.Subscription
	var startDateDB time.Time
	var endDateDB sql.NullTime
	err := row.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &startDateDB, &endDateDB)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warn("GetSubscriptionByID: subscription with id=%d not found", id)
			return nil, nil // если подписки по id нет, то возвращаем nil
		}
		logger.Error("GetSubscriptionByID: scan failed for id=%d: %v", id, err)
		return nil, err
	}
sub.StartDate = startDateDB.Format("01-2006")
if endDateDB.Valid {
    sub.EndDate = endDateDB.Time.Format("01-2006")
}
logger.Debug("GetSubscriptionByID: successfully retrieved subscription id=%d for user_id=%s", 
               sub.ID, sub.UserID)
	return &sub, nil
}


// UpdateSubscription : 3 Метод== обновление подписки
// Принимает уже готовые startDate и endDate (time.Time)
func (r *PostgresRepo) UpdateSubscription(ctx context.Context, sub models.Subscription, startDate time.Time, endDate *time.Time) error {
    query := `UPDATE subscriptions SET service_name = $1, price = $2, user_id = $3,
              start_date = $4, end_date = $5 WHERE id = $6`

    result, err := r.db.ExecContext(ctx, query,
        sub.ServiceName,
        sub.Price,
        sub.UserID,
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
    return nil
}
// DeleteSubscription : 4 Метод== -  удаляет подписку по ID     *************** Delete
func (r *PostgresRepo) DeleteSubscription(ctx context.Context,id int) error {
    query := `DELETE FROM subscriptions WHERE id = $1`
    result, err := r.db.ExecContext(ctx,query, id)
    if err != nil {
	 logger.Error("DeleteSubscription: exec failed for id %d: %v", id, err)	
        return err
    }
    exist, err := result.RowsAffected()
    if err != nil {
		   logger.Warn("DeleteSubscription: RowsAffected failed for id %d: %v", id, err)
        return err
    }
    if exist == 0 {
		logger.Warn("DeleteSubscription: no rows affected for id %d", id)
        return sql.ErrNoRows
    }
	 logger.Debug("DeleteSubscription: successfully deleted subscription id %d", id)
    return nil
}
// ListSubscriptions : 5 Метод== - получение списка подписок,
// отсортированный по user_id + по id, с пагинацией(limit, offset)  *************** List
// ListSubscriptions - возвращает список подписок с пагинацией, отсортированный по user_id и id
func (r *PostgresRepo) ListSubscriptions(ctx context.Context,limit, offset int) ([]models.Subscription, error) {
	query := `SELECT id, service_name, price, user_id, start_date, end_date 
              FROM subscriptions 
              ORDER BY user_id, id
              LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx,query, limit, offset)
	if err != nil {
		 logger.Error("ListSubscriptions: query failed with limit=%d, offset=%d: %v", limit, offset, err)
		return nil, err
	}
	defer rows.Close()

	var subscriptions []models.Subscription
	var startDate time.Time
	var endDate sql.NullTime
	for rows.Next() {
    if ctx.Err() != nil {
		return nil, ctx.Err() 
    }
		var sub models.Subscription
		err := rows.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &startDate, &endDate)
		if err != nil {
			    logger.Error("ListSubscriptions: scan failed: %v", err)
			return nil, err
		}

sub.StartDate = startDate.Format("01-2006")
if endDate.Valid {
    sub.EndDate = endDate.Time.Format("01-2006")
}
		subscriptions = append(subscriptions, sub)
	}
	// Проверяем ошибки после завершения итерации
    if err = rows.Err(); err != nil {
        logger.Error("ListSubscriptions: rows iteration error: %v", err)
        return nil, err
    }

    logger.Debug("ListSubscriptions: successfully fetched %d subscriptions (limit=%d, offset=%d)", 
               len(subscriptions), limit, offset)
    return subscriptions, nil
}

// GetTotalCost -: 6 Метод возвращает суммарную стоимость подписок за период с фильтрацией
func (r *PostgresRepo) GetTotalCost(ctx context.Context,userID, serviceName string, startDate, endDate time.Time ) (int, error) {
// startDate-стартовая дата, endDate-конечная дата просчитываемого периода, 
// указанного в задании на расчет- обязательные поля!!!
// startDateTimeDB-начало подписки, взятое из базы данных-обязательное поле
// endDateTimeDB-конец подписки, взятое из базы данных- не обязательное поле
	    query := `
        SELECT COALESCE
		(SUM
		     ( price * (EXTRACT(                MONTH FROM AGE(    LEAST     (COALESCE(end_date, 'infinity'), $2),
                                                                   GREATEST                      (start_date, $1)
											                    )
					             )+1   
					    )
             ),
		 0) AS total
                      FROM subscriptions WHERE start_date <= $2 AND (end_date IS NULL OR end_date >= $1)`

    args := []interface{}{startDate, endDate }

    if userID != "" {
        query += " AND user_id = $" + strconv.Itoa(len(args)+1)
        args = append(args, userID)
    }
    if serviceName != "" {
    query += " AND LOWER(service_name) LIKE LOWER($" + strconv.Itoa(len(args)+1) + ")"
    args = append(args, "%" + serviceName + "%")
}

    var total int
    err := r.db.QueryRowContext(ctx,query, args...).Scan(&total)
  if err != nil {
        logger.Error("GetTotalCost: query failed with userID=%s, serviceName=%s, startDate=%s, endDate=%s: %v", 
                   userID, serviceName, startDate, endDate, err)
        return 0, err
    }

    logger.Debug("GetTotalCost: successfully calculated total cost=%d for userID=%s, serviceName=%s, period=%s to %s", 
               total, userID, serviceName, startDate, endDate)
    return total, nil
}