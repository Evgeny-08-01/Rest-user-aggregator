//  Package database-инициализация базы данных
package database 

import (
	"database/sql"
	"log"
	_ "github.com/mattn/go-sqlite3"

)

var DB *sql.DB

func Init(dbPath string) error {
	var err error
	
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	err = DB.Ping()
	if err != nil {
		return err
	}
	log.Println("Database connected")
	return nil
}

func Close() error {
	if DB == nil {
		return nil
	}
	err := DB.Close()
	DB = nil
	return err
}