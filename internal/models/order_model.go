package models

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	OrderID       int         `json:"order_id"`
	OrderUUID     uuid.UUID   `json:"orderUUID"`
	ShopID        int         `json:"shop_id"`
	ShopName      string      `json:"shop_name,omitempty"`
	UserID        int         `json:"user_id"`
	TotalAmount   int         `json:"total_amount"`
	Status        string      `json:"status"`
	CustomerName    string      `json:"customer_name,omitempty"`
	CustomerPhone   string      `json:"customer_phone,omitempty"`
	PaymentMethod   string      `json:"payment_method,omitempty"`
	Notes           string      `json:"notes,omitempty"`
	DeliveryAddress string      `json:"delivery_address,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
	Items         []OrderItem `json:"items"`
}

type OrderItem struct {
	ItemID    int `json:"item_id"`
	OrderID   int `json:"order_id"`
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
	UnitPrice int `json:"unit_price"`
}
