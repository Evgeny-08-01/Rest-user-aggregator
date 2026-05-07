package database

// Файл database_CRUDL_func-файл с функциями CRUDL
import (
	"database/sql"
	"fmt"

	"strconv"
	"time"

	"github.com/Evgeny-08-01/Rest-user-aggregator/internal/models"
)

// CreateSubscription : 1 ФУНКЦИЯ== добавляет подписку в конец БД и ******** Create
// возвращает id+error
func CreateSubscription(sub models.Subscription) (int, error) {
	var id int
	query := `INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date) VALUES ($1,$2,$3,$4,$5) RETURNING id`
	startDate, err := time.Parse("01-2006", sub.StartDate)
	 if err != nil {
		return 0, err
		}
	var  endDate *time.Time
		if sub.EndDate != "" {
	 tempVar, err := time.Parse("01-2006", sub.EndDate)
	 if err != nil {
		return 0, err
		} 
		endDate = &tempVar}
	err = DB.QueryRow(query, sub.ServiceName, sub.Price, sub.UserID, startDate, endDate).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, err
}

// GetSubscriptionByID : 2 ФУНКЦИЯ==  получение подписки по ID***************** Read
func GetSubscriptionByID(id int) (*models.Subscription, error) {
	query := `SELECT id, service_name, price, user_id, start_date, end_date  FROM subscriptions WHERE id = $1`
	row := DB.QueryRow(query, id)
	var sub models.Subscription
	var startDateDB time.Time
	var endDateDB sql.NullTime
	err := row.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &startDateDB, &endDateDB)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // если подписки по id нет, то возвращаем nil
		}
		return nil, err
	}
sub.StartDate = startDateDB.Format("01-2006")
if endDateDB.Valid {
    sub.EndDate = endDateDB.Time.Format("01-2006")
}
	return &sub, nil
}


// UpdateSubscription : 3 ФУНКЦИЯ== - обновление подписки*********************** Update
func UpdateSubscription(sub models.Subscription) error {
  startDateDB, err := time.Parse("01-2006", sub.StartDate)
	 if err != nil {
		return err
		}
	var  endDateDB *time.Time
		if sub.EndDate != "" {
	 tempVar, err := time.Parse("01-2006", sub.EndDate)
	 if err != nil {
		return err
		} 
		endDateDB = &tempVar}
    query := `UPDATE subscriptions SET service_name = $1, price = $2, user_id = $3,
              start_date = $4, end_date = $5 WHERE id = $6`
    result, err := DB.Exec(query, sub.ServiceName, sub.Price, sub.UserID, startDateDB, endDateDB, sub.ID)
    if err != nil {
        return err
    }
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rowsAffected == 0 {
        return sql.ErrNoRows
    }
    return nil
}
// DeleteSubscription : 4 ФУНКЦИЯ== -  удаляет подписку по ID     *************** Delete
func DeleteSubscription(id int) error {
    query := `DELETE FROM subscriptions WHERE id = $1`
    result, err := DB.Exec(query, id)
    if err != nil {
        return err
    }
    exist, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if exist == 0 {
        return sql.ErrNoRows
    }
    return nil
}
// ListSubscriptions : 5 ФУНКЦИЯ== - получение списка подписок,
// отсортированный по user_id + по id, с пагинацией(limit, offset)  *************** List
// ListSubscriptions - возвращает список подписок с пагинацией, отсортированный по user_id и id
func ListSubscriptions(limit, offset int) ([]models.Subscription, error) {
	query := `SELECT id, service_name, price, user_id, start_date, end_date 
              FROM subscriptions 
              ORDER BY user_id, id
              LIMIT $1 OFFSET $2`

	rows, err := DB.Query(query, limit, offset)
	if err != nil {
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
			return nil, err
		}

sub.StartDate = startDate.Format("01-2006")
if endDate.Valid {
    sub.EndDate = endDate.Time.Format("01-2006")
}
		subscriptions = append(subscriptions, sub)
	}
	return subscriptions, nil
}

// GetTotalCost - возвращает суммарную стоимость подписок за период с фильтрацией
func GetTotalCost(userID, serviceName, startDate, endDate string) (int, error) {
// startDate-стартовая дата, endDate-конечная дата просчитываемого периода, 
// указанного в задании на расчет- обязательные поля!!!
// startDateTimeDB-начало подписки, взятое из базы данных-обязательное поле
// endDateTimeDB-конец подписки, взятое из базы данных- не обязательное поле

startDateTimeDB, err := time.Parse("01-2006", startDate)
	 if err != nil {
		return 0, fmt.Errorf("invalid startDate: %w", err)// startDate обязательное поле
		}
		var endDateTimeDB time.Time
		if endDate!="" {
			 tempVar, err2 := time.Parse("01-2006", endDate)
	 if err2 != nil {
		return 0, fmt.Errorf("invalid endDate: %w", err)// endDate передан, но не соответствует формату MM-YYYY
		} 	
		endDateTimeDB = tempVar	
endDateTimeDB = time.Date(tempVar.Year(), tempVar.Month()+1, 0, 0, 0, 0, 0, time.UTC)// Превращаем первый день в последний			
	} else {
//    endDateTimeDB,_ = time.Parse("2006-01", "2100-01")// присваиваем максимальное время, если данных нет в базе
endDateTimeDB = time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
}
if startDateTimeDB.After(endDateTimeDB) {
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

	//log.Printf("Query: %s", query)
  //  log.Printf("Args: startDateTimeDB=%v, endDateTimeDB=%v, userID=%s, serviceName=%s", 
 //   startDateTimeDB, endDateTimeDB, userID, serviceName)
    err = DB.QueryRow(query, args...).Scan(&total)
//	log.Printf("SQL error: %v", err) 
    return total, err
}
