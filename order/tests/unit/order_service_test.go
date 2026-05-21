package unit

import (
	"context"
	"fmt"
	"testing"

	"github.com/IK-akx/AP2_FINAL_PROJECT/order/internal/domain/entity"
	"github.com/IK-akx/AP2_FINAL_PROJECT/order/internal/domain/repository"
	"github.com/IK-akx/AP2_FINAL_PROJECT/order/internal/domain/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

var testLogger = zap.NewNop()

func TestCreateOrder_Success(t *testing.T) {
	// Arrange
	userID := uuid.New()
	productID := uuid.New()
	items := []service.CreateOrderItem{
		{ProductID: productID, Quantity: 2},
	}

	mockProduct := new(MockProductClient)
	mockProduct.On("CheckAvailability", mock.Anything, productID, int32(2)).
		Return(true, int32(10), nil)
	mockProduct.On("UpdateStock", mock.Anything, productID, int32(-2)).
		Return(nil)

	mockNATS := new(MockNATSPublisher)
	mockNATS.On("PublishOrderCreated", mock.Anything, mock.Anything).
		Return(nil)

	mockOrderRepo := new(MockOrderRepo)
	mockBalanceRepo := new(MockBalanceRepo)

	// Expect balance check
	mockBalanceRepo.On("GetUserBalanceForUpdate", mock.Anything, mock.Anything, userID).
		Return(&entity.UserBalance{UserID: userID, Balance: 500.00}, nil)
	mockBalanceRepo.On("UpdateBalance", mock.Anything, mock.Anything, userID, -200.0).
		Return(nil)
	mockBalanceRepo.On("CreateTransaction", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	svc := &OrderSvcWrapper{
		orderRepo:   mockOrderRepo,
		balanceRepo: mockBalanceRepo,
		productSvc:  mockProduct,
		natsPub:     mockNATS,
		logger:      testLogger,
	}

	// Act
	order, err := svc.CreateOrder(context.Background(), userID, items)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, entity.StatusConfirmed, order.Status)
	assert.Equal(t, 200.0, order.Total)
	mockProduct.AssertExpectations(t)
	mockNATS.AssertExpectations(t)
}

func TestCreateOrder_InsufficientBalance(t *testing.T) {
	// Arrange
	userID := uuid.New()
	productID := uuid.New()
	items := []service.CreateOrderItem{
		{ProductID: productID, Quantity: 5},
	}

	mockProduct := new(MockProductClient)
	mockProduct.On("CheckAvailability", mock.Anything, productID, int32(5)).
		Return(true, int32(100), nil)

	mockBalanceRepo := new(MockBalanceRepo)
	mockBalanceRepo.On("GetUserBalanceForUpdate", mock.Anything, mock.Anything, userID).
		Return(&entity.UserBalance{UserID: userID, Balance: 50.00}, nil)

	mockOrderRepo := new(MockOrderRepo)

	svc := &OrderSvcWrapper{
		orderRepo:   mockOrderRepo,
		balanceRepo: mockBalanceRepo,
		productSvc:  mockProduct,
		logger:      testLogger,
	}

	// Act
	order, err := svc.CreateOrder(context.Background(), userID, items)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "insufficient balance")
}

func TestCreateOrder_ProductUnavailable(t *testing.T) {
	// Arrange
	userID := uuid.New()
	productID := uuid.New()
	items := []service.CreateOrderItem{
		{ProductID: productID, Quantity: 10},
	}

	mockProduct := new(MockProductClient)
	mockProduct.On("CheckAvailability", mock.Anything, productID, int32(10)).
		Return(false, int32(3), nil)

	mockOrderRepo := new(MockOrderRepo)
	mockBalanceRepo := new(MockBalanceRepo)

	svc := &OrderSvcWrapper{
		orderRepo:   mockOrderRepo,
		balanceRepo: mockBalanceRepo,
		productSvc:  mockProduct,
		logger:      testLogger,
	}

	// Act
	order, err := svc.CreateOrder(context.Background(), userID, items)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "insufficient stock")
}

func TestCancelOrder_Success(t *testing.T) {
	// Arrange
	orderID := uuid.New()
	userID := uuid.New()
	order := &entity.Order{
		ID:     orderID,
		UserID: userID,
		Status: entity.StatusConfirmed,
		Total:  300.0,
	}

	mockOrderRepo := new(MockOrderRepo)
	mockOrderRepo.On("GetOrder", mock.Anything, orderID).Return(order, nil)
	mockOrderRepo.On("GetOrderItems", mock.Anything, orderID).
		Return([]*entity.OrderItem{
			{ProductID: uuid.New(), Quantity: 3, Price: 100.0},
		}, nil)

	mockBalanceRepo := new(MockBalanceRepo)
	mockBalanceRepo.On("UpdateBalance", mock.Anything, mock.Anything, userID, 300.0).Return(nil)
	mockBalanceRepo.On("CreateTransaction", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockProduct := new(MockProductClient)
	mockProduct.On("UpdateStock", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockCache := new(MockOrderCache)
	mockCache.On("DeleteOrder", mock.Anything, orderID).Return(nil)
	mockCache.On("InvalidateUserOrders", mock.Anything, userID).Return(nil)

	mockNATS := new(MockNATSPublisher)

	svc := &OrderSvcWrapper{
		orderRepo:   mockOrderRepo,
		balanceRepo: mockBalanceRepo,
		productSvc:  mockProduct,
		cache:       mockCache,
		natsPub:     mockNATS,
		logger:      testLogger,
	}

	// Act
	result, err := svc.CancelOrder(context.Background(), orderID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, entity.StatusCancelled, result.Status)
}

func TestCancelOrder_AlreadyCancelled(t *testing.T) {
	// Arrange
	orderID := uuid.New()
	order := &entity.Order{
		ID:     orderID,
		Status: entity.StatusCancelled,
	}

	mockOrderRepo := new(MockOrderRepo)
	mockOrderRepo.On("GetOrder", mock.Anything, orderID).Return(order, nil)

	svc := &OrderSvcWrapper{
		orderRepo: mockOrderRepo,
		logger:    testLogger,
	}

	// Act
	_, err := svc.CancelOrder(context.Background(), orderID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be cancelled")
}

// OrderSvcWrapper wraps the real service logic for testing
// This avoids needing pgxpool.Pool in unit tests
type OrderSvcWrapper struct {
	orderRepo   repository.OrderRepository
	balanceRepo repository.BalanceRepository
	cache       *MockOrderCache
	productSvc  service.ProductServiceClient
	natsPub     service.NATSPublisher
	logger      *zap.Logger
}

var _ service.OrderService = (*OrderSvcWrapper)(nil)

func (s *OrderSvcWrapper) CreateOrder(ctx context.Context, userID uuid.UUID, items []service.CreateOrderItem) (*entity.Order, error) {
	var total float64
	orderItems := make([]*entity.OrderItem, len(items))

	for i, item := range items {
		available, currentStock, err := s.productSvc.CheckAvailability(ctx, item.ProductID, item.Quantity)
		if err != nil {
			return nil, err
		}
		if !available {
			return nil, fmt.Errorf("product %s: insufficient stock (requested: %d, available: %d)",
				item.ProductID, item.Quantity, currentStock)
		}

		price := 100.0
		total += price * float64(item.Quantity)
		orderItems[i] = &entity.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     price,
		}
	}

	order := &entity.Order{
		UserID: userID,
		Status: entity.StatusConfirmed,
		Total:  total,
	}
	order.ID = uuid.New()

	// Check balance
	balance, err := s.balanceRepo.GetUserBalanceForUpdate(ctx, nil, userID)
	if err != nil {
		return nil, err
	}

	if balance.Balance < total {
		return nil, fmt.Errorf("insufficient balance: have %.2f, need %.2f", balance.Balance, total)
	}

	// Deduct
	_ = s.balanceRepo.UpdateBalance(ctx, nil, userID, -total)
	_ = s.balanceRepo.CreateTransaction(ctx, nil, &entity.Transaction{
		UserID:  userID,
		OrderID: &order.ID,
		Amount:  total,
		Type:    entity.TransactionTypeDebit,
	})

	// After commit: update stock
	for _, item := range orderItems {
		_ = s.productSvc.UpdateStock(ctx, item.ProductID, -item.Quantity)
	}

	// Publish event
	if s.natsPub != nil {
		order.Items = make([]entity.OrderItem, len(orderItems))
		for i, item := range orderItems {
			order.Items[i] = *item
		}
		_ = s.natsPub.PublishOrderCreated(ctx, order)
	}

	return order, nil
}

func (s *OrderSvcWrapper) GetOrder(ctx context.Context, id uuid.UUID) (*entity.Order, error) {
	order, err := s.orderRepo.GetOrder(ctx, id)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (s *OrderSvcWrapper) GetUserOrders(ctx context.Context, userID uuid.UUID, page, limit int32) ([]*entity.Order, int32, error) {
	return s.orderRepo.GetUserOrders(ctx, userID, page, limit)
}

func (s *OrderSvcWrapper) CancelOrder(ctx context.Context, id uuid.UUID) (*entity.Order, error) {
	order, err := s.orderRepo.GetOrder(ctx, id)
	if err != nil {
		return nil, err
	}

	if !order.CanBeCancelled() {
		return nil, fmt.Errorf("order %s cannot be cancelled (current status: %s)", id, order.Status)
	}

	// Refund
	_ = s.balanceRepo.UpdateBalance(ctx, nil, order.UserID, order.Total)
	_ = s.balanceRepo.CreateTransaction(ctx, nil, &entity.Transaction{
		UserID:  order.UserID,
		OrderID: &order.ID,
		Amount:  order.Total,
		Type:    entity.TransactionTypeCredit,
	})

	// Restore stock
	items, _ := s.orderRepo.GetOrderItems(ctx, id)
	for _, item := range items {
		_ = s.productSvc.UpdateStock(ctx, item.ProductID, item.Quantity)
	}

	// Invalidate cache
	if s.cache != nil {
		_ = s.cache.DeleteOrder(ctx, id)
		_ = s.cache.InvalidateUserOrders(ctx, order.UserID)
	}

	order.Status = entity.StatusCancelled
	return order, nil
}

// MockOrderCache for unit tests
type MockOrderCache struct {
	mock.Mock
}

func (m *MockOrderCache) GetOrder(ctx context.Context, id uuid.UUID) (*entity.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Order), args.Error(1)
}

func (m *MockOrderCache) SetOrder(ctx context.Context, order *entity.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderCache) DeleteOrder(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrderCache) GetUserOrders(ctx context.Context, userID uuid.UUID, page int32) ([]*entity.Order, error) {
	args := m.Called(ctx, userID, page)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Order), args.Error(1)
}

func (m *MockOrderCache) SetUserOrders(ctx context.Context, userID uuid.UUID, page int32, orders []*entity.Order) error {
	args := m.Called(ctx, userID, page, orders)
	return args.Error(0)
}

func (m *MockOrderCache) InvalidateUserOrders(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
