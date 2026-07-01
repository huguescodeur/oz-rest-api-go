package models

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ProductID       int        `json:"productID"`
	ProductUUID     uuid.UUID  `json:"productUUID"`
	ProductName     string     `json:"productName" `
	ProductCategory *Category  `json:"category" `
	UnitPrice       int        `json:"unitPrice" `
	Description     string     `json:"description" `
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	DeletedAt       *time.Time `json:"-"`
	OwnerID         int        `json:"ownerID"`
}
