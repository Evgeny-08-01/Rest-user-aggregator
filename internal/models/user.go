// Package models структура данных для работы с пользователями
package models

type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Password  string `json:"password"` // читаем из JSON-времянка
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}
