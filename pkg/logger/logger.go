// Package logger - логирование с уровнями для production
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

// Level - уровень логирования
// ============================================================================
// УРОВНИ ЛОГИРОВАНИЯ: КОГДА ЧТО ИСПОЛЬЗОВАТЬ
// ============================================================================
// УРОВЕНЬ | ЧТО ЛОГИРУЕМ                                | ПОЧЕМУ
// ----------------------------------------------------------------------------
// DEBUG   | Подписка создана / обновлена / удалена      | Только для отладки
// DEBUG   | Детали запроса к total-cost                 | Только для отладки
// ----------------------------------------------------------------------------
// INFO    | Сервер запущен / остановлен                 | Нормальное событие
// INFO    | БД подключена / миграции применены          | Нормальное событие
// INFO    | HTTP запрос (метод, путь, статус, время)    | Анализ трафика
// ----------------------------------------------------------------------------
// WARN    | Клиент прислал невалидный JSON              | Ошибка клиента
// WARN    | Клиент прислал неверный ID                  | Ошибка клиента
// WARN    | Подписка не найдена по ID                   | Бизнес-логика
// WARN    | Отрицательный offset в пагинации            | Клиентская ошибка
// WARN    | .env файл отсутствует                       | Есть дефолты
// WARN    | Не удалось применить миграции               | Таблицы могут существовать
// WARN    | Не удалось открыть лог-файл                 | Пишем в stdout
// ----------------------------------------------------------------------------
// ERROR   | Ошибка выполнения SQL запроса               | Сбой на стороне БД
// ERROR   | Разрыв соединения с БД                      | Сервер в опасности
// ERROR   | Не удалось создать подписку в БД            | Сбой записи
// ERROR   | Не удалось обновить/удалить подписку        | Сбой операции
// ERROR   | Парсинг даты внутри БД слоя                 | Внутренняя ошибка
// ERROR   | Не удалось закодировать JSON ответ          | Внутренняя ошибка
// ----------------------------------------------------------------------------
// FATAL   | Не удалось подключиться к БД при старте     | Без БД сервер не работает
// FATAL   | Не удалось прочитать файл миграции          | Без миграции таблицы не создать
// ============================================================================
type Level int

const (
	DEBUG Level = iota // 0 - для отладки
	INFO               // 1 - нормальные события
	WARN               // 2 - проблемы, не требующие остановки
	ERROR              // 3 - сбои, требующие внимания
	FATAL              // 4 - критические ошибки, сервер падает
)

// String возвращает имя уровня для красивого вывода--"метод, реализующий стандартный интерфейс fmt.Stringer".
func (l Level) String() string {
	names := []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
	if int(l) >= 0 && int(l) < len(names) {
		return names[l]
	}
	return "UNKNOWN"
}

var currentLevel Level = INFO

// Init инициализирует логгер с указанным уровнем
func Init(logFile string, level string) {
	// Устанавливаем уровень
	switch level {
	case "debug":
		currentLevel = DEBUG
	case "info":
		currentLevel = INFO
	case "warn":
		currentLevel = WARN
	case "error":
		currentLevel = ERROR
	case "fatal":
		currentLevel = FATAL
	default:
		currentLevel = INFO
	}

	// Открываем файл для записи
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
  Warn("Cannot open log file, using stdout only: %v", err) 
		log.SetOutput(os.Stdout)
		return
	}

	// Пишем и в файл, и в консоль
	multi := io.MultiWriter(os.Stdout, f)
	log.SetOutput(multi)
}

// internalLog - внутренняя функция для записи логов
func internalLog(level Level, msg string, args ...any) {
	if level < currentLevel {
		return
	}
	formattedMsg := fmt.Sprintf(msg, args...)
	log.Printf("[%s] %s", level.String(), formattedMsg)
}

// Debug - для отладки (только на dev окружении)
func Debug(msg string, args ...any) {
	internalLog(DEBUG, msg, args...)
}

// Info - важные события (запуск, остановка, миграции)
func Info(msg string, args ...any) {
	internalLog(INFO, msg, args...)
}

// Warn - проблемы, которые не остановили работу
func Warn(msg string, args ...any) {
	internalLog(WARN, msg, args...)
}

// Error - критические ошибки, но приложение продолжает работу
func Error(msg string, args ...any) {
	internalLog(ERROR, msg, args...)
}

// Fatal - критическая ошибка, после которой приложение завершается
func Fatal(msg string, args ...any) {
	internalLog(FATAL, msg, args...)
	os.Exit(1)
}