// Package models структура данных для работы с подписками
package models

type Subscription struct {
    ID          int    `json:"id"`
    ServiceName string `json:"service_name"` // пока оставляем для совместимости с хендлерами
    Price       int    `json:"price"`
    UserID      string `json:"user_id"`
    TemplateID  int    `json:"template_id"` // 
    StartDate   string `json:"start_date"`
    EndDate     string `json:"end_date"`
}