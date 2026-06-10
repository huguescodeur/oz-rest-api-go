package models

import "time"

type Stock struct {
	ProductID   int       `json:"product_id"`
	ProductName string    `json:"product_name"`
	ShopID      int       `json:"shop_id"`
	ShopName    string    `json:"shop_name"`
	Quantity    int       `json:"quantity"`
	UnitPrice   float64   `json:"unit_price"`
	MinStock    int       `json:"min_stock"`
	LowStock    bool      `json:"low_stock"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type StockMovement struct {
	MovementID     int       `json:"movement_id"`
	ProductID      int       `json:"product_id"`
	ProductName    string    `json:"product_name,omitempty"`
	ShopID         int       `json:"shop_id"`
	ShopName       string    `json:"shop_name,omitempty"`
	UserID         int       `json:"user_id"`
	Username       string    `json:"username,omitempty"`
	QuantityChange int       `json:"quantity_change"`
	MovementType   string    `json:"movement_type"`
	Comment        string    `json:"comment"`
	CreatedAt      time.Time `json:"created_at"`
}
