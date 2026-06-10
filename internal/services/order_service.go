package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/huguescodeur/oz-rest-api-go/internal/models"
	"github.com/huguescodeur/oz-rest-api-go/internal/store"
)

type OrderService struct {
	orderStore   store.OrderStore
	productStore store.ProductStore
	stockService *StockService
}

func NewOrderService(o store.OrderStore, p store.ProductStore, s *StockService) *OrderService {
	return &OrderService{orderStore: o, productStore: p, stockService: s}
}

func (s *OrderService) GetAllOrders(ctx context.Context, ownerID, shopID int, dateFrom, dateTo string, limit, offset int) ([]*models.Order, int, error) {
	return s.orderStore.GetAll(ctx, ownerID, shopID, dateFrom, dateTo, limit, offset)
}

func (s *OrderService) GetOrderByUUID(ctx context.Context, orderUUID uuid.UUID, ownerID int) (*models.Order, error) {
	return s.orderStore.GetByUUID(ctx, orderUUID, ownerID)
}

func (s *OrderService) CreateOrder(ctx context.Context, shopID, ownerID int, items []models.OrderItem, customerName, customerPhone, paymentMethod, notes, deliveryAddress string, immediateDelivery bool) (*models.Order, error) {
	if len(items) == 0 {
		return nil, errors.New("une commande doit contenir au moins un article")
	}

	var totalAmount int
	for i := range items {
		item := &items[i]
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("quantité invalide pour le produit %d", item.ProductID)
		}
		product, err := s.productStore.GetByID(ctx, item.ProductID, ownerID)
		if err != nil {
			return nil, fmt.Errorf("produit %d introuvable ou non autorisé", item.ProductID)
		}
		item.UnitPrice = product.UnitPrice
		totalAmount += item.UnitPrice * item.Quantity
	}

	status := "PENDING"
	if immediateDelivery {
		status = "DELIVERED"
	}

	order := &models.Order{
		OrderUUID:       uuid.New(),
		ShopID:          shopID,
		UserID:          ownerID,
		TotalAmount:     totalAmount,
		Status:          status,
		Items:           items,
		CustomerName:    customerName,
		CustomerPhone:   customerPhone,
		PaymentMethod:   paymentMethod,
		Notes:           notes,
		DeliveryAddress: deliveryAddress,
	}
	created, err := s.orderStore.Create(ctx, order)
	if err != nil {
		return nil, err
	}

	// Notifier si un article est passé en stock faible après la commande
	for _, item := range items {
		s.stockService.notifyLowStockAfterOrder(ctx, item.ProductID, shopID, ownerID)
	}

	return created, nil
}

func (s *OrderService) ConfirmOrder(ctx context.Context, orderUUID uuid.UUID, ownerID int) error {
	order, err := s.orderStore.GetByUUID(ctx, orderUUID, ownerID)
	if err != nil {
		return err
	}
	if order.Status != "PENDING" {
		return fmt.Errorf("impossible de confirmer une commande avec le statut '%s'", order.Status)
	}
	return s.orderStore.UpdateStatus(ctx, orderUUID, ownerID, "CONFIRMED")
}

func (s *OrderService) DeliverOrder(ctx context.Context, orderUUID uuid.UUID, ownerID int) error {
	order, err := s.orderStore.GetByUUID(ctx, orderUUID, ownerID)
	if err != nil {
		return err
	}
	if order.Status != "CONFIRMED" {
		return fmt.Errorf("impossible de livrer une commande avec le statut '%s' (doit être CONFIRMED)", order.Status)
	}
	return s.orderStore.UpdateStatus(ctx, orderUUID, ownerID, "DELIVERED")
}

func (s *OrderService) CancelOrder(ctx context.Context, orderUUID uuid.UUID, ownerID int) error {
	return s.orderStore.Cancel(ctx, orderUUID, ownerID)
}
