// Package handlers - пакет для обработки запросов, содержит 6 хэндлеров
package handlers

import (
	"net/http"
	"strconv"

	"Rest-user-agregator/internal/authentication"
	"Rest-user-agregator/internal/models"
	"Rest-user-agregator/internal/service"
	"Rest-user-agregator/pkg/logger"
)

// ============================================================
// Handler — структура, объединяющая все HTTP-обработчики
// ============================================================
// Содержит сервисы для работы с разными сущностями:
//   - SubscriptionService → работа с подписками (CRUD, total-cost)
//   - AuthService         → работа с пользователями (регистрация, логин, JWT)
//
// Все методы Handler используют эти сервисы для обработки запросов.
// ============================================================
type Handler struct {
	Service     *service.SubscriptionService // Сервис для подписок (уже был)
	AuthService *service.AuthService         // Сервис для авторизации (НОВЫЙ)
}

// ============================================================
// NewHandler — конструктор хендлера
// ============================================================
// Принимает оба сервиса и возвращает инициализированный Handler.
// ============================================================
func NewHandler(svc *service.SubscriptionService, authSvc *service.AuthService) *Handler {
	return &Handler{
		Service:     svc,     // Сервис подписок
		AuthService: authSvc, // Сервис авторизации
	}
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
// ============================================================
// CreateSubscriptionHandler — создание новой подписки
// ============================================================
// Аналог: POST /api/subscriptions
// ============================================================
// Логика работы:
//   1. Проверяет авторизацию (userID из JWT)
//   2. Парсит JSON из тела запроса (template_id, start_date, end_date)
//   3. Вызывает сервис (вся валидация внутри)
//   4. Возвращает ID созданной подписки или ошибку
// ============================================================
func (h *Handler) CreateSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
    // 1. Проверяем авторизацию (userID из JWT)
    //    Если user_id отсутствует или пустой — возвращаем 401 Unauthorized
    userID, ok := r.Context().Value(authentication.UserIDKey).(string)
    if !ok || userID == "" {
        logger.Warn("CreateSubscriptionHandler: user_id not found or invalid")
        writeJSONError(w, http.StatusUnauthorized, "unauthorized")
        return
    }

    ctx := r.Context()

    // 2. Парсим JSON из тела запроса
    //    Ожидаем: { "template_id": 1, "start_date": "08-2025", "end_date": "12-2025" }
    var req struct {
        TemplateID int    `json:"template_id"`
        StartDate  string `json:"start_date"`
        EndDate    string `json:"end_date"`
    }
    err := parseJSON(r, &req)
    if err != nil {
        logger.Warn("CreateSubscriptionHandler: failed to parse JSON: %v", err)
        writeJSONError(w, http.StatusBadRequest, "Invalid JSON")
        return
    }

    // 3. Вызываем сервис (ВСЯ ВАЛИДАЦИЯ ВНУТРИ)
    //    Сервис сам проверит:
    //    - template_id > 0 (ErrTemplateIDRequired)
    //    - user_id не пустой (ErrUserIDRequired)
    //    - start_date не пустой (ErrStartDateRequired)
    //    - Существует ли шаблон (ErrTemplateNotFound)
    //    - Корректны ли даты (ErrCannotChangeStartDate)
    id, err := h.Service.CreateSubscription(ctx, req.TemplateID, userID, req.StartDate, req.EndDate)
   if err != nil {
    status, msg := mapErrorToStatus(err)
    logger.Warn("Handler error: %v", err)  
    writeJSONError(w, status, msg)
    return
}

    // 5. Успешный ответ → 201 Created с ID созданной подписки
    logger.Debug("CreateSubscriptionHandler: successfully created subscription id=%d for user_id=%s, template_id=%d",
        id, userID, req.TemplateID)
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
func (h *Handler) GetSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // 1. Получаем ID из URL
    idStr := r.PathValue("id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        logger.Warn("GetSubscriptionHandler: invalid ID format: %s", idStr)
        writeJSONError(w, http.StatusBadRequest, "Invalid ID")
        return
    }

    // 2. Получаем данные пользователя из контекста (JWT)
    userID := authentication.GetUserID(ctx)
    role := authentication.GetRole(ctx)

    // 3. Вызываем сервис (проверка прав ВНУТРИ сервиса)
    sub, err := h.Service.GetSubscriptionByID(ctx, id, userID, role)
    if err != nil {
    status, msg := mapErrorToStatus(err)
    logger.Warn("Handler error: %v", err)  
    writeJSONError(w, status, msg)
    return
}
    if sub == nil {
        logger.Warn("GetSubscriptionHandler: subscription not found for id=%d", id)
        writeJSONError(w, http.StatusNotFound, "Subscription not found")
        return
    }

    // 6. Успешный ответ
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
// ============================================================
// UpdateSubscriptionHandler — обновление подписки (REST)
// ============================================================
// Аналог: PUT /api/subscriptions/{id}
// ============================================================
// Логика работы:
//   1. Проверяет авторизацию (userID из JWT)
//   2. Парсит ID из URL
//   3. Парсит JSON из тела запроса
//   4. Вызывает сервис (вся валидация и проверка прав внутри)
//   5. Возвращает статус обновления или ошибку
// ============================================================
func (h *Handler) UpdateSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // 1. Парсим ID из URL
    idStr := r.PathValue("id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        logger.Warn("UpdateSubscriptionHandler: invalid ID format: %s", idStr)
        writeJSONError(w, http.StatusBadRequest, "Invalid ID")
        return
    }

    // 2. Парсим JSON из тела запроса
    var req models.Subscription
    req.ID = id
    if err := parseJSON(r, &req); err != nil {
        logger.Warn("UpdateSubscriptionHandler: failed to parse JSON for id=%d: %v", id, err)
        writeJSONError(w, http.StatusBadRequest, "Invalid JSON")
        return
    }

    // 3. Получаем роль из контекста (для проверки прав в сервисе)
    role := authentication.GetRole(ctx)

    // 4. Вызов сервиса (ВСЯ ВАЛИДАЦИЯ И ПРОВЕРКА ПРАВ ВНУТРИ)
    //    Сервис сам проверит:
    //    - sub.ID > 0 (ErrInvalidID)
    //    - sub.UserID не пустой (ErrUserIDRequired)
    //    - sub.StartDate не пустой (ErrStartDateRequired)
    //    - Права доступа (ErrPermissionDenied)
    //    - Существует ли подписка (sql.ErrNoRows)
    //    - Корректны ли даты (ErrCannotChangeStartDate)
    err = h.Service.UpdateSubscription(ctx, req, role)
    if err != nil {
    status, msg := mapErrorToStatus(err)
    logger.Warn("Handler error: %v", err)  
    writeJSONError(w, status, msg)
    return
}

    // 6. Успешный ответ → 200 OK
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
// ============================================================
// DeleteSubscriptionHandler — удаление подписки (REST)
// ============================================================
// Аналог: DELETE /api/subscriptions/{id}
// ============================================================
// Логика работы:
//   1. Проверяет авторизацию (userID из JWT)
//   2. Парсит ID из URL
//   3. Вызывает сервис (вся валидация и проверка прав внутри)
//   4. Возвращает статус удаления или ошибку
// ============================================================
func (h *Handler) DeleteSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // 1. Парсим ID из URL
    idStr := r.PathValue("id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        logger.Warn("DeleteSubscriptionHandler: invalid ID format: %s", idStr)
        writeJSONError(w, http.StatusBadRequest, "Invalid ID")
        return
    }

    // 2. Получаем user_id и роль из контекста
    userID := authentication.GetUserID(ctx)
    role := authentication.GetRole(ctx)

    // 3. Вызов сервиса (ВСЯ ВАЛИДАЦИЯ И ПРОВЕРКА ПРАВ ВНУТРИ)
    //    Сервис сам проверит:
    //    - id > 0 (ErrInvalidID)
    //    - userID не пустой (ErrUserIDRequired)
    //    - Права доступа (ErrPermissionDenied)
    //    - Существует ли подписка (sql.ErrNoRows)
    err = h.Service.DeleteSubscription(ctx, id, userID, role)
    if err != nil {
    status, msg := mapErrorToStatus(err)
    logger.Warn("Handler error: %v", err)  
    writeJSONError(w, status, msg)
    return
}

    // 5. Успешный ответ → 200 OK
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
// ============================================================
// ListSubscriptionsHandler — получение списка подписок (REST)
// ============================================================
// Аналог: GET /api/subscriptions?limit=20&offset=0
// ============================================================
// Логика работы:
//   1. Проверяет авторизацию (userID и роль из JWT)
//   2. Парсит параметры пагинации (limit, offset)
//   3. Вызывает сервис (валидация внутри)
//   4. Возвращает список подписок или ошибку
// ============================================================
func (h *Handler) ListSubscriptionsHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // 1. Получаем user_id и роль из контекста
    userID := authentication.GetUserID(ctx)
    role := authentication.GetRole(ctx)

    // 2. Парсим параметры пагинации (с валидацией формата)
    limit := 20 // значение по умолчанию
    offset := 0 // значение по умолчанию

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

    // 3. Вызываем сервис (валидация прав внутри)
    list, err := h.Service.ListSubscriptions(ctx, userID, role, limit, offset)
    if err != nil {
    status, msg := mapErrorToStatus(err)
    logger.Warn("Handler error: %v", err)  
    writeJSONError(w, status, msg)
    return
}

    // 5. Если nil → пустой массив
    if list == nil {
        list = []models.Subscription{}
    }

    // 6. Успешный ответ → 200 OK
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
// ============================================================
// GetTotalCostHandler — расчёт суммарной стоимости подписок (REST)
// ============================================================
func (h *Handler) GetTotalCostHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    role := authentication.GetRole(ctx)
    var userID string
    if role == "admin" {
        userID = ""
    } else {
        userID = authentication.GetUserID(ctx)
    }

    serviceName := r.URL.Query().Get("service_name")
    startDate := r.URL.Query().Get("start_date")
    endDate := r.URL.Query().Get("end_date")

    logger.Debug("GetTotalCostHandler: request params - user_id=%s, service_name=%s, start_date=%s, end_date=%s",
        userID, serviceName, startDate, endDate)

    // Вызываем сервис (вся валидация внутри)
    total, err := h.Service.GetTotalCost(ctx, userID, serviceName, startDate, endDate)
    if err != nil {
    status, msg := mapErrorToStatus(err)
    logger.Warn("Handler error: %v", err)  
    writeJSONError(w, status, msg)
    return
}

    logger.Debug("GetTotalCostHandler: successfully calculated total=%d", total)
    writeJSON(w, http.StatusOK, map[string]int{"total": total})
}