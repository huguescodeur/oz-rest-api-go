package models

import (
	"time"

	"github.com/google/uuid"
)

type Shop struct {
	ShopID      int        `json:"shopID"`
	ShopUUID    uuid.UUID  `json:"shopUUID" `
	ShopName    string     `json:"shopName" validate:"required,min=3"`
	ShopAddress string     `json:"shopAddress" validate:"required"`
	ShopPhone   string     `json:"shopPhone" validate:"required"`
	ShopMail    string     `json:"shopMail" validate:"required"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `json:"-"`
	OwnerID     int        `json:"ownerID"`
	AssignedAt  time.Time  `json:"assignedAt,omitempty"`
}
