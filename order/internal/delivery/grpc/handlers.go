package grpc

import (
	"context"

	"github.com/IK-akx/AP2_FINAL_PROJECT/order/internal/domain/entity"
	"github.com/IK-akx/AP2_FINAL_PROJECT/order/internal/domain/service"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	orderpb "github.com/IK-akx/pharmacy-proto-gen/order"
)

// OrderHandler implements the gRPC OrderServiceServer
type OrderHandler struct {
	orderpb.UnimplementedOrderServiceServer
	orderSvc   service.OrderService
	balanceSvc service.BalanceService
	logger     *zap.Logger
}

func NewOrderHandler(orderSvc service.OrderService, balanceSvc service.BalanceService, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{
		orderSvc:   orderSvc,
		balanceSvc: balanceSvc,
		logger:     logger,
	}
}

// CreateOrder handles order creation
func (h *OrderHandler) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.OrderResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	items := make([]service.CreateOrderItem, len(req.Items))
	for i, item := range req.Items {
		productID, err := uuid.Parse(item.ProductId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid product_id in item %d: %v", i, err)
		}
		items[i] = service.CreateOrderItem{
			ProductID: productID,
			Quantity:  item.Quantity,
		}
	}

	order, err := h.orderSvc.CreateOrder(ctx, userID, items)
	if err != nil {
		h.logger.Error("failed to create order", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to create order: %v", err)
	}

	return orderToProto(order), nil
}

// GetOrder retrieves an order by ID
func (h *OrderHandler) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.OrderResponse, error) {
	orderID, err := uuid.Parse(req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid order_id: %v", err)
	}

	order, err := h.orderSvc.GetOrder(ctx, orderID)
	if err != nil {
		h.logger.Error("failed to get order", zap.Error(err))
		return nil, status.Errorf(codes.NotFound, "order not found: %v", err)
	}

	return orderToProto(order), nil
}

// GetUserOrders retrieves user's orders with pagination
func (h *OrderHandler) GetUserOrders(ctx context.Context, req *orderpb.GetUserOrdersRequest) (*orderpb.ListOrdersResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}

	orders, total, err := h.orderSvc.GetUserOrders(ctx, userID, page, limit)
	if err != nil {
		h.logger.Error("failed to get user orders", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get orders: %v", err)
	}

	resp := &orderpb.ListOrdersResponse{
		Orders:     make([]*orderpb.OrderResponse, len(orders)),
		TotalCount: total,
		Page:       page,
		Limit:      limit,
	}
	for i, o := range orders {
		resp.Orders[i] = orderToProto(o)
	}

	return resp, nil
}

// CancelOrder cancels a confirmed order
func (h *OrderHandler) CancelOrder(ctx context.Context, req *orderpb.CancelOrderRequest) (*orderpb.OrderResponse, error) {
	orderID, err := uuid.Parse(req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid order_id: %v", err)
	}

	order, err := h.orderSvc.CancelOrder(ctx, orderID)
	if err != nil {
		h.logger.Error("failed to cancel order", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to cancel order: %v", err)
	}

	return orderToProto(order), nil
}

// GetUserBalance retrieves user balance
func (h *OrderHandler) GetUserBalance(ctx context.Context, req *orderpb.GetBalanceRequest) (*orderpb.BalanceResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	balance, err := h.balanceSvc.GetUserBalance(ctx, userID)
	if err != nil {
		h.logger.Error("failed to get balance", zap.Error(err))
		return nil, status.Errorf(codes.NotFound, "balance not found: %v", err)
	}

	return &orderpb.BalanceResponse{
		UserId:    balance.UserID.String(),
		Balance:   balance.Balance,
		UpdatedAt: timestamppb.New(balance.UpdatedAt),
	}, nil
}

// === Helper functions ===

func orderToProto(order *entity.Order) *orderpb.OrderResponse {
	items := make([]*orderpb.OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		items[i] = &orderpb.OrderItemResponse{
			Id:        item.ID.String(),
			ProductId: item.ProductID.String(),
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	return &orderpb.OrderResponse{
		Id:        order.ID.String(),
		UserId:    order.UserID.String(),
		Status:    order.Status,
		Total:     order.Total,
		Items:     items,
		CreatedAt: timestamppb.New(order.CreatedAt),
		UpdatedAt: timestamppb.New(order.UpdatedAt),
	}
}
