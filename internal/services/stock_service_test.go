package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/huguescodeur/oz-rest-api-go/internal/models"
	"github.com/huguescodeur/oz-rest-api-go/internal/pkg/mailer"
	"github.com/huguescodeur/oz-rest-api-go/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ── Mock StockStore ─────────────────────────────────────────────────────────

type mockStockStore struct{ mock.Mock }

func (m *mockStockStore) GetAllStocks(ctx context.Context, ownerID int) ([]*models.Stock, error) {
	args := m.Called(ctx, ownerID)
	return args.Get(0).([]*models.Stock), args.Error(1)
}
func (m *mockStockStore) GetLowStocks(ctx context.Context, ownerID, shopID int) ([]*models.Stock, error) {
	args := m.Called(ctx, ownerID, shopID)
	return args.Get(0).([]*models.Stock), args.Error(1)
}
func (m *mockStockStore) SetMinStock(ctx context.Context, productID, shopID, ownerID, minStock int) error {
	return m.Called(ctx, productID, shopID, ownerID, minStock).Error(0)
}
func (m *mockStockStore) GetStockByProductAndShop(ctx context.Context, productID, shopID, ownerID int) (*models.Stock, error) {
	args := m.Called(ctx, productID, shopID, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Stock), args.Error(1)
}
func (m *mockStockStore) InitStock(ctx context.Context, productID, shopID, ownerID int) error {
	return m.Called(ctx, productID, shopID, ownerID).Error(0)
}
func (m *mockStockStore) AdjustStock(ctx context.Context, productID, shopID, ownerID, delta int, movType, comment string) error {
	return m.Called(ctx, productID, shopID, ownerID, delta, movType, comment).Error(0)
}
func (m *mockStockStore) GetMovements(ctx context.Context, ownerID int) ([]*models.StockMovement, error) {
	args := m.Called(ctx, ownerID)
	return args.Get(0).([]*models.StockMovement), args.Error(1)
}
func (m *mockStockStore) GetMovementsByProduct(ctx context.Context, productID, ownerID int) ([]*models.StockMovement, error) {
	args := m.Called(ctx, productID, ownerID)
	return args.Get(0).([]*models.StockMovement), args.Error(1)
}
func (m *mockStockStore) GetMovementsByShop(ctx context.Context, shopID, ownerID int) ([]*models.StockMovement, error) {
	args := m.Called(ctx, shopID, ownerID)
	return args.Get(0).([]*models.StockMovement), args.Error(1)
}
func (m *mockStockStore) GetLogs(ctx context.Context, ownerID, shopID int, dateFrom, dateTo string, limit, offset int) ([]*models.StockMovement, int, error) {
	args := m.Called(ctx, ownerID, shopID, dateFrom, dateTo, limit, offset)
	return args.Get(0).([]*models.StockMovement), args.Int(1), args.Error(2)
}

// ── Helpers ─────────────────────────────────────────────────────────────────

func newStockService(ss *mockStockStore, us *mockUserStore) *services.StockService {
	return services.NewStockService(ss, us, mailer.New())
}

// ── Tests ───────────────────────────────────────────────────────────────────

func TestStockIn_Success(t *testing.T) {
	ss := &mockStockStore{}
	us := &mockUserStore{}

	ss.On("AdjustStock", mock.Anything, 1, 2, 10, 50, services.MovementStockIn, "").Return(nil)

	svc := newStockService(ss, us)
	err := svc.StockIn(context.Background(), 1, 2, 10, 50, "")
	assert.NoError(t, err)
	ss.AssertExpectations(t)
}

func TestStockIn_ZeroQuantity(t *testing.T) {
	ss := &mockStockStore{}
	us := &mockUserStore{}

	svc := newStockService(ss, us)
	err := svc.StockIn(context.Background(), 1, 2, 10, 0, "")
	assert.Error(t, err)
	ss.AssertNotCalled(t, "AdjustStock")
}

func TestSellProduct_Success(t *testing.T) {
	ss := &mockStockStore{}
	us := &mockUserStore{}

	stock := &models.Stock{ProductID: 1, ShopID: 2, Quantity: 10, MinStock: 2, LowStock: false}
	ss.On("GetStockByProductAndShop", mock.Anything, 1, 2, 10).Return(stock, nil)
	ss.On("AdjustStock", mock.Anything, 1, 2, 10, -3, services.MovementSale, "").Return(nil)
	// notifyLowStockAsync will call GetStockByProductAndShop again in a goroutine
	ss.On("GetStockByProductAndShop", mock.Anything, 1, 2, 10).Return(
		&models.Stock{Quantity: 7, MinStock: 2, LowStock: false}, nil,
	).Maybe()

	svc := newStockService(ss, us)
	err := svc.SellProduct(context.Background(), 1, 2, 10, 3, "")
	assert.NoError(t, err)
}

func TestSellProduct_InsufficientStock(t *testing.T) {
	ss := &mockStockStore{}
	us := &mockUserStore{}

	stock := &models.Stock{ProductID: 1, ShopID: 2, Quantity: 2}
	ss.On("GetStockByProductAndShop", mock.Anything, 1, 2, 10).Return(stock, nil)

	svc := newStockService(ss, us)
	err := svc.SellProduct(context.Background(), 1, 2, 10, 5, "")
	assert.Error(t, err)
	ss.AssertNotCalled(t, "AdjustStock")
}

func TestAdjustStock_InvalidType(t *testing.T) {
	ss := &mockStockStore{}
	us := &mockUserStore{}

	svc := newStockService(ss, us)
	err := svc.AdjustStock(context.Background(), 1, 2, 10, -5, "INVALID_TYPE", "")
	assert.Error(t, err)
}

func TestAdjustStock_NegativeExceedsStock(t *testing.T) {
	ss := &mockStockStore{}
	us := &mockUserStore{}

	stock := &models.Stock{ProductID: 1, ShopID: 2, Quantity: 3}
	ss.On("GetStockByProductAndShop", mock.Anything, 1, 2, 10).Return(stock, nil)

	svc := newStockService(ss, us)
	err := svc.AdjustStock(context.Background(), 1, 2, 10, -10, services.MovementLoss, "")
	assert.Error(t, err)
}

func TestAdjustStock_StoreError(t *testing.T) {
	ss := &mockStockStore{}
	us := &mockUserStore{}

	ss.On("AdjustStock", mock.Anything, 1, 2, 10, 5, services.MovementStockIn, "").Return(errors.New("db error"))

	svc := newStockService(ss, us)
	err := svc.AdjustStock(context.Background(), 1, 2, 10, 5, services.MovementStockIn, "")
	assert.Error(t, err)
}
