// ============================================================
// ПАКЕТ: service
// ФАЙЛ: template_service.go
// НАЗНАЧЕНИЕ: бизнес-логика для работы с шаблонами подписок
// ============================================================
// Что здесь происходит:
//   1. CreateTemplate — создание шаблона (админ)
//   2. ListTemplates — список всех шаблонов (для всех пользователей)
//   3. GetTemplateByID — получить шаблон по ID
//   4. UpdateTemplate — обновить шаблон (админ)
//   5. DeleteTemplate — удалить шаблон (админ)
//
// ВАЖНО:
//   - Название уникально (регистр и пробелы игнорируются через БД)
//   - При обновлении названия проверяем, что новое название не занято
//   - При удалении проверяем, что нет подписок с этим шаблоном
// ============================================================

package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"Rest-user-agregator/internal/models"
	"Rest-user-agregator/internal/repository"
	"Rest-user-agregator/pkg/logger"
)

// TemplateService — сервис для работы с шаблонами подписок
type TemplateService struct {
	repo repository.TemplateRepository
}

// NewTemplateService — конструктор сервиса шаблонов
func NewTemplateService(repo repository.TemplateRepository) *TemplateService {
	return &TemplateService{repo: repo}
}

// CreateTemplate — создаёт новый шаблон с полной валидацией
// Параметры:
//   - serviceName: название сервиса (обязательно, не пустое)
//   - price: цена в рублях (>= 0)
// Возвращает:
//   - int: ID созданного шаблона
//   - error: ErrTemplateAlreadyExists если шаблон с таким названием уже существует
//   - error: ошибка валидации или БД
func (s *TemplateService) CreateTemplate(ctx context.Context, serviceName string, price int) (int, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    // ============================================================
    // 1. ВАЛИДАЦИЯ ВХОДНЫХ ДАННЫХ
    // ============================================================
    if serviceName == "" {
        return 0, fmt.Errorf("service_name is required")
    }
    if price < 0 {
        return 0, fmt.Errorf("price cannot be negative")
    }

    // ============================================================
    // 2. БИЗНЕС-ЛОГИКА
    // ============================================================

    // 2.1. Проверяем, существует ли шаблон с таким названием
    existing, err := s.repo.GetTemplateByName(ctx, serviceName)
    if err != nil {
        logger.Error("CreateTemplate: failed to check existing template: %v", err)
        return 0, err
    }
    if existing != nil {
        logger.Warn("CreateTemplate: template already exists: %s", serviceName)
        return 0, ErrTemplateAlreadyExists
    }

    // 2.2. Создаём шаблон
    id, err := s.repo.CreateTemplate(ctx, serviceName, price)
    if err != nil {
        logger.Error("CreateTemplate: failed to create template: %v", err)
        return 0, err
    }

    logger.Info("CreateTemplate: template created: id=%d, name=%s, price=%d", id, serviceName, price)
    return id, nil
}
// ============================================================
// 2. ПОЛУЧЕНИЕ СПИСКА ВСЕХ ШАБЛОНОВ (ListTemplates)
// ============================================================
// Что делает:
//   - Возвращает все шаблоны из БД
//   - Используется всеми пользователями для выбора шаблона
//
// ПАРАМЕТРЫ:
//   - ctx: контекст для отмены операций
//
// ВОЗВРАЩАЕТ:
//   - []models.Template: список шаблонов (может быть пустым)
//   - error: ошибка, если запрос не удался
// ============================================================
func (s *TemplateService) ListTemplates(ctx context.Context) ([]models.Template, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	templates, err := s.repo.ListTemplates(ctx)
	if err != nil {
		logger.Error("ListTemplates: failed to get templates: %v", err)
		return nil, err
	}

	logger.Debug("ListTemplates: returned %d templates", len(templates))
	return templates, nil
}

// ============================================================
// 3. ПОЛУЧЕНИЕ ШАБЛОНА ПО ID (GetTemplateByID)
// ============================================================
// Что делает:
//   - Возвращает шаблон по ID
//   - Возвращает nil, если шаблон не найден
//
// ПАРАМЕТРЫ:
//   - ctx: контекст для отмены операций
//   - id: ID шаблона
//
// ВОЗВРАЩАЕТ:
//   - *models.Template: шаблон или nil
//   - error: ошибка, если запрос не удался
// ============================================================
func (s *TemplateService) GetTemplateByID(ctx context.Context, id int) (*models.Template, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	template, err := s.repo.GetTemplateByID(ctx, id)
	if err != nil {
		logger.Error("GetTemplateByID: failed for id %d: %v", id, err)
		return nil, err
	}

	if template == nil {
		logger.Debug("GetTemplateByID: template not found for id %d", id)
		return nil, nil
	}

	logger.Debug("GetTemplateByID: found template id=%d, name=%s", id, template.ServiceName)
	return template, nil
}

// ============================================================
// 4. ОБНОВЛЕНИЕ ШАБЛОНА (UpdateTemplate)
// ============================================================
// Что делает:
//   - Проверяет, что шаблон существует
//   - Проверяет, что новое название не занято другим шаблоном
//   - Обновляет название и цену
//   - Используется админом
//
// ПАРАМЕТРЫ:
//   - ctx: контекст для отмены операций
//   - id: ID шаблона
//   - serviceName: новое название
//   - price: новая цена
//
// ВОЗВРАЩАЕТ:
//   - error: ошибка, если шаблон не найден или название занято
// ============================================================
func (s *TemplateService) UpdateTemplate(ctx context.Context, id int, serviceName string, price int) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Валидация: цена не может быть отрицательной
	if price < 0 {
		logger.Warn("UpdateTemplate: invalid price %d", price)
		return fmt.Errorf("price cannot be negative")
	}

	// Проверяем, существует ли шаблон
	existing, err := s.repo.GetTemplateByID(ctx, id)
	if err != nil {
		logger.Error("UpdateTemplate: failed to get template: %v", err)
		return err
	}
	if existing == nil {
		logger.Warn("UpdateTemplate: template not found for id %d", id)
		return sql.ErrNoRows
	}

	// Проверяем, что новое название не занято другим шаблоном
	if serviceName != existing.ServiceName {
		conflict, err := s.repo.GetTemplateByName(ctx, serviceName)
		if err != nil {
			logger.Error("UpdateTemplate: failed to check name conflict: %v", err)
			return err
		}
		if conflict != nil && conflict.ID != id {
			logger.Warn("UpdateTemplate: name '%s' already used by template %d", serviceName, conflict.ID)
			return fmt.Errorf("template with name '%s' already exists", serviceName)
		}
	}

	// Обновляем шаблон
	err = s.repo.UpdateTemplate(ctx, id, serviceName, price)
	if err != nil {
		logger.Error("UpdateTemplate: failed to update template %d: %v", id, err)
		return err
	}

	logger.Info("UpdateTemplate: template updated: id=%d, name=%s, price=%d", id, serviceName, price)
	return nil
}

// ============================================================
// 5. УДАЛЕНИЕ ШАБЛОНА (DeleteTemplate)
// ============================================================
// Что делает:
//   - Проверяет, что шаблон существует
//   - Проверяет, что нет подписок с этим шаблоном
//   - Удаляет шаблон
//   - Используется админом
//
// ПАРАМЕТРЫ:
//   - ctx: контекст для отмены операций
//   - id: ID шаблона
//
// ВОЗВРАЩАЕТ:
//   - error: ошибка, если шаблон не найден или есть подписки
// ============================================================
func (s *TemplateService) DeleteTemplate(ctx context.Context, id int) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Проверяем, существует ли шаблон
	existing, err := s.repo.GetTemplateByID(ctx, id)
	if err != nil {
		logger.Error("DeleteTemplate: failed to get template: %v", err)
		return err
	}
	if existing == nil {
		logger.Warn("DeleteTemplate: template not found for id %d", id)
		return sql.ErrNoRows
	}

	// Удаляем шаблон (репозиторий сам проверит наличие подписок)
	err = s.repo.DeleteTemplate(ctx, id)
	if err != nil {
		logger.Error("DeleteTemplate: failed to delete template %d: %v", id, err)
		return err
	}

	logger.Info("DeleteTemplate: template deleted: id=%d, name=%s", id, existing.ServiceName)
	return nil
}
