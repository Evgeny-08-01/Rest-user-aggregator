// ============================================================
// ПАКЕТ: handlers/rest
// ФАЙЛ: template_handlers.go
// НАЗНАЧЕНИЕ: HTTP-обработчики для работы с шаблонами подписок
// ============================================================
// Что здесь происходит:
//   1. CreateTemplateHandler — создание шаблона (только админ)
//   2. ListTemplatesHandler — список всех шаблонов (все пользователи)
//   3. GetTemplateHandler — получить шаблон по ID (все пользователи)
//   4. UpdateTemplateHandler — обновить шаблон (только админ)
//   5. DeleteTemplateHandler — удалить шаблон (только админ)
//
// ВАЖНО:
//   - Все хендлеры для шаблонов защищены middleware AuthMiddleware
//   - Админ проверяется через role из контекста
// ============================================================

package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"Rest-user-agregator/internal/authentication"
	"Rest-user-agregator/internal/models"
	"Rest-user-agregator/internal/service"
	"Rest-user-agregator/pkg/logger"
)

// TemplateHandler — структура для обработки запросов к шаблонам
type TemplateHandler struct {
	Service *service.TemplateService
}

// NewTemplateHandler — конструктор обработчика шаблонов
func NewTemplateHandler(svc *service.TemplateService) *TemplateHandler {
	return &TemplateHandler{Service: svc}
}

// ============================================================
// CreateTemplateHandler — создание шаблона (REST)
// ============================================================
// Аналог: POST /api/admin/templates
// ============================================================
// Логика работы:
//  1. Проверяет роль (только admin)
//  2. Парсит JSON (service_name, price)
//  3. Вызывает сервис (вся валидация внутри)
//  4. Возвращает ID созданного шаблона или ошибку
//
// ============================================================
func (h *TemplateHandler) CreateTemplateHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Проверка роли (только админ)
	role := r.Context().Value(authentication.RoleKey)
	if role == nil || role != "admin" {
		logger.Warn("CreateTemplateHandler: non-admin attempt to create template")
		writeJSONError(w, http.StatusForbidden, "admin only")
		return
	}

	// 2. Парсим JSON
	var req struct {
		ServiceName string `json:"service_name"`
		Price       int    `json:"price"`
	}
	if err := parseJSON(r, &req); err != nil {
		logger.Warn("CreateTemplateHandler: invalid JSON: %v", err)
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// 3. Вызов сервиса (ВСЯ ВАЛИДАЦИЯ ВНУТРИ)
	//    Сервис сам проверит:
	//    - service_name не пустой
	//    - price >= 0
	//    - Нет шаблона с таким названием (ErrTemplateAlreadyExists)
	id, err := h.Service.CreateTemplate(r.Context(), req.ServiceName, req.Price)
	if err != nil {
		status, msg := mapErrorToStatus(err)
		logger.Warn("Handler error: %v", err)
		writeJSONError(w, status, msg)
		return
	}

	// 5. Успешный ответ → 201 Created
	logger.Info("CreateTemplateHandler: template created: id=%d, name=%s", id, req.ServiceName)
	writeJSON(w, http.StatusCreated, map[string]int{"id": id})
}

// ============================================================
// 2. СПИСОК ВСЕХ ШАБЛОНОВ (ListTemplatesHandler)
// ============================================================
// Что делает:
//   - Возвращает список всех шаблонов
//   - Доступно всем авторизованным пользователям
//
// ОТВЕТ:
//
//	200 OK: [ { "id": 1, "service_name": "...", "price": 400 } ]
//	500 Internal Server Error: { "error": "..." }
//
// ============================================================
func (h *TemplateHandler) ListTemplatesHandler(w http.ResponseWriter, r *http.Request) {
	templates, err := h.Service.ListTemplates(r.Context())
	if err != nil {
		status, msg := mapErrorToStatus(err)
		logger.Warn("Handler error: %v", err)
		writeJSONError(w, status, msg)
		return
	}

	// ✅ Если nil — возвращаем пустой массив
	if templates == nil {
		templates = []models.Template{}
	}
	logger.Debug("ListTemplatesHandler: returned %d templates", len(templates))
	writeJSON(w, http.StatusOK, templates)
}

// ============================================================
// 3. ПОЛУЧЕНИЕ ШАБЛОНА ПО ID (GetTemplateHandler)
// ============================================================
// Что делает:
//   - Возвращает шаблон по ID
//   - Доступно всем авторизованным пользователям
//
// ПАРАМЕТРЫ URL:
//
//	/api/templates/{id}
//
// ОТВЕТ:
//
//	200 OK: { "id": 1, "service_name": "...", "price": 400 }
//	404 Not Found: { "error": "template not found" }
//
// ============================================================
func (h *TemplateHandler) GetTemplateHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем ID из URL
	idStr := strings.TrimSpace(r.PathValue("id"))
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Warn("GetTemplateHandler: invalid id: %s", idStr)
		writeJSONError(w, http.StatusBadRequest, "Invalid template ID")
		return
	}

	// Получаем шаблон
	template, err := h.Service.GetTemplateByID(r.Context(), id)
	if err != nil {
		status, msg := mapErrorToStatus(err)
		logger.Warn("Handler error: %v", err)
		writeJSONError(w, status, msg)
		return
	}
	if template == nil {
		logger.Warn("GetTemplateHandler: template not found: id=%d", id)
		writeJSONError(w, http.StatusNotFound, "template not found")
		return
	}

	logger.Debug("GetTemplateHandler: returned template id=%d", id)
	writeJSON(w, http.StatusOK, template)
}

// ============================================================
// 4. ОБНОВЛЕНИЕ ШАБЛОНА (UpdateTemplateHandler)
// ============================================================
// Что делает:
//   - Обновляет название и цену шаблона
//   - Только для админа
//   - Проверяет, что новое название не занято
//
// ТЕЛО ЗАПРОСА:
//
//	{
//	  "service_name": "Новое название",
//	  "price": 500
//	}
//
// ОТВЕТ:
//
//	200 OK: { "status": "updated" }
//	403 Forbidden: { "error": "admin only" }
//	404 Not Found: { "error": "template not found" }
//
// ============================================================
func (h *TemplateHandler) UpdateTemplateHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем, что пользователь — админ
	role := r.Context().Value(authentication.RoleKey)
	if role == nil || role != "admin" {
		logger.Warn("UpdateTemplateHandler: non-admin attempt to update template")
		writeJSONError(w, http.StatusForbidden, "admin only")
		return
	}

	// Получаем ID из URL
	idStr := strings.TrimSpace(r.PathValue("id"))
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Warn("UpdateTemplateHandler: invalid id: %s", idStr)
		writeJSONError(w, http.StatusBadRequest, "Invalid template ID")
		return
	}

	// Парсим JSON
	var req struct {
		ServiceName string `json:"service_name"`
		Price       int    `json:"price"`
	}
	if err := parseJSON(r, &req); err != nil {
		logger.Warn("UpdateTemplateHandler: invalid JSON: %v", err)
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Валидация
	req.ServiceName = strings.TrimSpace(req.ServiceName)
	if req.ServiceName == "" {
		logger.Warn("UpdateTemplateHandler: empty service_name")
		writeJSONError(w, http.StatusBadRequest, "service_name is required")
		return
	}
	if req.Price < 0 {
		logger.Warn("UpdateTemplateHandler: negative price: %d", req.Price)
		writeJSONError(w, http.StatusBadRequest, "price cannot be negative")
		return
	}

	// Обновляем шаблон
	err = h.Service.UpdateTemplate(r.Context(), id, req.ServiceName, req.Price)
	if err != nil {
		status, msg := mapErrorToStatus(err)
		logger.Warn("Handler error: %v", err)
		writeJSONError(w, status, msg)
		return
	}

	logger.Info("UpdateTemplateHandler: template updated: id=%d, name=%s", id, req.ServiceName)
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// ============================================================
// 5. УДАЛЕНИЕ ШАБЛОНА (DeleteTemplateHandler)
// ============================================================
// Что делает:
//   - Удаляет шаблон по ID
//   - Только для админа
//   - Проверяет, что нет подписок с этим шаблоном
//
// ПАРАМЕТРЫ URL:
//
//	/api/admin/templates/{id}
//
// ОТВЕТ:
//
//	200 OK: { "status": "deleted" }
//	403 Forbidden: { "error": "admin only" }
//	404 Not Found: { "error": "template not found" }
//	409 Conflict: { "error": "cannot delete template: X subscriptions are using it" }
//
// ============================================================
func (h *TemplateHandler) DeleteTemplateHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем, что пользователь — админ
	role := r.Context().Value(authentication.RoleKey)
	if role == nil || role != "admin" {
		logger.Warn("DeleteTemplateHandler: non-admin attempt to delete template")
		writeJSONError(w, http.StatusForbidden, "admin only")
		return
	}

	// Получаем ID из URL
	idStr := strings.TrimSpace(r.PathValue("id"))
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Warn("DeleteTemplateHandler: invalid id: %s", idStr)
		writeJSONError(w, http.StatusBadRequest, "Invalid template ID")
		return
	}

	// Удаляем шаблон
	err = h.Service.DeleteTemplate(r.Context(), id)
	if err != nil {
		status, msg := mapErrorToStatus(err)
		logger.Warn("Handler error: %v", err) // ← добавить логгер
		writeJSONError(w, status, msg)
		return
	}

	logger.Info("DeleteTemplateHandler: template deleted: id=%d", id)
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
