package database

// Файл database_CRUDL_func-файл с функциями CRUDL
import (
	"database/sql"
	"fmt"
	"strings"
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
	var startDateDB, endDateDB time.Time
	err := row.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &startDateDB, &endDateDB)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // если подписки по id нет, то возвращаем nil
		}
		return nil, err
	}
sub.StartDate = startDateDB.Format("01-2006")
if !endDateDB.IsZero() {
    sub.EndDate = endDateDB.Format("01-2006")
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
	var startDate, endDate time.Time
	for rows.Next() {
		var sub models.Subscription
		err := rows.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &startDate, &endDate)
		if err != nil {
			return nil, err
		}

sub.StartDate = startDate.Format("01-2006")
if !endDate.IsZero() {
    sub.EndDate = endDate.Format("01-2006")
}
		subscriptions = append(subscriptions, sub)
	}
	return subscriptions, nil
}

// GetTotalCost - возвращает суммарную стоимость подписок за период с фильтрацией
func GetTotalCost(userID, serviceName, startDate, endDate string) (int, error) {
// 	query := `SELECT id, service_name, price, user_id, start_date, end_date 
//              FROM subscriptions 
 //             ORDER BY user_id, id
//              LIMIT $1 OFFSET $2`

startDateDB, err := time.Parse("01-2006", startDate)
	 if err != nil {
		return 0, err
		}
	var  endDateDB *time.Time
		if endDate!="" {
			 tempVar, err := time.Parse("01-2006", endDate)
	 if err != nil {
		return 0, err
		} 
		endDateDB = &tempVar
	}
	query := `SELECT COALESCE(SUM(price), 0) FROM subscriptions WHERE 1=1`
	var args []any
	x := 1
	if userID != "" {
		query += fmt.Sprintf(" AND user_id = $%d", x)
		args = append(args, userID)
		x++
	}
	if serviceName != "" {
		query += fmt.Sprintf(" AND service_name = $%d", x)
		args = append(args, serviceName)
		x++
	}
	if startDate != "" {
		startDateDB := convertToDatabase(startDate)
		query += fmt.Sprintf(" AND start_date >= $%d", x)
		args = append(args, startDateDB)
		x++
	}
	if endDate != "" {
		endDateDB := convertToDatabase(endDate)
		query += fmt.Sprintf(" AND start_date  <= $%d", x)
		args = append(args, endDateDB)
	}

	var total int
	err := DB.QueryRow(query, args...).Scan(&total)
	return total, err
}

func convertToDatabase(date string) string {
	length := strings.Split(date, "-")
	return length[1] + "-" + length[0]
}
func convertFromDatabase(date string) string {
	length := strings.Split(date, "-")
	return length[1] + "-" + length[0]
}