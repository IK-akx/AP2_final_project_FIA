package unit

import (
	"context"

	"github.com/IK-akx/AP2_FINAL_PROJECT/order/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockOrderRepo implements repository.OrderRepository
type MockOrderRepo struct {
	mock.Mock
}

func (m *MockOrderRepo) CreateOrder(ctx context.Context, order *entity.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepo) GetOrder(ctx context.Context, id uuid.UUID) (*entity.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Order), args.Error(1)
}

func (m *MockOrderRepo) GetUserOrders(ctx context.Context, userID uuid.UUID, page, limit int32) ([]*entity.Order, int32, error) {
	args := m.Called(ctx, userID, page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*entity.Order), args.Get(1).(int32), args.Error(2)
}

func (m *MockOrderRepo) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockOrderRepo) CreateOrderItems(ctx context.Context, items []*entity.OrderItem) error {
	args := m.Called(ctx, items)
	return args.Error(0)
}

func (m *MockOrderRepo) GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]*entity.OrderItem, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.OrderItem), args.Error(1)
}

// MockBalanceRepo implements repository.BalanceRepository
type MockBalanceRepo struct {
	mock.Mock
}

func (m *MockBalanceRepo) GetUserBalance(ctx context.Context, userID uuid.UUID) (*entity.UserBalance, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserBalance), args.Error(1)
}

func (m *MockBalanceRepo) GetUserBalanceForUpdate(ctx context.Context, tx interface{}, userID uuid.UUID) (*entity.UserBalance, error) {
	args := m.Called(ctx, tx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserBalance), args.Error(1)
}

func (m *MockBalanceRepo) UpdateBalance(ctx context.Context, tx interface{}, userID uuid.UUID, amount float64) error {
	args := m.Called(ctx, tx, userID, amount)
	return args.Error(0)
}

func (m *MockBalanceRepo) CreateTransaction(ctx context.Context, tx interface{}, txRecord *entity.Transaction) error {
	args := m.Called(ctx, tx, txRecord)
	return args.Error(0)
}

func (m *MockBalanceRepo) TopUpBalance(ctx context.Context, userID uuid.UUID, amount float64, description string) (*entity.UserBalance, error) {
	args := m.Called(ctx, userID, amount, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserBalance), args.Error(1)
}

func (m *MockBalanceRepo) GetTransactionHistory(ctx context.Context, userID uuid.UUID, page, limit int32) ([]*entity.Transaction, int32, error) {
	args := m.Called(ctx, userID, page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*entity.Transaction), args.Get(1).(int32), args.Error(2)
}

// MockProductClient implements service.ProductServiceClient
type MockProductClient struct {
	mock.Mock
}

func (m *MockProductClient) CheckAvailability(ctx context.Context, productID uuid.UUID, quantity int32) (bool, int32, error) {
	args := m.Called(ctx, productID, quantity)
	return args.Bool(0), args.Get(1).(int32), args.Error(2)
}

func (m *MockProductClient) UpdateStock(ctx context.Context, productID uuid.UUID, delta int32) error {
	args := m.Called(ctx, productID, delta)
	return args.Error(0)
}

// MockNATSPublisher implements service.NATSPublisher
type MockNATSPublisher struct {
	mock.Mock
}

func (m *MockNATSPublisher) PublishOrderCreated(ctx context.Context, order *entity.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}
