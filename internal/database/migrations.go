// Package database - инициализация базы данных и миграции
package database

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"Rest-user-agregator/pkg/logger"
)

// ============================================================
// RunMigrations — применяет все миграции (up-файлы)
// ============================================================
// Логика работы:
//  1. Находит все .up.sql файлы в папке migrations/
//  2. Сортирует их по имени (чтобы 000001 шёл до 000002)
//  3. Выполняет каждый SQL-файл по порядку
//  4. Если один файл не выполнился — останавливается и возвращает ошибку
//
// Почему так:
//   - Автоматически подхватывает новые миграции
//   - Не нужно вручную обновлять список файлов
//   - Порядок выполнения гарантирован
//
// ============================================================
func RunMigrations() error {
	// 1. Находим все .up.sql файлы в папке migrations/
	files, err := filepath.Glob("migrations/*.up.sql")
	if err != nil {
		logger.Error("RunMigrations: failed to list migration files: %v", err)
		return err
	}

	// 2. Если файлов нет — выходим (не ошибка, просто нет миграций)
	if len(files) == 0 {
		logger.Warn("RunMigrations: no migration files found")
		return nil
	}

	// 3. Сортируем файлы по имени (по алфавиту)
	//    Это гарантирует, что 000001 выполнится раньше 000002
	//    Без сортировки порядок может быть случайным
	sort.Strings(files)

	// 4. Логируем, сколько файлов найдено
	logger.Info("RunMigrations: found %d migration files", len(files))

	// 5. Выполняем каждый файл по очереди
	for _, file := range files {
		// 5.1 Читаем SQL-код из файла
		sql, err := os.ReadFile(file)
		if err != nil {
			logger.Error("RunMigrations: failed to read file %s: %v", file, err)
			return err
		}

		// 5.2 Выполняем SQL
		if _, err := db.Exec(string(sql)); err != nil {
			// Если таблица уже существует — пропускаем (это не ошибка)
			if strings.Contains(err.Error(), "already exists") {
				logger.Warn("RunMigrations: table already exists, skipping %s", file)
				continue
			}
			logger.Error("RunMigrations: failed to execute migration %s: %v", file, err)
			return err
		}

		// 5.3 Логируем успешное выполнение
		logger.Info("RunMigrations: applied %s", file)
	}

	logger.Info("RunMigrations: all migrations applied successfully")
	return nil
}

// ============================================================
// RollbackMigrations — откатывает все миграции (down-файлы)
// ============================================================
// Логика работы:
//  1. Находит все .down.sql файлы в папке migrations/
//  2. Сортирует их в обратном порядке (сначала последний, потом первый)
//  3. Выполняет каждый SQL-файл по порядку
//  4. Если один файл не выполнился — останавливается и возвращает ошибку
//
// Почему обратный порядок:
//   - Сначала откатываем последнюю миграцию (000002), потом первую (000001)
//   - Так же, как и в других инструментах (golang-migrate, goose)
//
// ============================================================
func RollbackMigrations() error {
	// 1. Находим все .down.sql файлы
	files, err := filepath.Glob("migrations/*.down.sql")
	if err != nil {
		logger.Error("RollbackMigrations: failed to list migration files: %v", err)
		return err
	}

	// 2. Если файлов нет — выходим
	if len(files) == 0 {
		logger.Warn("RollbackMigrations: no down migration files found")
		return nil
	}

	// 3. Сортируем файлы по имени (по алфавиту)
	sort.Strings(files)

	// 4. Переворачиваем порядок (обратный)
	//    Сначала выполняем последний файл (000002), потом первый (000001)
	//    Это гарантирует, что откат происходит в правильном порядке
	sort.Sort(sort.Reverse(sort.StringSlice(files)))

	// 5. Выполняем каждый файл
	for _, file := range files {
		sql, err := os.ReadFile(file)
		if err != nil {
			logger.Error("RollbackMigrations: failed to read file %s: %v", file, err)
			return err
		}

		if _, err := db.Exec(string(sql)); err != nil {
			logger.Error("RollbackMigrations: failed to rollback %s: %v", file, err)
			return err
		}

		logger.Info("RollbackMigrations: rolled back %s", file)
	}

	logger.Info("RollbackMigrations: all migrations rolled back successfully")
	return nil
}
