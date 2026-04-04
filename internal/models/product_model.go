package models

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ProductID       int        `json:"productID"`
	ProductUUID     uuid.UUID  `json:"productUUID"`
	ProductName     string     `json:"productName" validate:"required"`
	ProductCategory *Category  `json:"category" validate:"required"`
	UnitPrice       int        `json:"unitPrice" validate:"required,gte=0"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	DeletedAt       *time.Time `json:"-"`
	OwnerID         int        `json:"ownerID"`
}
