package models

import "time"

type Stock struct {
	ProductID   int       `json:"product_id"`
	ProductName string    `json:"product_name"`
	ShopID      int       `json:"shop_id"`
	ShopName    string    `json:"shop_name"`
	Quantity    int       `json:"quantity"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type StockMovement struct {
	MovementID     int       `json:"movement_id"`
	ProductID      int       `json:"product_id"`
	ShopID         int       `json:"shop_id"`
	UserID         int       `json:"user_id"`
	QuantityChange int       `json:"quantity_change"`
	MovementType   string    `json:"movement_type"`
	Comment        string    `json:"comment"`
	CreatedAt      time.Time `json:"created_at"`
}
