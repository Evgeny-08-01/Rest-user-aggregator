package logger

import (
	"bytes"
	"log"
	"os"
	"testing"
)

// captureLogOutput временно подменяет log.SetOutput для захвата вывода
func captureLogOutput(f func()) string {
	// Сохраняем старый вывод
	oldOutput := log.Writer()
	
	// Создаём буфер
	var buf bytes.Buffer
	log.SetOutput(&buf)
	
	// Выполняем функцию
	f()
	
	// Восстанавливаем старый вывод
	log.SetOutput(oldOutput)
	
	return buf.String()
}

func TestInit(t *testing.T) {
	t.Run("successful init with debug level", func(t *testing.T) {
		Init("test.log", "debug")
		if currentLevel != DEBUG {
			t.Errorf("expected DEBUG, got %v", currentLevel)
		}
		os.Remove("test.log")
	})

	t.Run("successful init with info level", func(t *testing.T) {
		Init("test.log", "info")
		if currentLevel != INFO {
			t.Errorf("expected INFO, got %v", currentLevel)
		}
		os.Remove("test.log")
	})

	t.Run("init with invalid log file", func(t *testing.T) {
		// Должно быть предупреждение, но не паника
		Init("/nonexistent/dir/test.log", "info")
	})
}

func TestDefaultLevel(t *testing.T) {
	Init("test.log", "")
	defer os.Remove("test.log")

	if currentLevel != INFO {
		t.Errorf("expected default INFO, got %v", currentLevel)
	}
}

func TestStringMethod(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
		{FATAL, "FATAL"},
		{Level(99), "UNKNOWN"},
		{Level(-1), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestLogOutput(t *testing.T) {
	// Тест проверяет, что логи пишутся в правильном формате
	// НЕ вызываем Init() - используем стандартный log.SetOutput
	
	output := captureLogOutput(func() {
		// Временно устанавливаем уровень вручную для теста
		currentLevel = DEBUG
		
		// Пишем логи напрямую через internalLog
		internalLog(DEBUG, "debug message")
		internalLog(INFO, "info message")
		internalLog(WARN, "warn message")
		internalLog(ERROR, "error message")
	})
	
	expectedStrings := []string{
		"[DEBUG] debug message",
		"[INFO] info message",
		"[WARN] warn message",
		"[ERROR] error message",
	}
	
	for _, expected := range expectedStrings {
		if !bytes.Contains([]byte(output), []byte(expected)) {
			t.Errorf("Expected to contain '%s'\nGot:\n%s", expected, output)
		}
	}
}

func TestLevelFilteringOutput(t *testing.T) {
	tests := []struct {
		name       string
		level      Level
		shouldSee  []string
		shouldNotSee []string
	}{
		{
			name:       "debug sees everything",
			level:      DEBUG,
			shouldSee:  []string{"[DEBUG]", "[INFO]", "[WARN]", "[ERROR]"},
			shouldNotSee: []string{},
		},
		{
			name:       "info sees info,warn,error",
			level:      INFO,
			shouldSee:  []string{"[INFO]", "[WARN]", "[ERROR]"},
			shouldNotSee: []string{"[DEBUG]"},
		},
		{
			name:       "warn sees warn,error",
			level:      WARN,
			shouldSee:  []string{"[WARN]", "[ERROR]"},
			shouldNotSee: []string{"[DEBUG]", "[INFO]"},
		},
		{
			name:       "error sees only error",
			level:      ERROR,
			shouldSee:  []string{"[ERROR]"},
			shouldNotSee: []string{"[DEBUG]", "[INFO]", "[WARN]"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureLogOutput(func() {
				// Устанавливаем уровень фильтрации
				currentLevel = tt.level
				
				internalLog(DEBUG, "debug msg")
				internalLog(INFO, "info msg")
				internalLog(WARN, "warn msg")
				internalLog(ERROR, "error msg")
			})
			
			for _, s := range tt.shouldSee {
				if !bytes.Contains([]byte(output), []byte(s)) {
					t.Errorf("Expected to see '%s', but got:\n%s", s, output)
				}
			}
			for _, s := range tt.shouldNotSee {
				if bytes.Contains([]byte(output), []byte(s)) {
					t.Errorf("Expected NOT to see '%s', but got:\n%s", s, output)
				}
			}
		})
	}
}

func TestWarnOutput(t *testing.T) {
	output := captureLogOutput(func() {
		currentLevel = DEBUG
		Warn("test warning: %d", 42)
	})
	
	if !bytes.Contains([]byte(output), []byte("[WARN] test warning: 42")) {
		t.Errorf("Expected '[WARN] test warning: 42'\nGot:\n%s", output)
	}
}

func TestErrorOutput(t *testing.T) {
	output := captureLogOutput(func() {
		currentLevel = DEBUG
		Error("test error: %s", "something")
	})
	
	if !bytes.Contains([]byte(output), []byte("[ERROR] test error: something")) {
		t.Errorf("Expected '[ERROR] test error: something'\nGot:\n%s", output)
	}
}

func TestDebugOutput(t *testing.T) {
	output := captureLogOutput(func() {
		currentLevel = DEBUG
		Debug("debug value: %v", true)
	})
	
	if !bytes.Contains([]byte(output), []byte("[DEBUG] debug value: true")) {
		t.Errorf("Expected '[DEBUG] debug value: true'\nGot:\n%s", output)
	}
}

func TestInfoOutput(t *testing.T) {
	output := captureLogOutput(func() {
		currentLevel = DEBUG
		Info("info: %s", "started")
	})
	
	if !bytes.Contains([]byte(output), []byte("[INFO] info: started")) {
		t.Errorf("Expected '[INFO] info: started'\nGot:\n%s", output)
	}
}

func TestMultipleArguments(t *testing.T) {
	output := captureLogOutput(func() {
		currentLevel = DEBUG
		Debug("values: %s, %d, %v", "string", 42, true)
	})
	
	if !bytes.Contains([]byte(output), []byte("values: string, 42, true")) {
		t.Errorf("Expected 'values: string, 42, true'\nGot:\n%s", output)
	}
}