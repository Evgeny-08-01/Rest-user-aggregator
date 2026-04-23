// Package models структура данных для работы с подписками
package models

type Subscription struct {
	ID          int    `json:"id"`
	ServiceName string `json:"serviceName"`
	Price       int    `json:"price"`
	UserID      string `json:"userID"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
}
