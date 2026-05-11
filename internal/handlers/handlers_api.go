// Package handlers - пакет для обработки запросов, содержит 6 хэндлеров
package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/Evgeny-08-01/Rest-user-aggregator/internal/models"
	"github.com/Evgeny-08-01/Rest-user-aggregator/internal/repository"
	"github.com/Evgeny-08-01/Rest-user-aggregator/pkg/logger"
)

// Handler - структура хендлера, принимающая репозиторий
type Handler struct {
    Repo repository.SubscriptionRepository
}

// NewHandler - конструктор хендлера
func NewHandler(repo repository.SubscriptionRepository) *Handler {
    return &Handler{Repo: repo}
}
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
func (h *Handler) CreateSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req models.Subscription
	// 1. Распарсить JSON
	err := parseJSON(r, &req)
	if err != nil {
		logger.Warn("CreateSubscriptionHandler: failed to parse JSON: %v", err)
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// 2. Валидация
	err = validateSubscription(req)
	if err != nil {
		logger.Warn("CreateSubscriptionHandler: validation failed: %v", err)
		writeJSONError(w, http.StatusBadRequest, "Validation error")
		return
	}	

	// 3. Вызвать функцию database.CreateSubscription-создаем запись в базе данных
	id, err := h.Repo.CreateMtd(ctx,req)
	if err != nil {
		logger.Error("CreateSubscriptionHandler: database error for user_id=%s, service=%s: %v",
			req.UserID, req.ServiceName, err)
		writeJSONError(w, http.StatusInternalServerError, "Database error")
		return
	}
	// 4. Ответ-передаем id созданной записи в базе данных в виде JSON в теле ответа хэндлера
	logger.Debug("CreateSubscriptionHandler: successfully created subscription id=%d for user_id=%s",
		id, req.UserID)
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
func (h *Handler) GetSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// 1. Получить id
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Warn("GetSubscriptionHandler: invalid ID format: %s", idStr)
		writeJSONError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	// 2. Вызвать database.GetSubscription
	sub, err := h.Repo.GetByIDMtd(ctx,id)
	if err != nil {
		logger.Error("GetSubscriptionHandler: database error for id=%d: %v", id, err)
		writeJSONError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if sub == nil {
		logger.Warn("GetSubscriptionHandler: subscription not found for id=%d", id)
		writeJSONError(w, http.StatusNotFound, "Subscription not found")
		return
	}
	// 3. Ответ	
	logger.Debug("GetSubscriptionHandler: successfully retrieved subscription id=%d", id)
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
// @Failure      404   {object}  map[string]string 
// @Failure      500  {object}  map[string]string
// @Router        /subscriptions/{id} [put]
// 3. UpdateSubscriptionHandler-Хэндлер обновления одной строки
func (h *Handler)UpdateSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req models.Subscription
	// 1. Получить id из url
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Warn("UpdateSubscriptionHandler: invalid ID format: %s", idStr)
		writeJSONError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	req.ID = id
	// 2. Распарсить JSON
	err = parseJSON(r, &req)
	if err != nil {
		logger.Warn("UpdateSubscriptionHandler: failed to parse JSON for id=%d: %v", id, err)
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// 3. Валидация
	err = validateSubscription(req)
	if err != nil {
		logger.Warn("UpdateSubscriptionHandler: validation failed for id=%d: %v", id, err)
		writeJSONError(w, http.StatusBadRequest, "Validation error")
		return
	}
	// 4.  Вызвать database.UpdateSubscription
	err = h.Repo.UpdateMtd(ctx,req)
if err != nil {
    if err == sql.ErrNoRows {
	logger.Warn("UpdateSubscriptionHandler: subscription not found for id=%d", id)
        writeJSONError(w, http.StatusNotFound, "Subscription not found")
    } else {
		logger.Error("UpdateSubscriptionHandler: database error for id=%d: %v", id, err)
        writeJSONError(w, http.StatusInternalServerError, "Database error")
    }
    return
}
	// 5. Ответ
	logger.Debug("UpdateSubscriptionHandler: successfully updated subscription id=%d", id)
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// @Summary      Хэндлер удаления строки по id
// @Tags        subscriptions
// @Accept      json
// @Produce     json
// @Param        id   path      int  true  "ID подписки"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]string
// @Failure      404  {object}  map[string]string 
// @Failure      500 {object} map[string]string
// @Router       /subscriptions/{id} [delete]
// 4. DeleteSubscriptionHandler-Хэндлер удаления строки по id
func (h *Handler)DeleteSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// 1. Получить id
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Warn("DeleteSubscriptionHandler: invalid ID format: %s", idStr)
		writeJSONError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	// 2. Вызвать database.DeleteSubscription

	err = h.Repo.DeleteMtd(ctx,id)
if err != nil {
    if err == sql.ErrNoRows {
		logger.Warn("DeleteSubscriptionHandler: subscription not found for id=%d", id)
        writeJSONError(w, http.StatusNotFound, "Subscription not found")
    } else {
		logger.Error("DeleteSubscriptionHandler: database error for id=%d: %v", id, err)
        writeJSONError(w, http.StatusInternalServerError, "Database error")
    }
    return
}
		// 3. Ответ
	logger.Debug("DeleteSubscriptionHandler: successfully deleted subscription id=%d", id)
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
func (h *Handler)ListSubscriptionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// 1. Валидация
	// Получить параметры limit и offset из URL(если их нет,то предустановка в программе)
	limit := 20
	offset := 0
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil {
		logger.Warn("ListSubscriptionsHandler: invalid limit value: %s", limitStr)
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
	logger.Warn("ListSubscriptionsHandler: invalid offset value: %s", offsetStr)
        writeJSONError(w, http.StatusBadRequest, "Invalid offset")
        return
    }
    if parsed < 0 {
		logger.Warn("ListSubscriptionsHandler: negative offset: %d", parsed)
        writeJSONError(w, http.StatusBadRequest, "Negative offset")
        return
    }
    offset = parsed
	}
	// 2. Вызвать database.ListSubscriptions
	list, err := h.Repo.ListMtd(ctx,limit, offset)
	if err != nil {
		logger.Error("ListSubscriptionsHandler: database error (limit=%d, offset=%d): %v", limit, offset, err)
		writeJSONError(w, http.StatusInternalServerError, "Failed to get subscriptions")
		return
	}
	// 3. Ответ	
	logger.Debug("ListSubscriptionsHandler: successfully fetched %d subscriptions (limit=%d, offset=%d)",
		len(list), limit, offset)
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
func (h *Handler)GetTotalCostHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// 1. Получаем параметры из URL (не парсим JSON)
	userID := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	logger.Debug("GetTotalCostHandler: request params - user_id=%s, service_name=%s, start_date=%s, end_date=%s",
		userID, serviceName, startDate, endDate)

	//  2. Вызвать database.GetTotalCost
total, err := h.Repo.GetTotalCostMtd(ctx,userID, serviceName, startDate, endDate)
if err != nil {
    if err.Error() == "start_date > end_date" {
	logger.Warn("GetTotalCostHandler: invalid date range - start_date=%s, end_date=%s", startDate, endDate)
        writeJSONError(w, http.StatusBadRequest, "start_date > end_date")
    } else {
		logger.Error("GetTotalCostHandler: database error: %v", err)
        writeJSONError(w, http.StatusInternalServerError, "Database error")
    }
    return
}
	// 3. Ответ
logger.Debug("GetTotalCostHandler: successfully calculated total=%d", total)
	writeJSON(w, http.StatusOK, map[string]int{"total": total})
}
