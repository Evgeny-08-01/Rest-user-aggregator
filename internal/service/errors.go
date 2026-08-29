package service

import "errors"

var (
	// ============================================================
	// ОШИБКИ ШАБЛОНОВ
	// ============================================================
	ErrTemplateNotFound         = errors.New("template not found")
	ErrTemplateAlreadyExists    = errors.New("template with this name already exists")
	ErrTemplateHasSubscriptions = errors.New("cannot delete template: subscriptions are using it")
	ErrServiceNameRequired      = errors.New("service_name is required")
	ErrPriceNegative            = errors.New("price cannot be negative")

	// ============================================================
	// ОШИБКИ ПОДПИСОК (бизнес-логика)
	// ============================================================
	ErrCannotChangeStartDate = errors.New("cannot change start_date that is today or in the past")
	ErrPermissionDenied      = errors.New("permission denied")

	// ============================================================
	// ОШИБКИ ВАЛИДАЦИИ (формат входных данных)
	// ============================================================
	ErrInvalidID          = errors.New("invalid id")
	ErrUserIDRequired     = errors.New("user_id is required")
	ErrTemplateIDRequired = errors.New("template_id is required")
	ErrStartDateRequired  = errors.New("start_date is required")
	ErrEndDateRequired    = errors.New("end_date is required")
	ErrInvalidDateRange   = errors.New("start_date > end_date")
	ErrInvalidDateFormat  = errors.New("invalid date format (expected MM-YYYY)")
)
