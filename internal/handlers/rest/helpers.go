package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"Rest-user-agregator/internal/authentication"
	"Rest-user-agregator/internal/middleware"
	"Rest-user-agregator/internal/models"
	"Rest-user-agregator/internal/service"
	"Rest-user-agregator/pkg/logger"

	"github.com/google/uuid"
)

// parseJSON читает и декодирует JSON из тела запроса
func parseJSON(r *http.Request, dest any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dest)
}

// writeJSON отправляет JSON-ответ с указанным статусом
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Error("Failed to encode JSON response: %v", err)
	}
}

// writeJsonError отправляет JSON-ошибку с указанным статусом
func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// isValidDate проверяет формат даты MM-YYYY
func isValidDate(date string) bool {
	parts := strings.Split(date, "-")
	if len(parts) != 2 {
		return false
	}
	month, err := strconv.Atoi(parts[0])
	if err != nil {
		return false
	}
	year, err := strconv.Atoi(parts[1])
	if err != nil {
		return false
	}

	if month < 1 || month > 12 {
		return false
	}
	if year < 1900 || year > 2100 {
		return false
	}
	return true
}

// validateSubscription проверяет поля подписки и возвращает ошибку
func validateSubscription(req models.Subscription) error {
	if req.ServiceName == "" {
		return fmt.Errorf("service_name is required")
	}
	if req.Price < 0 {
		return fmt.Errorf("price cant be negative value")
	}
	if req.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	_, err := uuid.Parse(req.UserID)
	if err != nil {
		return fmt.Errorf("user_id: not valid-UUID")
	}
	if !isValidDate(req.StartDate) {
		return fmt.Errorf("start_date must be in format MM-YYYY")
	}
	if req.EndDate != "" && !isValidDate(req.EndDate) {
		return fmt.Errorf("end_date must be in format MM-YYYY")
	}
	return nil
}

// loggingResponseWriter обёртка для перехвата HTTP статуса
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader перехватывает статус код
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next(lrw, r)
		log.Printf("%s %s %d %v", r.Method, r.URL.Path, lrw.statusCode, time.Since(start))
	}
}

// HealthHandler — хендлер для проверки работоспособности сервера (healthcheck)
// Используется Docker для проверки, что сервер жив и отвечает на запросы- не требует подключения к БД
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
// ConfigHandler — возвращает конфигурацию для фронтенда
func (h *Handler) ConfigHandler(w http.ResponseWriter, r *http.Request) {
    port := os.Getenv("SERVER_PORT")
    if port == "" {
        port = "8087"
    }
    config := map[string]string{
        "apiBase": "http://localhost:" + port + "/api",
    }
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(config); err != nil {
        logger.Error("Failed to encode config: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
}
// ============================================================
// wrapHandler — обёртка для маршрутов с middleware
// ============================================================
// Применяет стандартные middleware в правильном порядке:
//   1. AuthMiddleware (авторизация)
//   2. CorsMiddleware (CORS)
//   3. MetricsMiddleware (метрики)
//   4. LoggingMiddleware (логирование)
// ============================================================
func WrapHandler(handler http.HandlerFunc) http.HandlerFunc {
    // Если требуется — добавляем AuthMiddleware
    return authentication.AuthMiddleware(
        middleware.CorsMiddleware(
            middleware.MetricsMiddleware(
               LoggingMiddleware(handler))))
}

// mapErrorToStatus — маппер ошибок в HTTP-статусы
func mapErrorToStatus(err error) (int, string) {
    if err == nil {
        return http.StatusOK, ""
    }

    switch {
    // ============================================================
    // 400 Bad Request — ошибки валидации
    // ============================================================
   case errors.Is(err, service.ErrInvalidID),
        errors.Is(err, service.ErrUserIDRequired),
        errors.Is(err, service.ErrTemplateIDRequired),
        errors.Is(err, service.ErrStartDateRequired),
        errors.Is(err, service.ErrEndDateRequired),
        errors.Is(err, service.ErrInvalidDateRange),
        errors.Is(err, service.ErrCannotChangeStartDate),
        errors.Is(err, service.ErrPriceNegative),
        errors.Is(err, service.ErrServiceNameRequired),
		errors.Is(err, service.ErrTemplateAlreadyExists),
        errors.Is(err, service.ErrTemplateHasSubscriptions),
		errors.Is(err, service.ErrInvalidDateFormat):  
        return http.StatusBadRequest, err.Error()

    // ============================================================
    // 403 Forbidden — нет прав
    // ============================================================
    case errors.Is(err, service.ErrPermissionDenied):
        return http.StatusForbidden, err.Error()

    // ============================================================
    // 404 Not Found — ресурс не найден
    // ============================================================
    case errors.Is(err, service.ErrTemplateNotFound),
        errors.Is(err, sql.ErrNoRows):
        return http.StatusNotFound, err.Error()

    // ============================================================
    // 409 Conflict — конфликт
    // ============================================================
    case errors.Is(err, service.ErrTemplateAlreadyExists),
        errors.Is(err, service.ErrTemplateHasSubscriptions):
        return http.StatusConflict, err.Error()

    // ============================================================
    // 500 Internal Server Error — всё остальное
    // ============================================================
    default:
        logger.Error("Unmapped error: %v", err)
        return http.StatusInternalServerError, "Internal server error"
    }
}

