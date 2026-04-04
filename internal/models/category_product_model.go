package models

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	CategoryID   int        `json:"id"`
	CategoryUUID uuid.UUID  `json:"categoryUUID"`
	CategoryName string     `json:"category" validate:"required"`
	DeletedAt    *time.Time `json:"-"`
	OwnerID      int        `json:"ownerID"`
}
