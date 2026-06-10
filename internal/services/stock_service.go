package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/huguescodeur/oz-rest-api-go/internal/models"
	"github.com/huguescodeur/oz-rest-api-go/internal/pkg/errs"
	"github.com/huguescodeur/oz-rest-api-go/internal/pkg/mailer"
	"github.com/huguescodeur/oz-rest-api-go/internal/store"
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
	userStore  store.UserStore
	mailer     *mailer.Mailer
}

func NewStockService(s store.StockStore, u store.UserStore, m *mailer.Mailer) *StockService {
	return &StockService{stockStore: s, userStore: u, mailer: m}
}

// notifyLowStockAfterOrder est appelé après création d'une commande dont le stock est déjà déduit par le store.
// On ne connaît pas l'état avant — on envoie l'alerte si le stock est maintenant faible.
func (s *StockService) notifyLowStockAfterOrder(_ context.Context, productID, shopID, ownerID int) {
	go func() {
		stock, err := s.stockStore.GetStockByProductAndShop(context.Background(), productID, shopID, ownerID)
		if err != nil || !stock.LowStock || stock.MinStock == 0 {
			return
		}
		owner, err := s.userStore.GetByID(context.Background(), ownerID)
		if err != nil {
			return
		}
		items := []mailer.LowStockItem{{
			ProductName: stock.ProductName,
			ShopName:    stock.ShopName,
			Quantity:    stock.Quantity,
			MinStock:    stock.MinStock,
		}}
		fmt.Printf("[STOCK] Alerte commande: %s qty=%d seuil=%d → %s\n", stock.ProductName, stock.Quantity, stock.MinStock, owner.Email)
		if err := s.mailer.SendLowStockAlert(owner.Email, owner.Firstname+" "+owner.Lastname, items); err != nil {
			fmt.Printf("[STOCK] Erreur envoi alerte: %v\n", err)
		}
	}()
}

// notifyLowStockAsync envoie une alerte seulement quand on PASSE le seuil
// (wasAlreadyLow = true → déjà en alerte avant ce mouvement → pas de nouvel email).
func (s *StockService) notifyLowStockAsync(ownerID, productID, shopID int, wasAlreadyLow bool) {
	if wasAlreadyLow {
		return
	}
	go func() {
		ctx := context.Background()
		stock, err := s.stockStore.GetStockByProductAndShop(ctx, productID, shopID, ownerID)
		if err != nil {
			fmt.Printf("[STOCK] Alerte: GetStock(%d,%d,owner=%d) erreur: %v\n", productID, shopID, ownerID, err)
			return
		}
		if !stock.LowStock || stock.MinStock == 0 {
			return
		}
		owner, err := s.userStore.GetByID(ctx, ownerID)
		if err != nil {
			fmt.Printf("[STOCK] Alerte: GetByID(%d) erreur: %v\n", ownerID, err)
			return
		}
		items := []mailer.LowStockItem{{
			ProductName: stock.ProductName,
			ShopName:    stock.ShopName,
			Quantity:    stock.Quantity,
			MinStock:    stock.MinStock,
		}}
		fmt.Printf("[STOCK] Envoi alerte stock faible à %s (%s qty=%d seuil=%d)\n", owner.Email, stock.ProductName, stock.Quantity, stock.MinStock)
		if err := s.mailer.SendLowStockAlert(owner.Email, owner.Firstname+" "+owner.Lastname, items); err != nil {
			fmt.Printf("[STOCK] Erreur envoi alerte à %s: %v\n", owner.Email, err)
		} else {
			fmt.Printf("[STOCK] Alerte envoyée avec succès à %s\n", owner.Email)
		}
	}()
}

func (s *StockService) GetAllStocks(ctx context.Context, ownerID int) ([]*models.Stock, error) {
	return s.stockStore.GetAllStocks(ctx, ownerID)
}

func (s *StockService) GetLowStocks(ctx context.Context, ownerID, shopID int) ([]*models.Stock, error) {
	return s.stockStore.GetLowStocks(ctx, ownerID, shopID)
}

func (s *StockService) SetMinStock(ctx context.Context, productID, shopID, ownerID, minStock int) error {
	if minStock < 0 {
		return errors.New("le seuil minimum ne peut pas être négatif")
	}
	return s.stockStore.SetMinStock(ctx, productID, shopID, ownerID, minStock)
}

func (s *StockService) GetStockByProductAndShop(ctx context.Context, productID, shopID, ownerID int) (*models.Stock, error) {
	if productID <= 0 || shopID <= 0 {
		return nil, errors.New("product_id et shop_id sont requis")
	}
	return s.stockStore.GetStockByProductAndShop(ctx, productID, shopID, ownerID)
}

func (s *StockService) InitStock(ctx context.Context, productID, shopID, ownerID int) error {
	if productID <= 0 || shopID <= 0 {
		return errors.New("product_id et shop_id sont requis")
	}
	return s.stockStore.InitStock(ctx, productID, shopID, ownerID)
}

func (s *StockService) StockIn(ctx context.Context, productID, shopID, ownerID, quantity int, comment string) error {
	if quantity <= 0 {
		return errors.New("la quantité doit être positive")
	}
	return s.stockStore.AdjustStock(ctx, productID, shopID, ownerID, quantity, MovementStockIn, comment)
}

func (s *StockService) SellProduct(ctx context.Context, productID, shopID, ownerID, quantity int, comment string) error {
	if quantity <= 0 {
		return errors.New("la quantité vendue doit être positive")
	}

	stock, err := s.stockStore.GetStockByProductAndShop(ctx, productID, shopID, ownerID)
	if err != nil {
		fmt.Printf("[SELL] GetStock(product=%d, shop=%d, owner=%d) → erreur: %v\n", productID, shopID, ownerID, err)
		return err
	}
	fmt.Printf("[SELL] GetStock(product=%d, shop=%d, owner=%d) → qty=%d min=%d lowStock=%v\n", productID, shopID, ownerID, stock.Quantity, stock.MinStock, stock.LowStock)
	if stock.Quantity < quantity {
		return errs.NewBadRequest(fmt.Sprintf("Stock insuffisant : disponible %d, demandé %d", stock.Quantity, quantity))
	}

	if err := s.stockStore.AdjustStock(ctx, productID, shopID, ownerID, -quantity, MovementSale, comment); err != nil {
		return err
	}
	s.notifyLowStockAsync(ownerID, productID, shopID, stock.LowStock)
	return nil
}

func (s *StockService) AdjustStock(ctx context.Context, productID, shopID, ownerID, quantityChange int, movementType, comment string) error {
	if !validMovementTypes[movementType] {
		return errs.NewBadRequest(fmt.Sprintf("Type de mouvement invalide : %s", movementType))
	}

	var wasAlreadyLow bool
	if quantityChange < 0 {
		stock, err := s.stockStore.GetStockByProductAndShop(ctx, productID, shopID, ownerID)
		if err != nil {
			return err
		}
		if stock.Quantity+quantityChange < 0 {
			return errs.NewBadRequest(fmt.Sprintf("Stock insuffisant : disponible %d, variation demandée %d", stock.Quantity, quantityChange))
		}
		wasAlreadyLow = stock.LowStock
	}

	if err := s.stockStore.AdjustStock(ctx, productID, shopID, ownerID, quantityChange, movementType, comment); err != nil {
		return err
	}
	if quantityChange < 0 {
		s.notifyLowStockAsync(ownerID, productID, shopID, wasAlreadyLow)
	}
	return nil
}

func (s *StockService) GetMovements(ctx context.Context, ownerID int) ([]*models.StockMovement, error) {
	return s.stockStore.GetMovements(ctx, ownerID)
}

func (s *StockService) GetMovementsByProduct(ctx context.Context, productID, ownerID int) ([]*models.StockMovement, error) {
	if productID <= 0 {
		return nil, errors.New("product_id invalide")
	}
	return s.stockStore.GetMovementsByProduct(ctx, productID, ownerID)
}

func (s *StockService) GetMovementsByShop(ctx context.Context, shopID, ownerID int) ([]*models.StockMovement, error) {
	if shopID <= 0 {
		return nil, errors.New("shop_id invalide")
	}
	return s.stockStore.GetMovementsByShop(ctx, shopID, ownerID)
}

func (s *StockService) GetLogs(ctx context.Context, ownerID, shopID int, dateFrom, dateTo string, limit, offset int) ([]*models.StockMovement, int, error) {
	return s.stockStore.GetLogs(ctx, ownerID, shopID, dateFrom, dateTo, limit, offset)
}
