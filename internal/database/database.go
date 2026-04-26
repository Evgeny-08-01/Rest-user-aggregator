// Package database-инициализация базы данных
package database

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq" // ← ЭТОТ ИМПОРТ ОБЯЗАТЕЛЕН
)

// DB - глобальное соединение с базой данных- используется при работе с таблицей
var DB *sql.DB

// Init открывает соединение с БД и проверяет его работоспособность
func Init(databasePath string) error {
	var err error
	DB, err = sql.Open("postgres", databasePath)
	if err != nil {
		return err
	}
	//  Проверяем подключение
	err = DB.Ping()
	if err != nil {
		return err
	}
	log.Println("Database connected")
	return nil
}

// Close закрывает соединение с базой данных
func Close() error {
	if DB == nil {
		return nil
	}
	err := DB.Close()
	DB = nil
	return err
}