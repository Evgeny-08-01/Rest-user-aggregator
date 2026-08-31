// ============================================================
// ПАКЕТ: grpcserver
// ФАЙЛ: grpc_api.go
// НАЗНАЧЕНИЕ: gRPC-сервер — транспортный слой
// ============================================================
//
// ЧТО ЗДЕСЬ ПРОИСХОДИТ:
//   1. Принимает gRPC-запросы (Protobuf)
//   2. Вызывает бизнес-сервис (вся логика там)
//   3. Возвращает gRPC-ответы (Protobuf) или gRPC-ошибки
//
// ВАЖНО:
//   - ВСЯ бизнес-логика (проверки, права, валидация) — в СЕРВИСЕ
//   - gRPC-хендлер — ТОЛЬКО адаптер между Protobuf и сервисом
//   - Ошибки маппятся из service.Err* в gRPC status codes
// ============================================================

package grpcserver

import (
	"context"

	"Rest-user-agregator/internal/authentication"
	"Rest-user-agregator/internal/models"
	"Rest-user-agregator/internal/service"
	"Rest-user-agregator/pkg/logger"
	pb "Rest-user-agregator/proto/subscription"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ============================================================
// 1. СТРУКТУРА СЕРВЕРА
// ============================================================

type SubscriptionServer struct {
	pb.UnimplementedSubscriptionServiceServer
	svc         *service.SubscriptionService
	templateSvc *service.TemplateService
}

func NewSubscriptionServer(
	svc *service.SubscriptionService,
	templateSvc *service.TemplateService,
) *SubscriptionServer {
	return &SubscriptionServer{
		svc:         svc,
		templateSvc: templateSvc,
	}
}

// ============================================================
// 2. ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ (работа с контекстом)
// ============================================================

// getUserIDFromContext — извлекает user_id из контекста (устанавливается interceptor-ом)
func getUserIDFromContext(ctx context.Context) (string, error) {
	// Добавляем логирование
	logger.Debug("getUserIDFromContext: checking user_id")
	
	userID, ok := ctx.Value(authentication.UserIDKey).(string)
	// Добавляем логирование
	logger.Debug("getUserIDFromContext: ok=%v, userID=%s", ok, userID)

	
	if !ok || userID == "" {
		logger.Warn("getUserIDFromContext: user not authenticated")
		return "", status.Error(codes.Unauthenticated, "user not authenticated")
	}
	return userID, nil
}

// getRoleFromContext — извлекает роль из контекста
func getRoleFromContext(ctx context.Context) string {
	role, ok := ctx.Value("role").(string)
	if !ok || role == "" {
		return "user" // по умолчанию
	}
	return role
}

// requireAdmin — проверяет, что пользователь — админ
func requireAdmin(ctx context.Context) error {
	role := getRoleFromContext(ctx)
	if role != "admin" {
		return status.Error(codes.PermissionDenied, "admin only")
	}
	return nil
}

// ============================================================
// 3. МЕТОДЫ ДЛЯ ПОДПИСОК (полный аналог REST)
// ============================================================
// ============================================================
// CreateSubscription — создание новой подписки (gRPC)
// ============================================================
// Аналог: POST /api/subscriptions
// ============================================================
// Логика работы:
//  1. Проверяет авторизацию (userID из JWT)
//  2. Вызывает сервис (вся валидация внутри)
//  3. Возвращает ID созданной подписки или gRPC-ошибку
//
// ============================================================
func (s *SubscriptionServer) CreateSubscription(ctx context.Context, req *pb.CreateRequest) (*pb.CreateResponse, error) {
	logger.Debug("gRPC CreateSubscription: template_id=%d, user_id=%s, start_date=%s, end_date=%s",
		req.TemplateId, req.UserId, req.StartDate, req.EndDate)

	// 1. Проверка авторизации (userID из контекста)
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		logger.Warn("gRPC CreateSubscription: unauthorized")
		return nil, err
	}

	// 2. Вызов сервиса (ВСЯ ВАЛИДАЦИЯ ВНУТРИ)
	//    Сервис сам проверит:
	//    - template_id > 0 (ErrTemplateIDRequired)
	//    - user_id не пустой (ErrUserIDRequired)
	//    - start_date не пустой (ErrStartDateRequired)
	//    - Существует ли шаблон (ErrTemplateNotFound)
	//    - Корректны ли даты (ErrCannotChangeStartDate)
	id, err := s.svc.CreateSubscription(ctx, int(req.TemplateId), userID, req.StartDate, req.EndDate)
	if err != nil {
		code, msg := mapErrorToGRPCStatus(err)
		logger.Warn("gRPC CreateSubscription: %v", err)
		return nil, status.Error(code, msg)
	}

	// 4. Успешный ответ
	logger.Debug("gRPC CreateSubscription: created id=%d for user=%s", id, userID)
	return &pb.CreateResponse{Id: int32(id)}, nil
}

// ============================================================
// GetSubscription — получение подписки по ID (gRPC)
// ============================================================
// Аналог: GET /api/subscriptions/{id}
// ============================================================
// Логика работы:
//  1. Проверяет авторизацию (userID из JWT)
//  2. Вызывает сервис (вся валидация и проверка прав внутри)
//  3. Возвращает подписку или gRPC-ошибку
//
// ============================================================
func (s *SubscriptionServer) GetSubscription(ctx context.Context, req *pb.GetRequest) (*pb.Subscription, error) {
	logger.Debug("gRPC GetSubscription: id=%d", req.Id)

	// 1. Проверка авторизации (userID и роль из контекста)
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		logger.Warn("gRPC GetSubscription: unauthorized")
		return nil, err
	}
	role := getRoleFromContext(ctx)

	// 2. Вызов сервиса (ВСЯ ВАЛИДАЦИЯ И ПРОВЕРКА ПРАВ ВНУТРИ)
	//    Сервис сам проверит:
	//    - id > 0 (ErrInvalidID)
	//    - user_id не пустой (ErrUserIDRequired)
	//    - Права доступа (ErrPermissionDenied)
	sub, err := s.svc.GetSubscriptionByID(ctx, int(req.Id), userID, role)
	if err != nil {
		code, msg := mapErrorToGRPCStatus(err)
		logger.Warn("gRPC GetSubscription: %v", err)
		return nil, status.Error(code, msg)
	}
	if sub == nil {
		logger.Warn("gRPC GetSubscription: subscription not found for id=%d", req.Id)
		return nil, status.Error(codes.NotFound, "Subscription not found")
	}

	// 4. Успешный ответ
	logger.Debug("gRPC GetSubscription: returned subscription id=%d for user=%s", req.Id, userID)
	return toProtoSubscription(sub), nil
}

// ============================================================
// GetSubscriptions — получение списка подписок (gRPC)
// ============================================================
// Аналог: GET /api/subscriptions?limit=20&offset=0
// ============================================================
// Логика работы:
//  1. Проверяет авторизацию (userID и роль из JWT)
//  2. Устанавливает значения пагинации по умолчанию
//  3. Вызывает сервис (валидация прав внутри)
//  4. Возвращает список подписок или gRPC-ошибку
//
// ============================================================
func (s *SubscriptionServer) GetSubscriptions(ctx context.Context, req *pb.GetSubscriptionsRequest) (*pb.SubscriptionList, error) {
	logger.Debug("gRPC GetSubscriptions: limit=%d, offset=%d", req.Limit, req.Offset)

	// 1. Проверка авторизации (userID и роль из контекста)
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		logger.Warn("gRPC GetSubscriptions: unauthorized")
		return nil, err
	}
	role := getRoleFromContext(ctx)

	// 2. Устанавливаем значения по умолчанию (как в REST)
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 20
	}
	offset := int(req.Offset)
	if offset < 0 {
		offset = 0
	}

	// 3. Вызываем сервис (валидация прав внутри)
	//    Сервис сам проверит:
	//    - Для не-админа: userID не пустой (ErrUserIDRequired)
	list, err := s.svc.ListSubscriptions(ctx, userID, role, limit, offset)
	if err != nil {
		code, msg := mapErrorToGRPCStatus(err)
		logger.Warn("gRPC GetSubscriptions: %v", err)
		return nil, status.Error(code, msg)
	}

	// 5. Если nil → пустой массив
	if list == nil {
		list = []models.Subscription{}
	}

	// 6. Успешный ответ → 200 OK
	logger.Debug("gRPC GetSubscriptions: returned %d subscriptions", len(list))
	return &pb.SubscriptionList{Subscriptions: toProtoSubscriptions(list)}, nil
}

// ============================================================
// UpdateSubscription — обновление подписки (gRPC)
// ============================================================
// Аналог: PUT /api/subscriptions/{id}
// ============================================================
// Логика работы:
//  1. Проверяет авторизацию (userID из JWT)
//  2. Собирает данные из запроса
//  3. Вызывает сервис (вся валидация и проверка прав внутри)
//  4. Возвращает Empty или gRPC-ошибку
//
// ============================================================
func (s *SubscriptionServer) UpdateSubscription(ctx context.Context, req *pb.UpdateRequest) (*pb.Empty, error) {
	logger.Debug("gRPC UpdateSubscription: id=%d, template_id=%d, user_id=%s, start_date=%s, end_date=%s",
		req.Id, req.TemplateId, req.UserId, req.StartDate, req.EndDate)

	// 1. Проверка авторизации (userID и роль из контекста)
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		logger.Warn("gRPC UpdateSubscription: unauthorized")
		return nil, err
	}
	role := getRoleFromContext(ctx)

	// 2. Собираем данные подписки из запроса
	sub := models.Subscription{
		ID:         int(req.Id),
		UserID:     req.UserId,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		TemplateID: int(req.TemplateId),
	}

	// 3. Вызов сервиса (ВСЯ ВАЛИДАЦИЯ И ПРОВЕРКА ПРАВ ВНУТРИ)
	//    Сервис сам проверит:
	//    - sub.ID > 0 (ErrInvalidID)
	//    - sub.UserID не пустой (ErrUserIDRequired)
	//    - sub.StartDate не пустой (ErrStartDateRequired)
	//    - Права доступа (ErrPermissionDenied)
	//    - Существует ли подписка (sql.ErrNoRows)
	//    - Корректны ли даты (ErrCannotChangeStartDate)
	err = s.svc.UpdateSubscription(ctx, sub, role)
	if err != nil {
		code, msg := mapErrorToGRPCStatus(err)
		logger.Warn("gRPC UpdateSubscription: user=%s, error=%v", userID, err)
		return nil, status.Error(code, msg)
	}

	// 5. Успешный ответ
	return &pb.Empty{}, nil
}

// ============================================================
// DeleteSubscription — удаление подписки (gRPC)
// ============================================================
// Аналог: DELETE /api/subscriptions/{id}
// ============================================================
// Логика работы:
//  1. Проверяет авторизацию (userID из JWT)
//  2. Вызывает сервис (вся валидация и проверка прав внутри)
//  3. Возвращает Empty или gRPC-ошибку
//
// ============================================================
func (s *SubscriptionServer) DeleteSubscription(ctx context.Context, req *pb.GetRequest) (*pb.Empty, error) {
	logger.Debug("gRPC DeleteSubscription: id=%d", req.Id)

	// 1. Проверка авторизации (userID и роль из контекста)
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		logger.Warn("gRPC DeleteSubscription: unauthorized")
		return nil, err
	}
	role := getRoleFromContext(ctx)

	// 2. Вызов сервиса (ВСЯ ВАЛИДАЦИЯ И ПРОВЕРКА ПРАВ ВНУТРИ)
	//    Сервис сам проверит:
	//    - id > 0 (ErrInvalidID)
	//    - userID не пустой (ErrUserIDRequired)
	//    - Права доступа (ErrPermissionDenied)
	//    - Существует ли подписка (sql.ErrNoRows)
	err = s.svc.DeleteSubscription(ctx, int(req.Id), userID, role)
	if err != nil {
		code, msg := mapErrorToGRPCStatus(err)
		logger.Warn("gRPC DeleteSubscription: %v", err)
		return nil, status.Error(code, msg)
	}

	// 4. Успешный ответ
	logger.Debug("gRPC DeleteSubscription: deleted subscription id=%d", req.Id)
	return &pb.Empty{}, nil
}

// ============================================================
// GetTotalCost — расчёт суммарной стоимости подписок (gRPC)
// ============================================================
// Аналог: GET /api/subscriptions/total-cost
// ============================================================
// Логика работы:
//  1. Проверяет авторизацию (userID из JWT)
//  2. Определяет userID для фильтрации:
//     - Админ: filterUserID = "" (все подписки)
//     - Пользователь: filterUserID = userID (только свои)
//  3. Вызывает сервис (вся валидация внутри)
//  4. Возвращает сумму или gRPC-ошибку
//
// ============================================================
func (s *SubscriptionServer) GetTotalCost(ctx context.Context, req *pb.TotalCostRequest) (*pb.TotalCostResponse, error) {
	logger.Debug("gRPC GetTotalCost: user_id=%s, service_name=%s, start_date=%s, end_date=%s",
		req.UserId, req.ServiceName, req.StartDate, req.EndDate)

	// 1. Проверка авторизации (userID из контекста)
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		logger.Warn("gRPC GetTotalCost: unauthorized")
		return nil, err
	}
	role := getRoleFromContext(ctx)

	// 2. Определяем userID для фильтрации
	var filterUserID string
	if role == "admin" {
		filterUserID = "" // админ видит все подписки
	} else {
		filterUserID = userID // пользователь — только свои
	}

	// 3. Вызов сервиса (ВСЯ ВАЛИДАЦИЯ ДАТ ВНУТРИ)
	//    Сервис сам проверит:
	//    - startDate не пустой (ErrStartDateRequired)
	//    - endDate не пустой (ErrEndDateRequired)
	//    - startDate > endDate (ErrInvalidDateRange)
	//    - Неправильный формат даты (ErrInvalidDateFormat)
	total, err := s.svc.GetTotalCost(ctx, filterUserID, req.ServiceName, req.StartDate, req.EndDate)
	if err != nil {
		code, msg := mapErrorToGRPCStatus(err)
		logger.Warn("gRPC GetTotalCost: %v", err)
		return nil, status.Error(code, msg)
	}

	// 5. Успешный ответ
	logger.Debug("gRPC GetTotalCost: total=%d for user=%s", total, filterUserID)
	return &pb.TotalCostResponse{Total: int32(total)}, nil
}

// ============================================================
// 4. МЕТОДЫ ДЛЯ ШАБЛОНОВ (полный аналог REST)
// ============================================================

// ============================================================
// 4.1. LIST TEMPLATES
// ============================================================
// Аналог: GET /api/templates
// ============================================================
func (s *SubscriptionServer) ListTemplates(ctx context.Context, req *pb.Empty) (*pb.TemplateList, error) {
	logger.Debug("gRPC ListTemplates")

	// 1. Проверка авторизации
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		logger.Warn("gRPC ListTemplates: unauthorized")
		return nil, err
	}

	// 2. Получаем шаблоны
	templates, err := s.templateSvc.ListTemplates(ctx)
	if err != nil {
		code, msg := mapErrorToGRPCStatus(err)
		logger.Warn("gRPC ListTemplates: %v", err)
		return nil, status.Error(code, msg)
	}

	// 3. Если nil → пустой массив
	if templates == nil {
		templates = []models.Template{}
	}

	logger.Debug("gRPC ListTemplates: returned %d templates", len(templates))
	return &pb.TemplateList{Templates: toProtoTemplates(templates)}, nil
}

// ============================================================
// 4.2. GET TEMPLATE BY ID
// ============================================================
// Аналог: GET /api/templates/{id}
// ============================================================
func (s *SubscriptionServer) GetTemplate(ctx context.Context, req *pb.GetRequest) (*pb.Template, error) {
	logger.Debug("gRPC GetTemplate: id=%d", req.Id)

	// 1. Проверка авторизации
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		logger.Warn("gRPC GetTemplate: unauthorized")
		return nil, err
	}

	// 2. Валидация
	if req.Id <= 0 {
		logger.Warn("gRPC GetTemplate: invalid id=%d", req.Id)
		return nil, status.Error(codes.InvalidArgument, "Invalid template ID")
	}

	// 3. Получаем шаблон
	template, err := s.templateSvc.GetTemplateByID(ctx, int(req.Id))
	if err != nil {
		code, msg := mapErrorToGRPCStatus(err)
		logger.Warn("gRPC GetTemplate: %v", err)
		return nil, status.Error(code, msg)
	}

	if template == nil {
		logger.Warn("gRPC GetTemplate: template not found for id=%d", req.Id)
		return nil, status.Error(codes.NotFound, "template not found")
	}

	logger.Debug("gRPC GetTemplate: returned template id=%d", req.Id)
	return toProtoTemplate(template), nil
}

// ============================================================
// CreateTemplate — создание нового шаблона (gRPC)
// ============================================================
// Аналог: POST /api/admin/templates
// ============================================================
// Логика работы:
//  1. Проверяет авторизацию (userID из JWT)
//  2. Проверяет роль (только admin)
//  3. Вызывает сервис (вся валидация внутри)
//  4. Возвращает ID созданного шаблона или gRPC-ошибку
//
// ============================================================
func (s *SubscriptionServer) CreateTemplate(ctx context.Context, req *pb.CreateTemplateRequest) (*pb.CreateResponse, error) {
	logger.Debug("gRPC CreateTemplate: service_name=%s, price=%d", req.ServiceName, req.Price)

	// 1. Проверка авторизации (userID из контекста)
	//    Если пользователь не авторизован — возвращаем Unauthenticated
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		logger.Warn("gRPC CreateTemplate: unauthorized")
		return nil, err
	}

	// 2. Проверка роли (только админ)
	//    Если роль не admin — возвращаем PermissionDenied
	if err := requireAdmin(ctx); err != nil {
		logger.Warn("gRPC CreateTemplate: non-admin attempt")
		return nil, err
	}

	// 3. Вызываем сервис (ВСЯ ВАЛИДАЦИЯ ВНУТРИ)
	//    Сервис сам проверит:
	//    - service_name не пустой (ErrServiceNameRequired)
	//    - price >= 0 (ErrPriceNegative)
	//    - Нет шаблона с таким названием (ErrTemplateAlreadyExists)
	id, err := s.templateSvc.CreateTemplate(ctx, req.ServiceName, int(req.Price))
	if err != nil {
		code, msg := mapErrorToGRPCStatus(err)
		logger.Warn("gRPC CreateTemplate: %v", err)
		return nil, status.Error(code, msg)
	}

	// 5. Успешный ответ → CreateResponse с ID созданного шаблона
	logger.Info("gRPC CreateTemplate: template created id=%d, name=%s", id, req.ServiceName)
	return &pb.CreateResponse{Id: int32(id)}, nil
}

// ============================================================
// 4.4. UPDATE TEMPLATE (только админ)
// ============================================================
// Аналог: PUT /api/admin/templates/{id}
// ============================================================
func (s *SubscriptionServer) UpdateTemplate(ctx context.Context, req *pb.UpdateTemplateRequest) (*pb.Empty, error) {
	logger.Debug("gRPC UpdateTemplate: id=%d, service_name=%s, price=%d", req.Id, req.ServiceName, req.Price)

	// 1. Проверка авторизации
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		logger.Warn("gRPC UpdateTemplate: unauthorized")
		return nil, err
	}

	// 2. Проверка роли (только админ)
	if err := requireAdmin(ctx); err != nil {
		logger.Warn("gRPC UpdateTemplate: non-admin attempt")
		return nil, err
	}

	// 3. Валидация
	if req.Id <= 0 {
		logger.Warn("gRPC UpdateTemplate: invalid id=%d", req.Id)
		return nil, status.Error(codes.InvalidArgument, "Invalid template ID")
	}
	if req.ServiceName == "" {
		logger.Warn("gRPC UpdateTemplate: empty service_name")
		return nil, status.Error(codes.InvalidArgument, "service_name is required")
	}
	if req.Price < 0 {
		logger.Warn("gRPC UpdateTemplate: negative price: %d", req.Price)
		return nil, status.Error(codes.InvalidArgument, "price cannot be negative")
	}

	// 4. Вызываем сервис
	err = s.templateSvc.UpdateTemplate(ctx, int(req.Id), req.ServiceName, int(req.Price))
	if err != nil {
		code, msg := mapErrorToGRPCStatus(err)
		logger.Warn("gRPC UpdateTemplate: %v", err)
		return nil, status.Error(code, msg)
	}

	logger.Info("gRPC UpdateTemplate: template updated id=%d, name=%s", req.Id, req.ServiceName)
	return &pb.Empty{}, nil
}

// ============================================================
// 4.5. DELETE TEMPLATE (только админ)
// ============================================================
// Аналог: DELETE /api/admin/templates/{id}
// ============================================================
func (s *SubscriptionServer) DeleteTemplate(ctx context.Context, req *pb.GetRequest) (*pb.Empty, error) {
	logger.Debug("gRPC DeleteTemplate: id=%d", req.Id)

	// 1. Проверка авторизации
	_, err := getUserIDFromContext(ctx)
	if err != nil {
		logger.Warn("gRPC DeleteTemplate: unauthorized")
		return nil, err
	}

	// 2. Проверка роли (только админ)
	if err := requireAdmin(ctx); err != nil {
		logger.Warn("gRPC DeleteTemplate: non-admin attempt")
		return nil, err
	}

	// 3. Валидация
	if req.Id <= 0 {
		logger.Warn("gRPC DeleteTemplate: invalid id=%d", req.Id)
		return nil, status.Error(codes.InvalidArgument, "Invalid template ID")
	}

	// 4. Вызываем сервис
	err = s.templateSvc.DeleteTemplate(ctx, int(req.Id))
	if err != nil {
		code, msg := mapErrorToGRPCStatus(err)
		logger.Warn("gRPC DeleteTemplate: %v", err)
		return nil, status.Error(code, msg)
	}
	logger.Info("gRPC DeleteTemplate: template deleted id=%d", req.Id)
	return &pb.Empty{}, nil
}

// ============================================================
// 5. ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ (маппинг моделей → protobuf)
// ============================================================

// toProtoSubscription — models.Subscription → pb.Subscription
func toProtoSubscription(sub *models.Subscription) *pb.Subscription {
	if sub == nil {
		return nil
	}
	return &pb.Subscription{
		Id:          int32(sub.ID),
		ServiceName: sub.ServiceName,
		Price:       int32(sub.Price),
		UserId:      sub.UserID,
		StartDate:   sub.StartDate,
		EndDate:     sub.EndDate,
	}
}

// toProtoSubscriptions — []models.Subscription → []*pb.Subscription
func toProtoSubscriptions(subs []models.Subscription) []*pb.Subscription {
	result := make([]*pb.Subscription, len(subs))
	for i := range subs {
		result[i] = toProtoSubscription(&subs[i])
	}
	return result
}

// toProtoTemplate — models.Template → pb.Template
func toProtoTemplate(t *models.Template) *pb.Template {
	if t == nil {
		return nil
	}
	return &pb.Template{
		Id:          int32(t.ID),
		ServiceName: t.ServiceName,
		Price:       int32(t.Price),
	}
}

// toProtoTemplates — []models.Template → []*pb.Template
func toProtoTemplates(templates []models.Template) []*pb.Template {
	result := make([]*pb.Template, len(templates))
	for i := range templates {
		result[i] = toProtoTemplate(&templates[i])
	}
	return result
}
