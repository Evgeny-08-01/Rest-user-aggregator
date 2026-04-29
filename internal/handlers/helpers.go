package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Evgeny-08-01/Rest-user-aggregator/internal/models"
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
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
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
	 _,err := uuid.Parse(req.UserID)
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
