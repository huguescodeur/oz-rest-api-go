package services

import (
	"errors"
	"fmt"

	"github.com/huguescodeur/zoro_rest_api_go/internal/models"
	"github.com/huguescodeur/zoro_rest_api_go/internal/store"
)

const (
	MovementSale       = "SALE"
	MovementStockIn    = "STOCK_IN"
	MovementGift       = "GIFT"
	MovementLoss       = "LOSS"
	MovementAdjustment = "ADJUSTMENT"
)

var validMovementTypes = map[string]bool{
	MovementSale:       true,
	MovementStockIn:    true,
	MovementGift:       true,
	MovementLoss:       true,
	MovementAdjustment: true,
}

type StockService struct {
	stockStore store.StockStore
}

func NewStockService(s store.StockStore) *StockService {
	return &StockService{stockStore: s}
}

func (s *StockService) GetAllStocks(ownerID int) ([]*models.Stock, error) {
	return s.stockStore.GetAllStocks(ownerID)
}

func (s *StockService) GetStockByProductAndShop(productID, shopID, ownerID int) (*models.Stock, error) {
	if productID <= 0 || shopID <= 0 {
		return nil, errors.New("product_id et shop_id sont requis")
	}
	return s.stockStore.GetStockByProductAndShop(productID, shopID, ownerID)
}

func (s *StockService) InitStock(productID, shopID, ownerID int) error {
	if productID <= 0 || shopID <= 0 {
		return errors.New("product_id et shop_id sont requis")
	}
	return s.stockStore.InitStock(productID, shopID, ownerID)
}

func (s *StockService) StockIn(productID, shopID, ownerID, quantity int, comment string) error {
	if quantity <= 0 {
		return errors.New("la quantité doit être positive")
	}
	return s.stockStore.AdjustStock(productID, shopID, ownerID, quantity, MovementStockIn, comment)
}

func (s *StockService) SellProduct(productID, shopID, ownerID, quantity int, comment string) error {
	if quantity <= 0 {
		return errors.New("la quantité vendue doit être positive")
	}

	stock, err := s.stockStore.GetStockByProductAndShop(productID, shopID, ownerID)
	if err != nil {
		return err
	}
	if stock.Quantity < quantity {
		return fmt.Errorf("stock insuffisant: disponible=%d, demandé=%d", stock.Quantity, quantity)
	}

	return s.stockStore.AdjustStock(productID, shopID, ownerID, -quantity, MovementSale, comment)
}

func (s *StockService) AdjustStock(productID, shopID, ownerID, quantityChange int, movementType, comment string) error {
	if !validMovementTypes[movementType] {
		return fmt.Errorf("type de mouvement invalide: %s", movementType)
	}

	if quantityChange < 0 {
		stock, err := s.stockStore.GetStockByProductAndShop(productID, shopID, ownerID)
		if err != nil {
			return err
		}
		if stock.Quantity+quantityChange < 0 {
			return fmt.Errorf("stock insuffisant: disponible=%d, variation=%d", stock.Quantity, quantityChange)
		}
	}

	return s.stockStore.AdjustStock(productID, shopID, ownerID, quantityChange, movementType, comment)
}

func (s *StockService) GetMovements(ownerID int) ([]*models.StockMovement, error) {
	return s.stockStore.GetMovements(ownerID)
}

func (s *StockService) GetMovementsByProduct(productID, ownerID int) ([]*models.StockMovement, error) {
	if productID <= 0 {
		return nil, errors.New("product_id invalide")
	}
	return s.stockStore.GetMovementsByProduct(productID, ownerID)
}

func (s *StockService) GetMovementsByShop(shopID, ownerID int) ([]*models.StockMovement, error) {
	if shopID <= 0 {
		return nil, errors.New("shop_id invalide")
	}
	return s.stockStore.GetMovementsByShop(shopID, ownerID)
}
