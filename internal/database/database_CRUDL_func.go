package database

// Файл database_CRUDL_func-файл с функциями CRUDL
import (
	"context"
	"database/sql"
	"fmt"

	"strconv"
	"time"

	"github.com/Evgeny-08-01/Rest-user-aggregator/internal/models"
	"github.com/Evgeny-08-01/Rest-user-aggregator/pkg/logger"
)

// CreateSubscription : 1 ФУНКЦИЯ== добавляет подписку в конец БД и ******** Create
// возвращает id+error
func CreateSubscription(ctx context.Context,sub models.Subscription) (int, error) {
	var id int
	query := `INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date) VALUES ($1,$2,$3,$4,$5) RETURNING id`
	startDate, err := time.Parse("01-2006", sub.StartDate)
	 if err != nil {
		    logger.Warn("CreateSubscription: failed to parse start_date %s: %v", sub.StartDate, err)
		return 0, err
		}
	var  endDate *time.Time
		if sub.EndDate != "" {
	 tempVar, err := time.Parse("01-2006", sub.EndDate)
	 if err != nil {
		 logger.Warn("CreateSubscription: failed to parse end_date %s: %v", sub.EndDate, err)
		return 0, err
		} 
		endDate = &tempVar}
	err = db.QueryRowContext(ctx,query, sub.ServiceName, sub.Price, sub.UserID, startDate, endDate).Scan(&id)
if err != nil {
    logger.Error("CreateSubscription: failed to insert subscription (service=%s, user_id=%s): %v", 
               sub.ServiceName, sub.UserID, err)
    return 0, err
}
logger.Debug("CreateSubscription: successfully created subscription id=%d for user_id=%s, service=%s", 
           id, sub.UserID, sub.ServiceName)
	return id, err
}

// GetSubscriptionByID : 2 ФУНКЦИЯ==  получение подписки по ID***************** Read
func GetSubscriptionByID(ctx context.Context,id int) (*models.Subscription, error) {
	query := `SELECT id, service_name, price, user_id, start_date, end_date  FROM subscriptions WHERE id = $1`
	row := db.QueryRowContext(ctx,query, id)
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


// UpdateSubscription : 3 ФУНКЦИЯ== - обновление подписки*********************** Update
func UpdateSubscription(ctx context.Context,sub models.Subscription) error {
  startDateDB, err := time.Parse("01-2006", sub.StartDate)
	 if err != nil {
		 logger.Warn("UpdateSubscription: failed to parse start_date %s: %v", sub.StartDate, err)
		return err
		}
	var  endDateDB *time.Time
		if sub.EndDate != "" {
	 tempVar, err := time.Parse("01-2006", sub.EndDate)
	 if err != nil {
		 logger.Warn("UpdateSubscription: failed to parse end_date %s: %v", sub.EndDate, err)
		return err
		} 
		endDateDB = &tempVar}
    query := `UPDATE subscriptions SET service_name = $1, price = $2, user_id = $3,
              start_date = $4, end_date = $5 WHERE id = $6`
    result, err := db.ExecContext(ctx,query, sub.ServiceName, sub.Price, sub.UserID, startDateDB, endDateDB, sub.ID)
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
// DeleteSubscription : 4 ФУНКЦИЯ== -  удаляет подписку по ID     *************** Delete
func DeleteSubscription(ctx context.Context,id int) error {
    query := `DELETE FROM subscriptions WHERE id = $1`
    result, err := db.ExecContext(ctx,query, id)
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
// ListSubscriptions : 5 ФУНКЦИЯ== - получение списка подписок,
// отсортированный по user_id + по id, с пагинацией(limit, offset)  *************** List
// ListSubscriptions - возвращает список подписок с пагинацией, отсортированный по user_id и id
func ListSubscriptions(ctx context.Context,limit, offset int) ([]models.Subscription, error) {
	query := `SELECT id, service_name, price, user_id, start_date, end_date 
              FROM subscriptions 
              ORDER BY user_id, id
              LIMIT $1 OFFSET $2`

	rows, err := db.QueryContext(ctx,query, limit, offset)
	if err != nil {
		 logger.Error("ListSubscriptions: query failed with limit=%d, offset=%d: %v", limit, offset, err)
		return nil, err
	}
	defer rows.Close()

	var subscriptions []models.Subscription
	var startDate time.Time
	var endDate sql.NullTime
	for rows.Next() {
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

// GetTotalCost - возвращает суммарную стоимость подписок за период с фильтрацией
func GetTotalCost(ctx context.Context,userID, serviceName, startDate, endDate string) (int, error) {
// startDate-стартовая дата, endDate-конечная дата просчитываемого периода, 
// указанного в задании на расчет- обязательные поля!!!
// startDateTimeDB-начало подписки, взятое из базы данных-обязательное поле
// endDateTimeDB-конец подписки, взятое из базы данных- не обязательное поле

startDateTimeDB, err := time.Parse("01-2006", startDate)
	 if err != nil {
		logger.Warn("GetTotalCost: failed to parse startDate %s: %v", startDate, err)
		return 0, fmt.Errorf("invalid startDate: %w", err)// startDate обязательное поле
		}
		var endDateTimeDB time.Time
		if endDate!="" {
			 tempVar, err2 := time.Parse("01-2006", endDate)
	 if err2 != nil {
		 logger.Warn("GetTotalCost: failed to parse endDate %s: %v", endDate, err2)
		return 0, fmt.Errorf("invalid endDate: %w", err)// endDate передан, но не соответствует формату MM-YYYY
		} 	
		endDateTimeDB = tempVar	
endDateTimeDB = time.Date(tempVar.Year(), tempVar.Month()+1, 0, 0, 0, 0, 0, time.UTC)// Превращаем первый день в последний			
	} else {
//    endDateTimeDB,_ = time.Parse("2006-01", "2100-01")// присваиваем максимальное время, если данных нет в базе
endDateTimeDB = time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
}
if startDateTimeDB.After(endDateTimeDB) {
	 logger.Warn("GetTotalCost: invalid date range: startDate=%s > endDate=%s", startDate, endDate)
    return 0, fmt.Errorf( "start_date > end_date")
}
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

    args := []interface{}{startDateTimeDB, endDateTimeDB }

    if userID != "" {
        query += " AND user_id = $" + strconv.Itoa(len(args)+1)
        args = append(args, userID)
    }
    if serviceName != "" {
        query += " AND service_name = $" + strconv.Itoa(len(args)+1)
        args = append(args, serviceName)
    }

    var total int
    err = db.QueryRowContext(ctx,query, args...).Scan(&total)
  if err != nil {
        logger.Error("GetTotalCost: query failed with userID=%s, serviceName=%s, startDate=%s, endDate=%s: %v", 
                   userID, serviceName, startDate, endDate, err)
        return 0, err
    }

    logger.Debug("GetTotalCost: successfully calculated total cost=%d for userID=%s, serviceName=%s, period=%s to %s", 
               total, userID, serviceName, startDate, endDate)
    return total, nil
}