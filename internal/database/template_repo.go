// ============================================================
// ПАКЕТ: database
// ФАЙЛ: template_repo.go
// НАЗНАЧЕНИЕ: работа с таблицей subscription_templates (шаблоны подписок)
// ============================================================
// Что здесь происходит:
//   1. CreateTemplate — создаёт новый шаблон (админ)
//   2. ListTemplates — список всех шаблонов (для всех пользователей)
//   3. GetTemplateByID — получить шаблон по ID (для проверки при создании подписки)
//   4. GetTemplateByName — проверить, существует ли шаблон с таким названием (при создании/обновлении)
//   5. UpdateTemplate — обновить шаблон (админ)
//   6. DeleteTemplate — удалить шаблон (админ)
//
// Шаблоны создаются админом и используются пользователями для создания подписок.
// Название уникально (регистр и пробелы игнорируются через индекс в БД).
//
// Связи:
//   - Шаблон → Подписки: один шаблон может быть использован во многих подписках
//   - При удалении шаблона: template_id в подписках становится NULL (ON DELETE SET NULL)
// ============================================================

package database

import (
	"context"
	"database/sql"
	"fmt"

	"Rest-user-agregator/internal/models"
	"Rest-user-agregator/pkg/logger"
)

// ============================================================
// 1. СОЗДАНИЕ НОВОГО ШАБЛОНА (CreateTemplate)
// ============================================================
// Что делает:
//   - Вставляет новый шаблон в таблицу subscription_templates
//   - Проверка уникальности названия происходит через UNIQUE INDEX в БД
//   - Возвращает ID созданного шаблона
//
// ПАРАМЕТРЫ:
//   - ctx: контекст для отмены операций
//   - serviceName: название сервиса (например, "Яндекс Плюс")
//   - price: цена в рублях (целое число, >= 0)
//
// ВОЗВРАЩАЕТ:
//   - int: ID созданного шаблона
//   - error: ошибка, если не удалось создать
//
// ВАЖНО:
//   - Если шаблон с таким названием уже существует — БД вернёт ошибку 23505 (duplicate key)
//   - Название нормализуется в БД через индекс (LOWER(TRIM(...))), в коде не нормализуем
//
// ============================================================
func (r *PostgresRepo) CreateTemplate(ctx context.Context, serviceName string, price int) (int, error) {
	// Переменная для хранения ID созданного шаблона
	var id int

	// SQL-запрос: вставляем шаблон и возвращаем его ID
	//   - $1 → serviceName (название сервиса)
	//   - $2 → price (цена)
	//   - RETURNING id → сразу получаем ID созданной записи
	query := `INSERT INTO subscription_templates (service_name, price) 
	          VALUES ($1, $2) 
	          RETURNING id`

	// Выполняем запрос и сканируем ID в переменную id
	err := r.db.QueryRowContext(ctx, query, serviceName, price).Scan(&id)
	if err != nil {
		// Логируем ошибку (например, duplicate key)
		logger.Error("CreateTemplate: failed to insert template %s: %v", serviceName, err)
		return 0, err
	}

	// Успешное создание — логируем ID и название
	logger.Debug("CreateTemplate: created template id=%d, name=%s", id, serviceName)
	return id, nil
}

// ============================================================
// 2. ПОЛУЧЕНИЕ СПИСКА ВСЕХ ШАБЛОНОВ (ListTemplates)
// ============================================================
// Что делает:
//   - Возвращает все шаблоны из таблицы, отсортированные по ID
//   - Используется пользователями для выбора шаблона при создании подписки
//   - Используется админом для просмотра всех шаблонов
//
// ПАРАМЕТРЫ:
//   - ctx: контекст для отмены операций
//
// ВОЗВРАЩАЕТ:
//   - []models.Template: список шаблонов (может быть пустым)
//   - error: ошибка, если запрос не удался
//
// ПРИМЕР ВЫЗОВА:
//
//	templates, err := repo.ListTemplates(ctx)
//	if err != nil { ... }
//	for _, t := range templates {
//	    fmt.Println(t.ServiceName, t.Price)
//	}
//
// ============================================================
func (r *PostgresRepo) ListTemplates(ctx context.Context) ([]models.Template, error) {
	// SQL-запрос: выбираем все поля, сортируем по ID (от старых к новым)
	query := `SELECT id, service_name, price FROM subscription_templates ORDER BY id`

	// Выполняем запрос — получаем строки результата
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		logger.Error("ListTemplates: query failed: %v", err)
		return nil, err
	}
	// Обязательно закрываем rows после завершения функции
	// Даже если возникнет ошибка — rows закроется (благодаря defer)
	defer rows.Close()

	// Создаём слайс для результатов (пустой, но готовый к добавлению)
	var templates []models.Template

	// Проходим по всем строкам результата
	for rows.Next() {
		var t models.Template
		// Сканируем поля из строки в структуру
		// Порядок сканирования должен совпадать с порядком полей в SELECT
		err := rows.Scan(&t.ID, &t.ServiceName, &t.Price)
		if err != nil {
			logger.Error("ListTemplates: scan failed: %v", err)
			return nil, err
		}
		// Добавляем шаблон в слайс
		templates = append(templates, t)
	}

	// Проверяем ошибки после завершения итерации
	// (например, обрыв соединения во время чтения)
	if err = rows.Err(); err != nil {
		logger.Error("ListTemplates: rows error: %v", err)
		return nil, err
	}

	// Логируем количество найденных шаблонов
	logger.Debug("ListTemplates: found %d templates", len(templates))
	return templates, nil
}

// ============================================================
// 3. ПОЛУЧЕНИЕ ШАБЛОНА ПО ID (GetTemplateByID)
// ============================================================
// Что делает:
//   - Ищет шаблон по его ID
//   - Возвращает nil, если шаблон не найден (без ошибки)
//   - Используется при:
//   - Создании подписки (проверить, что шаблон существует)
//   - Редактировании подписки админом (получить текущие данные шаблона)
//
// ПАРАМЕТРЫ:
//   - ctx: контекст для отмены операций
//   - id: ID шаблона (число)
//
// ВОЗВРАЩАЕТ:
//   - *models.Template: указатель на шаблон (или nil, если не найден)
//   - error: ошибка, если запрос не удался
//
// ВАЖНО:
//   - Возвращаем nil, а не пустую структуру, чтобы чётко обозначить "не найдено"
//   - Проверка: if template == nil { ... }
//
// ============================================================
func (r *PostgresRepo) GetTemplateByID(ctx context.Context, id int) (*models.Template, error) {
	// SQL-запрос: выбираем шаблон по ID
	query := `SELECT id, service_name, price FROM subscription_templates WHERE id = $1`

	// Создаём структуру для результата
	var t models.Template

	// Выполняем запрос и сканируем результат
	err := r.db.QueryRowContext(ctx, query, id).Scan(&t.ID, &t.ServiceName, &t.Price)
	if err != nil {
		// Если шаблон не найден — возвращаем nil (не ошибка)
		if err == sql.ErrNoRows {
			return nil, nil
		}
		// Другие ошибки логируем и возвращаем
		logger.Error("GetTemplateByID: failed for id %d: %v", id, err)
		return nil, err
	}

	// Шаблон найден — логируем и возвращаем
	logger.Debug("GetTemplateByID: found template id=%d, name=%s", id, t.ServiceName)
	return &t, nil
}

// ============================================================
// 4. ПРОВЕРКА СУЩЕСТВОВАНИЯ ШАБЛОНА ПО НАЗВАНИЮ (GetTemplateByName)
// ============================================================
// Что делает:
//   - Проверяет, существует ли шаблон с таким названием
//   - Используется при создании/обновлении шаблона, чтобы избежать дублирования
//   - Регистр и пробелы игнорируются (через LOWER(TRIM(...)) в запросе)
//
// ПАРАМЕТРЫ:
//   - ctx: контекст для отмены операций
//   - serviceName: название сервиса (например, "Яндекс Плюс")
//
// ВОЗВРАЩАЕТ:
//   - *models.Template: указатель на шаблон (если найден) или nil
//   - error: ошибка, если запрос не удался
//
// ПРИМЕР:
//
//	existing, err := repo.GetTemplateByName(ctx, "Яндекс Плюс")
//	if err != nil { ... }
//	if existing != nil {
//	    // Шаблон уже существует
//	}
//
// ============================================================
func (r *PostgresRepo) GetTemplateByName(ctx context.Context, serviceName string) (*models.Template, error) {
	// SQL-запрос: ищем шаблон по нормализованному названию
	// LOWER(TRIM(service_name)) = LOWER(TRIM($1)) — игнорируем регистр и пробелы
	query := `SELECT id, service_name, price FROM subscription_templates WHERE LOWER(TRIM(service_name)) = LOWER(TRIM($1))`

	var t models.Template
	err := r.db.QueryRowContext(ctx, query, serviceName).Scan(&t.ID, &t.ServiceName, &t.Price)
	if err != nil {
		// Если не найден — возвращаем nil
		if err == sql.ErrNoRows {
			return nil, nil
		}
		// Другие ошибки логируем
		logger.Error("GetTemplateByName: failed for name %s: %v", serviceName, err)
		return nil, err
	}

	// Шаблон найден
	logger.Debug("GetTemplateByName: found template id=%d, name=%s", t.ID, t.ServiceName)
	return &t, nil
}

// ============================================================
// 5. ОБНОВЛЕНИЕ ШАБЛОНА (UpdateTemplate)
// ============================================================
// Что делает:
//   - Обновляет название и цену шаблона по ID
//   - Используется админом при редактировании подписки (меняет шаблон для всех)
//   - Проверка уникальности названия происходит через UNIQUE INDEX в БД
//
// ПАРАМЕТРЫ:
//   - ctx: контекст для отмены операций
//   - id: ID шаблона
//   - serviceName: новое название
//   - price: новая цена
//
// ВОЗВРАЩАЕТ:
//   - error: ошибка, если обновление не удалось
//
// ОСОБЕННОСТИ:
//   - Если шаблон не найден — возвращает sql.ErrNoRows
//   - Если новое название уже существует — БД вернёт ошибку 23505 (duplicate key)
//
// ============================================================
func (r *PostgresRepo) UpdateTemplate(ctx context.Context, id int, serviceName string, price int) error {
	// SQL-запрос: обновляем название и цену по ID
	query := `UPDATE subscription_templates SET service_name = $1, price = $2 WHERE id = $3`

	// Выполняем запрос
	result, err := r.db.ExecContext(ctx, query, serviceName, price, id)
	if err != nil {
		logger.Error("UpdateTemplate: failed for id %d: %v", id, err)
		return err
	}

	// Проверяем, сколько строк было обновлено
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("UpdateTemplate: RowsAffected failed for id %d: %v", id, err)
		return err
	}

	// Если 0 — значит шаблон не найден
	if rowsAffected == 0 {
		logger.Warn("UpdateTemplate: no rows affected for id %d", id)
		return sql.ErrNoRows
	}

	// Успешное обновление
	logger.Debug("UpdateTemplate: updated template id=%d", id)
	return nil
}

// ============================================================
// 6. УДАЛЕНИЕ ШАБЛОНА (DeleteTemplate)
// ============================================================
// Что делает:
//   - Удаляет шаблон по ID
//   - Используется админом для удаления шаблона
//   - При удалении шаблона: template_id в подписках становится NULL
//     (благодаря ON DELETE SET NULL в БД)
//
// ПАРАМЕТРЫ:
//   - ctx: контекст для отмены операций
//   - id: ID шаблона
//
// ВОЗВРАЩАЕТ:
//   - error: ошибка, если удаление не удалось
//
// ОСОБЕННОСТИ:
//   - Если шаблон не найден — возвращает sql.ErrNoRows
//   - Подписки пользователей не удаляются (template_id становится NULL)
//
// ============================================================
// DeleteTemplate — удаляет шаблон, если на него нет подписок
func (r *PostgresRepo) DeleteTemplate(ctx context.Context, id int) error {
	// 1. Проверяем, есть ли подписки с этим template_id
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM subscriptions WHERE template_id = $1`, id).Scan(&count)
	if err != nil {
		logger.Error("DeleteTemplate: failed to check subscriptions for template %d: %v", id, err)
		return err
	}

	if count > 0 {
		logger.Warn("DeleteTemplate: cannot delete template %d, %d subscriptions exist", id, count)
		return fmt.Errorf("cannot delete template: %d subscriptions are using it", count)
	}

	// 2. Удаляем шаблон
	query := `DELETE FROM subscription_templates WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		logger.Error("DeleteTemplate: failed for id %d: %v", id, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("DeleteTemplate: RowsAffected failed for id %d: %v", id, err)
		return err
	}
	if rowsAffected == 0 {
		logger.Warn("DeleteTemplate: no rows affected for id %d", id)
		return sql.ErrNoRows
	}

	logger.Debug("DeleteTemplate: deleted template id=%d", id)
	return nil
}
