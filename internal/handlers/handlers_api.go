// Package handlers - пакет для обработки запросов, содержит 6 хэндлеров
package handlers

import (
	"net/http"
	"strconv"

	"github.com/Evgeny-08-01/Rest-user-aggregator/internal/database"
	"github.com/Evgeny-08-01/Rest-user-aggregator/internal/models"
)

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
	if err = validateSubscription(req); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 3. Вызвать функцию database.CreateSubscription
	id, err := database.CreateSubscription(req)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Database error")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int{"id": id})
}

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
	writeJSON(w, http.StatusOK, sub)
}

// 3. UpdateSubscriptionHandler-Хэндлер обновления одной строки
func UpdateSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	var req models.Subscription
	// 1. Получить id
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
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	// 4.  Вызвать database.UpdateSubscription
	err = database.UpdateSubscription(req)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Database error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// 4. DeleteSubscriptionHandler-Хэндлер удаления строки по id
func DeleteSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получить id
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	// 3. Вызвать database.DeleteSubscription

	err = database.DeleteSubscription(id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Database error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// 5. ListSubscriptionsHandler-Хэндлер чтения всех строк по фильтру
func ListSubscriptionsHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Получить параметры limit и offset из URL(если их нет,то предустановка в программе)
	limit := 20
	offset := 0
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err == nil && parsed > 0 {
			limit = parsed
		}
	}
	offsetStr := r.URL.Query().Get("offset")
	if offsetStr != "" {
		parsed, err := strconv.Atoi(offsetStr)
		if err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	// 2. Вызвать database.ListSubscriptions
	list, err := database.ListSubscriptions(limit, offset)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to encode JSON response:")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// 6. GetTotalCostHandlerHandler-Хэндлер чтения всех строк по фильтру
func GetTotalCostHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем параметры из URL (не парсим JSON)
	userID := r.URL.Query().Get("user_id")
	serviceName := r.URL.Query().Get("service_name")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	// 3. Вызвать database.CreateSubscription

	total, err := database.GetTotalCost(userID, serviceName, startDate, endDate)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to encode JSON response:")
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"total": total})
}
