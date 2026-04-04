package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             int        `json:"id"`
	UUID           uuid.UUID  `json:"UUID" `
	Username       string     `json:"username" validate:"required"`
	Firstname      string     `json:"firstname" validate:"required"`
	Lastname       string     `json:"lastname" validate:"required"`
	Email          string     `json:"email" validate:"required"`
	Phone          string     `json:"phone" validate:"required"`
	Password       string     `json:"password,omitempty" validate:"required,min=6"`
	PasswordHash   string     `json:"-"`
	Role           string     `json:"role" validate:"required"`
	ProfilePicture string     `json:"profilePicture"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	DeletedAt      *time.Time `json:"deletedAt"`
	ParentID       *int       `json:"parentID"`
}
