// Package handlers - пакет для обработки запросов, содержит 6 хэндлеров
package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/Evgeny-08-01/Rest-user-aggregator/internal/database"
	"github.com/Evgeny-08-01/Rest-user-aggregator/internal/models"
	"github.com/google/uuid"
)

// @Summary      Создать подписку
// @Description  Добавляет новую подписку в базу данных
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        request body models.Subscription true "Данные подписки"
// @Success      201  {object}  map[string]int
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /subscriptions [post]
// 1. CreateSubscriptionHandler-Хэндлер записи одной строки
func CreateSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	var req models.Subscription
	// 1. Распарсить JSON
	err := parseJSON(r, &req)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// 2. Валидация
	err = validateSubscription(req)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Validation error")
		return
	}	

	// 3. Вызвать функцию database.CreateSubscription-создаем запись в базе данных
	id, err := database.CreateSubscription(req)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Database error")
		return
	}
	// 4. Ответ-передаем id созданной записи в базе данных в виде JSON в теле ответа хэндлера 
	writeJSON(w, http.StatusCreated, map[string]int{"id": id})
}

// @Summary      Получить подписку по ID
// @Tags         subscriptions
// @Accept       json          
// @Produce      json          
// @Param        id   path      int  true  "ID подписки"
// @Success      200  {object}  models.Subscription
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /subscriptions/{id} [get]
// 2. GetSubscriptionHandler-Хэндлер чтения одной строки
func GetSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получить id
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	// 2. Вызвать database.GetSubscription
	sub, err := database.GetSubscriptionByID(id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if sub == nil {
		writeJSONError(w, http.StatusNotFound, "Subscription not found")
		return
	}
	// 3. Ответ	
	writeJSON(w, http.StatusOK, sub)
}
// @Summary     Хэндлер обновления одной строки
// @Tags        subscriptions
// @Accept      json
// @Produce     json
// @Param        id   path      int  true  "ID подписки"
// @Param        request body models.Subscription true  "Новые данные"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router        /subscriptions/{id} [put]
// 3. UpdateSubscriptionHandler-Хэндлер обновления одной строки
func UpdateSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	var req models.Subscription
	// 1. Получить id из url
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	req.ID = id
	// 2. Распарсить JSON
	err = parseJSON(r, &req)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// 3. Валидация
	err = validateSubscription(req)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Validation error")
		return
	}
	// 4.  Вызвать database.UpdateSubscription
	err = database.UpdateSubscription(req)
if err != nil {
    if err == sql.ErrNoRows {
        writeJSONError(w, http.StatusNotFound, "Subscription not found")
    } else {
        writeJSONError(w, http.StatusInternalServerError, "Database error")
    }
    return
}
	// 5. Ответ
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// @Summary      Хэндлер удаления строки по id
// @Tags        subscriptions
// @Accept      json
// @Produce     json
// @Param        id   path      int  true  "ID подписки"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /subscriptions/{id} [delete]
// 4. DeleteSubscriptionHandler-Хэндлер удаления строки по id
func DeleteSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получить id
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	// 2. Вызвать database.DeleteSubscription

	err = database.DeleteSubscription(id)
if err != nil {
    if err == sql.ErrNoRows {
        writeJSONError(w, http.StatusNotFound, "Subscription not found")
    } else {
        writeJSONError(w, http.StatusInternalServerError, "Database error")
    }
    return
}
		// 3. Ответ
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// @Summary      Хэндлер чтения всех строк по фильтру
// @Tags        subscriptions
// @Accept      json
// @Produce     json
// @Param        limit   query     int  false  "Лимит фильтрации (должен быть > 0)"
// @Param        offset  query     int  false  "Офсет фильтрации (должен быть >= 0)"
// @Success      200     {array}   models.Subscription  "Массив подписок"
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router        /subscriptions [get]
// 5. ListSubscriptionsHandler-Хэндлер чтения всех строк по фильтру
func ListSubscriptionsHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Валидация
	// Получить параметры limit и offset из URL(если их нет,то предустановка в программе)
	limit := 20
	offset := 0
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "Invalid limit")
			return
		}
		if parsed > 0 {
			limit = parsed
		}
	}
	offsetStr := r.URL.Query().Get("offset")
if offsetStr != "" {
    parsed, err := strconv.Atoi(offsetStr)
    if err != nil {
        writeJSONError(w, http.StatusBadRequest, "Invalid offset")
        return
    }
    if parsed < 0 {
        writeJSONError(w, http.StatusBadRequest, "Negative offset")
        return
    }
    offset = parsed
	}
	// 2. Вызвать database.ListSubscriptions
	list, err := database.ListSubscriptions(limit, offset)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to get subscriptions")
		return
	}
	// 3. Ответ	
	writeJSON(w, http.StatusOK, list)
}

// @Summary      Хэндлер для подсчета суммарной стоимости всех подписок за выбранный период
// @Tags        subscriptions
// @Accept      json
// @Produce     json
// @Param        user_id       query     string  false  "ID пользователя" format(uuid)
// @Param        service_name  query     string  false  "Название подписки"
// @Param        start_date    query     string  false  "Дата начала (MM-YYYY)"
// @Param        end_date      query     string  false  "Дата окончания (MM-YYYY) или пустое значение = без верхней границы"
// @Success      200  {object}  map[string]int  "суммарная стоимость всех подписок"
// @Failure      500 {object}  map[string]string
// @Router       /subscriptions/total-cost [get]
//  6. GetTotalCostHandler-Хэндлер для подсчета суммарной стоимости всех подписок за
//     выбранный период с фильтрацией по id пользователя и названию подписки
func GetTotalCostHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получаем параметры из URL (не парсим JSON)
	userID := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
// 2. Валидация
    _, err := uuid.Parse(userID)
	if err != nil {
        writeJSONError(w, http.StatusBadRequest, "user_id must always be a valid UUID")
        return
    }

	//  3. Вызвать database.GetTotalCost
	total, err := database.GetTotalCost(userID, serviceName, startDate, endDate)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to get subscriptions")
		return
	}
	// 3. Ответ
	writeJSON(w, http.StatusOK, map[string]int{"total": total})
}
