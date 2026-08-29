package grpcserver

import (
	"database/sql"
	"errors"

	"Rest-user-agregator/internal/service"
	"Rest-user-agregator/pkg/logger"

	"google.golang.org/grpc/codes"
)

// mapErrorToGRPCStatus — маппер ошибок в gRPC-коды
func mapErrorToGRPCStatus(err error) (codes.Code, string) {
	if err == nil {
		return codes.OK, ""
	}

	switch {
	// ============================================================
	// InvalidArgument (400) — ошибки валидации
	// ============================================================
	case errors.Is(err, service.ErrInvalidID),
		errors.Is(err, service.ErrUserIDRequired),
		errors.Is(err, service.ErrTemplateIDRequired),
		errors.Is(err, service.ErrStartDateRequired),
		errors.Is(err, service.ErrEndDateRequired),
		errors.Is(err, service.ErrInvalidDateRange),
		errors.Is(err, service.ErrInvalidDateFormat),
		errors.Is(err, service.ErrCannotChangeStartDate),
		errors.Is(err, service.ErrPriceNegative),
		errors.Is(err, service.ErrServiceNameRequired):
		return codes.InvalidArgument, err.Error()

	// ============================================================
	// PermissionDenied (403) — нет прав
	// ============================================================
	case errors.Is(err, service.ErrPermissionDenied):
		return codes.PermissionDenied, err.Error()

	// ============================================================
	// NotFound (404) — ресурс не найден
	// ============================================================
	case errors.Is(err, service.ErrTemplateNotFound),
		errors.Is(err, sql.ErrNoRows):
		return codes.NotFound, err.Error()

	// ============================================================
	// AlreadyExists (409) — конфликт
	// ============================================================
	case errors.Is(err, service.ErrTemplateAlreadyExists),
		errors.Is(err, service.ErrTemplateHasSubscriptions):
		return codes.AlreadyExists, err.Error()

	// ============================================================
	// Internal (500) — всё остальное
	// ============================================================
	default:
		logger.Error("Unmapped error: %v", err)
		return codes.Internal, "Internal server error"
	}
}
