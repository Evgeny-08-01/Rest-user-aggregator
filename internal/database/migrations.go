package database

import (
    "os"
)

func RunMigrations() error {
    migrationSQL, err := os.ReadFile("migrations/000001_create_subscriptions_table.up.sql")
    if err != nil {
        return err
    }
    _, err = db.Exec(string(migrationSQL))
    return err
}

func RollbackMigrations() error {
    downSQL, err := os.ReadFile("migrations/000001_create_subscriptions_table.down.sql")
    if err != nil {
        return err
    }
    _, err = db.Exec(string(downSQL))
    return err
}