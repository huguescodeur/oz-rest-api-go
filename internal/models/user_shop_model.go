package models

import "time"

type UserShop struct {
	UserID     int       `json:"user_id"`
	ShopID     int       `json:"shop_id"`
	AssignedAt time.Time `json:"assigned_at"`
}
