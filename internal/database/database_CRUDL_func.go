package database

// Файл database_CRUDL_func-файл с функциями CRUDL
import (
	"database/sql"
	"strings"

	"github.com/Evgeny-08-01/Rest-user-aggregator/internal/models"
)

// CreateSubscription : 1 ФУНКЦИЯ== добавляет подписку в конец БД и ******** Create
// возвращает id+error
func CreateSubscription(sub models.Subscription) (int, error) {
var id int64
	query := `INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date) VALUES (?, ?, ?, ?, ?)`
    startDate := convertToDatabase(sub.StartDate)
    endDate := ""
    if sub.EndDate != "" {
        endDate = convertToDatabase(sub.EndDate)
    }
     res,err := DB.Exec(query, sub.ServiceName, sub.Price, sub.UserID, startDate, endDate)
	if err != nil {
    return 0, err}
		id, err = res.LastInsertId()
	return int(id), err
}

// GetSubscriptionByID : 2 ФУНКЦИЯ==  получение подписки по ID***************** Read
func GetSubscriptionByID(id int) (*models.Subscription, error) {
	 query := `SELECT id, service_name, price, user_id, start_date, end_date FROM subscriptions WHERE id = ?`
    row := DB.QueryRow(query, id)
var sub models.Subscription
    var startDateDB, endDateDB string
err := row.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &startDateDB, &endDateDB)
if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil    // если подписки по id нет, то возвращаем nil
        }
        return nil, err
    }
sub.StartDate = convertFromDatabase(startDateDB)
if endDateDB != "" {
    sub.EndDate = convertFromDatabase(endDateDB)
}
	return &sub, nil
	}

// UpdateSubscription : 3 ФУНКЦИЯ== - обновление подписки*********************** Update
func UpdateSubscription(sub models.Subscription) error {
   startDateDB := convertToDatabase(sub.StartDate)          
    endDateDB := ""                                     
    if sub.EndDate != "" {                             
        endDateDB = convertToDatabase(sub.EndDate)           
    } 
    query := `UPDATE subscriptions SET service_name = ?, price = ?, user_id = ?, start_date = ?, end_date = ? WHERE id = ?`
    _, err := DB.Exec(query, sub.ServiceName, sub.Price, sub.UserID, startDateDB, endDateDB, sub.ID)
    return err
}

// DeleteSubscription : 4 ФУНКЦИЯ== -  удаляет подписку по ID     *************** Delete
func DeleteSubscription(id int) error {
    query := `DELETE FROM subscriptions WHERE id = ?`
    _, err := DB.Exec(query, id)
    return err
}

// ListSubscriptions : 5 ФУНКЦИЯ== - получение списка подписок,
// отсортированный по user_id + по id, с пагинацией(limit, offset)  *************** List 
// ListSubscriptions - возвращает список подписок с пагинацией, отсортированный по user_id и id
func ListSubscriptions(limit, offset int) ([]models.Subscription, error) {
    query := `SELECT id, service_name, price, user_id, start_date, end_date 
              FROM subscriptions 
              ORDER BY user_id, id
              LIMIT ? OFFSET ?`
    
    rows, err := DB.Query(query, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var subscriptions []models.Subscription
    for rows.Next() {
        var sub models.Subscription
        err := rows.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &sub.StartDate, &sub.EndDate)
        if err != nil {
            return nil, err
        }
           sub.StartDate = convertFromDatabase(sub.StartDate)
if sub.EndDate != "" {
    sub.EndDate = convertFromDatabase(sub.EndDate)
} 
        subscriptions = append(subscriptions, sub)
    }

    return subscriptions, nil
}

// GetTotalCost - возвращает суммарную стоимость подписок за период с фильтрацией
func GetTotalCost(userID, serviceName, startDate, endDate string) (int, error) {
    query := `SELECT COALESCE(SUM(price), 0) FROM subscriptions WHERE 1=1`
    var args []any

    if userID != "" {
        query += " AND user_id = ?"
        args = append(args, userID)
    }
    if serviceName != "" {
        query += " AND service_name = ?"
        args = append(args, serviceName)
    }
     if startDate != "" {
    startDateDB := convertToDatabase(startDate)
    query += " AND start_date >= ?"
    args = append(args, startDateDB)
}
    if endDate != "" {
    endDateDB := convertToDatabase(endDate)
    query += " AND start_date <= ?"
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
