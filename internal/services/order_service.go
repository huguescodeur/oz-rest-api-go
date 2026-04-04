package services

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/huguescodeur/oz-rest-api-go/internal/models"
	"github.com/huguescodeur/oz-rest-api-go/internal/store"
)

type OrderService struct {
	orderStore   store.OrderStore
	productStore store.ProductStore
}

func NewOrderService(o store.OrderStore, p store.ProductStore) *OrderService {
	return &OrderService{
		orderStore:   o,
		productStore: p,
	}
}

func (s *OrderService) GetAllOrders(ownerID int) ([]*models.Order, error) {
	return s.orderStore.GetAll(ownerID)
}

func (s *OrderService) GetOrderByUUID(orderUUID uuid.UUID, ownerID int) (*models.Order, error) {
	return s.orderStore.GetByUUID(orderUUID, ownerID)
}

func (s *OrderService) CreateOrder(shopID, ownerID int, items []models.OrderItem) (*models.Order, error) {
	if len(items) == 0 {
		return nil, errors.New("une commande doit contenir au moins un article")
	}

	var totalAmount int

	for i := range items {
		item := &items[i]

		if item.Quantity <= 0 {
			return nil, fmt.Errorf("quantité invalide pour le produit %d", item.ProductID)
		}

		product, err := s.productStore.GetByID(item.ProductID, ownerID)
		if err != nil {
			return nil, fmt.Errorf("produit %d introuvable ou non autorisé", item.ProductID)
		}

		item.UnitPrice = product.UnitPrice
		totalAmount += item.UnitPrice * item.Quantity
	}

	order := &models.Order{
		OrderUUID:   uuid.New(),
		ShopID:      shopID,
		UserID:      ownerID,
		TotalAmount: totalAmount,
		Items:       items,
	}

	return s.orderStore.Create(order)
}

func (s *OrderService) CancelOrder(orderUUID uuid.UUID, ownerID int) error {
	_, err := s.orderStore.GetByUUID(orderUUID, ownerID)
	if err != nil {
		return err
	}

	return s.orderStore.Cancel(orderUUID, ownerID)
}
