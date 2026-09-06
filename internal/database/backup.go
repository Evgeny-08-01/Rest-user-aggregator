package database

import (
	"compress/gzip"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"Rest-user-agregator/pkg/logger"
)

const (
    containerName = "subscription-db"
    dbUser        = "postgres"
    dbName        = "subscriptions"
    backupDir     = "./backups"
)

// BackupDB создаёт сжатый бэкап базы данных.
// Возвращает путь к созданному .sql.gz или ошибку.
func BackupDB() (string, error) {
    if err := os.MkdirAll(backupDir, 0755); err != nil {
        return "", fmt.Errorf("failed to create backup dir: %w", err)
    }

    timestamp := time.Now().Format("20060102_150405.000")
    backupFile := filepath.Join(backupDir, timestamp+".sql.gz")
    rawFile := filepath.Join(backupDir, "backup_raw.sql")

    // 1. pg_dump через stdout (без -f)
    cmd := exec.Command(
        "docker", "exec",
        containerName,
        "pg_dump",
        "-U", dbUser,
        "-d", dbName,
    )

    // 2. Сохраняем stdout в файл
    out, err := os.Create(rawFile)
    if err != nil {
        return "", fmt.Errorf("failed to create temp file: %w", err)
    }
    defer out.Close()
    cmd.Stdout = out

    if err := cmd.Run(); err != nil {
        return "", fmt.Errorf("pg_dump failed: %w", err)
    }

    // 3. Сжимаем
    if err := compressFile(rawFile, backupFile); err != nil {
        return "", fmt.Errorf("compression failed: %w", err)
    }

    // 4. Удаляем временный файл
    _ = os.Remove(rawFile)

    // 5. Ротация
    if err := rotateBackups(backupDir); err != nil {
        logger.Warn("Rotation failed: %v", err)
    }

    return backupFile, nil
}

// compressFile сжимает файл в gzip
func compressFile(src, dst string) error {
    srcFile, err := os.Open(src)
    if err != nil {
        return err
    }
    defer srcFile.Close()

    dstFile, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer dstFile.Close()

    gz := gzip.NewWriter(dstFile)
    defer gz.Close()

    _, err = srcFile.WriteTo(gz)
    return err
}

// rotateBackups оставляет только последние BACKUP_KEEP бэкапов
func rotateBackups(dir string) error {
	const defaultKeep = 7

	keep := defaultKeep
	if v := os.Getenv("BACKUP_KEEP"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			keep = i
		}
	}

	pattern := filepath.Join(dir, "*.sql.gz")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	if len(files) <= keep {
		return nil
	}

	sort.Strings(files)
	toDelete := files[:len(files)-keep]

	for _, f := range toDelete {
		if err := os.Remove(f); err != nil {
			logger.Warn("Failed to remove %s: %v", f, err)
		} else {
			logger.Debug("Removed old backup: %s", f)
		}
	}

	return nil
}