package grpcserver

import (
	"Rest-user-agregator/internal/models"
	"Rest-user-agregator/internal/service"
	pb "Rest-user-agregator/proto/subscription"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SubscriptionServer struct {
	pb.UnimplementedSubscriptionServiceServer
	svc *service.SubscriptionService
}

func NewSubscriptionServer(svc *service.SubscriptionService) *SubscriptionServer {
	return &SubscriptionServer{svc: svc}
}
// GetSubscriptions — список всех подписок
func (s *SubscriptionServer) GetSubscriptions(ctx context.Context, req *pb.GetSubscriptionsRequest) (*pb.SubscriptionList, error) {
    limit := int(req.Limit)
    offset := int(req.Offset)

    if limit <= 0 {
        limit = 20
    }
    if offset < 0 {
        offset = 0
    }

    subs, err := s.svc.ListSubscriptions(ctx, limit, offset)
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }
    return &pb.SubscriptionList{Subscriptions: toProtoSubscriptions(subs)}, nil
}

// CreateSubscription — создание подписки
func (s *SubscriptionServer) CreateSubscription(ctx context.Context, req *pb.CreateRequest) (*pb.CreateResponse, error) {
	sub := models.Subscription{
		ServiceName: req.ServiceName,
		Price:       int(req.Price),
		UserID:      req.UserId,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}
	id, err := s.svc.CreateSubscription(ctx, sub)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.CreateResponse{Id: int32(id)}, nil
}

// GetSubscription — получить одну подписку по ID
func (s *SubscriptionServer) GetSubscription(ctx context.Context, req *pb.GetRequest) (*pb.Subscription, error) {
	sub, err := s.svc.GetSubscriptionByID(ctx, int(req.Id))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if sub == nil {
		return nil, status.Error(codes.NotFound, "subscription not found")
	}
	return toProtoSubscription(sub), nil
}

// UpdateSubscription — обновление подписки
func (s *SubscriptionServer) UpdateSubscription(ctx context.Context, req *pb.UpdateRequest) (*pb.Empty, error) {
	sub := models.Subscription{
		ID:          int(req.Id),
		ServiceName: req.ServiceName,
		Price:       int(req.Price),
		UserID:      req.UserId,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}
	err := s.svc.UpdateSubscription(ctx, sub)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Empty{}, nil
}

// DeleteSubscription — удаление подписки
func (s *SubscriptionServer) DeleteSubscription(ctx context.Context, req *pb.GetRequest) (*pb.Empty, error) {
	err := s.svc.DeleteSubscription(ctx, int(req.Id))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Empty{}, nil
}

// GetTotalCost — суммарная стоимость за период
func (s *SubscriptionServer) GetTotalCost(ctx context.Context, req *pb.TotalCostRequest) (*pb.TotalCostResponse, error) {
	total, err := s.svc.GetTotalCost(ctx, req.UserId, req.ServiceName, req.StartDate, req.EndDate)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.TotalCostResponse{Total: int32(total)}, nil
}

// ===== ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ (маппинг) =====

// toProtoSubscription — преобразует models.Subscription → pb.Subscription
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

// toProtoSubscriptions — преобразует []models.Subscription → []*pb.Subscription
func toProtoSubscriptions(subs []models.Subscription) []*pb.Subscription {
	result := make([]*pb.Subscription, len(subs))
	for i := range subs {
		result[i] = toProtoSubscription(&subs[i])
	}
	return result
}
