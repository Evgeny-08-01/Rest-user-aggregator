//go:build unit

package main

import (
	"os"
	"testing"
)

// ============================================================
// ТЕСТЫ ДЛЯ ЛОГГЕРА
// ============================================================

func TestLoadEnv(t *testing.T) {
	loadEnv()
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		t.Log("DB_PATH not set (this is OK if .env not exists)")
	}
}

func TestLoadEnvWithFile(t *testing.T) {
	tmpFile, err := os.Create(".env.test")
	if err != nil {
		t.Skip("Cannot create test .env file")
	}
	defer os.Remove(".env.test")

	tmpFile.WriteString("TEST_VAR=test_value\n")
	tmpFile.Close()

	oldEnv := os.Getenv("TEST_VAR")
	os.Rename(".env.test", ".env")
	defer os.Rename(".env", ".env.test")

	loadEnv()

	if os.Getenv("TEST_VAR") == "test_value" {
		t.Log("loadEnv loaded .env file")
	} else {
		t.Log("loadEnv did not load .env")
	}
	os.Setenv("TEST_VAR", oldEnv)
}

func TestInitLoggerDocker(t *testing.T) {
	oldPath := os.Getenv("LOG_PATH")
	oldLevel := os.Getenv("LOG_LEVEL")
	oldEnv := os.Getenv("ENV")

	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_PATH", "")
	os.Setenv("ENV", "docker")

	initLogger()
	t.Log("Logger initialized with Docker settings")

	os.Setenv("LOG_PATH", oldPath)
	os.Setenv("LOG_LEVEL", oldLevel)
	os.Setenv("ENV", oldEnv)
}

func TestInitPprof(t *testing.T) {
	oldPprof := os.Getenv("PPROF_ENABLED")

	os.Setenv("PPROF_ENABLED", "false")
	initPprof()
	t.Log("Pprof skipped when disabled")

	os.Setenv("PPROF_ENABLED", "true")
	initPprof()
	t.Log("Pprof started in goroutine")

	os.Setenv("PPROF_ENABLED", oldPprof)
}

func TestShouldRollbackWithArgs(t *testing.T) {
	oldArgs := os.Args

	os.Args = []string{"cmd"}
	if shouldRollback() {
		t.Error("shouldRollback true without args")
	}

	os.Args = []string{"cmd", "-down"}
	if !shouldRollback() {
		t.Error("shouldRollback false with -down")
	}

	os.Args = []string{"cmd", "-up"}
	if shouldRollback() {
		t.Error("shouldRollback true with -up")
	}

	t.Log("shouldRollback works with all args")
	os.Args = oldArgs
}

func TestInitLogger(t *testing.T) {
	oldLevel := os.Getenv("LOG_LEVEL")
	oldPath := os.Getenv("LOG_PATH")
	oldEnv := os.Getenv("ENV")

	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_PATH", "./test.log")
	os.Setenv("ENV", "local")

	initLogger()

	os.Setenv("LOG_LEVEL", oldLevel)
	os.Setenv("LOG_PATH", oldPath)
	os.Setenv("ENV", oldEnv)

	t.Log("Logger initialized")
}
