package models

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	OrderID     int         `json:"order_id"`
	OrderUUID   uuid.UUID   `json:"orderUUID"`
	ShopID      int         `json:"shop_id"`
	UserID      int         `json:"user_id"`
	TotalAmount int         `json:"total_amount"`
	CreatedAt   time.Time   `json:"created_at"`
	Items       []OrderItem `json:"items"`
}

type OrderItem struct {
	ItemID    int `json:"item_id"`
	OrderID   int `json:"order_id"`
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
	UnitPrice int `json:"unit_price"`
}
